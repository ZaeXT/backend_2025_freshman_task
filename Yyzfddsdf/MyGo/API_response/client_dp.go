package API_response

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeepSeekClient 是一个封装了DeepSeek API调用的客户端结构体
type DeepSeekClient struct {
	ServerURL string
	APIKey    string
	Client    *http.Client
}

// DeepSeekRequest 表示发送给DeepSeek API的请求体
type DeepSeekRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Stream           bool      `json:"stream"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
}

// DeepSeekResponse 表示DeepSeek API返回的流式响应
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}

// NewDeepSeekClient 创建一个新的DeepSeek客户端实例
func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	// 创建带有连接池配置的HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个主机最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
	}

	return &DeepSeekClient{
		ServerURL: "https://api.deepseek.com/chat/completions",
		APIKey:    apiKey,
		Client: &http.Client{
			Transport: transport,
			Timeout:   300 * time.Second,
		},
	}
}

// GenerateWithContext 使用对话历史发送请求到DeepSeek API并以流式方式获取响应
func (dc *DeepSeekClient) GenerateWithContext(messages []Message, callback func(string) bool) error {
	// 创建请求体
	reqBody := DeepSeekRequest{
		Model:            "deepseek-chat",
		Messages:         messages,
		Stream:           true,
		Temperature:      1.3,
		MaxTokens:        4096,
		TopP:             1.0,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	req, err := http.NewRequest("POST", dc.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+dc.APIKey)

	resp, err := dc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 使用bufio.Scanner逐行读取响应流
	scanner := bufio.NewScanner(resp.Body)
	const maxScanTokenSize = 1024 * 1024 // 增加扫描缓冲区大小到1MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Text()

		// DeepSeek流式响应格式: data: {...}\n\n
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var deepseekResp DeepSeekResponse
		if err := json.Unmarshal([]byte(data), &deepseekResp); err != nil {
			// 忽略解析错误，继续处理下一行
			continue
		}

		// 如果有响应内容，调用回调函数
		if len(deepseekResp.Choices) > 0 && deepseekResp.Choices[0].Delta.Content != "" {
			shouldContinue := callback(deepseekResp.Choices[0].Delta.Content)
			if !shouldContinue {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取响应流失败: %w", err)
	}

	return nil
}

// SetAPIKey 设置API密钥
func (dc *DeepSeekClient) SetAPIKey(apiKey string) {
	dc.APIKey = apiKey
}
