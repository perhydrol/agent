package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// JobQueue 定义异步任务队列的标准接口
type JobQueue interface {
	// Push 将订单ID推入待核保队列
	PushUnderwritingJob(ctx context.Context, order TaskPayload) error
}

type TaskPayload struct {
	ID        int64 `json:"order_id,string"`
	UserID    int64 `json:"user_id,string"`
	ProductID int64 `json:"product_id,string"`
}

type RedisQueue struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisQueue(client *redis.Client, logger *zap.Logger) JobQueue {
	return &RedisQueue{
		client: client,
		logger: logger,
	}
}

func (q *RedisQueue) PushUnderwritingJob(ctx context.Context, order TaskPayload) error {
	// TODO
	return nil
}
