package routes

import (
	"ai-qa-system/controllers"
	"ai-qa-system/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}

	qa := r.Group("/qa").Use(middleware.JWTAuth())
	{
		qa.POST("/ask", controllers.AskQuestion)
		qa.GET("/history", controllers.GetHistory)
		qa.DELETE("/delete/:id", controllers.DeleteQuestion)
		qa.DELETE("/clear", controllers.ClearHistory)
	}
}
