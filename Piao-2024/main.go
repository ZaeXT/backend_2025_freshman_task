package main

import (
	"bufio"         // 用于按行读取数据（流式响应）
	"bytes"         // 用于处理字节缓冲区
	"database/sql"  // 数据库操作的标准接口
	"encoding/json" // JSON数据的编码和解码
	"fmt"           // 格式化输入输出
	"io"            // 基本的输入输出接口
	"log"           // 日志记录
	"net/http"      // HTTP客户端和服务器
	"os"            // 操作系统功能（环境变量等）
	"strings"       // 字符串处理
	"time"          // 时间相关功能

	_ "github.com/go-sql-driver/mysql" // MySQL驱动（下划线表示只导入初始化，不直接使用）
	"github.com/golang-jwt/jwt/v5"     // JWT token生成和验证
	"github.com/joho/godotenv"         // 加载.env环境变量文件
	"golang.org/x/crypto/bcrypt"       // 密码加密
)

// ============================================
// 数据结构定义
// ============================================

// User 用户结构体，对应数据库中的users表
type User struct {
	ID       int    `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
	Password string `json:"-"`        // 密码（json:"-"表示JSON序列化时忽略该字段，保护隐私）
	Level    int    `json:"level"`    // 用户等级（1=普通用户，2=高级用户）
}

// LoginRequest 登录请求的数据结构
type LoginRequest struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码（注意这里没有"-"，需要接收密码）
}

// RegisterRequest 注册请求的数据结构
type RegisterRequest struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
}

// Conversation 对话结构体，一个用户可以有多个对话
type Conversation struct {
	ID        int       `json:"id"`         // 对话ID
	UserID    int       `json:"user_id"`    // 所属用户ID
	Title     string    `json:"title"`      // 对话标题
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// Message 消息结构体，每条消息属于某个对话
type Message struct {
	ID             int       `json:"id"`              // 消息ID
	ConversationID int       `json:"conversation_id"` // 所属对话ID
	Role           string    `json:"role"`            // 角色（user=用户，assistant=AI助手）
	Content        string    `json:"content"`         // 消息内容
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
}

// Claims JWT token中存储的用户信息
type Claims struct {
	UserID               int    `json:"user_id"`  // 用户ID
	Username             string `json:"username"` // 用户名
	Level                int    `json:"level"`    // 用户等级
	jwt.RegisteredClaims        // 嵌入标准JWT字段（过期时间等）
}

// ============================================
// 火山引擎API相关结构
// ============================================

// VolcengineRequest 发送给火山引擎API的请求结构
type VolcengineRequest struct {
	Model    string                   `json:"model"`            // 使用的AI模型名称
	Messages []map[string]interface{} `json:"messages"`         // 对话历史消息列表
	Stream   bool                     `json:"stream,omitempty"` // 是否使用流式输出（逐字返回）
}

// VolcengineResponse 火山引擎API的标准响应结构（非流式）
type VolcengineResponse struct {
	ID      string     `json:"id"`      // 响应ID
	Object  string     `json:"object"`  // 对象类型
	Created int64      `json:"created"` // 创建时间戳
	Model   string     `json:"model"`   // 使用的模型
	Choices []struct { // AI生成的回复选项（通常只有一个）
		Index   int      `json:"index"` // 选项索引
		Message struct { // 消息内容
			Role    string `json:"role"`    // 角色
			Content string `json:"content"` // 内容
		} `json:"message"`
		FinishReason string `json:"finish_reason"` // 结束原因
	} `json:"choices"`
	Usage struct { // token使用统计
		PromptTokens     int `json:"prompt_tokens"`     // 输入token数
		CompletionTokens int `json:"completion_tokens"` // 输出token数
		TotalTokens      int `json:"total_tokens"`      // 总token数
	} `json:"usage"`
	Error *struct { // 错误信息（如果有）
		Message string `json:"message"` // 错误消息
		Type    string `json:"type"`    // 错误类型
		Code    string `json:"code"`    // 错误代码
	} `json:"error,omitempty"`
}

// VolcengineStreamResponse 流式响应结构（逐字返回时使用）
type VolcengineStreamResponse struct {
	ID      string     `json:"id"`      // 响应ID
	Object  string     `json:"object"`  // 对象类型
	Created int64      `json:"created"` // 创建时间戳
	Model   string     `json:"model"`   // 使用的模型
	Choices []struct { // 回复选项
		Index int      `json:"index"` // 选项索引
		Delta struct { // 增量内容（每次只返回新增的部分）
			Role    string `json:"role,omitempty"`    // 角色（可选）
			Content string `json:"content,omitempty"` // 内容增量
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"` // 结束原因（最后一条时出现）
	} `json:"choices"`
}

// ============================================
// 全局变量
// ============================================

var db *sql.DB       // 数据库连接对象（全局共享）
var jwtSecret []byte // JWT密钥（用于签名和验证token）

// ============================================
// 主函数 - 程序入口
// ============================================

func main() {
	// 1. 加载环境变量文件（.env）
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found")
	}

	// 2. 读取JWT密钥
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Fatal("❌ JWT_SECRET not set in .env file") // 如果没有密钥，程序无法运行
	}

	// 3. 连接MySQL数据库
	dbUser := os.Getenv("DB_USER")     // 数据库用户名
	dbPass := os.Getenv("DB_PASSWORD") // 数据库密码
	if dbUser == "" || dbPass == "" {
		log.Fatal("❌ DB_USER or DB_PASSWORD not set in .env file")
	}

	// 构建数据库连接字符串（DSN）
	// 格式：用户名:密码@tcp(主机:端口)/数据库名?参数
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass)
	var err error
	db, err = sql.Open("mysql", dsn) // 打开数据库连接
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}
	defer db.Close() // 程序结束时关闭数据库连接

	// 4. 测试数据库连接是否正常
	if err := db.Ping(); err != nil {
		log.Fatal("❌ 数据库ping失败:", err)
	}
	log.Println("✅ 数据库连接成功")

	// 5. 初始化数据库表结构
	initDatabase()

	// 6. 注册HTTP路由（URL路径与处理函数的映射）
	http.HandleFunc("/api/register", registerHandler)                                      // 用户注册
	http.HandleFunc("/api/login", loginHandler)                                            // 用户登录
	http.HandleFunc("/api/conversations", authMiddleware(conversationsHandler))            // 获取对话列表（需要认证）
	http.HandleFunc("/api/conversation/create", authMiddleware(createConversationHandler)) // 创建新对话（需要认证）
	http.HandleFunc("/api/messages", authMiddleware(messagesHandler))                      // 获取消息列表（需要认证）
	http.HandleFunc("/api/chat", authMiddleware(chatHandler))                              // 普通聊天（需要认证）
	http.HandleFunc("/api/chat/stream", authMiddleware(chatStreamHandler))                 // 流式聊天（需要认证）
	http.HandleFunc("/api/upgrade", authMiddleware(upgradeHandler))                        // 用户升级（需要认证）
	http.HandleFunc("/", serveHTML)                                                        // 根路径返回HTML页面

	// 7. 启动HTTP服务器
	fmt.Println("🚀 服务器启动在 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil)) // 监听8080端口，阻塞运行
}

// ============================================
// 数据库初始化函数
// ============================================

// initDatabase 创建数据库表（如果不存在）
func initDatabase() {
	// 定义三张表的SQL语句
	queries := []string{
		// users表：存储用户信息
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,      -- 主键，自增
			username VARCHAR(50) UNIQUE NOT NULL,   -- 用户名，唯一且不能为空
			password VARCHAR(255) NOT NULL,         -- 加密后的密码
			level INT DEFAULT 1,                    -- 用户等级，默认为1（普通用户）
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- 创建时间
		)`,
		// conversations表：存储对话
		`CREATE TABLE IF NOT EXISTS conversations (
			id INT AUTO_INCREMENT PRIMARY KEY,      -- 主键
			user_id INT NOT NULL,                   -- 所属用户ID
			title VARCHAR(255) NOT NULL,            -- 对话标题
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
			FOREIGN KEY (user_id) REFERENCES users(id)  -- 外键，关联到users表
		)`,
		// messages表：存储消息
		`CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,      -- 主键
			conversation_id INT NOT NULL,           -- 所属对话ID
			role VARCHAR(20) NOT NULL,              -- 角色（user或assistant）
			content TEXT NOT NULL,                  -- 消息内容
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 创建时间
			FOREIGN KEY (conversation_id) REFERENCES conversations(id)  -- 外键
		)`,
	}

	// 执行每个SQL语句
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Fatal("❌ 数据库初始化失败:", err)
		}
	}
	log.Println("✅ 数据库表初始化成功")
}

// ============================================
// 用户注册处理函数
// ============================================

// registerHandler 处理用户注册请求
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法，只接受POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 解析请求体中的JSON数据
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ 注册请求解析失败: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 3. 记录日志（用于调试）
	log.Printf("📝 注册请求: username=%s\n", req.Username)
	log.Printf("📝 密码长度: %d\n", len(req.Password))

	// 4. 使用bcrypt加密密码（单向加密，无法解密）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ 密码加密失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("📝 加密后的hash长度: %d\n", len(hashedPassword))

	// 5. 将用户信息插入数据库
	_, err = db.Exec("INSERT INTO users (username, password, level) VALUES (?, ?, ?)",
		req.Username, string(hashedPassword), 1) // 新用户默认等级为1
	if err != nil {
		log.Printf("❌ 用户注册失败: %v\n", err)
		http.Error(w, "Username already exists", http.StatusConflict) // 用户名已存在
		return
	}

	// 6. 返回成功响应
	log.Printf("✅ 用户注册成功: %s\n", req.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "注册成功"})
}

// ============================================
// 用户登录处理函数
// ============================================

// loginHandler 处理用户登录请求
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 解析登录请求
	var credentials LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		log.Printf("❌ 登录请求解析失败: %v\n", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 3. 记录登录尝试（用于调试和安全审计）
	log.Printf("🔐 登录请求: username=%s\n", credentials.Username)
	log.Printf("🔐 输入的密码长度: %d\n", len(credentials.Password))
	log.Printf("🔐 输入的密码: [%s]\n", credentials.Password)

	// 4. 从数据库查询用户信息
	var user User
	err := db.QueryRow("SELECT id, username, password, level FROM users WHERE username = ?",
		credentials.Username).Scan(&user.ID, &user.Username, &user.Password, &user.Level)
	if err != nil {
		log.Printf("❌ 用户不存在: %s, error: %v\n", credentials.Username, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 5. 验证密码（使用bcrypt比对）
	log.Printf("🔑 开始验证密码: username=%s\n", credentials.Username)
	log.Printf("🔑 数据库中的hash长度: %d\n", len(user.Password))

	compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if compareErr != nil {
		log.Printf("❌ 密码错误: username=%s, error=%v\n", credentials.Username, compareErr)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ 密码验证通过: %s\n", credentials.Username)

	// 6. 生成JWT Token（用于后续请求的身份认证）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   user.ID,
		Username: user.Username,
		Level:    user.Level,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // token有效期24小时
		},
	})

	// 7. 签名token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("❌ Token生成失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// 8. 返回token和用户信息
	log.Printf("✅ 登录成功: %s (Level: %d)\n", user.Username, user.Level)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    tokenString, // 返回给前端，后续请求需携带
		"username": user.Username,
		"level":    user.Level,
	})
}

// ============================================
// 认证中间件
// ============================================

// authMiddleware 身份认证中间件（包装需要登录才能访问的接口）
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 从请求头获取Authorization字段
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("❌ 缺少Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. 提取token（格式："Bearer <token>"）
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		// 3. 验证token的签名和有效期
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil // 返回密钥用于验证签名
		})

		if err != nil || !token.Valid {
			log.Printf("❌ Token验证失败: %v\n", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 4. 将用户信息添加到请求头，传递给下一个处理函数
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", claims.UserID))
		r.Header.Set("X-User-Level", fmt.Sprintf("%d", claims.Level))

		// 5. 调用实际的业务处理函数
		next(w, r)
	}
}

// ============================================
// 对话列表处理函数
// ============================================

// conversationsHandler 获取用户的所有对话列表
func conversationsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 从请求头获取用户ID（由authMiddleware添加）
	userID := r.Header.Get("X-User-ID")

	// 2. 查询该用户的所有对话，按创建时间降序排列
	rows, err := db.Query("SELECT id, title, created_at FROM conversations WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		log.Printf("❌ 查询对话列表失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close() // 确保关闭结果集

	// 3. 遍历结果集，构建对话列表
	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		rows.Scan(&conv.ID, &conv.Title, &conv.CreatedAt)
		conversations = append(conversations, conv)
	}

	// 4. 返回JSON格式的对话列表
	log.Printf("✅ 查询对话列表成功: userID=%s, count=%d\n", userID, len(conversations))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// ============================================
// 创建对话处理函数
// ============================================

// createConversationHandler 创建新的对话
func createConversationHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 获取用户ID
	userID := r.Header.Get("X-User-ID")

	// 3. 解析请求体，获取对话标题
	var data struct {
		Title string `json:"title"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	// 4. 插入新对话到数据库
	result, err := db.Exec("INSERT INTO conversations (user_id, title) VALUES (?, ?)", userID, data.Title)
	if err != nil {
		log.Printf("❌ 创建对话失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// 5. 获取新创建的对话ID
	id, _ := result.LastInsertId()
	log.Printf("✅ 创建对话成功: id=%d, title=%s\n", id, data.Title)

	// 6. 返回新对话的ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// ============================================
// 消息列表处理函数
// ============================================

// messagesHandler 获取指定对话的所有消息
func messagesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 从URL查询参数获取对话ID
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		http.Error(w, "conversation_id required", http.StatusBadRequest)
		return
	}

	// 2. 查询该对话的所有消息，按时间升序排列
	rows, err := db.Query("SELECT id, role, content, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC", conversationID)
	if err != nil {
		log.Printf("❌ 查询消息失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 3. 构建消息列表
	var messages []Message
	for rows.Next() {
		var msg Message
		rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.CreatedAt)
		messages = append(messages, msg)
	}

	// 4. 返回消息列表
	log.Printf("✅ 查询消息成功: conversationID=%s, count=%d\n", conversationID, len(messages))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// ============================================
// 普通聊天处理函数
// ============================================

// chatHandler 处理普通聊天请求（一次性返回完整回复）
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 获取用户等级
	userLevel := r.Header.Get("X-User-Level")

	// 3. 解析请求体
	var data struct {
		ConversationID int    `json:"conversation_id"` // 对话ID
		Message        string `json:"message"`         // 用户消息
		Model          string `json:"model"`           // 要使用的AI模型
	}
	json.NewDecoder(r.Body).Decode(&data)

	log.Printf("💬 收到聊天请求: conversationID=%d, model=%s, userLevel=%s\n", data.ConversationID, data.Model, userLevel)

	// 4. 权限检查：普通用户（level=1）不能使用高级模型
	if strings.Contains(data.Model, "ADVANCED") && userLevel == "1" {
		log.Printf("⛔ 权限不足: 用户level=%s 尝试使用高级模型\n", userLevel)
		http.Error(w, "权限不足，高级模型需要高级用户", http.StatusForbidden)
		return
	}

	// 5. 保存用户消息到数据库
	db.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "user", data.Message)

	// 6. 获取该对话的所有历史消息（用于提供上下文）
	rows, _ := db.Query("SELECT role, content FROM messages WHERE conversation_id = ? ORDER BY created_at ASC", data.ConversationID)
	var messages []map[string]interface{}
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		messages = append(messages, map[string]interface{}{"role": role, "content": content})
	}
	rows.Close()

	log.Printf("📚 加载历史消息: count=%d\n", len(messages))

	// 7. 调用火山引擎API获取AI回复
	response, err := callVolcengineAPI(data.Model, messages)
	if err != nil {
		log.Printf("❌ AI调用失败: %v\n", err)
		http.Error(w, "AI调用失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 8. 保存AI回复到数据库
	db.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "assistant", response)

	// 9. 返回AI回复
	log.Printf("✅ AI回复成功: length=%d\n", len(response))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

// ============================================
// 流式聊天处理函数
// ============================================

// chatStreamHandler 处理流式聊天请求（逐字返回，像ChatGPT那样）
func chatStreamHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 获取用户等级
	userLevel := r.Header.Get("X-User-Level")

	// 3. 解析请求体
	var data struct {
		ConversationID int    `json:"conversation_id"`
		Message        string `json:"message"`
		Model          string `json:"model"`
	}
	json.NewDecoder(r.Body).Decode(&data)

	log.Printf("💬 收到流式聊天请求: conversationID=%d, model=%s, userLevel=%s\n", data.ConversationID, data.Model, userLevel)

	// 4. 权限检查
	if strings.Contains(data.Model, "ADVANCED") && userLevel == "1" {
		log.Printf("⛔ 权限不足: 用户level=%s 尝试使用高级模型\n", userLevel)
		http.Error(w, "权限不足，高级模型需要高级用户", http.StatusForbidden)
		return
	}

	// 5. 保存用户消息
	db.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "user", data.Message)

	// 6. 获取历史消息
	rows, _ := db.Query("SELECT role, content FROM messages WHERE conversation_id = ? ORDER BY created_at ASC", data.ConversationID)
	var messages []map[string]interface{}
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		messages = append(messages, map[string]interface{}{"role": role, "content": content})
	}
	rows.Close()

	log.Printf("📚 加载历史消息: count=%d\n", len(messages))

	// 7. 设置SSE（Server-Sent Events）响应头，用于流式传输
	w.Header().Set("Content-Type", "text/event-stream") // SSE标准内容类型
	w.Header().Set("Cache-Control", "no-cache")         // 禁止缓存
	w.Header().Set("Connection", "keep-alive")          // 保持连接
	w.Header().Set("Access-Control-Allow-Origin", "*")  // 允许跨域

	// 8. 调用流式API，逐字返回
	fullResponse, err := callVolcengineStreamAPI(data.Model, messages, w)
	if err != nil {
		log.Printf("❌ AI流式调用失败: %v\n", err)
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		return
	}

	// 9. 保存完整的AI回复到数据库
	db.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "assistant", fullResponse)

	log.Printf("✅ AI流式回复成功: length=%d\n", len(fullResponse))
}

// ============================================
// 调用火山引擎API（普通模式）
// ============================================

// callVolcengineAPI 调用火山引擎API，一次性返回完整回复
func callVolcengineAPI(model string, messages []map[string]interface{}) (string, error) {
	// 1. 获取API密钥
	apiKey := os.Getenv("VOLCENGINE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("VOLCENGINE_API_KEY未配置")
	}

	// 2. API端点URL
	endpoint := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"

	// 3. 构建请求体
	reqBody := VolcengineRequest{
		Model:    model,
		Messages: messages,
	}

	// 4. 将请求体转换为JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	log.Printf("📤 发送API请求: endpoint=%s, model=%s\n", endpoint, model)
	log.Printf("📤 请求体: %s\n", string(jsonData))

	// 5. 创建HTTP请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 6. 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey) // API认证

	// 7. 发送请求（超时时间60秒）
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 8. 读取响应体
	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("📥 API响应状态: %d\n", resp.StatusCode)
	log.Printf("📥 API响应体: %s\n", string(bodyBytes))

	// 9. 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误状态: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 10. 解析JSON响应
	var result VolcengineResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v, body: %s", err, string(bodyBytes))
	}

	// 11. 检查API返回的错误
	if result.Error != nil {
		return "", fmt.Errorf("API错误: %s (type: %s, code: %s)", result.Error.Message, result.Error.Type, result.Error.Code)
	}

	// 12. 检查是否有回复内容
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API返回空结果")
	}

	// 13. 提取AI回复的内容
	content := result.Choices[0].Message.Content
	log.Printf("✅ API调用成功: tokens=%d\n", result.Usage.TotalTokens)
	return content, nil
}

// ============================================
// 调用火山引擎API（流式模式）
// ============================================

// callVolcengineStreamAPI 调用火山引擎API，流式返回（逐字输出）
func callVolcengineStreamAPI(model string, messages []map[string]interface{}, w http.ResponseWriter) (string, error) {
	// 1. 获取API密钥
	apiKey := os.Getenv("VOLCENGINE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("VOLCENGINE_API_KEY未配置")
	}

	endpoint := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"

	// 2. 构建请求体（注意Stream设置为true）
	reqBody := VolcengineRequest{
		Model:    model,
		Messages: messages,
		Stream:   true, // 开启流式输出
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	log.Printf("📤 发送流式API请求: endpoint=%s, model=%s\n", endpoint, model)

	// 3. 创建HTTP请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 4. 发送请求（超时时间120秒，因为流式可能较慢）
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 5. 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ 流式API返回错误: %d, body: %s\n", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("API返回错误状态: %d", resp.StatusCode)
	}

	// 6. 获取Flusher接口（用于立即发送数据）
	flusher, ok := w.(http.Flusher)
	if !ok {
		return "", fmt.Errorf("Streaming不支持")
	}

	// 7. 逐行读取流式响应
	scanner := bufio.NewScanner(resp.Body)
	fullResponse := "" // 累积完整的回复内容
	chunkCount := 0    // 统计收到的数据块数量

	for scanner.Scan() {
		line := scanner.Text()

		// SSE格式：每行以"data: "开头
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 提取数据部分
		data := strings.TrimPrefix(line, "data: ")

		// 检查是否结束
		if data == "[DONE]" {
			log.Printf("✅ 流式输出完成: chunks=%d, length=%d\n", chunkCount, len(fullResponse))
			break
		}

		// 8. 解析每一块的JSON数据
		var streamData VolcengineStreamResponse
		if err := json.Unmarshal([]byte(data), &streamData); err != nil {
			log.Printf("⚠️  解析流式数据失败: %v, data: %s\n", err, data)
			continue
		}

		// 9. 提取增量内容
		if len(streamData.Choices) > 0 && streamData.Choices[0].Delta.Content != "" {
			content := streamData.Choices[0].Delta.Content
			fullResponse += content // 累积到完整回复中
			chunkCount++

			// 10. 转发给前端（实时显示）
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush() // 立即发送，不等待缓冲区满
		}
	}

	// 11. 检查读取过程中的错误
	if err := scanner.Err(); err != nil {
		log.Printf("❌ 读取流式数据错误: %v\n", err)
		return fullResponse, err
	}

	// 12. 发送结束标记
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	return fullResponse, nil
}

// ============================================
// 用户升级处理函数
// ============================================

// upgradeHandler 处理用户升级请求（从level 1升级到level 2）
func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. 获取用户信息
	userID := r.Header.Get("X-User-ID")
	currentLevel := r.Header.Get("X-User-Level")

	// 3. 检查是否已经是高级用户
	if currentLevel != "1" {
		http.Error(w, "您已经是高级用户了", http.StatusBadRequest)
		return
	}

	// 4. 解析请求体，获取用户的答案
	var data struct {
		Answer string `json:"answer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 5. 验证答案（这里是一个简单的验证题）
	correctAnswer := "杭电助手"
	userAnswer := strings.TrimSpace(data.Answer)

	if userAnswer != correctAnswer {
		log.Printf("❌ 升级失败: userID=%s, 错误答案=%s\n", userID, userAnswer)
		http.Error(w, "答案错误", http.StatusUnauthorized)
		return
	}

	// 6. 更新用户等级为2
	result, err := db.Exec("UPDATE users SET level = 2 WHERE id = ?", userID)
	if err != nil {
		log.Printf("❌ 升级失败: %v\n", err)
		http.Error(w, "升级失败", http.StatusInternalServerError)
		return
	}

	// 7. 检查是否更新成功
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "用户不存在", http.StatusNotFound)
		return
	}

	// 8. 返回成功响应
	log.Printf("✅ 用户升级成功: userID=%s\n", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "升级成功！",
		"level":   2,
	})
}

// ============================================
// HTML页面服务函数
// ============================================

// serveHTML 提供HTML页面（用于前端界面）
func serveHTML(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径请求
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// 返回index.html文件
	http.ServeFile(w, r, "index.html")
}
