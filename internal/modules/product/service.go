package product

import (
	"context"
	"encoding/json"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"go.uber.org/zap"
)

type Service interface {
	ListProducts(ctx context.Context, page, size int, category string) ([]*PResp, int64, error)
	GetProduct(ctx context.Context, id int64) (*PResp, error)
}

type productService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) Service {
	return &productService{repo: repo}
}

// domainToDTO 辅助转换函数
func (s *productService) domainToDTO(p *domain.Product) *PResp {
	featuresBytes, err := json.Marshal(p.Features)
	if err != nil {
		logger.Log.Panic("failed to marshal product features to JSON",
			zap.Int64("product_id", p.ID),
			zap.Binary("raw_data", json.RawMessage(featuresBytes)),
			zap.Error(err),
		)
	}
	return &PResp{
		ID:          p.ID,
		Name:        p.Name,
		Category:    p.Category,
		BasePrice:   p.BasePrice,
		Description: p.Description,
		// json.RawMessage 可以直接透传给前端，不需要在此 Unmarshal
		Features: json.RawMessage(featuresBytes),
	}
}

func (s *productService) ListProducts(
	ctx context.Context,
	page,
	size int,
	category string,
) ([]*PResp, int64, error) {
	offset := (page - 1) * size
	list, total, err := s.repo.List(ctx, offset, size, category)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*PResp, 0, len(list))
	for _, p := range list {
		dtos = append(dtos, s.domainToDTO(p))
	}
	return dtos, total, nil
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*PResp, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return s.domainToDTO(p), nil
}
