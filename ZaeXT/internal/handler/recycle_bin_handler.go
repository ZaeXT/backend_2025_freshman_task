package handler

import (
	"ai-qa-backend/internal/handler/response"
	"ai-qa-backend/internal/pkg/e"
	"ai-qa-backend/internal/service"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RecycleBinHandler struct {
	recycleBinService service.RecycleBinService
}

func NewRecycleBinHandler(recycleBinService service.RecycleBinService) *RecycleBinHandler {
	return &RecycleBinHandler{recycleBinService: recycleBinService}
}

func (h *RecycleBinHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")

	convs, err := h.recycleBinService.List(userID.(uint))
	if err != nil {
		response.Fail(c, e.Error, "获取回收站列表失败")
		return
	}

	convInfos := make([]*response.ConversationInfo, len(convs))
	for i, conv := range convs {
		var deletedAt *time.Time
		if conv.DeletedAt.Valid {
			deletedAt = &conv.DeletedAt.Time
		}
		convInfos[i] = &response.ConversationInfo{
			ID:          conv.ID,
			Title:       conv.Title,
			IsTemporary: conv.IsTemporary,
			CategoryID:  conv.CategoryID,
			CreatedAt:   conv.CreatedAt,
			UpdatedAt:   conv.UpdatedAt,
			DeletedAt:   deletedAt,
		}
	}

	response.Success(c, convInfos)
}

func (h *RecycleBinHandler) Restore(c *gin.Context) {
	userID, _ := c.Get("userID")
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	if err := h.recycleBinService.Restore(uint(convID), userID.(uint)); err != nil {
		response.Fail(c, e.Error, "恢复对话失败")
		return
	}

	response.Success(c, nil)
}

func (h *RecycleBinHandler) PermanentDelete(c *gin.Context) {
	userID, _ := c.Get("userID")
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, e.InvalidParams, "无效的对话ID")
		return
	}

	if err := h.recycleBinService.PermanentDelete(uint(convID), userID.(uint)); err != nil {
		response.Fail(c, e.Error, "永久删除对话失败")
		return
	}

	response.Success(c, nil)
}
