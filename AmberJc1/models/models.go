package models

import "time"

type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"size:50;uniqueIndex" json:"username"`
	Password  string    `gorm:"size:255" json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Chat struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `json:"user_id"`
	Question  string    `gorm:"type:text" json:"question"`
	Answer    string    `gorm:"type:text" json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}
