package middleware

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"yyz.com/MyGo/controllers"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// Redis客户端
var redisClient *redis.Client

// Redis连接状态
var redisConnected = false

// 初始化Redis客户端
func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis服务器地址
		DB:   0,                // 使用默认数据库
	})

	// 测试Redis连接
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("警告: 无法连接到Redis服务器: %v，nonce验证功能将不可用", err)
	} else {
		redisConnected = true
		log.Println("Redis服务器连接成功")
	}
}

// AuthMiddleware 是一个 Gin 中间件，用于验证请求中的 JWT Token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权 Token"})
			c.Abort()
			return
		}

		// Token 格式通常为 "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &controllers.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return controllers.GetJWTSecret(), nil // 从 controllers 引入密钥
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的 Token"})
			c.Abort()
			return
		}

		// 新增：验证请求头中的时间戳，防止重放攻击
		timestampStr := c.GetHeader("X-Timestamp")
		if timestampStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少时间戳"})
			c.Abort()
			return
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的时间戳格式"})
			c.Abort()
			return
		}

		// 验证时间窗口（30秒）
	now := time.Now().Unix()
	if math.Abs(float64(now-timestamp)) > 30 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请求已过期"})
		c.Abort()
		return
	}

		// 新增：使用nonce防止重放攻击
		nonceStr := c.GetHeader("X-Nonce")
		if nonceStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少nonce"})
			c.Abort()
			return
		}

		// 验证nonce格式（32位十六进制字符串）
		if len(nonceStr) != 32 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的nonce格式"})
			c.Abort()
			return
		}

		// 强制执行nonce验证机制
		// 如果无法连接到Redis服务，应向客户端返回"服务器内部错误"响应
		if !redisConnected {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			c.Abort()
			return
		}

		// 使用Redis验证nonce是否已使用（防重放攻击）
		ctx := context.Background()
		redisKey := fmt.Sprintf("nonce:%s", nonceStr)

		// 检查nonce是否已存在于缓存中（已使用则拒绝）
		exists, err := redisClient.Exists(ctx, redisKey).Result()
		if err != nil {
			// Redis连接错误，记录日志并拒绝请求
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			c.Abort()
			return
		}

		if exists > 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的nonce"})
			c.Abort()
			return
		}

		// 存储nonce到Redis，设置10分钟过期时间
		err = redisClient.Set(ctx, redisKey, "1", 10*time.Minute).Err()
		if err != nil {
			// Redis存储错误，记录日志并拒绝请求
			log.Printf("错误: 无法存储nonce到Redis: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
			c.Abort()
			return
		}

		// 验证用户是否存在于数据库中
		var user models.User
		if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
			c.Abort()
			return
		}

		// 新增：确保数据库中的用户信息与Token声明一致
		if user.Email != claims.Email || user.Username != claims.Username {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不匹配，Token可能已失效"})
			c.Abort()
			return
		}

		// 将 UserID 和完整的用户信息存储在 Context 中，供后续处理器使用
		c.Set("userID", claims.UserID)
		c.Set("user", user) // 存储完整的用户信息
		c.Next()            // 继续处理请求
	}
}
