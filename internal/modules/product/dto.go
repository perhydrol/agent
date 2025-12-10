package product

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type TaskPayload struct {
	ID        int64 `json:"order_id,string"`
	UserID    int64 `json:"user_id,string"`
	ProductID int64 `json:"product_id,string"`
}

type PResp struct {
	ID          int64           `json:"id,string" example:"1"`
	Name        string          `json:"name" example:"Travel Insurance"`
	Category    string          `json:"category" example:"travel"`
	BasePrice   decimal.Decimal `json:"base_price" example:"100.00"`
	Description string          `json:"description" example:"Comprehensive travel insurance"`

	// json.RawMessage 在校验层通常只需要保证格式正确，
	// 但 validator 对 RawMessage 支持有限，通常由 json Unmarshal 自动处理，
	// 这里 omitempty 表示允许不传。
	Features json.RawMessage `json:"features" swaggertype:"string" example:"{\"coverage\": 100000}"`
}

type PListReq struct {
	Page     int    `form:"page,default=1" binding:"omitempty,min=1" example:"1"`
	PageSize int    `form:"page_size,default=10" binding:"omitempty,min=1,max=100" example:"10"`
	Category string `form:"category" binding:"omitempty" example:"travel"`
}

type PListResp struct {
	List     []PResp `json:"list"`
	Total    int64   `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}
