package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"houduan_from/config" // 你的本地 config 包
	"houduan_from/routes" // 你的本地 routes 包
)

func main() {
	// 1️⃣ 自动加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  未找到 .env 文件，将使用系统环境变量（例如火山引擎部署环境）")
	} else {
		log.Println("✅ 已成功加载 .env 文件")
	}

	// 2️⃣ 从环境变量读取配置（适配本地和火山引擎）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认本地端口
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("⚠️  未检测到 DATABASE_URL，将使用 config 包内默认配置")
	} else {
		log.Printf("🔗 使用外部数据库连接: %s\n", dbURL)
	}

	// 3️⃣ 初始化数据库（config.InitDB 内部可以使用 os.Getenv 来动态加载配置）
	config.InitDB()

	// 4️⃣ 初始化 Gin 实例
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 5️⃣ 注册路由
	routes.InitRoutes(r)

	// 6️⃣ 启动服务
	log.Printf("🚀 服务已启动，监听端口 %s ...", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("❌ 启动失败: %v", err)
	}
	fmt.Println("API Key:", os.Getenv("ARK_API_KEY"))
}
