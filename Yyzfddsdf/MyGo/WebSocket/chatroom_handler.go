package WebSocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"yyz.com/MyGo/controllers"
	"yyz.com/MyGo/database"
	"yyz.com/MyGo/models"
)

// 公共聊天室WebSocket连接升级器
var chatroomUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源连接
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	// 启用压缩（减少带宽使用）
	EnableCompression: true,
	// 设置握手超时时间
	HandshakeTimeout: 10 * time.Second,
}

// 聊天室客户端结构
type ChatroomClient struct {
	Conn     *websocket.Conn
	Username string
	UserID   uint
}

// 聊天室消息结构
type ChatroomMessage struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username"`
	Content  string `json:"content"`
	Time     string `json:"time"`
	Type     string `json:"type"` // "message", "join", "leave", "system"
}

// 聊天室全局变量
var (
	// 存储所有连接的客户端
	chatroomClients = make(map[*websocket.Conn]*ChatroomClient)
	chatroomMutex   sync.RWMutex

	// 存储聊天记录（临时变量，限制100条）
	chatroomMessages []ChatroomMessage
	messagesMutex    sync.RWMutex
	maxMessages      = 100
)

// 系统消息
func systemMessage(content string) ChatroomMessage {
	return ChatroomMessage{
		Username: "系统",
		Content:  content,
		Time:     time.Now().Format("15:04:05"),
		Type:     "system",
	}
}

// 广播消息给所有客户端
func broadcastChatroomMessage(message ChatroomMessage) {
	messagesMutex.Lock()
	// 添加新消息，如果超过限制则移除最旧的消息
	chatroomMessages = append(chatroomMessages, message)
	if len(chatroomMessages) > maxMessages {
		chatroomMessages = chatroomMessages[len(chatroomMessages)-maxMessages:]
	}
	messagesMutex.Unlock()

	chatroomMutex.RLock()
	defer chatroomMutex.RUnlock()

	messageJSON, _ := json.Marshal(message)

	for _, client := range chatroomClients {
		err := client.Conn.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Printf("广播消息失败: %v", err)
			client.Conn.Close()
			delete(chatroomClients, client.Conn)
		}
	}
}

// 发送聊天记录给新用户
func sendChatroomHistory(client *websocket.Conn) {
	messagesMutex.RLock()
	defer messagesMutex.RUnlock()

	for _, msg := range chatroomMessages {
		messageJSON, _ := json.Marshal(msg)
		err := client.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Printf("发送历史消息失败: %v", err)
			return
		}
	}
}

// 公共聊天室WebSocket处理函数
func HandleChatroomWebSocket(w http.ResponseWriter, r *http.Request) {
	// 增强的CORS响应头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Timestamp, X-Nonce, Origin, Accept, Connection, Host, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")

	// 处理OPTIONS预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 首先进行JWT token认证
	// 从URL参数中获取认证信息
	queryParams := r.URL.Query()

	// 优先从URL参数获取，如果没有则从请求头获取
	authHeader := queryParams.Get("Authorization")
	if authHeader == "" {
		authHeader = r.Header.Get("Authorization")
	}

	if authHeader == "" {
		sendChatroomAuthError(w, "未提供授权Token")
		return
	}

	// Token格式通常为 "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		sendChatroomAuthError(w, "Token格式错误")
		return
	}

	tokenString := parts[1]
	claims := &controllers.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return controllers.GetJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		sendChatroomAuthError(w, "无效或过期的Token")
		return
	}

	// 验证时间戳，防止重放攻击
	timestampStr := queryParams.Get("X-Timestamp")
	if timestampStr == "" {
		timestampStr = r.Header.Get("X-Timestamp")
	}

	if timestampStr == "" {
		sendChatroomAuthError(w, "缺少时间戳")
		return
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		sendChatroomAuthError(w, "无效的时间戳格式")
		return
	}

	// 验证时间窗口（30秒）
	now := time.Now().Unix()
	if abs(now-timestamp) > 30 {
		sendChatroomAuthError(w, "请求已过期")
		return
	}

	// 验证nonce
	nonceStr := queryParams.Get("X-Nonce")
	if nonceStr == "" {
		nonceStr = r.Header.Get("X-Nonce")
	}

	if nonceStr == "" {
		sendChatroomAuthError(w, "缺少nonce")
		return
	}

	// 验证nonce格式（32位十六进制字符串）
	if len(nonceStr) != 32 {
		sendChatroomAuthError(w, "无效的nonce格式")
		return
	}

	// Redis nonce验证（防止重放攻击）
	if redisConnected {
		// 检查nonce是否已存在
		nonceKey := "nonce:" + nonceStr
		exists, err := redisClient.Exists(context.Background(), nonceKey).Result()
		if err != nil {
			log.Printf("Redis查询错误: %v", err)
			sendChatroomAuthError(w, "服务器内部错误")
			return
		}

		if exists > 0 {
			sendChatroomAuthError(w, "重复的nonce，可能为重放攻击")
			return
		}

		// 存储nonce，10分钟过期
		err = redisClient.Set(context.Background(), nonceKey, "1", 10*time.Minute).Err()
		if err != nil {
			log.Printf("Redis存储错误: %v", err)
			sendChatroomAuthError(w, "服务器内部错误")
			return
		}
	} else {
		log.Printf("警告: Redis未连接，跳过nonce验证")
	}

	// 验证用户是否存在于数据库中
	var user models.User
	if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		sendChatroomAuthError(w, "用户不存在或已被删除")
		return
	}

	// 额外验证：确保数据库中的用户信息与Token声明一致
	if user.Email != claims.Email || user.Username != claims.Username {
		sendChatroomAuthError(w, "用户信息不匹配，Token可能已失效")
		return
	}

	// 认证通过，升级HTTP连接到WebSocket连接
	conn, err := chatroomUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("升级连接失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("用户进入公共聊天室: %s (ID: %d)", user.Username, claims.UserID)

	// 创建客户端对象
	client := &ChatroomClient{
		Conn:     conn,
		Username: user.Username,
		UserID:   claims.UserID,
	}

	// 添加新客户端
	chatroomMutex.Lock()
	chatroomClients[conn] = client
	chatroomMutex.Unlock()

	// 发送用户信息给前端
	userInfo := map[string]interface{}{
		"type":     "user_info",
		"username": user.Username,
		"userID":   claims.UserID,
	}
	userInfoJSON, _ := json.Marshal(userInfo)
	conn.WriteMessage(websocket.TextMessage, userInfoJSON)

	// 发送欢迎消息和聊天记录
	broadcastChatroomMessage(systemMessage(fmt.Sprintf("%s 加入了聊天室", user.Username)))
	sendChatroomHistory(conn)

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
		var messageData map[string]interface{}
		// 读取JSON格式的请求
		err := conn.ReadJSON(&messageData)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket连接意外关闭: %v", err)
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("WebSocket正常关闭")
			} else {
				log.Printf("读取消息失败: %v", err)
			}
			break
		}

		// 处理消息
		messageType, _ := messageData["type"].(string)

		if messageType == "message" {
			content, exists := messageData["content"].(string)
			if !exists || content == "" {
				continue
			}

			// 创建消息对象
			message := ChatroomMessage{
				ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
				Username: user.Username,
				Content:  content,
				Time:     time.Now().Format("15:04:05"),
				Type:     "message",
			}

			log.Printf("公共聊天室 - %s: %s", user.Username, content)
			broadcastChatroomMessage(message)
		}
	}

	// 用户断开连接
	chatroomMutex.Lock()
	delete(chatroomClients, conn)
	chatroomMutex.Unlock()

	broadcastChatroomMessage(systemMessage(fmt.Sprintf("%s 离开了聊天室", user.Username)))
	log.Printf("用户离开公共聊天室: %s (ID: %d)", user.Username, claims.UserID)
}

// 发送认证错误响应
func sendChatroomAuthError(w http.ResponseWriter, errorMsg string) {
	response := map[string]interface{}{
		"error":   errorMsg,
		"success": false,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(response)
}

// 绝对值函数
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
