package repository

import "time"

// Conversation 手写会话表模型，避免与 gorm-gen 冲突
type Conversation struct {
	ID             int64  `gorm:"primaryKey"`
	ConversationID string `gorm:"uniqueIndex;size:64"`
	UserID         int64  `gorm:"index"`
	HistoryJSON    []byte `gorm:"type:jsonb"`
	MessageCount   int
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (Conversation) TableName() string { return "conversations" }
