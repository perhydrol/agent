package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

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
	sf     singleflight.Group
}

func NewOrderRepo(db *gorm.DB, cache cache.OrderCache, locker cache.Locker) domain.OrderRepository {
	return &orderRepo{
		db:     db,
		cache:  cache,
		locker: locker,
	}
}

func (r *orderRepo) Create(ctx context.Context, order *domain.Order) error {
	_, err, _ := r.sf.Do(
		fmt.Sprintf("order:Create:%d:%d:%d", order.UserID, order.ProductID, int(order.Status)),
		func() (any, error) {
			return nil, r.db.WithContext(ctx).Create(order).Error
		},
	)
	if err != nil {
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
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
			if delErr := r.cache.DelUserOrders(bgCtx, order.UserID); delErr != nil {
				logger.Log.Error("failed to del user's orderID list",
					zap.Int64("user_id", order.UserID),
					zap.Error(delErr),
					zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
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
	v, err, _ := r.sf.Do(fmt.Sprintf("order:FindByID:%d", id), func() (any, error) {
		var o domain.Order
		if e := r.db.WithContext(ctx).First(&o, id).Error; e != nil {
			return nil, e
		}
		return &o, nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get Order (id %d) from DB: %w", id, err)
	}
	o, ok := v.(*domain.Order)
	if !ok {
		return nil, fmt.Errorf("type assert *domain.Order failed")
	}
	order := *o
	go func(order domain.Order) {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		orderCacheLockerKey := fmt.Sprintf("order:id:%d:lock", id)
		locker, err := r.locker.Obtain(bgCtx, orderCacheLockerKey, 200*time.Millisecond)
		if err != nil {
			logger.Log.Error(
				"failed to get Order Cache Lock",
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
			return
		}
		defer func() {
			if relErr := locker.Release(bgCtx); relErr != nil {
				logger.Log.Error(
					"locker release error",
					zap.Error(relErr),
					zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
				)
			}
		}()
		if v := r.cache.GetByID(bgCtx, id); v != nil {
			return
		}
		go func() {
			if refErr := locker.Refresh(bgCtx, 200*time.Millisecond, nil); refErr != nil {
				logger.Log.Error(
					"locker refresh error",
					zap.Error(refErr),
					zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
				)
			}
		}()
		if err := r.cache.SetOrder(bgCtx, &order); err != nil {
			logger.Log.Error("failed to set orderID to cache",
				zap.Int64("user_id", order.UserID),
				zap.Int64("order_id", order.ID),
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
		}
	}(order)
	return &order, nil
}

func (r *orderRepo) FindUserAllOrderID(ctx context.Context, userID int64) ([]int64, error) {
	idsP := r.cache.GetUserAllOrderID(ctx, userID)
	if idsP != nil {
		return idsP, nil
	}
	v, err, _ := r.sf.Do(fmt.Sprintf("order:FindUserAllOrderID:%d", userID), func() (any, error) {
		var ids []int64
		e := r.db.WithContext(ctx).Model(&domain.Order{}).
			Where("user_id = ?", userID).
			Pluck("id", &ids).Error
		if e != nil {
			return nil, e
		}
		return ids, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find user's order list: %w", err)
	}
	ids, ok := v.([]int64)
	if !ok {
		return nil, fmt.Errorf("type assert []int64 failed")
	}
	go func() {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()
		if err := r.cache.DelUserOrders(bgCtx, userID); err != nil {
			logger.Log.Error("failed to del user's order list",
				zap.Int64("user_id", userID),
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
			return
		}
		if err := r.cache.SetUserOrderIDs(bgCtx, userID, ids); err != nil {
			logger.Log.Error("failed to set user's order list",
				zap.Int64("user_id", userID),
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
			return
		}
	}()
	return ids, nil
}

func (r *orderRepo) UpdateStatus(
	ctx context.Context,
	id int64,
	status domain.OrderStatus,
	prevStatus domain.OrderStatus,
) error {
	// TODO
	return nil
}

func (r *orderRepo) UpdatePolicy(ctx context.Context, id int64, policyNumber string) error {
	// TODO
	return nil
}
