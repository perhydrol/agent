package queue

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/config"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 返回 nil 表示处理成功 (会自动 ACK)
// 返回 error 表示处理失败 (不会 ACK，等待重试)
type Handler func(ctx context.Context, task any) error

// JobQueue 定义异步任务队列的标准接口
type JobQueue interface {
	Push(ctx context.Context, streamName string, task any) error
	Consume(ctx context.Context, streamName string, consumerName string, handler Handler)
}

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) JobQueue {
	return &RedisQueue{
		client: client,
	}
}

func (q *RedisQueue) Push(ctx context.Context, streamName string, task any) error {
	err := q.client.XGroupCreateMkStream(ctx, streamName, streamName+"group", "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return errno.ErrStreamCreateGroupFailed.WithCause(fmt.Errorf("streamName: %s: %w", streamName, err))
	}

	args := &redis.XAddArgs{
		Stream: streamName, // 流的名称
		ID:     "*",        // * 表示由 Redis 自动生成 ID (时间戳-序号)
		Values: task,       // 消息体
	}

	_, addErr := q.client.XAdd(ctx, args).Result()
	if addErr != nil {
		return errno.ErrStreamAddTaskFailed.WithCause(addErr)
	}
	return nil
}

func (q *RedisQueue) Consume(ctx context.Context, streamName string, consumerName string, handler Handler) {
	groupName := streamName + "group"
	err := q.client.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logger.Log.Error("failed to create consumer group",
			zap.String("streamName", streamName),
			zap.Error(err),
		)
	}

	// 启动处理 PEL 队列的goroutine
	go q.monitorPending(ctx, streamName, consumerName, handler)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// 阻塞读取消息
		streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,              // 消费者名称，分布式下要唯一
			Streams:  []string{streamName, ">"}, // ">" 表示读取该组尚未处理的新消息
			Count:    1,                         // 每次读取的数量
			Block:    2 * time.Second,           // 阻塞 2 秒，避免死循环占用过多 CPU，同时允许检查 ctx.Done
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue // 超时无消息
			}
			logger.Log.Error("failed to read streams",
				zap.String("streamName", streamName),
				zap.Error(err),
			)
			// 出错后稍微 sleep 一下，防止疯狂重试
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				logger.Log.Info("get a new message", zap.String("message_id", message.ID))
				value := message.Values
				if err := handler(ctx, value); err != nil {
					logger.Log.Error("failed to process message",
						zap.String("msg_id", message.ID),
						zap.Error(err),
					)
					continue
				}
				q.client.XAck(ctx, streamName, groupName, message.ID)
			}
		}
	}
}

func (q *RedisQueue) monitorPending(ctx context.Context, streamName, consumerName string, handler Handler) {
	//nolint:gosec
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(config.AppConfig.Redis.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.processPendingMessages(ctx, streamName, consumerName, handler)
		}
	}
}

func (q *RedisQueue) processPendingMessages(ctx context.Context, streamName, consumerName string, handler Handler) {
	groupName := streamName + "group"
	pendingCmd := q.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamName,
		Group:  groupName,
		Start:  "-",
		End:    "+",
		Count:  int64(config.AppConfig.Redis.BatchSize),
	})

	redisConfig := &config.AppConfig.Redis
	pendingMsg, err := pendingCmd.Result()
	if err != nil {
		logger.Log.Error("failed to get pending messages", zap.Error(err))
		return
	}

	if len(pendingMsg) == 0 {
		return
	}

	for _, msg := range pendingMsg {
		if msg.RetryCount > int64(redisConfig.MaxRetries) {
			logger.Log.Warn("message reached max retries, dropping",
				zap.String("msg_id", msg.ID),
				zap.Int64("retry_count", msg.RetryCount),
			)
			// 在实际生产环境中，这里应该先把消息 XADD 到死信队列 (DLQ) 再 ACK
			q.client.XAck(ctx, streamName, groupName, msg.ID)
			continue
		}

		idleThreshold := time.Duration(redisConfig.IdleThreshold) * time.Second
		if msg.Idle >= idleThreshold {
			logger.Log.Info("claiming idle message",
				zap.String("msg_id", msg.ID),
				zap.Duration("idle", msg.Idle),
			)

			// XClaim 会将消息的所有权转移给当前消费者，并重置 Idle 时间，增加 DeliveryCount
			claimArgs := &redis.XClaimArgs{
				Stream:   streamName,
				Group:    groupName,
				Consumer: consumerName,
				MinIdle:  idleThreshold,
				Messages: []string{msg.ID},
			}
			claimedMsgs, err := q.client.XClaim(ctx, claimArgs).Result()
			if err != nil {
				logger.Log.Error("failed to claim message", zap.String("msg_id", msg.ID), zap.Error(err))
				continue
			}

			for _, xmsg := range claimedMsgs {
				value := xmsg.Values
				if err := handler(ctx, value); err != nil {
					logger.Log.Error("failed to process message",
						zap.String("msg_id", xmsg.ID),
						zap.Error(err),
					)
					continue
				}
				q.client.XAck(ctx, streamName, groupName, xmsg.ID)
			}
		}
	}
}
