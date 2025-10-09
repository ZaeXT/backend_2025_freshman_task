package middleware

import (
	"ai-qa-system/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Gin 上下文获取用户 ID（JWT 解码时放进去的）
		userID := c.GetString("user_id") // 注意用 string，和策略里的 user0、user1 对应

		// 从请求参数或 body 里获取要调用的 AI 模型
		model := c.Query("model") // 如果是 query 参数，比如 /qa/ask?model=deepseek-v3.1
		if model == "" {
			model = c.PostForm("model") // 如果是 POST 表单
		}
		if model == "" {
			// 也可以从 JSON 里取
			var body struct {
				Model string `json:"model"`
			}
			if err := c.ShouldBindJSON(&body); err == nil {
				model = body.Model
			}
		}

		// 如果 model 为空，直接拒绝
		if model == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model parameter required"})
			c.Abort()
			return
		}

		// Casbin 鉴权：是否允许该用户使用该模型
		allowed, _ := utils.Enforcer.Enforce(userID, model, "allow")

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort()
			return
		}

		// 通过鉴权，继续请求
		c.Next()
	}
}
