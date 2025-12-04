package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	// 这里的 ID 使用 int64 对应数据库的 bigint
	ID int64 `gorm:"primaryKey;autoIncrement:false" json:"id,string"`
	// json:"id,string" 是为了防止前端 JS 丢失 int64 精度，转成字符串传输

	Username string `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`

	PasswordHash string `gorm:"type:varchar(128);not null" json:"-"`

	Email string `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (o *User) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == 0 {
		o.ID = GenID()
	}
	return
}

type UserRepository interface {
	RegisterUser(ctx context.Context, username, passwordHash, email string) error
	GetUser(ctx context.Context, username string) (User, error)
}
