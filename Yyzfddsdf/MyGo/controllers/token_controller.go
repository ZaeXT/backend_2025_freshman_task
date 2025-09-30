package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// RechargeRequest 充值请求结构体
type RechargeRequest struct {
	TokenAmount int `json:"tokenAmount" binding:"required,min=1"`
}

// RechargeHandle 处理token充值
func RechargeHandle(c *gin.Context) {
	// 从中间件获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 查询当前用户
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 增加token数量
	newTokenCount := user.TokenCount + req.TokenAmount
	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("token_count", newTokenCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "充值失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "充值成功",
		"tokenCount": newTokenCount,
		"added":      req.TokenAmount,
	})
}

// GetTokenInfoHandle 获取用户token信息
func GetTokenInfoHandle(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokenCount": user.TokenCount,
		"username":   user.Username,
		"email":      user.Email,
	})
}
