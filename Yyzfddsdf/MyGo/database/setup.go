package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"yyz.com/MyGo/models"
)

var DB *gorm.DB

func ConnectDB() {
	// 从环境变量获取数据库连接信息，如果未设置则使用默认值
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// 默认数据库连接信息
		dsn = "root:Yyz123456@tcp(127.0.0.1:3306)/User_System?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// 连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	// 配置数据库连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("获取数据库连接池失败: ", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	// 自动迁移模型
	err = DB.AutoMigrate(&models.User{}, &models.Conversation{}, &models.Message{})
	if err != nil {
		log.Fatal("自动迁移模型失败: ", err)
	}

	fmt.Println("数据库连接成功")
}
