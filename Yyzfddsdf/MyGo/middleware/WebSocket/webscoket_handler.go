package WebSocket

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"yyz.com/MyGo/API_response"
	"yyz.com/MyGo/controllers"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// WebSocket连接升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源连接
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Redis客户端
var redisClient *redis.Client
// Redis连接状态
var redisConnected = false

// 初始化Redis客户端
func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis服务器地址
		DB:   0,                // 使用默认数据库
	})
	
	// 测试Redis连接
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("警告: 无法连接到Redis服务器: %v，nonce验证功能将不可用", err)
	} else {
		redisConnected = true
		log.Println("Redis服务器连接成功")
	}
}

// 请求结构体
type AIRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	UseContext     bool   `json:"useContext,omitempty"`     // 是否使用上下文
	ClearContext   bool   `json:"clearContext,omitempty"`   // 是否清除上下文
	ConversationID uint   `json:"conversationId,omitempty"` // 可选的对话ID
}

// 响应结构体
type AIResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
	Done  bool   `json:"done"`
}

// 认证响应结构体
type AuthResponse struct {
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// buildSignString 构建签名字符串
func buildSignString(method, path, timestamp, nonce string, queryParams url.Values, body []byte) string {
	// 1. 添加HTTP方法
	signParts := []string{method}

	// 2. 添加路径
	signParts = append(signParts, path)

	// 3. 添加时间戳
	signParts = append(signParts, timestamp)

	// 4. 添加nonce
	signParts = append(signParts, nonce)

	// 5. 添加查询参数（按键排序）
	if len(queryParams) > 0 {
		var queryParts []string
		for key, values := range queryParams {
			for _, value := range values {
				queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
			}
		}
		sort.Strings(queryParts)
		signParts = append(signParts, strings.Join(queryParts, "&"))
	}

	// 6. 添加请求体（如果有）
	if len(body) > 0 {
		signParts = append(signParts, string(body))
	}

	// 使用&连接所有部分
	return strings.Join(signParts, "&")
}

// verifySignature 验证签名
func verifySignature(signString, signature, secretKey string) bool {
	// 创建HMAC签名
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(signString))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// 比较签名（使用安全的字符串比较防止时序攻击）
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// WebSocket处理函数
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 首先进行JWT token认证
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		sendAuthError(w, "未提供授权Token")
		return
	}

	// Token格式通常为 "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		sendAuthError(w, "Token格式错误")
		return
	}

	tokenString := parts[1]
	claims := &controllers.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return controllers.GetJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		sendAuthError(w, "无效或过期的Token")
		return
	}

	// 新增：验证请求头中的时间戳，防止重放攻击
	timestampStr := r.Header.Get("X-Timestamp")
	if timestampStr == "" {
		sendAuthError(w, "缺少时间戳")
		return
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		sendAuthError(w, "无效的时间戳格式")
		return
	}

	// 验证时间窗口（10秒）
	now := time.Now().Unix()
	if math.Abs(float64(now-timestamp)) > 10 {
		sendAuthError(w, "请求已过期")
		return
	}

	// 新增：使用nonce防止重放攻击
	nonceStr := r.Header.Get("X-Nonce")
	if nonceStr == "" {
		sendAuthError(w, "缺少nonce")
		return
	}

	// 验证nonce格式（32位十六进制字符串）
	if len(nonceStr) != 32 {
		sendAuthError(w, "无效的nonce格式")
		return
	}

	// 新增：验证签名
	signature := r.Header.Get("X-Signature")
	if signature == "" {
		sendAuthError(w, "缺少签名")
		return
	}

	// 构建签名字符串
	signString := buildSignString("GET", "/ws/ai", timestampStr, nonceStr, r.URL.Query(), nil)

	// 验证签名（这里需要使用客户端和服务端共享的密钥）
	// 注意：在实际应用中，应该根据客户端ID或其他标识获取对应的密钥
	secretKey := "your_shared_secret_key" // 这应该从配置或数据库中获取

	if !verifySignature(signString, signature, secretKey) {
		sendAuthError(w, "签名验证失败")
		return
	}

	// 强制执行nonce验证机制
	// 如果无法连接到Redis服务，应向客户端返回"服务器内部错误"响应
	if !redisConnected {
		sendAuthError(w, "服务器内部错误")
		return
	}

	// 使用Redis验证nonce是否已使用（防重放攻击）
	ctx := context.Background()
	redisKey := fmt.Sprintf("nonce:%s", nonceStr)

	// 检查nonce是否已存在于缓存中（已使用则拒绝）
	exists, err := redisClient.Exists(ctx, redisKey).Result()
	if err != nil {
		// Redis连接错误，记录日志并拒绝请求
		sendAuthError(w, "服务器内部错误")
		return
	}

	if exists > 0 {
		sendAuthError(w, "无效的nonce")
		return
	}

	// 存储nonce到Redis，设置10分钟过期时间
	err = redisClient.Set(ctx, redisKey, "1", 10*time.Minute).Err()
	if err != nil {
		// Redis存储错误，记录日志并拒绝请求
		log.Printf("错误: 无法存储nonce到Redis: %v", err)
		sendAuthError(w, "服务器内部错误")
		return
	}

	// 验证用户是否存在于数据库中
	var user models.User
	if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		sendAuthError(w, "用户不存在或已被删除")
		return
	}

	// 额外验证：确保数据库中的用户信息与Token声明一致
	if user.Email != claims.Email || user.Username != claims.Username {
		sendAuthError(w, "用户信息不匹配，Token可能已失效")
		return
	}

	// 认证通过，升级HTTP连接到WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级连接失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("客户端已连接，用户ID: %d", claims.UserID)

	// 创建Ollama客户端（默认使用deepseek-r1:8b模型）
	ollamaClient := API_response.NewOllamaClient("http://10.150.28.241:11434", "deepseek-r1:8b")

	// 为每个连接维护一个对话历史，包含系统消息
	chatHistory := []API_response.Message{
		{Role: "system", Content: "你是小玉，一个活泼可爱的助手，不要说自己是AI"},
	}

	// 创建对话记录变量
	var conversation models.Conversation

	// 发送认证成功消息
	authSuccess := AuthResponse{
		Success: true,
		Message: "认证成功，可以开始聊天",
	}
	if err := conn.WriteJSON(authSuccess); err != nil {
		log.Printf("发送认证成功消息失败: %v", err)
		return
	}

	// 持续读取客户端消息
	for {
		var req AIRequest
		// 读取JSON格式的请求
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			break
		}

		// 检查是否提供了对话ID且不为0（仅在还没有加载对话时处理）
		if req.ConversationID != 0 && conversation.ID == 0 {
			// 尝试加载指定的对话记录
			if err := database.DB.Where("id = ? AND user_id = ?", req.ConversationID, claims.UserID).First(&conversation).Error; err != nil {
				// 对话不存在，返回错误
				response := AIResponse{
					Error: "指定的对话不存在",
					Done:  true,
				}
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("发送错误消息失败: %v", err)
				}
				// 继续等待下一条消息而不是断开连接
				continue
			}

			// 加载对话历史记录
			var messages []models.Message
			database.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages)

			// 将历史消息转换为chatHistory格式（包含系统消息）
			chatHistory = []API_response.Message{
				{Role: "system", Content: "你是小玉，一个活泼可爱的助手，不要说自己是AI"},
			}
			for _, msg := range messages {
				chatHistory = append(chatHistory, API_response.Message{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}

			log.Printf("已加载对话ID %d 的历史记录，共 %d 条消息", conversation.ID, len(messages))

			// 如果请求中包含prompt，则立即处理该问题
			if req.Prompt != "" {
				// 直接处理消息，不需要等待下一轮循环
				processMessage(conn, &req, claims, &user, &conversation, &chatHistory, ollamaClient)
				continue
			} else {
				// 发送确认消息给客户端
				response := AIResponse{
					Text: fmt.Sprintf("已加载对话: %s", conversation.Title),
					Done: true,
				}
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("发送确认消息失败: %v", err)
				}
				continue
			}
		}

		// 如果还没有创建对话记录且不是加载现有对话的请求，则创建新对话
		if conversation.ID == 0 && req.ConversationID == 0 {
			// 创建新对话记录
			conversation = models.Conversation{
				UserID: claims.UserID,
				Title:  "新对话", // 默认标题，后续可以更新
			}

			// 保存对话记录到数据库
			if err := database.DB.Create(&conversation).Error; err != nil {
				log.Printf("创建对话记录失败: %v", err)
				// 即使创建对话记录失败，也继续处理，不影响主要功能
			}
		}

		// 处理消息
		processMessage(conn, &req, claims, &user, &conversation, &chatHistory, ollamaClient)
	}

	log.Printf("客户端已断开连接，用户ID: %d", claims.UserID)
}

// 处理单条消息的函数
func processMessage(conn *websocket.Conn, req *AIRequest, claims *controllers.Claims, user *models.User, conversation *models.Conversation, chatHistory *[]API_response.Message, ollamaClient *API_response.OllamaClient) {
	// 每次请求前重新查询用户信息，获取最新的token数量
	if err := database.DB.Where("id = ?", claims.UserID).First(user).Error; err != nil {
		response := AIResponse{
			Error: "用户信息获取失败",
			Done:  true,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("发送错误消息失败: %v", err)
		}
		return
	}

	if user.TokenCount <= 0 {
		response := AIResponse{
			Error: "token数量不足，请充值",
			Done:  true,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("发送token不足消息失败: %v", err)
		}
		return
	}
	log.Printf("收到用户 %d 的请求: %s, 剩余token: %d", claims.UserID, req.Prompt, user.TokenCount)

	// 如果指定了模型，则使用指定的模型
	if req.Model != "" {
		ollamaClient = API_response.NewOllamaClient("http://10.150.28.241:11434", req.Model)
	}

	// 如果请求清除上下文
	if req.ClearContext {
		// 清空内存中的对话历史，但保留系统消息
		*chatHistory = []API_response.Message{
			{Role: "system", Content: "你是小玉，一个活泼可爱的助手，不要说自己是AI"},
		}

		// 同时清空数据库中的消息记录（保留对话记录本身）
		if conversation.ID != 0 {
			if err := database.DB.Where("conversation_id = ?", conversation.ID).Delete(&models.Message{}).Error; err != nil {
				log.Printf("清空对话消息失败: %v", err)
			} else {
				// 重置对话统计信息
				if err := database.DB.Model(conversation).Updates(map[string]interface{}{
					"message_count": 0,
					"token_used":    0,
				}).Error; err != nil {
					log.Printf("重置对话统计信息失败: %v", err)
				} else {
					// 更新内存中的对话对象
					conversation.MessageCount = 0
					conversation.TokenUsed = 0
				}
			}
		}

		response := AIResponse{
			Text: "上下文已清除，历史记录已同步删除",
			Done: true,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("发送消息失败: %v", err)
		}
		return
	}

	// 统一使用GenerateWithContext方法
	// 添加当前用户消息到临时历史
	userMessage := API_response.Message{
		Role:    "user",
		Content: req.Prompt,
	}

	// 创建用户消息记录
	userMsg := models.Message{
		ConversationID: conversation.ID, // 使用当前对话ID
		Role:           "user",
		Content:        req.Prompt,
		TokenCount:     0, // 用户消息不计算token
	}
	database.DB.Create(&userMsg)

	// 将用户消息添加到对话历史
	*chatHistory = append(*chatHistory, userMessage)

	// 创建临时历史副本（注意：不要重复添加用户消息）
	tempHistory := make([]API_response.Message, len(*chatHistory))
	copy(tempHistory, *chatHistory)

	// 使用带上下文的流式调用
	fullResponse := ""
	chunkCount := 0 // 计算chunk数量
	err := ollamaClient.GenerateWithContext(tempHistory, func(chunk string) bool {
		fullResponse += chunk
		chunkCount++ // 每收到一个chunk就计数
		// 发送响应块到客户端
		response := AIResponse{
			Text: chunk,
			Done: false,
		}
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("发送消息失败: %v", err)
			return false
		}
		// 继续接收下一个响应块
		return true
	})

	if chunkCount > 0 {
		log.Printf("本次回答使用 %d 个token", chunkCount)
	}

	// 创建助手消息记录
	assistantMsg := models.Message{
		ConversationID: conversation.ID, // 使用当前对话ID
		Role:           "assistant",
		Content:        fullResponse,
		TokenCount:     chunkCount,
	}
	database.DB.Create(&assistantMsg)

	if chunkCount > 0 {
		newTokenCount := user.TokenCount - chunkCount
		if newTokenCount < 0 {
			newTokenCount = 0
		}

		user.TokenCount = newTokenCount

		if err := database.DB.Model(&models.User{}).
			Where("id = ?", claims.UserID).
			Update("token_count", newTokenCount).Error; err != nil {
			log.Printf("更新token数量失败: %v", err)
		} else {
			log.Printf("用户 %d 使用 %d 个token，剩余 %d 个", claims.UserID, chunkCount, newTokenCount)
		}

		// 更新对话记录的统计信息
		conversation.MessageCount += 2 // 用户消息+助手消息
		conversation.TokenUsed += chunkCount
		database.DB.Model(conversation).Updates(map[string]interface{}{
			"message_count": conversation.MessageCount,
			"token_used":    conversation.TokenUsed,
		})
	}

	// 更新对话历史，添加助手响应
	if err == nil && fullResponse != "" {
		cleanedResponse := API_response.CleanResponseContent(fullResponse)
		*chatHistory = append(*chatHistory,
			API_response.Message{Role: "assistant", Content: cleanedResponse},
		)

		// 如果是第一条助手消息，设置对话标题
		if conversation.Title == "新对话" && len(*chatHistory) >= 2 { // 用户消息+助手消息
			// 找到第一条用户消息作为标题
			var firstUserPrompt string
			for _, msg := range *chatHistory {
				if msg.Role == "user" {
					firstUserPrompt = msg.Content
					break
				}
			}

			if firstUserPrompt != "" {
				// 使用用户的第一条消息作为标题，限制长度
				title := firstUserPrompt
				if len(title) > 50 {
					title = title[:50] + "..."
				}
				if err := database.DB.Model(&models.Conversation{}).Where("id = ?", conversation.ID).Update("title", title).Error; err != nil {
					log.Printf("更新对话标题失败: %v", err)
				} else {
					conversation.Title = title // 更新内存中的标题
				}
			}
		}
	}

	// 发送完成标记
	response := AIResponse{
		Text: "",
		Done: true,
	}
	if err != nil {
		response.Error = err.Error()
		log.Printf("生成响应失败: %v", err)
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("发送完成标记失败: %v", err)
	}
}

// 发送认证错误响应
func sendAuthError(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	response := AuthResponse{
		Error:   errorMsg,
		Success: false,
	}
	json.NewEncoder(w).Encode(response)
}
