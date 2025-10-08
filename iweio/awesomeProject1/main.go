package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var volcAPIURL, volcAPIKey string

// 请求结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateConversationRequest struct {
	UserID  int    `json:"user_id" binding:"required"`
	AiModel string `json:"ai_model" binding:"required"` // 传入 model_id
	Title   string `json:"title"`
}

type SendMessageRequest struct {
	ConversationID int64  `json:"conversation_id" binding:"required"`
	Sender         string `json:"sender" binding:"required"`
	Content        string `json:"content" binding:"required"`
}

type AssignPermissionRequest struct {
	UserID    int `json:"user_id" binding:"required"`
	AIModelID int `json:"ai_model_id" binding:"required"`
}

type AIChatRequest struct {
	ConversationID int64  `json:"conversation_id" binding:"required"`
	UserID         int    `json:"user_id" binding:"required"`
	Content        string `json:"content" binding:"required"`
}

// 火山引擎 API 请求结构体
type VolcChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// 火山引擎 API 消息结构体
type Message struct {
	Role             string `json:"role"`                        // 角色（user/assistant）
	Content          string `json:"content"`                     // 消息内容
	ReasoningContent string `json:"reasoning_content,omitempty"` // 火山扩展：思考过程（可选）
}

// 火山引擎 API 响应结构体（适配实际返回格式）
type VolcChatResponse struct {
	ID      string `json:"id"`      // 请求唯一标识
	Object  string `json:"object"`  // 响应类型（如 "chat.completion"）
	Created int64  `json:"created"` // 时间戳（秒级）
	Model   string `json:"model"`   // 模型标识（如 "doubao-seed-1-6-250615"）
	Choices []struct {
		Index        int     `json:"index"`         // 结果序号（通常为 0）
		Message      Message `json:"message"`       // 消息体
		FinishReason string  `json:"finish_reason"` // 结束原因（如 "stop" 表示正常结束）
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`     // 提问的 Token 数
		CompletionTokens int `json:"completion_tokens"` // AI 回复的 Token 数
		TotalTokens      int `json:"total_tokens"`      // 总 Token 数
		// 火山引擎扩展的 Token 明细（按需保留）
		PromptTokensDetails struct {
			CachedTokens int `json:"cached_tokens"`
		} `json:"prompt_tokens_details,omitempty"`
		CompletionTokensDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details,omitempty"`
	} `json:"usage"`
	Error struct { // 错误信息（接口失败时存在）
		Code    string `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

// 初始化数据库
func initDB() error {
	var err error
	// 注意：dsn 中指定 charset=utf8mb4 以支持 emoji 等特殊字符
	dsn := "root:Budonbej@614@tcp(127.0.0.1:3306)/ai_qa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库 ping 失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功")
	return nil
}

// 用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户名或密码"})
		return
	}
	result, err := db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败: " + err.Error()})
		return
	}
	userID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "注册成功", "data": gin.H{"user_id": userID}})
}

// 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户名或密码"})
		return
	}
	var userID int
	err := db.QueryRow("SELECT id FROM user WHERE username=? AND password=?", req.Username, req.Password).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "登录成功", "data": gin.H{"user_id": userID}})
}

// 检查权限
func CheckAIModelPermission(userID, aiModelID int) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user_ai_permission WHERE user_id = ? AND ai_model_id = ?", userID, aiModelID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 分配权限
func AssignAIModelPermission(c *gin.Context) {
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID或AI模型ID"})
		return
	}
	var userExists, modelExists int
	db.QueryRow("SELECT COUNT(*) FROM user WHERE id = ?", req.UserID).Scan(&userExists)
	db.QueryRow("SELECT COUNT(*) FROM ai_model WHERE id = ?", req.AIModelID).Scan(&modelExists)
	if userExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在"})
		return
	}
	if modelExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI模型不存在"})
		return
	}
	_, err := db.Exec("INSERT IGNORE INTO user_ai_permission (user_id, ai_model_id) VALUES (?, ?)", req.UserID, req.AIModelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配权限失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "权限分配成功"})
}

// 创建对话
func CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少用户ID或AI模型ID"})
		return
	}
	var aiModelID int
	err := db.QueryRow("SELECT id FROM ai_model WHERE model_id = ?", req.AiModel).Scan(&aiModelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "AI模型接入点不存在: " + req.AiModel})
		return
	}
	hasPermission, err := CheckAIModelPermission(req.UserID, aiModelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "权限检查失败: " + err.Error()})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有该AI模型的使用权限: " + req.AiModel})
		return
	}
	result, err := db.Exec("INSERT INTO conversation (user_id, ai_model, title) VALUES (?, ?, ?)", req.UserID, req.AiModel, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建对话失败: " + err.Error()})
		return
	}
	convID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "对话创建成功", "data": gin.H{"conversation_id": convID}})
}

// 发送消息
func SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少对话ID、发送者或消息内容"})
		return
	}
	// 根据 sender 自动设置 role 字段（匹配数据库枚举约束）
	role := "user"
	if req.Sender == "ai" {
		role = "assistant"
	}
	_, err := db.Exec("INSERT INTO message (conversation_id, sender, role, content) VALUES (?, ?, ?, ?)",
		req.ConversationID, req.Sender, role, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存消息失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "消息发送成功"})
}

// 获取消息
func GetMessages(c *gin.Context) {
	userID := c.Param("user_id")
	rows, err := db.Query(`
		SELECT m.sender, m.content, m.create_time 
		FROM message m
		JOIN conversation c ON m.conversation_id = c.id
		WHERE c.user_id = ?
		ORDER BY m.create_time ASC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询消息失败: " + err.Error()})
		return
	}
	defer rows.Close()
	var messages []map[string]interface{}
	for rows.Next() {
		var sender, content, createTime string
		if err := rows.Scan(&sender, &content, &createTime); err != nil {
			log.Println("扫描消息失败:", err)
			continue
		}
		messages = append(messages, map[string]interface{}{
			"sender":      sender,
			"content":     content,
			"create_time": createTime,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": messages})
}

// 调用火山引擎模型
func CallVolcModel(modelID string, messages []Message) (string, error) {
	reqBody := VolcChatRequest{
		Model:       modelID,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2048,
	}
	jsonData, _ := json.MarshalIndent(reqBody, "", "  ")
	log.Printf("调用火山引擎API请求体: %s", jsonData)

	req, err := http.NewRequest("POST", volcAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+volcAPIKey)
	req.Header.Set("X-Volc-Region", "cn-beijing")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Printf("火山引擎响应: %s", respBody)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败，状态码: %d，响应内容: %s", resp.StatusCode, string(respBody))
	}

	var volcResp VolcChatResponse
	if err := json.Unmarshal(respBody, &volcResp); err != nil {
		return "", fmt.Errorf("响应解析失败: %v，原始响应: %s", err, string(respBody))
	}

	// 检查火山引擎返回的错误
	if volcResp.Error.Code != "" {
		return "", fmt.Errorf("火山引擎错误: %s（错误码: %s）", volcResp.Error.Message, volcResp.Error.Code)
	}

	// 检查是否有AI回复
	if len(volcResp.Choices) == 0 {
		return "", fmt.Errorf("未获取到AI回复，原始响应: %s", string(respBody))
	}

	return volcResp.Choices[0].Message.Content, nil
}

// AI 对话
func ChatWithAI(c *gin.Context) {
	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少对话ID、用户ID或消息内容"})
		return
	}
	var modelID string
	err := db.QueryRow("SELECT ai_model FROM conversation WHERE id = ? AND user_id = ?", req.ConversationID, req.UserID).Scan(&modelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "对话不存在或不属于该用户，错误: " + err.Error()})
		return
	}

	rows, err := db.Query("SELECT sender, content FROM message WHERE conversation_id = ? ORDER BY create_time ASC", req.ConversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询历史消息失败: " + err.Error()})
		return
	}
	defer rows.Close()

	var historyMessages []Message
	for rows.Next() {
		var sender, content string
		if err := rows.Scan(&sender, &content); err != nil {
			log.Printf("扫描历史消息失败: %v", err)
			continue
		}
		role := "user"
		if sender == "ai" {
			role = "assistant"
		}
		historyMessages = append(historyMessages, Message{Role: role, Content: content})
	}

	historyMessages = append(historyMessages, Message{Role: "user", Content: req.Content})

	aiReply, err := CallVolcModel(modelID, historyMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用AI失败: " + err.Error()})
		return
	}

	// 插入用户消息到数据库（显式指定所有字段）
	userMsgResult, err := db.Exec(
		"INSERT INTO message (conversation_id, sender, role, content) VALUES (?, ?, ?, ?)",
		req.ConversationID, "user", "user", req.Content,
	)
	if err != nil {
		log.Printf("插入用户消息失败，详细错误：%+v", err) // 打印完整错误堆栈
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户消息时数据库出错：" + err.Error()})
		return
	}
	userMsgID, _ := userMsgResult.LastInsertId()
	log.Printf("用户消息插入成功，ID: %d", userMsgID)

	// 插入AI回复到数据库（显式指定所有字段）
	aiMsgResult, err := db.Exec(
		"INSERT INTO message (conversation_id, sender, role, content) VALUES (?, ?, ?, ?)",
		req.ConversationID, "ai", "assistant", aiReply,
	)
	if err != nil {
		log.Printf("插入AI回复失败，详细错误：%+v", err) // 打印完整错误堆栈
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存AI回复时数据库出错：" + err.Error()})
		return
	}
	aiMsgID, _ := aiMsgResult.LastInsertId()
	log.Printf("AI回复插入成功，ID: %d", aiMsgID)

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "AI回复成功", "data": gin.H{"ai_reply": aiReply}})
}

// 路由
func setupRoutes(r *gin.Engine) {
	r.POST("/register", Register)
	r.POST("/login", Login)
	r.POST("/permissions/assign", AssignAIModelPermission)
	r.POST("/conversations", CreateConversation)
	r.POST("/messages", SendMessage)
	r.GET("/conversations/:user_id/messages", GetMessages)
	r.POST("/chat", ChatWithAI)
}

// 主函数
func main() {
	// 从环境变量读取火山引擎 API 配置（需提前设置）
	volcAPIURL = os.Getenv("AI_API_URL")
	volcAPIKey = os.Getenv("AI_API_KEY")
	if volcAPIURL == "" || volcAPIKey == "" {
		log.Fatal("❌ 请先设置环境变量 AI_API_URL 和 AI_API_KEY")
	}

	// 初始化数据库
	if err := initDB(); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 启动 Gin 服务
	r := gin.Default()
	setupRoutes(r)

	log.Println("✅ 服务启动成功: http://127.0.0.1:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}
