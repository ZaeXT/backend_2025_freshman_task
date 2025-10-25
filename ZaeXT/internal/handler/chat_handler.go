package handler

import (
	"ai-qa-backend/internal/handler/request"
	"ai-qa-backend/internal/handler/response"
	"ai-qa-backend/internal/model"
	"ai-qa-backend/internal/pkg/e"
	"ai-qa-backend/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) AutoClassify(c *gin.Context) {
	userID, _ := c.Get("userID")
	conv, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	if err := h.chatService.AutoClassify(uint(conv), userID.(uint)); err != nil {
		response.Fail(c, e.Error, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ChatHandler) UpdateConversationCategory(c *gin.Context) {
	userID, _ := c.Get("userID")
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	var req request.UpdateConversationCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	if err := h.chatService.UpdateConversationCategory(uint(convID), userID.(uint), req.CategoryID); err != nil {
		response.Fail(c, e.PermissionDenied, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req request.CreateConversation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	conv, err := h.chatService.CreateConversation(userID.(uint), req.IsTemporary, req.CategoryID)
	if err != nil {
		response.Fail(c, e.Error, "创建对话失败")
		return
	}

	response.Success(c, response.ConversationInfo{
		ID:          conv.ID,
		Title:       conv.Title,
		IsTemporary: conv.IsTemporary,
		CategoryID:  conv.CategoryID,
		CreatedAt:   conv.CreatedAt,
		UpdatedAt:   conv.UpdatedAt,
	})
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, _ := c.Get("userID")

	convs, err := h.chatService.ListConversations(userID.(uint))
	if err != nil {
		response.Fail(c, e.Error, "获取对话列表失败")
		return
	}

	convInfos := h.transformConversationsToDTO(convs)

	response.Success(c, convInfos)
}

func (h *ChatHandler) UpdateTitle(c *gin.Context) {
	userID, _ := c.Get("userID")
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	var req request.UpdateTitle
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	if err := h.chatService.UpdateConversationTitle(uint(convID), userID.(uint), req.Title); err != nil {
		response.Fail(c, e.Error, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID, _ := c.Get("userID")
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	if err := h.chatService.DeleteConversation(uint(convID), userID.(uint)); err != nil {
		response.Fail(c, e.Error, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *ChatHandler) ProcessMessage(c *gin.Context) {
	conv, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)
	userTierVal, _ := c.Get("userTier")
	userTier := userTierVal.(string)

	var req request.ChatMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	responseChan, errChan := h.chatService.ProcessUserMessage(uint(conv), userID, userTier, req.Message, req.ModelID, req.EnableThinking)
	select {
	case initialError := <-errChan:
		if strings.Contains(initialError.Error(), "permission denied") {
			response.Fail(c, e.PermissionDenied, initialError.Error())
		} else if strings.Contains(initialError.Error(), "not found") {
			response.Fail(c, e.NotFound, initialError.Error())
		} else {
			response.Fail(c, e.Error, initialError.Error())
		}
		return
	case firstChunk, ok := <-responseChan:
		if !ok {
			c.Status(http.StatusOK)
			return
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.WriteHeader(http.StatusOK)

		fmt.Fprintf(c.Writer, "data: %s\n\n", firstChunk)
		c.Writer.Flush()

		for {
			select {
			case chunk, ok := <-responseChan:
				if !ok {
					fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
					c.Writer.Flush()
					return
				}
				fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
				c.Writer.Flush()
			case streamError, ok := <-errChan:
				if !ok {
					fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
					c.Writer.Flush()
					return
				}
				if streamError != nil {
					errorPayload := gin.H{"error": streamError.Error()}
					jsonData, _ := json.Marshal(errorPayload)
					fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
					c.Writer.Flush()
				}
				return
			case <-c.Request.Context().Done():
				return
			}
		}
	}
}

func (h *ChatHandler) transformConversationsToDTO(convs []*model.Conversation) []response.ConversationInfo {
	convInfos := make([]response.ConversationInfo, len(convs))
	for i, conv := range convs {
		convInfos[i] = response.ConversationInfo{
			ID:          conv.ID,
			Title:       conv.Title,
			IsTemporary: conv.IsTemporary,
			CategoryID:  conv.CategoryID,
			CreatedAt:   conv.CreatedAt,
			UpdatedAt:   conv.UpdatedAt,
		}
	}
	return convInfos
}

func (h *ChatHandler) ListModels(c *gin.Context) {
	userTierVal, _ := c.Get("userTier")
	userTier := userTierVal.(string)

	models := h.chatService.ListAvailableModels(userTier)

	response.Success(c, models)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, _ := c.Get("userID")
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	messages, err := h.chatService.GetMessagesByConversationID(uint(convID), userID.(uint))
	if err != nil {
		response.Fail(c, e.PermissionDenied, err.Error())
		return
	}

	messageInfo := make([]response.MessageInfo, len(messages))
	for i, msg := range messages {
		messageInfo[i] = response.MessageInfo{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		}
	}

	response.Success(c, messageInfo)
}
