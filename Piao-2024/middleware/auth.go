package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"Piao/config"
	"Piao/models"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 身份认证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("❌ 缺少Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &models.Claims{}

		// 验证token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return config.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			log.Printf("❌ Token验证失败: %v\n", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 将用户信息添加到请求头
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		r.Header.Set("X-User-Level", fmt.Sprintf("%d", claims.Level))
		r.Header.Set("X-Username", claims.Username)

		// 调用下一个处理函数
		next(w, r)
	}
}
