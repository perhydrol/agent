package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/queue"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"go.uber.org/zap"
)

type Service interface {
	CreateOrder(ctx context.Context, userID int64, productID int64) (*domain.Order, error)
	PayOrder(ctx context.Context, userID int64, orderID int64) error
	processOrder(ctx context.Context)
}

type orderService struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	queue       queue.JobQueue
}

func NewService(ctx context.Context, o domain.OrderRepository, p domain.ProductRepository, q queue.JobQueue) Service {
	service := orderService{orderRepo: o, productRepo: p, queue: q}
	go service.processOrder(ctx)
	return &service
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

	// 创建订单
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

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errno.ErrServerMarshalFailed.WithCause(fmt.Errorf("failed to marshal payload: %w", err))
	}

	err = s.queue.Push(ctx, orderQueueKey, map[string]interface{}{
		"payload": string(payloadBytes),
	})

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

func (s *orderService) processOrder(ctx context.Context) {
	s.queue.Consume(ctx, orderQueueKey, orderQueueKey, func(ctx context.Context, o any) error {
		msgMap, ok := o.(map[string]any)
		if !ok {
			logger.Log.Error("invalid message format: expected map[string]interface{}", zap.Any("received", o))
			// 返回 nil 表示虽然格式错误，但这是一个不可恢复的错误，我们确认掉它，避免死循环
			return nil
		}

		payloadStr, ok := msgMap["payload"].(string)
		if !ok {
			logger.Log.Error("invalid payload format: expected string in 'payload' field",
				zap.Any("msg_map", msgMap),
			)
			return nil
		}

		var task taskPayload
		if err := json.Unmarshal([]byte(payloadStr), &task); err != nil {
			logger.Log.Error("failed to unmarshal task payload", zap.String("payload", payloadStr), zap.Error(err))
			return nil
		}

		// 模拟核保
		return s.handleUnderwriting(ctx, &task)
	})
}

func (s *orderService) handleUnderwriting(ctx context.Context, task *taskPayload) error {
	logger.Log.Info("Start underwriting...", zap.Int64("order_id", task.ID))

	// 模拟耗时操作
	time.Sleep(500 * time.Millisecond)

	// 模拟核保通过，生成保单号
	policyNumber := fmt.Sprintf("POL-%s-%d", uuid.New().String()[:8], task.ID)

	// processOrder 传入的通常是 Background context，所以无需担心过期问题
	err := s.orderRepo.UpdatePolicy(ctx, task.ID, policyNumber)
	if err != nil {
		logger.Log.Error("Failed to update policy number",
			zap.Int64("order_id", task.ID),
			zap.Error(err),
		)
		// 告知 redis 发生了错误
		return err
	}

	logger.Log.Info("Underwriting completed, policy issued",
		zap.Int64("order_id", task.ID),
		zap.String("policy_number", policyNumber),
	)
	return nil
}
