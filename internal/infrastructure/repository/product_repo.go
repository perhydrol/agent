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
			logger.Log.Error(
				"redis set error",
				zap.Int64("product_id", p.ID),
				zap.Error(err),
				zap.String(traceid.TraceIDKey, traceid.GetTraceID(bgCtx)),
			)
		}
	}()
}

func (r *productRepo) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	productP := r.cache.GetByID(ctx, id)
	if productP != nil {
		return productP, nil
	}
	v, err, _ := r.sf.Do(
		fmt.Sprintf("product:FindByID:%d", id),
		func() (any, error) {
			var p domain.Product
			e := r.db.WithContext(ctx).First(&p, id).Error
			if e != nil {
				return nil, e
			}
			return &p, nil
		},
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("Unable to get data with Product ID %d from DB, err:%w", id, err)
	}
	prod, ok := v.(*domain.Product)
	if !ok {
		return nil, fmt.Errorf("type assert *domain.Product failed")
	}
	r.asyncCacheProduct(ctx, prod)
	return prod, nil
}

func (r *productRepo) List(ctx context.Context, offset, limit int, category string) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Product{})
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 先 Count (必须在 Offset 之前)
	tRes, errCount, _ := r.sf.Do(fmt.Sprintf("product:List:count:%s", category), func() (any, error) {
		var c int64
		if e := query.Count(&c).Error; e != nil {
			return nil, e
		}
		return c, nil
	})
	if errCount != nil {
		return nil, 0, errCount
	}
	t, ok := tRes.(int64)
	if !ok {
		return nil, 0, fmt.Errorf("type assert int64 failed")
	}
	total = t

	// Order by id desc 是为了让新上架的产品在前面
	dRes, errData, _ := r.sf.Do(
		fmt.Sprintf("product:List:data:%d:%d:%s", offset, limit, category),
		func() (any, error) {
			var list []*domain.Product
			if e := query.Order("id desc").Offset(offset).Limit(limit).Find(&list).Error; e != nil {
				return nil, e
			}
			return list, nil
		},
	)

	if errData != nil {
		return nil, 0, errData
	}
	list, ok := dRes.([]*domain.Product)
	if !ok {
		return nil, 0, fmt.Errorf("type assert []*domain.Product failed")
	}
	products = list
	return products, total, nil
}
