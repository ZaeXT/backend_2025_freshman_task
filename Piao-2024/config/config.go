package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// 全局配置变量
var (
	DB        *sql.DB // 数据库连接
	JWTSecret []byte  // JWT密钥

	// API配置
	VolcengineAPIKey   string
	VolcengineEndpoint string
)

// Init 初始化配置
func Init() error {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found")
	}

	// 加载JWT密钥
	JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(JWTSecret) == 0 {
		log.Fatal("❌ JWT_SECRET not set in .env file")
	}

	// 加载火山引擎API配置
	VolcengineAPIKey = os.Getenv("VOLCENGINE_API_KEY")
	VolcengineEndpoint = "https://ark.cn-beijing.volces.com/api/v3/chat/completions"

	log.Println("✅ 配置加载成功")
	return nil
}

// GetDBConfig 获取数据库配置
func GetDBConfig() (user, password string) {
	return os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")
}
