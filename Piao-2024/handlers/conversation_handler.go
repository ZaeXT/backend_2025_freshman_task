package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"Piao/config"
	"Piao/models"
)

// GetConversations 获取用户的对话列表
func GetConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	rows, err := config.DB.Query(
		"SELECT id, title, created_at FROM conversations WHERE user_id = ? ORDER BY created_at DESC",
		userID)
	if err != nil {
		log.Printf("❌ 查询对话列表失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		rows.Scan(&conv.ID, &conv.Title, &conv.CreatedAt)
		conversations = append(conversations, conv)
	}

	log.Printf("✅ 查询对话列表成功: userID=%s, count=%d\n", userID, len(conversations))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// CreateConversation 创建新对话
func CreateConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")

	var data struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := config.DB.Exec(
		"INSERT INTO conversations (user_id, title) VALUES (?, ?)",
		userID, data.Title)
	if err != nil {
		log.Printf("❌ 创建对话失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	log.Printf("✅ 创建对话成功: id=%d, title=%s\n", id, data.Title)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// GetMessages 获取对话的消息列表
func GetMessages(w http.ResponseWriter, r *http.Request) {
	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		http.Error(w, "conversation_id required", http.StatusBadRequest)
		return
	}

	rows, err := config.DB.Query(
		"SELECT id, role, content, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at ASC",
		conversationID)
	if err != nil {
		log.Printf("❌ 查询消息失败: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.CreatedAt)
		messages = append(messages, msg)
	}

	log.Printf("✅ 查询消息成功: conversationID=%s, count=%d\n", conversationID, len(messages))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
