package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 定义一个通用的 JSON Map 类型
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("domain.JSONMap Scan: expected []byte, got %T", value)
	}
	return json.Unmarshal(b, m)
}

type Product struct {
	ID       int64  `gorm:"primaryKey;autoIncrement:false" json:"id,string"`
	Name     string `gorm:"type:varchar(128);not null" json:"name"`
	Category string `gorm:"type:varchar(32);index;not null" json:"category"`

	// 金额严禁使用 float/double，必须用 decimal
	BasePrice   decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"base_price"`
	Description string          `gorm:"type:text" json:"description"`

	Features JSONMap `gorm:"type:json;not null" json:"features"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductRepository interface {
	FindByID(ctx context.Context, id int64) (*Product, error)
	List(ctx context.Context, offset, limit int, category string) ([]*Product, int64, error)
}

func (o *Product) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == 0 {
		o.ID = GenID()
	}
	return
}
