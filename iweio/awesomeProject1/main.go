package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// 用户注册请求体
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 用户登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 创建对话请求体
type CreateConversationRequest struct {
	UserID  int    `json:"user_id" binding:"required"`
	AiModel string `json:"ai_model" binding:"required"`
	Title   string `json:"title"`
}

// 发送消息请求体
type SendMessageRequest struct {
	ConversationID int64  `json:"conversation_id" binding:"required"`
	Sender         string `json:"sender" binding:"required"` // "user" 或 "ai"
	Content        string `json:"content" binding:"required"`
}

// AIModel 用于表示AI模型信息
type AIModel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 给用户分配AI模型权限的请求体
type AssignPermissionRequest struct {
	UserID    int `json:"user_id" binding:"required"`
	AIModelID int `json:"ai_model_id" binding:"required"`
}

func main() {
	// 1. 连接 MySQL（适配你的数据库名 ai_qa 和密码 Budonbej@614）
	var err error
	db, err = sql.Open("mysql", "root:Budonbej@614@tcp(127.0.0.1:3306)/ai_qa?charset=utf8mb4")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库 ping 失败:", err)
	}
	fmt.Println("✅ 数据库连接成功")

	// 2. 初始化 Gin 路由
	r := gin.Default()

	// 3. API 路由
	r.POST("/register", Register)
	r.POST("/login", Login)
	r.POST("/conversations", CreateConversation)
	r.POST("/messages", SendMessage)
	r.GET("/conversations/:user_id/messages", GetMessages)
	r.POST("/permissions/assign", AssignAIModelPermission)
	// 4. 启动服务
	r.Run(":8080")
}

// 用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	result, err := db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	userID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	var userID int
	err := db.QueryRow("SELECT id FROM user WHERE username=? AND password=?", req.Username, req.Password).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

// 检查用户是否有使用指定AI模型的权限
func CheckAIModelPermission(userID, aiModelID int) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM user_ai_permission WHERE user_id = ? AND ai_model_id = ?",
		userID, aiModelID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 给用户分配AI模型权限
func AssignAIModelPermission(c *gin.Context) {
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 检查用户和模型是否存在
	var userExists, modelExists int
	db.QueryRow("SELECT COUNT(*) FROM user WHERE id = ?", req.UserID).Scan(&userExists)
	db.QueryRow("SELECT COUNT(*) FROM ai_model WHERE id = ?", req.AIModelID).Scan(&modelExists)
	if userExists == 0 || modelExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户或AI模型不存在"})
		return
	}

	// 插入权限记录（重复分配不会报错）
	_, err := db.Exec("INSERT IGNORE INTO user_ai_permission (user_id, ai_model_id) VALUES (?, ?)", req.UserID, req.AIModelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配权限失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 步骤1：根据模型名称查ID
	var aiModelID int
	err := db.QueryRow("SELECT id FROM ai_model WHERE model_name = ?", req.AiModel).Scan(&aiModelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI模型不存在"})
		return
	}

	// 步骤2：检查权限
	hasPermission, err := CheckAIModelPermission(req.UserID, aiModelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败"})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有使用该AI模型的权限"})
		return
	}

	// 步骤3：创建对话
	result, err := db.Exec("INSERT INTO conversation (user_id, ai_model, title) VALUES (?, ?, ?)", req.UserID, req.AiModel, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建对话失败"})
		return
	}

	convID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"conversation_id": convID})
}

// 发送消息
func SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	_, err := db.Exec("INSERT INTO message (conversation_id, sender, content) VALUES (?, ?, ?)", req.ConversationID, req.Sender, req.Content)
	fmt.Println(err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// 获取某个用户的所有消息
func GetMessages(c *gin.Context) {
	userID := c.Param("user_id")

	rows, err := db.Query(`
		SELECT m.sender, m.content, m.create_time 
		FROM message m
		JOIN conversation c ON m.conversation_id = c.id
		WHERE c.user_id = ?
		ORDER BY m.create_time ASC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var sender, content, createTime string
		err := rows.Scan(&sender, &content, &createTime)
		if err != nil {
			log.Println("扫描消息失败:", err)
			continue
		}
		messages = append(messages, map[string]interface{}{
			"sender":      sender,
			"content":     content,
			"create_time": createTime,
		})
	}

	c.JSON(http.StatusOK, messages)
}
