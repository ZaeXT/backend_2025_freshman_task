package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	APIKey      string
	BaseURL     string
	MaxTokens   int
	Temperature float64
	Provider    string
}

// AI模型配置
type ModelConfig struct {
	Name        string
	Description string
	MaxTokens   int
	Temperature float64
	APIEndpoint string
}

func NewConfig() *Config {
	// 方法1：从环境变量读取
	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	// 方法2：如果环境变量没有，尝试从配置文件读取
	if apiKey == "" {
		apiKey = readAPIKeyFromConfigFile()
	}

	// 方法3：如果还是没有，使用默认值（模拟模式）
	if apiKey == "" {
		apiKey = "free-api-key"
		fmt.Println("⚠️  使用模拟响应模式（未配置API Key）")
	} else {
		fmt.Printf("✅ 已读取API Key，长度: %d\n", len(apiKey))
	}

	// 决定使用真实API还是模拟响应
	provider := "mock"
	if apiKey != "free-api-key" && strings.HasPrefix(apiKey, "sk-") {
		provider = "deepseek"
		fmt.Println("✅ API Key格式正确，启用DeepSeek API")
	} else if apiKey != "free-api-key" {
		fmt.Printf("❌ API Key格式不正确，当前: %s...\n", safeSubstring(apiKey, 10))
		fmt.Println("⚠️  将使用模拟响应模式")
		apiKey = "free-api-key" // 强制使用模拟模式
	}

	return &Config{
		APIKey:      apiKey,
		BaseURL:     "https://api.deepseek.com",
		MaxTokens:   1024,
		Temperature: 0.7,
		Provider:    provider,
	}
}

// 从配置文件读取API Key
func readAPIKeyFromConfigFile() string {
	configFiles := []string{
		"config.json",
		"config/config.json",
		"./config.json",
	}

	for _, filename := range configFiles {
		if data, err := os.ReadFile(filename); err == nil {
			var config struct {
				APIKey string `json:"api_key"`
			}
			if json.Unmarshal(data, &config) == nil && config.APIKey != "" {
				fmt.Printf("✅ 从配置文件 %s 读取API Key\n", filename)
				return config.APIKey
			}
		}
	}

	return ""
}

// 获取模型配置
func GetModelConfig(model string) *ModelConfig {
	switch model {
	case "basic":
		return &ModelConfig{
			Name:        "基础模型",
			Description: "适合日常对话和简单问答",
			MaxTokens:   1024,
			Temperature: 0.7,
			APIEndpoint: "/chat/completions",
		}
	case "advanced":
		return &ModelConfig{
			Name:        "高级模型",
			Description: "适合复杂问题分析和创意写作",
			MaxTokens:   2048,
			Temperature: 0.7,
			APIEndpoint: "/chat/completions",
		}
	case "premium":
		return &ModelConfig{
			Name:        "旗舰模型",
			Description: "适合专业领域和深度思考",
			MaxTokens:   4096,
			Temperature: 0.8,
			APIEndpoint: "/chat/completions",
		}
	default:
		return &ModelConfig{
			Name:        "基础模型",
			Description: "适合日常对话和简单问答",
			MaxTokens:   1024,
			Temperature: 0.7,
			APIEndpoint: "/chat/completions",
		}
	}
}

// 安全截取字符串
func safeSubstring(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}
