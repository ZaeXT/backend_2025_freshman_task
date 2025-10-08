package main

import (
	"fmt"
	"log"
	"net/http"

	"Piao/config"
	"Piao/database"
	"Piao/handlers"
	"Piao/middleware"
)

func main() {
	// 1. 初始化配置
	if err := config.Init(); err != nil {
		log.Fatal("❌ 配置初始化失败:", err)
	}

	// 2. 初始化数据库
	dbUser, dbPassword := config.GetDBConfig()
	if dbUser == "" || dbPassword == "" {
		log.Fatal("❌ 数据库配置未设置")
	}

	db, err := database.Init(dbUser, dbPassword)
	if err != nil {
		log.Fatal("❌ 数据库初始化失败:", err)
	}
	defer db.Close()

	// 将数据库连接保存到config中
	config.DB = db

	// 3. 注册路由
	// 公开接口（不需要认证）
	http.HandleFunc("/api/register", handlers.Register)
	http.HandleFunc("/api/login", handlers.Login)
	http.HandleFunc("/", handlers.ServeHTML)

	// 需要认证的接口
	http.HandleFunc("/api/conversations", middleware.AuthMiddleware(handlers.GetConversations))
	http.HandleFunc("/api/conversation/create", middleware.AuthMiddleware(handlers.CreateConversation))
	http.HandleFunc("/api/messages", middleware.AuthMiddleware(handlers.GetMessages))
	http.HandleFunc("/api/chat", middleware.AuthMiddleware(handlers.Chat))
	http.HandleFunc("/api/chat/stream", middleware.AuthMiddleware(handlers.ChatStream))
	http.HandleFunc("/api/upgrade", middleware.AuthMiddleware(handlers.Upgrade))

	// 4. 启动服务器
	fmt.Println("🚀 服务器启动在 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
