package order

import (
	"context"
	"fmt"

	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/queue"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"go.uber.org/zap"
)

type Service interface {
	CreateOrder(ctx context.Context, userID int64, productID int64) (*domain.Order, error)
	PayOrder(ctx context.Context, userID int64, orderID int64) error
}

type orderService struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	queue       queue.JobQueue
}

func NewService(o domain.OrderRepository, p domain.ProductRepository, q queue.JobQueue) Service {
	return &orderService{orderRepo: o, productRepo: p, queue: q}
}

func (s *orderService) CreateOrder(ctx context.Context, userID int64, productID int64) (*domain.Order, error) {
	prod, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	order := &domain.Order{
		UserID:               userID,
		ProductID:            productID,
		ProductNameSnapshot:  prod.Name,
		ProductPriceSnapshot: prod.BasePrice,
		TotalAmount:          prod.BasePrice,
		Status:               domain.OrderStatusPending,
	}

	// 3.2 创建订单
	if err := s.orderRepo.Create(ctx, order); err != nil {
		// 实际还需要回滚库存，或者在一个 DB 事务中完成
		return nil, err
	}

	return order, nil
}

func (s *orderService) PayOrder(ctx context.Context, userID int64, orderID int64) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return errno.ErrOrderUnauthorized.WithCause(
			fmt.Errorf("get user id:%d, expect user id:%d", userID, order.UserID),
		)
	}

	if !order.CanPay() {
		return errno.ErrOrderCannotPay
	}

	// TODO mock 一个支付接口
	// 模拟支付逻辑（这里我们假设支付肯定成功）
	// 更新数据库状态 Pending -> Paid
	// UpdateStatus 必须校验 prevStatus，防止并发重复支付
	err = s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPaid, domain.OrderStatusPending)
	if err != nil {
		return errno.ErrOrderCannotPay.WithCause(err)
	}

	// 发送消息到 Queue (触发异步核保)
	payload := taskPayload{
		ID:        order.ID,
		UserID:    order.UserID,
		ProductID: order.ProductID,
	}

	err = s.queue.Push(ctx, orderQueueKey, payload)
	if err != nil {
		logger.Log.Error("Order paid but failed to queue underwriting task!",
			zap.Int64("user_id", userID),
			zap.Int64("order_id", orderID),
			zap.Int64("product_id", order.ProductID),
			zap.Error(err),
		)
		return nil // 返回 nil，因为对用户来说支付是成功的
	}
	return nil
}
