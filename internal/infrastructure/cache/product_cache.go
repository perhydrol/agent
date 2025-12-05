package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ProductCache interface {
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	SetByID(ctx context.Context, id int64, p *domain.Product) error
	DelByID(ctx context.Context, id int64) error
}

type productCache struct {
	rdb *redis.Client
}

func NewProductCache(rdb *redis.Client) ProductCache {
	return &productCache{rdb: rdb}
}

func (c *productCache) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	key := fmt.Sprintf("product:id:%d", id)
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product %d from cache: %w", id, err)
	}
	var p domain.Product
	if err := json.Unmarshal(data, &p); err != nil {
		logger.Log.Error("Failed to unmarshal product from Redis",
			zap.String("key", key),
			zap.ByteString("data", data), // 记录原始数据有助于调试
			zap.Error(err),
		)

		_ = c.rdb.Del(ctx, key)

		// 返回 nil, nil 假装缓存没命中，让上层去查 DB 重新回填
		return nil, nil
	}
	return &p, nil
}

func (c *productCache) SetByID(ctx context.Context, id int64, p *domain.Product) error {
	key := fmt.Sprintf("product:id:%d", id)
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal product %d: %w", id, err)
	}
	//nolint:gosec
	ttl := time.Hour + time.Duration(rand.Intn(180))*time.Second
	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set product %d in cache: %w", id, err)
	}
	return nil
}
func (c *productCache) DelByID(ctx context.Context, id int64) error {
	key := fmt.Sprintf("product:id:%d", id)
	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete product %d from cache: %w", id, err)
	}
	return nil
}
