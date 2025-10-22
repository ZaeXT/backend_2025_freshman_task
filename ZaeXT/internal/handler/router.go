package handler

import (
	"ai-qa-backend/internal/handler/middleware"
	"ai-qa-backend/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(services *service.Service) *gin.Engine {
	router := gin.Default()
	config := cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:   []string{"Content-Length"},
		MaxAge:          12 * time.Hour,
	}
	router.Use(cors.New(config))

	apiV1 := router.Group("/api/v1")

	userHandler := NewUserHandler(services.User)
	chatHandler := NewChatHandler(services.Chat)
	categoryHandler := NewCategoryHandler(services.Category)
	recycleBinHandler := NewRecycleBinHandler(services.RecycleBin)

	apiV1.POST("/register", userHandler.Register)
	apiV1.POST("/login", userHandler.Login)

	authGroup := apiV1.Group("")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.GET("/profile", userHandler.GetProfile)
		authGroup.PUT("/profile/memory", userHandler.UpdateMemory)
		authGroup.GET("/models", chatHandler.ListModels)
		authGroup.POST("/conversations", chatHandler.CreateConversation)
		authGroup.GET("/conversations", chatHandler.ListConversations)
		authGroup.POST("/conversations/:id/messages", chatHandler.ProcessMessage)
		authGroup.PUT("/conversations/:id/title", chatHandler.UpdateTitle)
		authGroup.DELETE("/conversations/:id", chatHandler.DeleteConversation)
		authGroup.POST("/conversations/:id/auto-classify", chatHandler.AutoClassify)
		authGroup.POST("/categories", categoryHandler.Create)
		authGroup.GET("/categories", categoryHandler.List)
		authGroup.PUT("/categories/:id", categoryHandler.Update)
		authGroup.DELETE("/categories/:id", categoryHandler.Delete)
		authGroup.GET("/recycle-bin", recycleBinHandler.List)
		authGroup.POST("/recycle-bin/restore/:id", recycleBinHandler.Restore)
		authGroup.DELETE("/recycle-bin/permanent/:id", recycleBinHandler.PermanentDelete)
	}

	return router
}
