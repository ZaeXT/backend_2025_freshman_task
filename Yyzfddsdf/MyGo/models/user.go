package models

import (
	"time"

	"gorm.io/gorm"
)

// User 模型，用于 GORM 映射到数据库表
type User struct {
	gorm.Model        // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `gorm:"unique;not null" json:"username"`
	Email      string `gorm:"unique;not null" json:"email"`
	Password   string `gorm:"not null" json:"-"`            // 密码字段存储哈希值，JSON 序列化时忽略
	TokenCount int    `gorm:"default:10" json:"tokenCount"` // 新增：token数量，默认10个

	// 关联关系
	Conversations []Conversation `gorm:"foreignKey:UserID" json:"-"` // 用户的对话记录
}

// LoginPayload 登录请求的结构体
type LoginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterPayload 注册请求的结构体
type RegisterPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserResponse 响应给前端的用户信息（不包含密码）
type UserResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	TokenCount int    `json:"tokenCount"` // 新增：返回token数量
}

// ConversationResponse 对话记录响应结构体
type ConversationResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	MessageCount int       `json:"messageCount"`
	TokenUsed    int       `json:"tokenUsed"`
	CreatedAt    time.Time `json:"createdAt"`
}

// MessageResponse 消息响应结构体
type MessageResponse struct {
	ID         uint      `json:"id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	TokenCount int       `json:"tokenCount"`
	CreatedAt  time.Time `json:"createdAt"`
}
