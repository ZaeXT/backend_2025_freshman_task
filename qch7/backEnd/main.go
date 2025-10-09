package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"backEnd/internal/config"
	"backEnd/internal/db"
	"backEnd/internal/httpapi"
	"backEnd/internal/models"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize MongoDB
	if _, err := db.Init(context.Background()); err != nil {
		log.Fatalf("MongoDB 初始化失败: %v", err)
	}

	// Ensure DB indexes
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := models.EnsureIndexes(ctx, db.DB()); err != nil {
			log.Fatalf("索引初始化失败: %v", err)
		}
	}

	// Setup router
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:5500"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	httpapi.MountRoutes(r)

	log.Printf("服务启动于 %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
