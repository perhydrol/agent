package product

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type TaskPayload struct {
	ID        string `json:"order_id"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
}

type PResp struct {
	ID          int64           `json:"id,string"`
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	BasePrice   decimal.Decimal `json:"base_price"`
	Description string          `json:"description"`

	// 这里使用 json.RawMessage
	// json.RawMessage 本质上就是 []byte。
	// 当 gin/json 序列化这个结构体时，它不会对 Features 里的内容进行二次转义，而是直接拼接到最终的 JSON 串里。
	Features json.RawMessage `json:"features"`
}
