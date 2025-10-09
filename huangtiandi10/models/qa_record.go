package models

type QARecord struct {
	ID            int64  `gorm:"primaryKey" json:"ID"`
	UserID        int64  `gorm:"not null" json:"UserID"`
	Question      string `gorm:"type:text;not null" json:"Question"`
	Answer        string `gorm:"type:text;not null" json:"Answer"`
	ModelUsed     string `json:"ModelUsed"`
	QuestionCount int    `gorm:"-" json:"-"` // 不返回前端
}
