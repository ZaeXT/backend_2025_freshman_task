package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User 用户结构体
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // 不在JSON中显示
	Level    int    `json:"level"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims JWT token中的用户信息
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	jwt.RegisteredClaims
}

// Conversation 对话结构体
type Conversation struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

// Message 消息结构体
type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
