package main

import (
	"fmt"
	"log"
	"net/http"

	"yyz.com/MyGo/WebSocket" // 你的 WebSocket 包
	"yyz.com/MyGo/controllers"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/middleware"

	"github.com/gin-gonic/gin"
)

// ginHandlerWrapper 包装 Gin 处理器以适配 net/http.Handler
// 这样可以让你在标准的 net/http.ListenAndServe 中使用 Gin 的路由。
// 注意：你也可以直接使用 r.Run(fmt.Sprintf(":%d", port)) 来启动 Gin
func ginHandlerWrapper(r *gin.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.ServeHTTP(w, req)
	}
}

func main() {
	// 初始化数据库连接
	database.ConnectDB()

	// 使用 Gin 替代标准库的 ServeMux 来处理 API 路由
	r := gin.Default()

	// 添加请求频率限制中间件
	r.Use(middleware.RateLimitMiddleware())

	r.Use(middleware.SignatureMiddleware()) // 应用签名验证中间件

	// 1. 公开 API 路由：注册和登录
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/register", controllers.RegisterHandle)
		authRoutes.POST("/login", controllers.LoginHandle)
	}

	// API路由组（需要认证）
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) // 应用认证中间件
	// api.Use(middleware.SignatureMiddleware()) // 应用签名验证中间件
	{
		api.GET("/user", func(c *gin.Context) {
			user, _ := c.Get("user")
			c.JSON(http.StatusOK, user)
		})
		api.DELETE("/user", controllers.DeleteUserHandle)      // 删除用户数据接口
		api.POST("/recharge", controllers.RechargeHandle)      // 充值接口
		api.GET("/token-info", controllers.GetTokenInfoHandle) // 获取token信息

		// 模型管理API
		api.GET("/models", controllers.GetAvailableModels) // 获取可用模型列表

		// 对话记录管理API
		api.GET("/conversations", controllers.GetConversations)                  // 获取对话记录列表
		api.GET("/conversations/:id", controllers.GetConversation)               // 获取对话详情
		api.DELETE("/conversations/:id", controllers.DeleteConversation)         // 删除对话记录
		api.PUT("/conversations/:id/title", controllers.UpdateConversationTitle) // 更新对话标题
	}

	// 将 Gin 路由集成到标准的 http.ServeMux 中
	mux := http.NewServeMux()
	mux.HandleFunc("/", ginHandlerWrapper(r)) // 将所有非 WebSocket 请求交给 Gin 处理

	// 3. WebSocket 路由（保持不变）
	mux.HandleFunc("/ws/ai", WebSocket.HandleWebSocket)

	// 4. 静态文件服务（保持不变）
	// 注意：由于 Gin 接管了 "/", 静态文件可能需要通过 Gin 的 Static 方法配置，
	// 或者调整 mux.HandleFunc("/", ...) 的逻辑。
	// 为了简化，我们让 Gin 负责所有非 WebSocket 的路由。

	// 启动服务器
	port := 8080
	log.Printf("Server running on http://localhost:%d\n", port)
	log.Printf("WebSocket endpoint: ws://localhost:%d/ws/ai\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
