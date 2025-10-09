package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"Piao/config"
	"Piao/models"
)

// CallVolcengineAPI 调用火山引擎API（普通模式）
func CallVolcengineAPI(model string, messages []map[string]interface{}) (string, error) {
	if config.VolcengineAPIKey == "" {
		return "", fmt.Errorf("VOLCENGINE_API_KEY未配置")
	}

	// 构建请求
	reqBody := models.VolcengineRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	log.Printf("📤 发送API请求: model=%s\n", model)

	req, err := http.NewRequest("POST", config.VolcengineEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.VolcengineAPIKey)

	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("📥 API响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误状态: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var result models.VolcengineResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API错误: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API返回空结果")
	}

	content := result.Choices[0].Message.Content
	log.Printf("✅ API调用成功: tokens=%d\n", result.Usage.TotalTokens)
	return content, nil
}

// CallVolcengineStreamAPI 调用火山引擎API（流式模式）
func CallVolcengineStreamAPI(model string, messages []map[string]interface{}, w http.ResponseWriter) (string, error) {
	if config.VolcengineAPIKey == "" {
		return "", fmt.Errorf("VOLCENGINE_API_KEY未配置")
	}

	// 构建请求
	reqBody := models.VolcengineRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	log.Printf("📤 发送流式API请求: model=%s\n", model)

	req, err := http.NewRequest("POST", config.VolcengineEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.VolcengineAPIKey)

	// 发送请求
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API返回错误状态: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 获取Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		return "", fmt.Errorf("Streaming不支持")
	}

	// 读取流式响应
	scanner := bufio.NewScanner(resp.Body)
	fullResponse := ""
	chunkCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			log.Printf("✅ 流式输出完成: chunks=%d\n", chunkCount)
			break
		}

		var streamData models.VolcengineStreamResponse
		if err := json.Unmarshal([]byte(data), &streamData); err != nil {
			continue
		}

		if len(streamData.Choices) > 0 && streamData.Choices[0].Delta.Content != "" {
			content := streamData.Choices[0].Delta.Content
			fullResponse += content
			chunkCount++

			// 转发给前端
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		return fullResponse, err
	}

	// 发送结束标记
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	return fullResponse, nil
}
