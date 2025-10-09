package httpapi

import (
	"github.com/gin-gonic/gin"

	"backEnd/internal/middleware"
)

// MountRoutes 注册所有 HTTP 路由与中间件。
func MountRoutes(r *gin.Engine) {
	auth := NewAuthHandlers()
	users := NewUserHandlers()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", auth.Register)
		v1.POST("/auth/login", auth.Login)

		// protected
		chat := NewChatHandlers()
		ap := v1.Group("")
		ap.Use(middleware.AuthRequired())
		{
			ap.POST("/chat", chat.Chat)
			ap.GET("/conversations/:id", chat.GetConversation)
			// admin-only APIs
			admin := ap.Group("")
			admin.Use(middleware.RequireRoles("admin"))
			{
				admin.PUT("/users/:id/role", users.SetRole)
			}
		}
	}
}
