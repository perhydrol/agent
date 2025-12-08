package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
)

type OrderStatus int

const (
	OrderStatusPending OrderStatus = 0
	OrderStatusPaid    OrderStatus = 1
	OrderStatusActive  OrderStatus = 2
	OrderStatusFailed  OrderStatus = 3
)

type Order struct {
	ID        int64 `gorm:"primaryKey;autoIncrement:false" json:"id,string"`
	UserID    int64 `gorm:"index;not null" json:"user_id,string"`
	ProductID int64 `gorm:"index;not null" json:"product_id,string"`

	// 2. 快照字段 (Snapshot)
	// 即使 Product 表里的数据变了，这里的数据绝对不能变
	//nolint:revive
	ProductNameSnapshot string `gorm:"column:product_name_snapshot;type:varchar(128);not null" json:"product_name"`
	//nolint:revive
	ProductPriceSnapshot decimal.Decimal `gorm:"column:product_price_snapshot;type:decimal(10,2);not null" json:"unit_price"`

	// 订单总金额 (可能包含后续的折扣逻辑，所以单独存)
	TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`

	// 3. 状态机
	Status OrderStatus `gorm:"type:tinyint;default:0;index" json:"status"`

	// 4. 异步核保结果
	// omitempty: 如果还没生成保单号，JSON里就不返回这个字段，保持接口整洁
	PolicyNumber string `gorm:"type:varchar(64)" json:"policy_number,omitempty"`

	// 5. 乐观锁版本号
	// json:"-": 这是一个内部技术字段，不需要暴露给前端 API
	Version optimisticlock.Version `gorm:"default:1" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (o *Order) CanPay() bool {
	return o.Status == OrderStatusPending
}

func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid || o.Status == OrderStatusActive
}

type OrderRepository interface {
	// 下单
	Create(ctx context.Context, order *Order) error
	// 详情 & 轮询
	FindByID(ctx context.Context, id int64) (*Order, error)
	FindUserAllOrderID(ctx context.Context, userID int64) ([]int64, error)
	// 支付成功回调：将状态改为 Paid
	// 必须校验当前状态 (prevStatus)，防止重复支付
	UpdateStatus(ctx context.Context, id int64, status OrderStatus, prevStatus OrderStatus) error

	// --- 异步 Worker 专用接口 ---

	// 核保成功：填入保单号，状态改为 Active
	UpdatePolicy(ctx context.Context, id int64, policyNumber string) error
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == 0 {
		o.ID = GenID()
	}
	return
}
