package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/redis/go-redis/v9"
)

type UserCache interface {
	Get(ctx context.Context, username string) (*domain.User, error)
	Set(ctx context.Context, user *domain.User, ttl time.Duration) error
	Delete(ctx context.Context, username string) error
}

type userCache struct {
	rdb *redis.Client
}

func NewUserCache(rdb *redis.Client) UserCache {
	return &userCache{rdb: rdb}
}

func (c *userCache) Get(ctx context.Context, username string) (*domain.User, error) {
	key := fmt.Sprintf("user:username:%s", username)
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}

	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *userCache) Set(ctx context.Context, user *domain.User, ttl time.Duration) error {
	key := fmt.Sprintf("user:username:%s", user.Username)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *userCache) Delete(ctx context.Context, username string) error {
	key := fmt.Sprintf("user:username:%s", username)
	return c.rdb.Del(ctx, key).Err()
}
