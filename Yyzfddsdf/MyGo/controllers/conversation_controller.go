package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// GetConversations 获取用户的对话记录列表
func GetConversations(c *gin.Context) {
	// 从中间件获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 确保分页参数有效
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var conversations []models.Conversation
	var total int64

	// 查询用户的所有对话记录（按创建时间倒序）
	query := database.DB.Where("user_id = ?", userID).Order("created_at DESC")

	// 获取总数
	query.Model(&models.Conversation{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	query.Offset(offset).Limit(pageSize).Find(&conversations)

	// 转换为响应结构体
	conversationResponses := make([]models.ConversationResponse, len(conversations))
	for i, conv := range conversations {
		conversationResponses[i] = models.ConversationResponse{
			ID:           conv.ID,
			Title:        conv.Title,
			MessageCount: conv.MessageCount,
			TokenUsed:    conv.TokenUsed,
			CreatedAt:    conv.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversationResponses,
		"total":         total,
		"page":          page,
		"pageSize":      pageSize,
		"totalPages":    (int(total) + pageSize - 1) / pageSize,
	})
}

// GetConversation 获取特定对话记录的详细信息
func GetConversation(c *gin.Context) {
	// 从中间件获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	// 获取对话ID
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少对话ID"})
		return
	}

	var conversation models.Conversation
	// 查询对话记录并验证用户权限
	if err := database.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	// 获取对话中的所有消息（按创建时间正序）
	var messages []models.Message
	database.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages)

	// 转换为响应结构体
	messageResponses := make([]models.MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = models.MessageResponse{
			ID:         msg.ID,
			Role:       msg.Role,
			Content:    msg.Content,
			TokenCount: msg.TokenCount,
			CreatedAt:  msg.CreatedAt,
		}
	}

	// 构造响应
	response := gin.H{
		"id":           conversation.ID,
		"title":        conversation.Title,
		"messageCount": conversation.MessageCount,
		"tokenUsed":    conversation.TokenUsed,
		"createdAt":    conversation.CreatedAt,
		"messages":     messageResponses,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteConversation 删除对话记录
func DeleteConversation(c *gin.Context) {
	// 从中间件获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	// 获取对话ID
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少对话ID"})
		return
	}

	// 查询对话记录并验证用户权限
	var conversation models.Conversation
	if err := database.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "事务开始失败"})
		return
	}

	// 先删除关联的消息记录（硬删除）
	if err := tx.Unscoped().Where("conversation_id = ?", conversationID).Delete(&models.Message{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除消息记录失败"})
		return
	}

	// 再删除对话记录（硬删除）
	if err := tx.Unscoped().Delete(&conversation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除对话记录失败"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// UpdateConversationTitle 更新对话标题
func UpdateConversationTitle(c *gin.Context) {
	// 从中间件获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证的用户"})
		return
	}

	// 获取对话ID
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少对话ID"})
		return
	}

	// 获取请求参数
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 查询对话记录并验证用户权限
	var conversation models.Conversation
	if err := database.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	// 更新标题
	if err := database.DB.Model(&conversation).Update("title", req.Title).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"title":   req.Title,
	})
}
