package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"Piao/config"
	"Piao/services"
)

// Chat 普通聊天（一次性返回）
func Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userLevel := r.Header.Get("X-User-Level")

	var data struct {
		ConversationID int    `json:"conversation_id"`
		Message        string `json:"message"`
		Model          string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("💬 收到聊天请求: conversationID=%d, model=%s, userLevel=%s\n",
		data.ConversationID, data.Model, userLevel)

	// 权限检查
	if strings.Contains(data.Model, "ADVANCED") && userLevel == "1" {
		log.Printf("⛔ 权限不足: 用户level=%s 尝试使用高级模型\n", userLevel)
		http.Error(w, "权限不足，高级模型需要高级用户", http.StatusForbidden)
		return
	}

	// 保存用户消息
	config.DB.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "user", data.Message)

	// 获取历史消息
	messages := getConversationMessages(data.ConversationID)
	log.Printf("📚 加载历史消息: count=%d\n", len(messages))

	// 调用AI API
	response, err := services.CallVolcengineAPI(data.Model, messages)
	if err != nil {
		log.Printf("❌ AI调用失败: %v\n", err)
		http.Error(w, "AI调用失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 保存AI回复
	config.DB.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "assistant", response)

	log.Printf("✅ AI回复成功: length=%d\n", len(response))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

// ChatStream 流式聊天
func ChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userLevel := r.Header.Get("X-User-Level")

	var data struct {
		ConversationID int    `json:"conversation_id"`
		Message        string `json:"message"`
		Model          string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("💬 收到流式聊天请求: conversationID=%d, model=%s\n", data.ConversationID, data.Model)

	// 权限检查
	if strings.Contains(data.Model, "ADVANCED") && userLevel == "1" {
		log.Printf("⛔ 权限不足\n")
		http.Error(w, "权限不足，高级模型需要高级用户", http.StatusForbidden)
		return
	}

	// 保存用户消息
	config.DB.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "user", data.Message)

	// 获取历史消息
	messages := getConversationMessages(data.ConversationID)
	log.Printf("📚 加载历史消息: count=%d\n", len(messages))

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 调用流式API
	fullResponse, err := services.CallVolcengineStreamAPI(data.Model, messages, w)
	if err != nil {
		log.Printf("❌ AI流式调用失败: %v\n", err)
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		return
	}

	// 保存完整回复
	config.DB.Exec("INSERT INTO messages (conversation_id, role, content) VALUES (?, ?, ?)",
		data.ConversationID, "assistant", fullResponse)

	log.Printf("✅ AI流式回复成功: length=%d\n", len(fullResponse))
}

// getConversationMessages 获取对话的历史消息
func getConversationMessages(conversationID int) []map[string]interface{} {
	rows, _ := config.DB.Query(
		"SELECT role, content FROM messages WHERE conversation_id = ? ORDER BY created_at ASC",
		conversationID)
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": content,
		})
	}
	return messages
}
