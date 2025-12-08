package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type OrderCache interface {
	GetByID(ctx context.Context, id int64) *domain.Order
	SetOrder(ctx context.Context, p *domain.Order) error
	GetUserAllOrderID(ctx context.Context, userID int64) []int64
	AddUserOrderID(ctx context.Context, userID int64, orderID int64) error
	SetUserOrderIDs(ctx context.Context, userID int64, orderIDs []int64) error
	DelByID(ctx context.Context, id int64) error
	DelUserOrders(ctx context.Context, userID int64) error
}

type orderCache struct {
	rdb *redis.Client
}

func NewOrderCache(rdb *redis.Client) OrderCache {
	return &orderCache{rdb: rdb}
}

func (c *orderCache) GetByID(ctx context.Context, id int64) *domain.Order {
	key := fmt.Sprintf("order:id:%d", id)
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		logger.Log.Error("failed to get Order from redis",
			zap.Int64("OrderId", id),
			zap.Error(err),
		)
		return nil
	}
	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		logger.Log.Error("Failed to unmarshal Order from Redis",
			zap.String("key", key),
			zap.ByteString("data", data), // 记录原始数据有助于调试
			zap.Error(err),
		)

		_ = c.rdb.Del(ctx, key)

		// 返回 nil, nil 假装缓存没命中，让上层去查 DB 重新回填
		return nil
	}
	return &order
}

func (c *orderCache) GetUserAllOrderID(ctx context.Context, userID int64) []int64 {
	key := fmt.Sprintf("order:user:%d", userID)
	valStrings, err := c.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		logger.Log.Error("failed to get User's OrderIDs from redis",
			zap.Int64("userID", userID),
			zap.Error(err),
		)
		return nil
	}
	ids := make([]int64, 0, len(valStrings))
	for _, s := range valStrings {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			logger.Log.Error("Failed to pares OrderID from Redis",
				zap.String("key", key),
				zap.String("bad_data", s), // 记录原始数据有助于调试
				zap.Error(err),
			)
			return nil
		}
		ids = append(ids, id)
	}
	return ids
}

func (c *orderCache) AddUserOrderID(ctx context.Context, userID int64, orderID int64) error {
	key := fmt.Sprintf("order:user:%d", userID)
	// RPushX 只有当 Key 存在时才追加；如果 Key 不存在，什么都不做，返回 0。
	if err := c.rdb.RPushX(ctx, key, orderID).Err(); err != nil {
		return fmt.Errorf("failed to add Order ID %d to User ID %d cache: %w", orderID, userID, err)
	}
	return nil
}

func (c *orderCache) SetUserOrderIDs(ctx context.Context, userID int64, orderIDs []int64) error {
	key := fmt.Sprintf("order:user:%d", userID)

	// 如果用户没有任何订单，需要确保缓存中没有这个 Key，防止脏数据
	if len(orderIDs) == 0 {
		return c.rdb.Del(ctx, key).Err()
	}

	// go-redis 的 RPush 需要 interface{} 类型的变长参数
	values := make([]interface{}, len(orderIDs))
	for i, v := range orderIDs {
		values[i] = v
	}

	pipe := c.rdb.Pipeline()

	pipe.RPush(ctx, key, values...)

	//nolint:gosec
	ttl := time.Hour + time.Duration(rand.Intn(180))*time.Second
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set order list for user %d: %w", userID, err)
	}

	return nil
}

func (c *orderCache) DelUserOrders(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("order:user:%d", userID)
	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete Order %d from cache: %w", userID, err)
	}
	return nil
}

func (c *orderCache) SetOrder(ctx context.Context, order *domain.Order) error {
	key := fmt.Sprintf("order:id:%d", order.ID)
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal Order %d: %w", order.ID, err)
	}
	//nolint:gosec
	ttl := time.Hour + time.Duration(rand.Intn(180))*time.Second
	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set Order %d in cache: %w", order.ID, err)
	}
	return nil
}

func (c *orderCache) DelByID(ctx context.Context, id int64) error {
	key := fmt.Sprintf("order:id:%d", id)
	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete Order %d from cache: %w", id, err)
	}
	return nil
}
