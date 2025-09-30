package models

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 对话记录模型
type Conversation struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"userId"`   // 关联的用户ID
	Title        string         `gorm:"type:varchar(255)" json:"title"` // 对话标题
	MessageCount int            `gorm:"default:0" json:"messageCount"`  // 消息数量
	TokenUsed    int            `gorm:"default:0" json:"tokenUsed"`     // 消耗的token数量
	CreatedAt    time.Time      `json:"createdAt"`                      // 创建时间
	UpdatedAt    time.Time      `json:"updatedAt"`                      // 更新时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                 // 软删除

	// 关联关系
	User     User      `gorm:"foreignKey:UserID" json:"-"`                // 关联的用户
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages"` // 关联的消息
}
