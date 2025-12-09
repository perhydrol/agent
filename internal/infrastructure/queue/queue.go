package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// JobQueue 定义异步任务队列的标准接口
type JobQueue interface {
	Push(ctx context.Context, streamName string, task any) error
	Consume(ctx context.Context, streamName string, consumerName string, retChan chan<- any)
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

func (q *RedisQueue) Consume(ctx context.Context, streamName string, consumerName string, retChan chan<- any) {
	groupName := streamName + "group"
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// 阻塞读取消息
		streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,             // 消费者名称，分布式下要唯一
			Streams:  []string{groupName, ">"}, // ">" 表示读取该组尚未处理的新消息
			Count:    1,                        // 每次读取的数量
			Block:    2 * time.Second,          // 阻塞 2 秒，避免死循环占用过多 CPU，同时允许检查 ctx.Done
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
				select {
				case retChan <- message.Values:
					q.client.XAck(ctx, streamName, groupName, message.ID)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
