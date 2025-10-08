package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// TODO: **请在生产环境中替换为复杂且保密的密钥！**
var jwtSecret = []byte("your_very_secret_jwt_key")

// Claims 定义 JWT 的 payload
// 修改Claims结构体，添加更多用户标识
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`    // 添加邮箱
	Username string `json:"username"` // 添加用户名
	jwt.RegisteredClaims
}

// generateJWT 生成 JWT token
func generateJWT(userID uint, email, username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token 24小时后过期

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// RegisterHandle 处理用户注册
func RegisterHandle(c *gin.Context) {
	var payload models.RegisterPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 检查邮箱是否已被注册
	var existingUser models.User
	if err := database.DB.Where("email = ?", payload.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "该邮箱已被注册"})
		return
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := models.User{
		Username:   payload.Username,
		Email:      payload.Email,
		Password:   string(hashedPassword),
		TokenCount: 10, // 明确设置默认token数量
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败: " + err.Error()})
		return
	}

	// 注册成功，返回脱敏后的用户信息
	c.JSON(http.StatusCreated, models.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		TokenCount: user.TokenCount, // 添加这一行
	})
}

// LoginHandle 处理用户登录
func LoginHandle(c *gin.Context) {
	var payload models.LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var user models.User
	// 查找用户
	if err := database.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		// 无论是邮箱不存在还是密码错误，都返回统一的错误信息，以提高安全性
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码不正确"})
		return
	}
	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码不正确"})
		return
	}

	// 生成 JWT Token
	tokenString, err := generateJWT(user.ID, user.Email, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
		return
	}

	// 登录成功，返回用户信息和 Token
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   tokenString,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"tokenCount": user.TokenCount,
		},
	})

}

// DeleteUserHandle 处理用户数据删除请求
func DeleteUserHandle(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法开始数据库事务"})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户数据时发生错误"})
		}
	}()

	// 1. 删除用户的所有消息
	// 首先需要找到用户的所有对话
	var conversations []models.Conversation
	if err := tx.Where("user_id = ?", userID).Find(&conversations).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查找用户对话时出错: " + err.Error()})
		return
	}

	// 删除所有对话中的消息（硬删除）
	for _, conversation := range conversations {
		if err := tx.Unscoped().Where("conversation_id = ?", conversation.ID).Delete(&models.Message{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除对话消息时出错: " + err.Error()})
			return
		}
	}

	// 2. 删除用户的所有对话（硬删除）
	if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.Conversation{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户对话时出错: " + err.Error()})
		return
	}

	// 3. 删除用户
	// 3. 硬删除用户（绕过软删除）
	if err := tx.Unscoped().Where("id = ?", userID).Delete(&models.User{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户时出错: " + err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务时出错: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户数据已成功删除"})
}

// GetJWTSecret 用于从其他包（如中间件）安全地获取 JWT 密钥
func GetJWTSecret() []byte {
	return jwtSecret
}
