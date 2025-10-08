package main

import (
	"log"
	"webtest/database"
	"webtest/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 连接数据库
	database.ConnectDatabase()

	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		v1.POST("/users", handlers.RegisterUser)
		v1.POST("/login", handlers.LoginHandler)
		v1.POST("/change-password", handlers.ChangePasswordHandler)
		v1.GET("/check-username", handlers.CheckUsernameAvailability)
		// 获取可用模型需要Token
		v1.GET("/available-models", handlers.GetAvailableModels)
		// 聊天记录相关路由，需要Token
		v1.GET("/chat-history", handlers.GetChatHistoryHandler)
		v1.GET("/user-ChatContentGet", handlers.GetUserChatContent)

		// AI聊天请求路由，Token作为参数
		v1.GET("/AIChatRequest", handlers.AIChatRequest)
	}

	log.Println("Server starting on localhost:8080")
	err := r.Run("localhost:8080")
	if err != nil {
		return
	}
}
