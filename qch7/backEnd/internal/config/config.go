package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

// AppConfig holds all runtime configuration values.
type AppConfig struct {
	Port           string
	MongoURI       string
	MongoDBName    string
	JWTSecret      string
	AIAPIKey       string
	AIBaseURL      string
	AIModel        string
	RequestTimeout time.Duration
}

var cfg AppConfig

// Load loads environment variables and prepares application config.
func Load() AppConfig {
	cfg = AppConfig{
		Port:        getEnv("PORT", ":8080"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB", "qa_app"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		// 兼容阿里云百炼的DashScope兼容模式（OpenAI风格）
		AIAPIKey:       getEnv("DASHSCOPE_API_KEY", getEnv("AI_API_KEY", "")),
		AIBaseURL:      getEnv("AI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
		AIModel:        getEnv("AI_MODEL", "qwen-plus"),
		RequestTimeout: getEnvDuration("REQUEST_TIMEOUT_SECONDS", 60),
	}

	if cfg.AIAPIKey == "" {
		log.Println("[warn] AI_API_KEY 未配置，AI接口将不可用")
	}

	return cfg
}

// Get returns loaded configuration. Ensure Load() is called during startup.
func Get() AppConfig { return cfg }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallbackSeconds int) time.Duration {
	if v := os.Getenv(key); v != "" {
		// simple parse as seconds
		// ignore error and fallback to default
		var n int
		_, err := fmt.Sscanf(v, "%d", &n)
		if err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(fallbackSeconds) * time.Second
}
