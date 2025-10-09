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

type ChatHandlers struct {
	chatRepo *repo.ChatRepository
	aiClient *ai.Client
}

func NewChatHandlers() *ChatHandlers {
	return &ChatHandlers{chatRepo: repo.NewChatRepository(), aiClient: ai.NewClient()}
}

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
	userID, err := primitive.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uid"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	// Ensure conversation
	var conv *models.Conversation
	if req.ConversationID == "" {
		title := req.Title
		if title == "" {
			title = "新对话"
		}
		model := req.Model
		if model == "" {
			model = h.aiClient.Model()
		}
		conv, err = h.chatRepo.UpsertConversation(ctx, userID, title, model)
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
		reply, err := h.aiClient.Chat(ctx, toAIMessages(req.Messages))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = h.chatRepo.InsertMessage(ctx, &models.ChatMessage{ConversationID: conv.ID, UserID: userID, Role: models.MsgAssistant, Content: reply})
		c.JSON(http.StatusOK, gin.H{"conversationId": conv.ID.Hex(), "content": reply})
		return
	}

	// stream via SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	var assembled string
	err = h.aiClient.ChatStream(ctx, toAIMessages(req.Messages), func(delta string) error {
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
