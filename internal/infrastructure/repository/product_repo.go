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

type productRepo struct {
	db    *gorm.DB
	cache cache.ProductCache
	sf    singleflight.Group
}

func NewProductRepository(db *gorm.DB, cache cache.ProductCache) domain.ProductRepository {
	return &productRepo{
		db:    db,
		cache: cache,
	}
}

func (r *productRepo) asyncCacheProduct(ctx context.Context, p *domain.Product) {
	go func() {
		tempCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		bgCtx := context.WithValue(tempCtx, traceid.ContextTraceIDKey, traceid.GetTraceID(ctx))
		defer cancel()

		if err := r.cache.Set(bgCtx, p); err != nil {
			logger.Log.Error("redis set error", zap.Int64("product_id", p.ID), zap.Error(err))
		}
	}()
}

func (r *productRepo) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	productP := r.cache.GetByID(ctx, id)
	if productP != nil {
		return productP, nil
	}
	var p domain.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("Unable to get data with Product ID %d from DB, err:%w", id, err)
	}
	r.asyncCacheProduct(ctx, &p)
	return &p, err
}

func (r *productRepo) List(ctx context.Context, offset, limit int, category string) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Product{})
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 先 Count (必须在 Offset 之前)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Order by id desc 是为了让新上架的产品在前面
	err := query.Order("id desc").Offset(offset).Limit(limit).Find(&products).Error

	return products, total, err
}
