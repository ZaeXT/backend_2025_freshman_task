package handler

import (
	"ai-qa-backend/internal/handler/request"
	"ai-qa-backend/internal/handler/response"
	"ai-qa-backend/internal/pkg/e"
	"ai-qa-backend/internal/service"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req request.UserRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	if err := h.userService.Register(req.Username, req.Password); err != nil {
		response.Fail(c, e.Error, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req request.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, e.Unauthorized, err.Error())
		return
	}

	response.Success(c, gin.H{"token": token})

}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, e.NotFound, "用户不存在")
		} else {
			response.Fail(c, e.Error, "获取用户信息失败")
		}
		return
	}

	userProfile := response.UserProfile{
		ID:         user.ID,
		Username:   user.Username,
		Tier:       user.Tier,
		MemoryInfo: user.MemoryInfo,
		CreatedAt:  user.CreatedAt,
	}
	response.Success(c, userProfile)
}

func (h *UserHandler) UpdateMemory(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	var req request.UpdateUserMemory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, e.InvalidParams, err.Error())
		return
	}

	if err := h.userService.UpdateUserMemory(userID, req.MemoryInfo); err != nil {
		response.Fail(c, e.Error, "更新用户记忆信息失败")
		return
	}

	response.Success(c, nil)
}
