package middleware

import (
	"ai-qa-backend/internal/configs"
	"ai-qa-backend/internal/handler/response"
	"ai-qa-backend/internal/pkg/e"
	"ai-qa-backend/internal/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	jwtHelper := jwt.NewJWT(configs.Conf.JWT.Secret)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, e.Unauthorized, "请求未携带token")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Fail(c, e.Unauthorized, "token格式不正确")
			c.Abort()
			return
		}

		claims, err := jwtHelper.ParseToken(parts[1])
		if err != nil {
			response.Fail(c, e.Unauthorized, "token无效或已过期")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userTier", claims.UserTier)
		c.Next()
	}
}
