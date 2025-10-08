package handlers

import (
	"net/http"
	"strconv"
	"webtest/database"
	"webtest/utils"

	"github.com/gin-gonic/gin"
)

// GetChatHistoryHandler 处理获取聊天历史记录的HTTP请求
func GetChatHistoryHandler(c *gin.Context) {
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

	// 从查询参数中获取其他参数
	cid := c.Query("cid")
	countStr := c.Query("count")

	if cid == "" || countStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters: cid, or count"})
		return
	}
	// 转换count参数为整数
	count, err := strconv.Atoi(countStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid count parameter"})
		return
	}

	chatRecords, err := database.GetChatHistory(username, cid, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": chatRecords,
	})
}
