package httpapi

import (
	"github.com/gin-gonic/gin"

	"backEnd/internal/middleware"
)

func MountRoutes(r *gin.Engine) {
	auth := NewAuthHandlers()
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
		}
	}
}
