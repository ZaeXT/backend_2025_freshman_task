package models

// User 用户模型
type User struct {
	Username   string `gorm:"primaryKey;column:username;not null" json:"username"`
	Password   string `gorm:"column:passwordmd5;not null" json:"-"`
	Permission int    `gorm:"column:permission;default:0" json:"permission"`
}
