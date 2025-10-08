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

	// 添加CORS中间件，允许所有跨域请求
	r.Use(func(c *gin.Context) {
		// 允许所有来源的跨域请求
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With, X-Timestamp, X-Nonce, Accept, Accept-Language, Content-Language")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 预检请求缓存24小时

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 添加安全头中间件
	r.Use(middleware.SecurityHeadersMiddleware())

	// 添加请求频率限制中间件
	r.Use(middleware.RateLimitMiddleware())

	// 1. 公开 API 路由：注册和登录
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/register", controllers.RegisterHandle)
		authRoutes.POST("/login", controllers.LoginHandle)
	}

	// API路由组（需要认证）
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) // 应用认证中间件
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

	// 添加测试路由，显示杨艺哲三个大字（不受签名验证中间件影响）
	r.GET("/test", func(c *gin.Context) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>测试页面</title>
    <style>
        body {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f0f0f0;
        }
        .name {
            font-size: 100px;
            font-weight: bold;
            color: #333;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
            font-family: 'Microsoft YaHei', Arial, sans-serif;
        }
    </style>
</head>
<body>
    <div class="name">杨艺哲</div>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// 将 Gin 路由集成到标准的 http.ServeMux 中
	mux := http.NewServeMux()
	mux.HandleFunc("/", ginHandlerWrapper(r)) // 将所有非 WebSocket 请求交给 Gin 处理

	// 3. WebSocket 路由
	mux.HandleFunc("/ws/ai", WebSocket.HandleWebSocket)
	mux.HandleFunc("/ws/chatroom", WebSocket.HandleChatroomWebSocket) // 公共聊天室WebSocket路由

	// 4. 静态文件服务（保持不变）
	// 注意：由于 Gin 接管了 "/", 静态文件可能需要通过 Gin 的 Static 方法配置，
	// 或者调整 mux.HandleFunc("/", ...) 的逻辑。
	// 为了简化，我们让 Gin 负责所有非 WebSocket 的路由。

	// 启动服务器
	httpPort := 8080
	httpsPort := 8443

	// 证书文件路径
	certFile := "20735405_www.yyzyyz.click_nginx/www.yyzyyz.click.pem"
	keyFile := "20735405_www.yyzyyz.click_nginx/www.yyzyyz.click.key"

	// 同时启动HTTP和HTTPS服务
	go func() {
		// HTTP模式
		log.Printf("HTTP Server running on http://0.0.0.0:%d\n", httpPort)
		log.Printf("WebSocket endpoint: ws://0.0.0.0:%d/ws/ai\n", httpPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", httpPort), mux))
	}()

	// HTTPS模式
	log.Printf("HTTPS Server running on https://0.0.0.0:%d\n", httpsPort)
	log.Printf("Secure WebSocket endpoint: wss://0.0.0.0:%d/ws/ai\n", httpsPort)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", httpsPort), certFile, keyFile, mux))
}
