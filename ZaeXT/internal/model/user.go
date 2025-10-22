package model

type User struct {
	BaseModel
	Username     string `gorm:"unique;not null;size:64"`
	PasswordHash string `gorm:"not null"`
	Tier         string `gorm:"size:20;default:'free';not null"`
	MemoryInfo   string `gorm:"type:text"`

	Conversations []Conversation `gorm:"foreignKey:UserID"`
	Categories    []Category     `gorm:"foreignKey:UserID"`
}
