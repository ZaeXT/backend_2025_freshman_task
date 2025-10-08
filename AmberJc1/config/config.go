package config

import (
	"houduan_from/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 连接数据库（密码改成你自己的）
	dsn := "root:123456@tcp(127.0.0.1:3306)/ai_chat?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
	}
	DB = db
	// 创建数据表
	db.AutoMigrate(&models.User{}, &models.Chat{})
}