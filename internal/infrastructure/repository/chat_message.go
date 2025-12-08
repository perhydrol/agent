package repository

import (
	"context"
	"fmt"

	"time"

	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/cache"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	traceid "github.com/perhydrol/insurance-agent-backend/pkg/traceID"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type chatRepo struct {
	db    *gorm.DB
	cache cache.ChatMessageCache
	sf    singleflight.Group
}

func NewChatRepository(db *gorm.DB, cache cache.ChatMessageCache) domain.ChatRepository {
	return &chatRepo{db: db, cache: cache}
}

func (r *chatRepo) Save(ctx context.Context, msg *domain.ChatMessage) error {
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return errno.ErrRepoDB.WithCause(err)
	}
	go func(m *domain.ChatMessage) {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		if err := r.cache.Append(bgCtx, m.SessionID, m); err != nil {
			logger.Log.Error(
				"chat cache append error",
				zap.String("session_id", m.SessionID),
				zap.Int64("msg_id", m.ID),
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
		}
	}(msg)
	return nil
}

func (r *chatRepo) GetHistory(ctx context.Context, sessionID string, limit int) ([]*domain.ChatMessage, error) {
	if r.cache != nil {
		if list := r.cache.GetRecent(ctx, sessionID, limit); list != nil {
			return list, nil
		}
	}
	v, err, _ := r.sf.Do(
		fmt.Sprintf("chat:GetHistory:%s:%d", sessionID, limit),
		func() (any, error) {
			var list []*domain.ChatMessage
			e := r.db.WithContext(ctx).
				Where("session_id = ?", sessionID).
				Order("id desc").
				Limit(limit).
				Find(&list).Error
			if e != nil {
				return nil, e
			}
			return list, nil
		},
	)
	if err != nil {
		return nil, errno.ErrRepoDB.WithCause(err)
	}
	msgs, ok := v.([]*domain.ChatMessage)
	if !ok {
		return nil, errno.ErrRepoTypeAssert.WithCause(fmt.Errorf("type assert []*domain.ChatMessage failed"))
	}
	n := len(msgs)
	for i := 0; i < n/2; i++ {
		msgs[i], msgs[n-1-i] = msgs[n-1-i], msgs[i]
	}
	go func(data []*domain.ChatMessage) {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		if err := r.cache.SetSessionMessages(bgCtx, sessionID, data); err != nil {
			logger.Log.Error(
				"chat cache set session messages error",
				zap.String("session_id", sessionID),
				zap.Int("count", len(data)),
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
		}
	}(msgs)
	return msgs, nil
}
