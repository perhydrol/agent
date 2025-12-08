package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/cache"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	traceid "github.com/perhydrol/insurance-agent-backend/pkg/traceID"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type orderRepo struct {
	db     *gorm.DB
	cache  cache.OrderCache
	locker cache.Locker
}

func NewOrderRepo(db *gorm.DB, cache cache.OrderCache, locker cache.Locker) domain.OrderRepository {
	return &orderRepo{
		db:     db,
		cache:  cache,
		locker: locker,
	}
}

func (r *orderRepo) Create(ctx context.Context, order *domain.Order) error {
	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return fmt.Errorf("failed to create new order: %w", err)
	}
	go func() {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		if addErr := r.cache.AddUserOrderID(bgCtx, order.UserID, order.ID); addErr != nil {
			logger.Log.Error("failed to add new orderID to user in cache",
				zap.Int64("user_id", order.UserID),
				zap.Int64("order_id", order.ID),
				zap.Error(addErr),
			)
			if delErr := r.cache.DelUserOrders(bgCtx, order.UserID); delErr != nil {
				logger.Log.Error("failed to del user's orderID list",
					zap.Int64("user_id", order.UserID),
					zap.Error(delErr),
				)
			}
		}
	}()
	return nil
}

func (r *orderRepo) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	orderP := r.cache.GetByID(ctx, id)
	if orderP != nil {
		return orderP, nil
	}
	var order domain.Order
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get Order (id %d) from DB: %w", id, err)
	}
	go func(order domain.Order) {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		orderCacheLockerKey := fmt.Sprintf("order:id:%d:lock", id)
		locker, err := r.locker.Obtain(bgCtx, orderCacheLockerKey, 200*time.Millisecond)
		if err != nil {
			logger.Log.Error("failed to get Order Cache Lock", zap.Error(err))
			return
		}
		defer func() { _ = locker.Release(bgCtx) }()
		if v := r.cache.GetByID(bgCtx, id); v != nil {
			return
		}
		go func() {
			_ = locker.Refresh(bgCtx, 200*time.Millisecond, nil)
		}()
		if err := r.cache.SetOrder(bgCtx, &order); err != nil {
			logger.Log.Error("failed to set orderID to cache",
				zap.Int64("user_id", order.UserID),
				zap.Int64("order_id", order.ID),
				zap.Error(err),
			)
		}
		return
	}(order)
	return &order, nil
}

func (r *orderRepo) FindUserAllOrderID(ctx context.Context, userID int64) ([]int64, error) {
	idsP := r.cache.GetUserAllOrderID(ctx, userID)
	if idsP != nil {
		return idsP, nil
	}
	var ids []int64
	err := r.db.WithContext(ctx).Model(&domain.Order{}).
		Where("user_id = ?", userID).
		Pluck("id", &ids).Error
	if err != nil {
		return ids, fmt.Errorf("failed to find user's order list: %w", err)
	}
	go func() {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		if err := r.cache.DelUserOrders(bgCtx, userID); err != nil {
			logger.Log.Error("failed to del user's order list",
				zap.Int64("user_id", userID),
				zap.Error(err),
			)
			return
		}
		if err := r.cache.SetUserOrderIDs(bgCtx, userID, ids); err != nil {
			logger.Log.Error("failed to set user's order list",
				zap.Int64("user_id", userID),
				zap.Error(err),
			)
			return
		}
	}()
	return ids, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus, prevStatus domain.OrderStatus) error {
	// TODO
	return nil
}

func (r *orderRepo) UpdatePolicy(ctx context.Context, id int64, policyNumber string) error {
	// TODO
	return nil
}
