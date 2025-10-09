package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backEnd/internal/ai"
	"backEnd/internal/middleware"
	"backEnd/internal/models"
	"backEnd/internal/repo"
)

// ChatHandlers 处理聊天相关的 HTTP 请求。
type ChatHandlers struct {
	chatRepo *repo.ChatRepository
	aiClient *ai.Client
}

// NewChatHandlers 创建 ChatHandlers。
func NewChatHandlers() *ChatHandlers {
	return &ChatHandlers{chatRepo: repo.NewChatRepository(), aiClient: ai.NewClient()}
}

// chatReq 聊天请求体。
type chatReq struct {
	ConversationID string       `json:"conversationId"`
	Title          string       `json:"title"`
	Model          string       `json:"model"`
	Messages       []ai.Message `json:"messages" binding:"required,min=1"`
	Stream         bool         `json:"stream"`
}

// POST /api/v1/chat
func (h *ChatHandlers) Chat(c *gin.Context) {
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uidStr := c.GetString(middleware.CtxUserID)
	roleStr := c.GetString(middleware.CtxUserRole)
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uid"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	// 根据角色决策可用模型（并设置默认值）
	// free: qwen-plus only
	// pro: qwen3-max allowed (and qwen-plus)
	// admin: any
	requestedModel := req.Model
	switch roleStr {
	case string(models.RoleFree):
		if requestedModel == "" {
			requestedModel = "qwen-plus"
		}
		if requestedModel != "qwen-plus" {
			c.JSON(http.StatusForbidden, gin.H{"error": "model not allowed for role"})
			return
		}
	case string(models.RolePro):
		if requestedModel == "" {
			requestedModel = "qwen3-max"
		}
		if requestedModel != "qwen3-max" && requestedModel != "qwen-plus" {
			c.JSON(http.StatusForbidden, gin.H{"error": "model not allowed for role"})
			return
		}
	case string(models.RoleAdmin):
		if requestedModel == "" {
			requestedModel = h.aiClient.Model()
		}
		// admin: no restriction
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "unknown role"})
		return
	}

	// 创建或获取会话（新建时会写入最终模型）
	var conv *models.Conversation
	if req.ConversationID == "" {
		title := req.Title
		if title == "" {
			title = "新对话"
		}
		conv, err = h.chatRepo.UpsertConversation(ctx, userID, title, requestedModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// reuse conversation id (no extra check for brevity)
		cid, _ := primitive.ObjectIDFromHex(req.ConversationID)
		conv = &models.Conversation{ID: cid, UserID: userID}
	}

	// save user message
	last := req.Messages[len(req.Messages)-1]
	_ = h.chatRepo.InsertMessage(ctx, &models.ChatMessage{ConversationID: conv.ID, UserID: userID, Role: models.MessageRole(last.Role), Content: last.Content})

	if !req.Stream {
		reply, err := h.aiClient.ChatWithModel(ctx, requestedModel, toAIMessages(req.Messages))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = h.chatRepo.InsertMessage(ctx, &models.ChatMessage{ConversationID: conv.ID, UserID: userID, Role: models.MsgAssistant, Content: reply})
		c.JSON(http.StatusOK, gin.H{"conversationId": conv.ID.Hex(), "content": reply})
		return
	}

	// SSE 流式返回
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	var assembled string
	err = h.aiClient.ChatStreamWithModel(ctx, requestedModel, toAIMessages(req.Messages), func(delta string) error {
		assembled += delta
		c.SSEvent("message", delta)
		c.Writer.Flush()
		return nil
	})
	if err != nil {
		return
	}
	_ = h.chatRepo.InsertMessage(ctx, &models.ChatMessage{ConversationID: conv.ID, UserID: userID, Role: models.MsgAssistant, Content: assembled})
}

func toAIMessages(msgs []ai.Message) []ai.Message { return msgs }

// GET /api/v1/conversations/:id
func (h *ChatHandlers) GetConversation(c *gin.Context) {
	cidStr := c.Param("id")
	uidStr := c.GetString(middleware.CtxUserID)
	convID, err := primitive.ObjectIDFromHex(cidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uid"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	conv, err := h.chatRepo.FindConversationByIDAndUser(ctx, convID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": conv.ID.Hex(), "title": conv.Title, "model": conv.Model, "createdAt": conv.CreatedAt, "updatedAt": conv.UpdatedAt})
}
