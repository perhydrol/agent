package repository

import (
	"context"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"gorm.io/gorm"
)

type productRepo struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepo{
		db: db,
	}
}

func (r *productRepo) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	var p domain.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
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
