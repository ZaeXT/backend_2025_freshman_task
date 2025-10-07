package models

import "time"

type User struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"unique;not null" json:"username"`
	Password      string    `gorm:"not null" json:"-"`
	VipLevel      int       `gorm:"default:0" json:"vip_level"`
	QuestionCount int       `gorm:"default:0" json:"question_count"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}
