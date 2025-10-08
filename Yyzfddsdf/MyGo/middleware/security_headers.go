package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersMiddleware 安全头中间件，设置各种安全相关的HTTP头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		
		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		
		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 内容安全策略（CSP）- 基本配置，可根据需要调整
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self';")
		
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// HSTS (HTTP Strict Transport Security) - 仅在HTTPS中启用
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		c.Next()
	}
}