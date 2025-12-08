package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ChatMessageCache interface {
	Append(ctx context.Context, sessionID string, msg *domain.ChatMessage) error
	GetRecent(ctx context.Context, sessionID string, limit int) []*domain.ChatMessage
	SetSessionMessages(ctx context.Context, sessionID string, msgs []*domain.ChatMessage) error
	DelSession(ctx context.Context, sessionID string) error
}

type chatMessageCache struct {
	rdb *redis.Client
}

func NewChatMessageCache(rdb *redis.Client) ChatMessageCache {
	return &chatMessageCache{rdb: rdb}
}

func (c *chatMessageCache) Append(ctx context.Context, sessionID string, msg *domain.ChatMessage) error {
	key := fmt.Sprintf("chat:session:%s", sessionID)
	data, err := json.Marshal(msg)
	if err != nil {
		return errno.ErrCacheMarshalFailed.WithCause(err)
	}
	pipe := c.rdb.Pipeline()
	pipe.RPush(ctx, key, data)
	//nolint:gosec
	ttl := time.Hour + time.Duration(rand.Intn(180))*time.Second
	pipe.Expire(ctx, key, ttl)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return errno.ErrCacheSetFailed.WithCause(err)
	}
	return nil
}

func (c *chatMessageCache) GetRecent(ctx context.Context, sessionID string, limit int) []*domain.ChatMessage {
	if limit <= 0 {
		return nil
	}
	key := fmt.Sprintf("chat:session:%s", sessionID)
	vals, err := c.rdb.LRange(ctx, key, int64(-limit), -1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		logger.Log.Error("failed to get chat history from redis",
			zap.String("sessionID", sessionID),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		return nil
	}
	res := make([]*domain.ChatMessage, 0, len(vals))
	for _, s := range vals {
		var m domain.ChatMessage
		if uErr := json.Unmarshal([]byte(s), &m); uErr != nil {
			logger.Log.Error("Failed to unmarshal ChatMessage from Redis",
				zap.String("key", key),
				zap.String("bad_data", s),
				zap.Error(uErr),
			)
			return nil
		}
		res = append(res, &m)
	}
	return res
}

func (c *chatMessageCache) SetSessionMessages(ctx context.Context, sessionID string, msgs []*domain.ChatMessage) error {
	key := fmt.Sprintf("chat:session:%s", sessionID)
	if len(msgs) == 0 {
		return c.rdb.Del(ctx, key).Err()
	}
	values := make([]interface{}, len(msgs))
	for i, m := range msgs {
		data, err := json.Marshal(m)
		if err != nil {
			return errno.ErrCacheMarshalFailed.WithCause(err)
		}
		values[i] = data
	}
	pipe := c.rdb.Pipeline()
	pipe.Del(ctx, key)
	pipe.RPush(ctx, key, values...)
	//nolint:gosec
	ttl := time.Hour + time.Duration(rand.Intn(180))*time.Second
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errno.ErrCacheSetFailed.WithCause(err)
	}
	return nil
}

func (c *chatMessageCache) DelSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("chat:session:%s", sessionID)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return errno.ErrCacheDelFailed.WithCause(err)
	}
	return nil
}
