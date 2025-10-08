package API_response

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// 默认Ollama服务器URL
const DefaultOllamaServerURL = "http://10.150.123.93:11434"

// OllamaClient 是一个封装了Ollama API调用的客户端结构体
type OllamaClient struct {
	ServerURL string
	Model     string
	Client    *http.Client
}

// Message 表示对话中的一条消息
type Message struct {
	Role    string `json:"role"`    // "user" 或 "assistant"
	Content string `json:"content"` // 消息内容
}

// OllamaRequest 表示发送给Ollama API的请求体
type OllamaRequest struct {
	Model    string    `json:"model"`
	Prompt   string    `json:"prompt,omitempty"`
	Messages []Message `json:"messages,omitempty"` // 对话历史
	Stream   bool      `json:"stream"`
}

// OllamaResponse 表示Ollama API返回的响应
type OllamaResponse struct {
	Model     string `json:"model,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func CleanResponseContent(content string) string {
	// 移除<think>和</think>标签及其内容
	cleaned := content
	// 处理嵌套的<think>标签
	for {
		start := strings.Index(cleaned, "<think>")
		end := strings.Index(cleaned, "</think>")
		if start == -1 || end == -1 || end < start {
			break
		}
		// 只保留标签外的内容
		cleaned = cleaned[:start] + cleaned[end+len("</think>"):]
	}
	// 修剪多余的空白字符
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// NewOllamaClient 创建一个新的Ollama客户端实例
func NewOllamaClient(serverURL string, model string) *OllamaClient {
	// 创建带有连接池配置的HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个主机最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
	}

	return &OllamaClient{
		ServerURL: serverURL,
		Model:     model,
		Client: &http.Client{
			Transport: transport,
			Timeout:   300 * time.Second,
		},
	}
}

// GenerateWithContext 使用对话历史发送请求到Ollama API并以流式方式获取响应
func (oc *OllamaClient) GenerateWithContext(messages []Message, callback func(string) bool) error {
	// 创建包含对话历史的请求
	reqBody := OllamaRequest{
		Model:    oc.Model,
		Messages: messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}
	resp, err := oc.Client.Post(
		oc.ServerURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 使用bufio.Scanner逐行读取响应流
	scanner := bufio.NewScanner(resp.Body)
	const maxScanTokenSize = 1024 * 1024 // 增加扫描缓冲区大小到1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		var ollamaResp OllamaResponse
		if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
			return fmt.Errorf("解析响应行失败: %w, 内容: %s", err, line)
		}

		// 如果有响应内容，调用回调函数
		if ollamaResp.Message.Content != "" {
			shouldContinue := callback(ollamaResp.Message.Content)
			if !shouldContinue {
				break
			}
		}

		// 检查是否完成
		if ollamaResp.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取响应流失败: %w", err)
	}

	return nil
}
