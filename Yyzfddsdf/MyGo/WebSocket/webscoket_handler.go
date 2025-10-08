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
	"github.com/pkoukk/tiktoken-go"
	"yyz.com/MyGo/API_response"
	"yyz.com/MyGo/controllers"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// 系统提示词
const SystemPrompt = "你是小玉，一个活泼可爱的助手，不要说自己是AI"

// DeepSeek API密钥配置
const DeepSeekAPIKey = "" // 请在此处设置您的DeepSeek API密钥

// WebSocket连接升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源连接
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	// 启用压缩（减少带宽使用）
	EnableCompression: true,
	// 设置握手超时时间
	HandshakeTimeout: 30 * time.Second,
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
	Prompt                string `json:"prompt"`
	Model                 string `json:"model"`
	ClearContext          bool   `json:"clearContext,omitempty"`          // 是否清除上下文
	ConversationID        uint   `json:"conversationId,omitempty"`        // 可选的对话ID
	CreateNewConversation bool   `json:"createNewConversation,omitempty"` // 是否创建新对话
}

// 使用tiktoken-go库计算文本的token数量
func calculateTokenCount(text string) int {
	// 初始化tiktoken编码器，使用gpt-3.5-turbo模型的编码方式
	// 这里使用cl100k_base编码，适用于大多数OpenAI模型
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		log.Printf("初始化tiktoken编码器失败: %v，使用字符数作为替代", err)
		return len(text)
	}

	// 编码文本并返回token数量
	tokens := encoding.Encode(text, nil, nil)
	return len(tokens)
}

// 提取第一句话的最多前N个字
func extractFirstSentence(text string, maxWords int) string {
	// 定义句子分隔符
	sentenceSeparators := []string{".", "!", "?", "。", "！", "？", "\n"}

	// 找到第一个句子分隔符的位置
	firstSentenceEnd := len(text)
	for _, sep := range sentenceSeparators {
		if pos := strings.Index(text, sep); pos != -1 && pos < firstSentenceEnd {
			firstSentenceEnd = pos + len(sep)
		}
	}

	// 提取第一句话
	firstSentence := text[:firstSentenceEnd]

	// 按空格分割成单词
	words := strings.Fields(firstSentence)

	// 限制单词数量
	if len(words) > maxWords {
		words = words[:maxWords]
	}

	// 重新组合成字符串
	result := strings.Join(words, " ")

	// 如果原文本没有句子分隔符，且结果比原文本短，添加省略号
	if firstSentenceEnd == len(text) && len(result) < len(text) {
		result += "..."
	}

	return result
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
	// 增强的CORS响应头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Timestamp, X-Nonce, Origin, Accept, Connection, Host, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// 特别为WebSocket添加的头部
	w.Header().Set("Access-Control-Expose-Headers", "Authorization")

	// 安全头设置
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' ws: wss:;")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// HSTS (HTTP Strict Transport Security) - 仅在HTTPS中启用
	if r.TLS != nil {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// 处理OPTIONS预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 首先进行JWT token认证
	// 从URL参数中获取认证信息（因为WebSocket无法在连接时设置请求头）
	queryParams := r.URL.Query()

	// 优先从URL参数获取，如果没有则从请求头获取（向后兼容）
	authHeader := queryParams.Get("Authorization")
	if authHeader == "" {
		authHeader = r.Header.Get("Authorization")
	}

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
	// 优先从URL参数获取，如果没有则从请求头获取
	timestampStr := queryParams.Get("X-Timestamp")
	if timestampStr == "" {
		timestampStr = r.Header.Get("X-Timestamp")
	}

	if timestampStr == "" {
		sendAuthError(w, "缺少时间戳")
		return
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		sendAuthError(w, "无效的时间戳格式")
		return
	}

	// 验证时间窗口（30秒）
	now := time.Now().Unix()
	if math.Abs(float64(now-timestamp)) > 30 {
		sendAuthError(w, "请求已过期")
		return
	}

	// 新增：使用nonce防止重放攻击
	// 优先从URL参数获取，如果没有则从请求头获取
	nonceStr := queryParams.Get("X-Nonce")
	if nonceStr == "" {
		nonceStr = r.Header.Get("X-Nonce")
	}

	if nonceStr == "" {
		sendAuthError(w, "缺少nonce")
		return
	}

	// 验证nonce格式（32位十六进制字符串）
	if len(nonceStr) != 32 {
		sendAuthError(w, "无效的nonce格式")
		return
	}

	// 已移除签名验证，仅保留时间戳和nonce验证

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
	ollamaClient := API_response.NewOllamaClient(API_response.DefaultOllamaServerURL, "deepseek-r1:8b")

	// 创建DeepSeek客户端（需要设置API密钥）
	deepseekClient := API_response.NewDeepSeekClient(DeepSeekAPIKey) // 使用顶部定义的API密钥常量

	// 为每个连接维护一个对话历史，包含系统消息
	chatHistory := []API_response.Message{
		{Role: "system", Content: SystemPrompt},
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

	// 设置连接保活机制
	conn.SetPongHandler(func(string) error {
		return nil
	})

	// 启动心跳检测
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	go func() {
		for {
			select {
			case <-heartbeatTicker.C:
				// 发送Ping消息保持连接
				if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
					log.Printf("发送Ping消息失败: %v", err)
					return
				}
			}
		}
	}()

	// 持续读取客户端消息
	for {
		var req AIRequest
		// 读取JSON格式的请求
		err := conn.ReadJSON(&req)
		if err != nil {
			// 更详细的错误处理
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket连接意外关闭: %v", err)
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("WebSocket正常关闭")
			} else {
				log.Printf("读取消息失败: %v", err)
			}
			break
		}

		// 检查是否提供了对话ID且不为0（需要检查是否需要切换对话）
		if req.ConversationID != 0 {
			// 如果当前对话ID与请求的对话ID不同，或者当前没有加载对话，则需要加载新对话
			if conversation.ID != req.ConversationID || conversation.ID == 0 {
				// 尝试加载指定的对话记录
				var targetConversation models.Conversation
				if err := database.DB.Where("id = ? AND user_id = ?", req.ConversationID, claims.UserID).First(&targetConversation).Error; err != nil {
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

				// 切换到目标对话
				conversation = targetConversation

				// 加载对话历史记录
				var messages []models.Message
				database.DB.Where("conversation_id = ?", conversation.ID).Order("created_at ASC").Find(&messages)

				// 将历史消息转换为chatHistory格式（包含系统消息）
				chatHistory = []API_response.Message{
					{Role: "system", Content: SystemPrompt},
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
					processMessage(conn, &req, claims, &user, &conversation, &chatHistory, ollamaClient, deepseekClient)
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
		}

		// 检查是否是创建新对话的请求
		if req.CreateNewConversation {
			// 创建新对话记录
			newConversation := models.Conversation{
				UserID: claims.UserID,
				Title:  "新对话", // 默认标题，后续可以更新
			}

			// 保存新对话记录到数据库
			if err := database.DB.Create(&newConversation).Error; err != nil {
				log.Printf("创建新对话记录失败: %v", err)
				response := AIResponse{
					Error: "创建新对话失败",
					Done:  true,
				}
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("发送错误消息失败: %v", err)
				}
				continue
			}

			// 切换到新对话
			conversation = newConversation
			// 重置对话历史
			chatHistory = []API_response.Message{
				{Role: "system", Content: SystemPrompt},
			}

			// 发送成功响应
			response := AIResponse{
				Text: "已创建新对话",
				Done: true,
			}
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("发送成功消息失败: %v", err)
			}
			continue
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
		processMessage(conn, &req, claims, &user, &conversation, &chatHistory, ollamaClient, deepseekClient)
	}

	log.Printf("客户端已断开连接，用户ID: %d", claims.UserID)
}

// 处理单条消息的函数
func processMessage(conn *websocket.Conn, req *AIRequest, claims *controllers.Claims, user *models.User, conversation *models.Conversation, chatHistory *[]API_response.Message, ollamaClient *API_response.OllamaClient, deepseekClient *API_response.DeepSeekClient) {
	// 检查连接是否仍然有效
	if err := conn.WriteMessage(websocket.PingMessage, []byte("health-check")); err != nil {
		log.Printf("连接健康检查失败，连接可能已断开: %v", err)
		return
	}

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
	var currentClient interface{}
	if req.Model == "DeepSeek-V3.2" {
		// 使用DeepSeek客户端
		currentClient = deepseekClient
		log.Printf("用户 %d 选择使用DeepSeek-V3.2模型", claims.UserID)
	} else if req.Model != "" {
		// 使用Ollama客户端
		ollamaClient = API_response.NewOllamaClient(API_response.DefaultOllamaServerURL, req.Model)
		currentClient = ollamaClient
		log.Printf("用户 %d 选择使用Ollama模型: %s", claims.UserID, req.Model)
	} else {
		// 默认使用Ollama客户端
		currentClient = ollamaClient
		log.Printf("用户 %d 使用默认Ollama模型", claims.UserID)
	}

	// 如果请求清除上下文
	if req.ClearContext {
		// 清空内存中的对话历史，但保留系统消息
		*chatHistory = []API_response.Message{
			{Role: "system", Content: "你是小玉，一个活泼可爱的助手，不要说自己是AI"},
		}

		// 硬删除对话记录和所有消息记录
		if conversation.ID != 0 {
			// 开始事务
			tx := database.DB.Begin()
			if tx.Error != nil {
				log.Printf("事务开始失败: %v", tx.Error)
			} else {
				// 硬删除关联的消息记录
				if err := tx.Unscoped().Where("conversation_id = ?", conversation.ID).Delete(&models.Message{}).Error; err != nil {
					tx.Rollback()
					log.Printf("硬删除消息记录失败: %v", err)
				} else {
					// 硬删除对话记录
					if err := tx.Unscoped().Delete(&conversation).Error; err != nil {
						tx.Rollback()
						log.Printf("硬删除对话记录失败: %v", err)
					} else {
						// 提交事务
						if err := tx.Commit().Error; err != nil {
							log.Printf("事务提交失败: %v", err)
						} else {
							// 重置内存中的对话对象
							conversation.ID = 0
							conversation.Title = "新对话"
							conversation.MessageCount = 0
							conversation.TokenUsed = 0
							log.Printf("对话记录 %d 已硬删除", conversation.ID)
						}
					}
				}
			}
		}

		response := AIResponse{
			Text: "对话记录已完全删除，下次发送消息将创建新对话",
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

	// 计算上下文中的token数量（包括系统提示词和所有历史消息）
	contextTokenCount := 0
	for _, msg := range *chatHistory {
		contextTokenCount += calculateTokenCount(msg.Content)
	}

	// 加上当前用户输入的token数量
	inputTokenCount := calculateTokenCount(req.Prompt)
	totalInputTokenCount := contextTokenCount + inputTokenCount

	// 创建用户消息记录
	userMsg := models.Message{
		ConversationID: conversation.ID, // 使用当前对话ID
		Role:           "user",
		Content:        req.Prompt,
		TokenCount:     inputTokenCount, // 记录用户输入的token数量（基于tiktoken-go库计算）
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
	var err error

	// 根据选择的模型调用不同的客户端
	switch client := currentClient.(type) {
	case *API_response.OllamaClient:
		err = client.GenerateWithContext(tempHistory, func(chunk string) bool {
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
	case *API_response.DeepSeekClient:
		// 检查API密钥是否已设置
		if client.APIKey == "" {
			err = fmt.Errorf("DeepSeek API密钥未设置，请先配置API密钥")
		} else {
			err = client.GenerateWithContext(tempHistory, func(chunk string) bool {
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
		}
	default:
		err = fmt.Errorf("不支持的客户端类型")
	}

	// 计算总token消耗（上下文+当前输入+输出）
	totalTokenCount := totalInputTokenCount + chunkCount
	log.Printf("本次请求消耗 %d 个token (上下文: %d, 输入: %d, 输出: %d) [基于tiktoken-go库计算]", totalTokenCount, contextTokenCount, inputTokenCount, chunkCount)

	// 创建助手消息记录 - 保存清理后的内容
	cleanedResponse := API_response.CleanResponseContent(fullResponse)
	assistantMsg := models.Message{
		ConversationID: conversation.ID, // 使用当前对话ID
		Role:           "assistant",
		Content:        cleanedResponse,
		TokenCount:     chunkCount, // 输出token按chunk数量计算
	}
	database.DB.Create(&assistantMsg)

	if totalTokenCount > 0 {
		newTokenCount := user.TokenCount - totalTokenCount
		if newTokenCount < 0 {
			newTokenCount = 0
		}

		user.TokenCount = newTokenCount

		if err := database.DB.Model(&models.User{}).
			Where("id = ?", claims.UserID).
			Update("token_count", newTokenCount).Error; err != nil {
			log.Printf("更新token数量失败: %v", err)
		} else {
			log.Printf("用户 %d 使用 %d 个token，剩余 %d 个 [上下文: %d (tiktoken), 输入: %d (tiktoken), 输出: %d (chunk)]",
				claims.UserID, totalTokenCount, newTokenCount, contextTokenCount, inputTokenCount, chunkCount)
		}

		// 更新对话记录的统计信息
		conversation.MessageCount += 2 // 用户消息+助手消息
		conversation.TokenUsed += totalTokenCount
		database.DB.Model(conversation).Updates(map[string]interface{}{
			"message_count": conversation.MessageCount,
			"token_used":    conversation.TokenUsed,
		})
	}

	// 更新对话历史，添加助手响应
	if err == nil && fullResponse != "" {
		// 使用已经清理过的cleanedResponse变量，避免重复调用CleanResponseContent
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
				// 提取第一句话的最多前10个字作为标题
				title := extractFirstSentence(firstUserPrompt, 10)
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
