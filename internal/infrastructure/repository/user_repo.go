package repository

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/cache"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userRepo struct {
	db    *gorm.DB
	redis cache.UserCache
}

func NewUerRepository(db *gorm.DB, redis cache.UserCache) domain.UserRepository {
	return &userRepo{
		db:    db,
		redis: redis,
	}
}

func (r *userRepo) RegisterUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
func (r *userRepo) GetUserName(ctx context.Context, username string) (*domain.User, error) {
	userP, err := r.redis.Get(ctx, username)
	if err == nil && userP != nil {
		return userP, nil
	} else if err != nil {
		logger.Log.Error("redis error", zap.Error(err))
	}
	var user domain.User
	err = r.db.WithContext(ctx).Where("username=?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	go func() {
		// 注意：这里新建一个 context，因为请求的 ctx 可能很快就 cancel 了
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, middleware.ContextTraceIDKey, middleware.GetTraceID(ctx))
		defer cancel()

		//nolint:gosec // G404: Use of weak random number generator is acceptable for cache jitter.
		if setErr := r.redis.Set(bgCtx, &user, time.Hour+time.Duration(rand.Intn(180))*time.Second); setErr != nil {
			logger.Log.Error("redis error", zap.Error(setErr))
		}
	}()
	return &user, err
}
