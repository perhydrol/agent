package repository

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/cache"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	traceid "github.com/perhydrol/insurance-agent-backend/pkg/traceID"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userRepo struct {
	db    *gorm.DB
	cache cache.UserCache
	sf    singleflight.Group
}

func NewUerRepository(db *gorm.DB, cache cache.UserCache) domain.UserRepository {
	return &userRepo{
		db:    db,
		cache: cache,
	}
}

func (r *userRepo) RegisterUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) asyncCacheUser(ctx context.Context, user *domain.User) {
	go func() {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()

		if err := r.cache.Set(bgCtx, user); err != nil {
			logger.Log.Error("cache set error", zap.Error(err))
		}
	}()
}

func (r *userRepo) GetUserByName(ctx context.Context, username string) (*domain.User, error) {
	user, err := r.cache.GetByName(ctx, username)
	if err == nil && user != nil {
		return user, nil
	}
	if err != nil {
		logger.Log.Error("cache get error", zap.Error(err))
	}

	var dbUser domain.User
	err = r.db.WithContext(ctx).Where("username = ?", username).First(&dbUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	r.asyncCacheUser(ctx, &dbUser)

	return &dbUser, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := r.cache.GetByID(ctx, id)
	if err == nil && user != nil {
		return user, nil
	}
	if err != nil {
		logger.Log.Error("cache get error", zap.Error(err))
	}

	var dbUser domain.User
	err = r.db.WithContext(ctx).First(&dbUser, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	r.asyncCacheUser(ctx, &dbUser)

	return &dbUser, nil
}
