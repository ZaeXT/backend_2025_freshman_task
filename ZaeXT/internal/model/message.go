package model

type Message struct {
	BaseModel
	ConversationID uint   `gorm:"not null;index"`
	Role           string `gorm:"size:20;not null"`
	Content        string `gorm:"type:text;not null"`

	Conversation Conversation `gorm:"foreignKey:ConversationID"`
}
