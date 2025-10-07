package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"AIBackend/pkg/auth"
)

// Allowed models by role (exported for reuse)
var AllowedModelsByRole = map[string][]string{
	"free":  {"mock-mini", "gpt-4o-mini"},
	"pro":   {"mock-mini", "mock-pro", "gpt-4o-mini", "gpt-4o"},
	"admin": {"mock-mini", "mock-pro", "mock-admin", "gpt-4o-mini", "gpt-4o", "gpt-4.1"},
}

// CheckModelAccess returns true if the role is allowed to use the model.
func CheckModelAccess(role, model string) bool {
	if model == "" {
		return true
	}
	list := AllowedModelsByRole[role]
	for _, m := range list {
		if m == model {
			return true
		}
	}
	return false
}

// AuthRequired validates JWT and sets user info in context.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := auth.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// ModelAccess enforces role-based access to models using query parameter if present.
func ModelAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("user_role")
		roleStr := "free"
		if r, ok := role.(string); ok && r != "" {
			roleStr = r
		}
		reqModel := c.Query("model")
		if reqModel == "" {
			// body may contain model; handler should validate with CheckModelAccess
			c.Next()
			return
		}
		if !CheckModelAccess(roleStr, reqModel) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "model access denied for role"})
			return
		}
		c.Next()
	}
}
