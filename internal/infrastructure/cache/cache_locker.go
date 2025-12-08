package cache

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type Locker interface {
	// Obtain 尝试获取锁，ttl 是锁的生命周期
	Obtain(ctx context.Context, key string, ttl time.Duration) (*redislock.Lock, error)
}

type redisLocker struct {
	client *redislock.Client
}

func NewRedisLocker(rdb *redis.Client) Locker {
	// 初始化 redislock 客户端
	return &redisLocker{
		client: redislock.New(rdb),
	}
}

func (l *redisLocker) Obtain(ctx context.Context, key string, ttl time.Duration) (*redislock.Lock, error) {
	// 线性重试策略：每 100ms 重试一次，最多重试 1 秒
	// 这对应了“等待锁”的场景，防止请求直接失败
	opt := &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 10),
	}
	return l.client.Obtain(ctx, key, ttl, opt)
}
