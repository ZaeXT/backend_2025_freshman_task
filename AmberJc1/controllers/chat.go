package controllers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"houduan_from/config"
	"houduan_from/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ====================== 用户注册 ======================
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if user.Username == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码不能为空"})
		return
	}

	// 密码 MD5 加密
	h := md5.New()
	h.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(h.Sum(nil))

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// ====================== 用户登录 ======================
func Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 对密码做 MD5 加密再验证
	h := md5.New()
	h.Write([]byte(loginData.Password))
	password := hex.EncodeToString(h.Sum(nil))

	var user models.User
	if err := config.DB.Where("username = ? AND password = ?", loginData.Username, password).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// ====================== JWT 鉴权中间件 ======================
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供token"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token解析失败"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", int(claims["user_id"].(float64)))
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效"})
			c.Abort()
		}
	}
}

// ====================== 聊天接口（接入火山引擎） ======================
func Chat(c *gin.Context) {
	var chatData struct {
		Question string `json:"question"`
		ImageURL string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&chatData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID := c.GetInt("user_id")

	apiKey := os.Getenv("VOLC_API_KEY")
	apiURL := os.Getenv("VOLC_API_URL")
	model := os.Getenv("VOLC_MODEL_ID")

	if apiKey == "" || apiURL == "" || model == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "火山引擎配置缺失"})
		return
	}

	// 构建消息内容
	content := []map[string]interface{}{}
	if chatData.ImageURL != "" {
		content = append(content, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": chatData.ImageURL,
			},
		})
	}
	content = append(content, map[string]interface{}{
		"type": "text",
		"text": chatData.Question,
	})

	// 构建请求体
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用火山引擎失败", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  fmt.Sprintf("火山接口错误: %s", resp.Status),
			"detail": string(body),
		})
		return
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析火山返回失败", "detail": err.Error()})
		return
	}

	// 提取回答内容
	answer := "AI暂无回复"
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if contentArr, ok := message["content"].([]interface{}); ok && len(contentArr) > 0 {
					if first, ok := contentArr[0].(map[string]interface{}); ok {
						if text, ok := first["text"].(string); ok {
							answer = text
						}
					}
				}
			}
		}
	}

	// 保存聊天记录
	chat := models.Chat{
		UserID:   uint(userID),
		Question: chatData.Question,
		Answer:   answer,
	}
	config.DB.Create(&chat)

	c.JSON(http.StatusOK, gin.H{
		"question": chatData.Question,
		"answer":   answer,
	})
}

// ====================== 获取聊天记录 ======================
func GetChatHistory(c *gin.Context) {
	userID := c.GetInt("user_id")
	var chats []models.Chat
	config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&chats)
	c.JSON(http.StatusOK, chats)
}
