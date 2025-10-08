package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Init 初始化数据库连接
func Init(dbUser, dbPassword string) (*sql.DB, error) {
	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword)

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库ping失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	// 初始化表结构
	if err := initTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

// initTables 创建数据库表
func initTables(db *sql.DB) error {
	queries := []string{
		// users表
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			level INT DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// conversations表
		`CREATE TABLE IF NOT EXISTS conversations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			title VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		// messages表
		`CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			conversation_id INT NOT NULL,
			role VARCHAR(20) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("数据库初始化失败: %v", err)
		}
	}

	log.Println("✅ 数据库表初始化成功")
	return nil
}
