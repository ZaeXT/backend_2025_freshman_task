// web.go
package main

import (
	"fmt"
	"net/http"
	"time"

	"ciallo/config"
	"ciallo/models"

	"github.com/gin-gonic/gin"
)

type WebServer struct {
	aiClient     *AIClient
	router       *gin.Engine
	sessionStore map[string]*Session
}

type Session struct {
	UserID    string
	Username  string
	ExpiresAt time.Time
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Confirm  string `json:"confirm"`
}

type MessageRequest struct {
	Content string `json:"content"`
	Model   string `json:"model,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewWebServer() *WebServer {
	cfg := config.NewConfig()
	fmt.Printf("🌐 Web服务器配置 - Provider: %s, API Key: %s...\n",
		cfg.Provider, safeSubstring(cfg.APIKey, 10))

	aiClient := NewAIClient(cfg)

	router := gin.Default()

	// 设置模板
	router.LoadHTMLGlob("templates/*")

	// 静态文件
	router.Static("/static", "./static")

	server := &WebServer{
		aiClient:     aiClient,
		router:       router,
		sessionStore: make(map[string]*Session),
	}

	server.setupRoutes()
	return server
}

// 辅助函数：安全截取字符串
func safeSubstring(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length]
}

func (ws *WebServer) setupRoutes() {
	// 页面路由
	ws.router.GET("/", ws.homePage)
	ws.router.GET("/login", ws.loginPage)
	ws.router.GET("/register", ws.registerPage)
	ws.router.GET("/chat", ws.authMiddleware(), ws.chatPage)
	ws.router.GET("/profile", ws.authMiddleware(), ws.profilePage)

	// API路由
	api := ws.router.Group("/api")
	{
		api.POST("/login", ws.apiLogin)
		api.POST("/register", ws.apiRegister)
		api.POST("/logout", ws.apiAuthMiddleware(), ws.apiLogout)
		api.GET("/user", ws.apiAuthMiddleware(), ws.apiGetUser)
		api.POST("/message", ws.apiAuthMiddleware(), ws.apiSendMessage)
		api.GET("/conversations", ws.apiAuthMiddleware(), ws.apiGetConversations)
		api.POST("/conversation", ws.apiAuthMiddleware(), ws.apiCreateConversation)
		api.PUT("/conversation/:id", ws.apiAuthMiddleware(), ws.apiSwitchConversation)
		api.PUT("/model", ws.apiAuthMiddleware(), ws.apiSwitchModel)
		api.PUT("/profile", ws.apiAuthMiddleware(), ws.apiUpdateProfile)
		api.POST("/upgrade", ws.apiAuthMiddleware(), ws.apiUpgradeUser)
	}
}

// 认证中间件
func (ws *WebServer) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		session, exists := ws.sessionStore[sessionID]
		if !exists || time.Now().After(session.ExpiresAt) {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 更新会话过期时间
		session.ExpiresAt = time.Now().Add(24 * time.Hour)

		// 设置用户信息到上下文
		c.Set("userID", session.UserID)
		c.Set("username", session.Username)
		c.Next()
	}
}

// API认证中间件
func (ws *WebServer) apiAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "未登录",
			})
			c.Abort()
			return
		}

		session, exists := ws.sessionStore[sessionID]
		if !exists || time.Now().After(session.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "会话已过期",
			})
			c.Abort()
			return
		}

		session.ExpiresAt = time.Now().Add(24 * time.Hour)
		c.Set("userID", session.UserID)
		c.Set("username", session.Username)
		c.Next()
	}
}

// 创建会话
func (ws *WebServer) createSession(user *models.User) string {
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	ws.sessionStore[sessionID] = &Session{
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return sessionID
}

// 页面路由处理
func (ws *WebServer) homePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "AI问答系统",
	})
}

func (ws *WebServer) loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (ws *WebServer) registerPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func (ws *WebServer) chatPage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	currentConv := user.GetCurrentConversation()

	c.HTML(http.StatusOK, "chat.html", gin.H{
		"username": user.Username,
		"level":    user.Level,
		"model":    user.CurrentModel,
		"conversation": gin.H{
			"id":    currentConv.ID,
			"title": currentConv.Title,
		},
	})
}

func (ws *WebServer) profilePage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	level, levelInfo := user.GetLevelInfo()

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"user":        user,
		"level":       level,
		"levelInfo":   levelInfo,
		"modelConfig": models.AIModelConfig,
	})
}

// API路由处理
func (ws *WebServer) apiLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	user, err := ws.aiClient.userManager.VerifyPassword(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "用户名或密码错误",
		})
		return
	}

	sessionID := ws.createSession(user)
	user.UpdateLoginTime()
	ws.aiClient.userManager.SaveUsers()

	c.SetCookie("session_id", sessionID, 3600*24, "/", "", false, true)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "登录成功",
		Data: gin.H{
			"username": user.Username,
			"level":    user.Level,
		},
	})
}

func (ws *WebServer) apiRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	if req.Password != req.Confirm {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "两次密码不一致",
		})
		return
	}

	if ws.aiClient.userManager.FindUserByUsername(req.Username) != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "用户名已存在",
		})
		return
	}

	var user *models.User
	var err error

	if req.Username == "xunzhu" {
		user, err = ws.aiClient.userManager.CreateUser(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "创建用户失败: " + err.Error(),
			})
			return
		}
		ws.aiClient.userManager.UpdateUserLevel(user.ID, models.UserLevelAdmin)
	} else {
		user, err = ws.aiClient.userManager.CreateUser(req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "创建用户失败: " + err.Error(),
			})
			return
		}
	}

	sessionID := ws.createSession(user)
	c.SetCookie("session_id", sessionID, 3600*24, "/", "", false, true)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "注册成功",
		Data: gin.H{
			"username": user.Username,
			"level":    user.Level,
		},
	})
}

func (ws *WebServer) apiLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("session_id")
	delete(ws.sessionStore, sessionID)
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "注销成功",
	})
}

func (ws *WebServer) apiGetUser(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	level, levelInfo := user.GetLevelInfo()
	currentConv := user.GetCurrentConversation()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"username":     user.Username,
			"nickname":     user.Nickname,
			"level":        level,
			"levelInfo":    levelInfo,
			"currentModel": user.CurrentModel,
			"conversation": gin.H{
				"id":       currentConv.ID,
				"title":    currentConv.Title,
				"messages": currentConv.Messages,
			},
			"conversations": user.Conversations,
			"isSpecial":     user.IsSpecialUser(),
			"greeting":      user.GetGreeting(),
		},
	})
}

func (ws *WebServer) apiSendMessage(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	// 添加用户消息
	currentConv := user.GetCurrentConversation()
	err := user.AddMessageToCurrentConversation("user", req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 获取AI响应
	response, err := ws.aiClient.SendMessage(currentConv.Messages, user.CurrentModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "AI响应失败: " + err.Error(),
		})
		return
	}

	// 添加AI响应
	err = user.AddMessageToCurrentConversation("assistant", response)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ws.aiClient.userManager.SaveUsers()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"response": response,
			"conversation": gin.H{
				"id":       currentConv.ID,
				"title":    currentConv.Title,
				"messages": currentConv.Messages,
			},
		},
	})
}

func (ws *WebServer) apiGetConversations(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"conversations": user.Conversations,
		},
	})
}

func (ws *WebServer) apiCreateConversation(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		title = "新对话"
	}

	conv, err := user.CreateNewConversation(title)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ws.aiClient.userManager.SaveUsers()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "对话创建成功",
		Data: gin.H{
			"conversation": conv,
		},
	})
}

func (ws *WebServer) apiSwitchConversation(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	convID := c.Param("id")

	// 查找对话并切换到当前
	for i, conv := range user.Conversations {
		if conv.ID == convID {
			// 将选中的对话移到列表末尾
			user.Conversations = append(
				append(user.Conversations[:i], user.Conversations[i+1:]...),
				conv,
			)
			ws.aiClient.userManager.SaveUsers()

			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Message: "切换对话成功",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: "对话不存在",
	})
}

func (ws *WebServer) apiSwitchModel(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	model := c.PostForm("model")
	if model == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "模型不能为空",
		})
		return
	}

	err := ws.aiClient.userManager.UpdateUserModel(user.ID, model)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ws.aiClient.userManager.SaveUsers()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "模型切换成功",
		Data: gin.H{
			"model": model,
		},
	})
}

func (ws *WebServer) apiUpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	nickname := c.PostForm("nickname")
	gender := c.PostForm("gender")

	if nickname != "" {
		err := ws.aiClient.userManager.UpdateUserNickname(user.ID, nickname)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}
	}

	if gender != "" {
		err := ws.aiClient.userManager.UpdateUserGender(user.ID, gender)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}
	}

	ws.aiClient.userManager.SaveUsers()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "个人信息更新成功",
	})
}

func (ws *WebServer) apiUpgradeUser(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	user := ws.aiClient.userManager.FindUserByID(userID)
	if user == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	targetLevel := c.PostForm("level")
	password := c.PostForm("password")

	if targetLevel == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "目标等级不能为空",
		})
		return
	}

	if password != models.UpgradePassword {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "升级密码错误",
		})
		return
	}

	err := ws.aiClient.userManager.UpdateUserLevel(user.ID, targetLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	ws.aiClient.userManager.SaveUsers()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "升级成功",
		Data: gin.H{
			"level": targetLevel,
		},
	})
}

func (ws *WebServer) Run(addr string) error {
	return ws.router.Run(addr)
}

func main() {
	server := NewWebServer()
	fmt.Println("服务器启动在 http://localhost:8080")
	server.Run(":8080")
}
