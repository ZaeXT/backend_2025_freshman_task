package models

import (
	"time"

	"gorm.io/gorm"
)

// Message 消息模型
type Message struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ConversationID uint           `gorm:"not null;index" json:"conversationId"`  // 关联的对话ID
	Role           string         `gorm:"type:varchar(20);not null" json:"role"` // 角色: user, assistant, system
	Content        string         `gorm:"type:text;not null" json:"content"`     // 消息内容
	TokenCount     int            `gorm:"default:0" json:"tokenCount"`           // 消息消耗的token数量
	CreatedAt      time.Time      `json:"createdAt"`                             // 创建时间
	UpdatedAt      time.Time      `json:"updatedAt"`                             // 更新时间
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`                        // 软删除

	// 关联关系
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"` // 关联的对话
}
