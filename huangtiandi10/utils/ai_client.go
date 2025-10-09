package utils

import (
	"ai-qa-system/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// AI 请求体
type AIRequest struct {
	Model    string      `json:"model"`
	Messages []AIMessage `json:"messages"`
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AI 响应体
type AIResponse struct {
	Choices []struct {
		Message AIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// CallAI 通过模型名调用 AI
func CallAI(modelName, question string) (string, error) {
	modelConfig, ok := config.AppConfig.AIModels[modelName]
	if !ok {
		return "", fmt.Errorf("未找到模型配置: %s", modelName)
	}

	// 构造请求
	reqBody := AIRequest{
		Model: modelConfig.Endpoint, // ep-xxxxx
		Messages: []AIMessage{
			{Role: "user", Content: question},
		},
	}
	body, _ := json.Marshal(reqBody)

	// 发送请求
	req, err := http.NewRequest("POST", config.AppConfig.AI.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+modelConfig.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	var aiResp AIResponse
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		return "", fmt.Errorf("unmarshal error: %v, body=%s", err, string(respBody))
	}

	if aiResp.Error != nil {
		return "", fmt.Errorf("AI call failed code=%s msg=%s", aiResp.Error.Code, aiResp.Error.Message)
	}

	if len(aiResp.Choices) > 0 {
		return aiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("AI response empty: %s", string(respBody))
}
