package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// SignatureMiddleware 是一个 Gin 中间件，用于验证请求签名
func SignatureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取签名相关的请求头
		signature := c.GetHeader("X-Signature")
		timestamp := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")

		if signature == "" || timestamp == "" || nonce == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要的签名参数"})
			c.Abort()
			return
		}

		// 获取请求方法和路径
		method := c.Request.Method
		path := c.Request.URL.Path

		// 获取查询参数
		queryParams := c.Request.URL.Query()

		// 获取请求体
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取请求体"})
			c.Abort()
			return
		}

		// 重新设置请求体，确保后续处理可以正常读取
		c.Request.Body = nil

		// 构建签名字符串
		signString := buildSignString(method, path, timestamp, nonce, queryParams, body)

		// 验证签名（这里需要使用客户端和服务端共享的密钥）
		// 注意：在实际应用中，应该根据客户端ID或其他标识获取对应的密钥
		secretKey := "your_shared_secret_key" // 这应该从配置或数据库中获取

		if !verifySignature(signString, signature, secretKey) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "签名验证失败"})
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// buildSignString 构建签名字符串
func buildSignString(method, path, timestamp, nonce string, queryParams url.Values, body []byte) string {
	// 1. 添加HTTP方法
	signParts := []string{method}

	// 2. 添加路径
	signParts = append(signParts, path)

	// 3. 添加时间戳
	signParts = append(signParts, timestamp)

	// 4. 添加nonce
	signParts = append(signParts, nonce)

	// 5. 添加查询参数（按键排序）
	if len(queryParams) > 0 {
		var queryParts []string
		for key, values := range queryParams {
			for _, value := range values {
				queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
			}
		}
		sort.Strings(queryParts)
		signParts = append(signParts, strings.Join(queryParts, "&"))
	}

	// 6. 添加请求体（如果有）
	if len(body) > 0 {
		signParts = append(signParts, string(body))
	}

	// 使用&连接所有部分
	return strings.Join(signParts, "&")
}

// verifySignature 验证签名
func verifySignature(signString, signature, secretKey string) bool {
	// 创建HMAC签名
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(signString))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// 比较签名（使用安全的字符串比较防止时序攻击）
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
