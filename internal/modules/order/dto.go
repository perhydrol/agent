package order

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateOrderReq 下单请求参数
type CreateOrderReq struct {
	ProductID int64 `json:"product_id,string" binding:"required,gt=0"`
}

type OrderResp struct {
	ID        int64 `json:"id,string"`
	UserID    int64 `json:"user_id,string"`
	ProductID int64 `json:"product_id,string"`

	// 快照信息
	ProductName  string          `json:"product_name"`
	ProductPrice decimal.Decimal `json:"unit_price"`
	TotalAmount  decimal.Decimal `json:"total_amount"`

	Status int `json:"status"`

	// 只有当有保单号时才返回 (omitempty)
	PolicyNumber string `json:"policy_number,omitempty"`

	// 格式化时间
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
