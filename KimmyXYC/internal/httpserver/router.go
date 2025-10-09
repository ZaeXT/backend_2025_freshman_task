package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"AIBackend/internal/provider"
	"AIBackend/internal/services"
	"AIBackend/pkg/middleware"
)

type Server struct {
	Auth *services.AuthService
	Chat *services.ChatService
}

func NewRouter(db *gorm.DB, llm provider.LLMProvider) *gin.Engine {
	g := gin.Default()

	server := &Server{
		Auth: services.NewAuthService(db),
		Chat: services.NewChatService(db, llm),
	}

	g.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := g.Group("/api")
	{
		auth := api.Group("/auth")
		auth.POST("/register", server.handleRegister)
		auth.POST("/login", server.handleLogin)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/me", server.handleMe)
		protected.GET("/conversations", server.handleListConversations)
		protected.GET("/conversations/:id/messages", server.handleGetMessages)
		protected.POST("/chat", middleware.ModelAccess(), server.handleChat)
	}

	// Serve static frontend files without conflicting wildcard
	g.StaticFile("/", "./web/index.html")
	g.Static("/css", "./web/css")
	g.Static("/js", "./web/js")

	return g
}

type registerReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *Server) handleRegister(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, token, err := s.Auth.Register(req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (s *Server) handleLogin(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, token, err := s.Auth.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func (s *Server) handleMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id":    c.GetUint("user_id"),
		"user_email": c.GetString("user_email"),
		"user_role":  c.GetString("user_role"),
	})
}

func (s *Server) handleListConversations(c *gin.Context) {
	uid := c.GetUint("user_id")
	convs, err := s.Chat.ListConversations(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": convs})
}

func (s *Server) handleGetMessages(c *gin.Context) {
	uid := c.GetUint("user_id")
	idStr := c.Param("id")
	id64, _ := strconv.ParseUint(idStr, 10, 64)
	msgs, err := s.Chat.GetMessages(uid, uint(id64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": msgs})
}

type chatReq struct {
	ConversationID uint   `json:"conversation_id"`
	Model          string `json:"model"`
	Message        string `json:"message" binding:"required"`
	Stream         *bool  `json:"stream"`
}

func (s *Server) handleChat(c *gin.Context) {
	uid := c.GetUint("user_id")
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Fallback to query param model for middleware check compatibility
	if req.Model == "" {
		req.Model = c.Query("model")
	}
	// Enforce model access if provided in body
	role := c.GetString("user_role")
	if !middleware.CheckModelAccess(role, req.Model) {
		c.JSON(http.StatusForbidden, gin.H{"error": "model access denied for role"})
		return
	}
	streaming := false
	if req.Stream != nil {
		streaming = *req.Stream
	}
	if c.Query("stream") == "1" || c.Query("stream") == "true" {
		streaming = true
	}
	if !streaming {
		convID, reply, err := s.Chat.SendMessage(c.Request.Context(), uid, req.ConversationID, req.Model, req.Message, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"conversation_id": convID, "reply": reply})
		return
	}
	// Streaming via SSE
	w := c.Writer
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	sentAny := false
	convID, _, err := s.Chat.SendMessage(c.Request.Context(), uid, req.ConversationID, req.Model, req.Message, func(chunk string) error {
		sentAny = true
		_, err := w.Write([]byte("data: " + chunk + "\n\n"))
		if err == nil && flusher != nil {
			flusher.Flush()
		}
		return err
	})
	if err != nil {
		// send error as SSE comment and 0-length event end
		_, _ = w.Write([]byte(": error: " + err.Error() + "\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
		return
	}
	if !sentAny {
		// send at least one empty event to keep clients happy
		_, _ = w.Write([]byte("data: \n\n"))
	}
	// end event
	_, _ = w.Write([]byte("event: done\n" + "data: {\"conversation_id\": " + strconv.FormatUint(uint64(convID), 10) + "}\n\n"))
	if flusher != nil {
		flusher.Flush()
	}
	// allow connection to close shortly after
	time.Sleep(50 * time.Millisecond)
}
