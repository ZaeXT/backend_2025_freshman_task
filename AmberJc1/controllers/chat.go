package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"time"

	"houduan_from/config"
	"houduan_from/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 用户注册
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

	// 对密码做 MD5 加密
	h := md5.New()
	h.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(h.Sum(nil))

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// 用户登录
func Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 把输入的密码 MD5 加密
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

// JWT 鉴权中间件
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

// 聊天接口
func Chat(c *gin.Context) {
	var chatData struct {
		Question string `json:"question"`
	}
	if err := c.ShouldBindJSON(&chatData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	userID := c.GetInt("user_id")
	answer := "这是AI的回答：" + chatData.Question

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

// 获取聊天记录
func GetChatHistory(c *gin.Context) {
	userID := c.GetInt("user_id")
	var chats []models.Chat
	config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&chats)
	c.JSON(http.StatusOK, chats)
}
