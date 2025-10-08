package handlers

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"webtest/database"
	"webtest/models"
	"webtest/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 对输入的密码进行MD5加密
	hash := md5.Sum([]byte(input.Password))
	passwordMD5 := fmt.Sprintf("%x", hash)

	user := models.User{
		Username: input.Username,
		Password: passwordMD5,
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"data": gin.H{
			"username": user.Username,
		},
	})
}

func RegisterUser(c *gin.Context) {
	CreateUser(c)
}

// CheckUsernameAvailability 检查用户名是否可用
func CheckUsernameAvailability(c *gin.Context) {
	// 从查询参数中获取用户名
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username parameter is required"})
		return
	}

	// 查询数据库中是否已存在该用户名
	var user models.User
	result := database.DB.Where("username = ?", username).First(&user)

	// 根据查询结果返回相应的响应
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{
			"available": true,
			"message":   "Username is available",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   "Username is already taken",
		})
	}
}

// LoginHandler 处理用户登录请求
func LoginHandler(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 查询用户
	var user models.User
	result := database.DB.Select("username", "passwordmd5").Where("username = ?", input.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	// 对输入的密码进行MD5加密
	hash := md5.Sum([]byte(input.Password))
	passwordMD5 := fmt.Sprintf("%x", hash)
	// 比较加密后的密码
	if user.Password == passwordMD5 {
		// 生成JWT令牌
		token, err := utils.GenerateToken(user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"data": gin.H{
				"username": user.Username,
				"token":    token,
			},
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	}
}

// ChangePasswordHandler 处理用户修改密码请求
func ChangePasswordHandler(c *gin.Context) {
	var input struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户
	var user models.User
	result := database.DB.Select("username", "passwordmd5").Where("username = ?", input.Username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 对输入的原密码进行MD5加密
	hash := md5.Sum([]byte(input.Password))
	passwordMD5 := fmt.Sprintf("%x", hash)

	// 验证原密码是否正确
	if user.Password != passwordMD5 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 对新密码进行MD5加密
	newHash := md5.Sum([]byte(input.NewPassword))
	newPasswordMD5 := fmt.Sprintf("%x", newHash)

	// 更新数据库中的密码
	updateResult := database.DB.Model(&models.User{}).Where("username = ?", input.Username).Update("passwordmd5", newPasswordMD5)
	if updateResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// GetUserPermission 根据用户名获取用户权限
func GetUserPermission(username string) (int, error) {
	var user models.User
	result := database.DB.Select("permission").Where("username = ?", username).First(&user)
	if result.Error != nil {
		return -1, result.Error
	}
	return user.Permission, nil
}

// GetUserChatContent 获取用户聊天记录对话
func GetUserChatContent(c *gin.Context) {
	// 从Authorization头部获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// 检查Bearer前缀
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	// 提取token
	tokenString := authHeader[7:]

	// 验证token并获取用户名
	username, err := utils.VerifyToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 查询数据库获取该用户的所有对话ID
	cids, err := database.GetUserConversationIds(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user conversations"})
		return
	}

	// 返回用户的所有对话ID
	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"cids":     cids,
	})
}

// GetAvailableModels 获取用户可用的模型列表
func GetAvailableModels(c *gin.Context) {
	// 从Authorization头部获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// 检查Bearer前缀
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	// 提取token
	tokenString := authHeader[7:]

	// 验证token并获取用户名
	username, err := utils.VerifyToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 查询数据库获取用户权限
	userPermission, err := GetUserPermission(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user permission"})
		return
	}

	// 从配置文件加载模型配置
	modelsConfig, err := utils.LoadModelsConfig("models_config.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load models configuration"})
		return
	}

	// 根据用户权限获取可用模型
	availableModels := modelsConfig.GetModelsByPermission(userPermission)

	// 返回可用模型列表
	c.JSON(http.StatusOK, gin.H{
		"username":   username,
		"permission": userPermission,
		"models":     availableModels,
	})
}
