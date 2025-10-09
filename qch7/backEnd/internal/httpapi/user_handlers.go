package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backEnd/internal/models"
	"backEnd/internal/repo"
)

// UserHandlers 处理用户管理相关的 HTTP 请求。
type UserHandlers struct {
    users *repo.UserRepository
}

// NewUserHandlers 创建 UserHandlers。
func NewUserHandlers() *UserHandlers {
    return &UserHandlers{users: repo.NewUserRepository()}
}

// setRoleReq 设定角色的请求体。
type setRoleReq struct {
    Role string `json:"role" binding:"required"`
}

// PUT /api/v1/users/:id/role
func (h *UserHandlers) SetRole(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req setRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// validate role
	switch models.UserRole(req.Role) {
	case models.RoleFree, models.RolePro, models.RoleAdmin:
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := h.users.UpdateRole(ctx, oid, models.UserRole(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
