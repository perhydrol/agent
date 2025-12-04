package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID string    `gorm:"type:varchar(64);index;not null" json:"session_id"`
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	Role      string    `gorm:"type:varchar(10);not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (o *ChatMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == 0 {
		o.ID = GenID()
	}
	return
}

// 负责对话历史记录 (存储于 MySQL 或 Redis List)
type ChatRepository interface {
	// 保存一条消息 (User 或 AI 的)
	Save(ctx context.Context, msg *ChatMessage) error

	// 获取最近的 k 条历史记录，用于构建 Prompt Context
	GetHistory(ctx context.Context, sessionID string, limit int) ([]*ChatMessage, error)
}

// 负责 RAG 向量检索 (这是 clean architecture 中的 "Port")
// 它的实现将在 infrastructure/vector/qdrant 或 memory 中
type VectorRepository interface {
	// 根据用户 query，返回相关的文档片段(string)和相似度
	SearchSimilarDocs(ctx context.Context, query string, topK int) ([]string, error)
}
