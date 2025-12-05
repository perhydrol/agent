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
	GetByName(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	Set(ctx context.Context, user *domain.User, ttl time.Duration) error
	Del(ctx context.Context, user *domain.User) error
}

type userCache struct {
	rdb *redis.Client
}

func NewUserCache(rdb *redis.Client) UserCache {
	return &userCache{rdb: rdb}
}

func (c *userCache) get(ctx context.Context, key string) (*domain.User, error) {
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

func (c *userCache) GetByName(ctx context.Context, username string) (*domain.User, error) {
	key := fmt.Sprintf("user:username:%s", username)
	return c.get(ctx, key)
}

func (c *userCache) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	key := fmt.Sprintf("user:id:%d", id)
	return c.get(ctx, key)
}

func (c *userCache) Set(ctx context.Context, user *domain.User, ttl time.Duration) error {
	keyName := fmt.Sprintf("user:username:%s", user.Username)
	keyId := fmt.Sprintf("user:id:%d", user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, keyId, data, ttl)
	pipe.Set(ctx, keyName, data, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *userCache) Del(ctx context.Context, user *domain.User) error {
	keyName := fmt.Sprintf("user:username:%s", user.Username)
	keyId := fmt.Sprintf("user:id:%d", user.ID)

	pipe := c.rdb.Pipeline()
	pipe.Del(ctx, keyName)
	pipe.Del(ctx, keyId)

	_, err := pipe.Exec(ctx)
	return err
}
