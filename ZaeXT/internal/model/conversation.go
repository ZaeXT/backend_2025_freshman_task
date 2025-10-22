package model

import "gorm.io/gorm"

type Conversation struct {
	BaseModel
	UserID              uint           `gorm:"not null;index"`
	Title               string         `gorm:"size:255;default:'New Chat'"`
	IsTitleUserModified bool           `gorm:"default:false"`
	CategoryID          *uint          `gorm:"index"`
	IsTemporary         bool           `gorm:"default:false"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`

	User     User       `gorm:"foreignKey:UserID"`
	Messages []*Message `gorm:"foreignKey:ConversationID"`
}
