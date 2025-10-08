package routes

import (
	"houduan_from/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", controllers.Register)
		v1.POST("/login", controllers.Login)

		auth := v1.Group("/")
		auth.Use(controllers.AuthMiddleware())
		{
			auth.POST("/chat", controllers.Chat)
			auth.GET("/chat/history", controllers.GetChatHistory)
		}
	}
}
