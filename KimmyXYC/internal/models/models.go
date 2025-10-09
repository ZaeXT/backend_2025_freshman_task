package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered account.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email        string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string `json:"-"`
	Role         string `gorm:"size:20;not null;default:free" json:"role"` // free, pro, admin
}

// Conversation stores a chat session for a user.
type Conversation struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint   `gorm:"index" json:"user_id"`
	Title  string `gorm:"size:255" json:"title"`
	Model  string `gorm:"size:100" json:"model"`

	Messages []Message `json:"messages"`
}

// Message stores individual messages in a conversation.
type Message struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	ConversationID uint   `gorm:"index" json:"conversation_id"`
	Role           string `gorm:"size:20;not null" json:"role"` // user or assistant
	Content        string `gorm:"type:text" json:"content"`
}
