package volcengine

import (
	"ai-qa-backend/internal/configs"
	"ai-qa-backend/internal/model"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type apiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiChatRequest struct {
	Model    string           `json:"model"`
	Messages []apiChatMessage `json:"messages"`
	Stream   bool             `json:"stream"`
}

type apiStreamChoiceDelta struct {
	Content string `json:"content"`
}

type apiStreamChoice struct {
	Delta apiStreamChoiceDelta `json:"delta"`
}

type apiStreamResponse struct {
	Choice []apiStreamChoice `json:"choices"`
}

type ChatRequest struct {
	SystemPrompt string
	Messages     []*model.Message
}

type AvailableModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type VolcengineAdapter struct {
	client          *http.Client
	apiKey          string
	baseURL         string
	AvailableModels []configs.ModelInfo
	tierLevels      map[string]int
}

func NewVolcengineAdapter() *VolcengineAdapter {
	cfg := configs.Conf.VolcEngine
	tierLevels := map[string]int{
		"free":    0,
		"premium": 1,
		"pro":     2,
	}
	return &VolcengineAdapter{
		client:          &http.Client{Timeout: 3 * time.Minute},
		apiKey:          cfg.APIKey,
		AvailableModels: cfg.AvailableModels,
		baseURL:         cfg.BaseURL,
		tierLevels:      tierLevels,
	}
}

func (a *VolcengineAdapter) GetAvailableModelsForTier(userTier string) []AvailableModel {
	userLevel, ok := a.tierLevels[userTier]
	if !ok {
		userLevel = 0
	}

	var result []AvailableModel
	for _, modelInfo := range a.AvailableModels {
		modelLevel, ok := a.tierLevels[modelInfo.Tier]
		if !ok {
			continue
		}
		if modelLevel <= userLevel {
			result = append(result, AvailableModel{
				ID:   modelInfo.ID,
				Name: modelInfo.Name,
			})
		}
	}
	return result
}

func (a *VolcengineAdapter) isValidModelForTier(modelID, userTier string) bool {
	userLevel, ok := a.tierLevels[userTier]
	if !ok {
		userLevel = 0
	}

	for _, model := range a.AvailableModels {
		if model.ID == modelID {
			modelLevel, ok := a.tierLevels[model.Tier]
			if !ok {
				return false
			}
			return modelLevel <= userLevel
		}
	}
	return false
}

func (a *VolcengineAdapter) ChatStream(req ChatRequest, userTier, modelID string) (<-chan string, <-chan error) {
	responseChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errChan)

		if !a.isValidModelForTier(modelID, userTier) {
			errChan <- errors.New("permission denied for the selected model")
			return
		}

		apiMessages := make([]apiChatMessage, 0, len(req.Messages)+1)
		if req.SystemPrompt != "" {
			apiMessages = append(apiMessages, apiChatMessage{
				Role:    "system",
				Content: req.SystemPrompt,
			})
		}

		for _, msg := range req.Messages {
			apiMessages = append(apiMessages, apiChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		apiRequest := apiChatRequest{
			Model:    modelID,
			Messages: apiMessages,
			Stream:   true,
		}
		requestBody, err := json.Marshal(apiRequest)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request body: %w", err)
			return
		}

		url := fmt.Sprintf("%s/chat/completions", a.baseURL)
		httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			errChan <- fmt.Errorf("failed to create HTTP request: %w", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

		resp, err := a.client.Do(httpReq)
		if err != nil {
			errChan <- fmt.Errorf("failed to send http request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					break
				}

				var streamResp apiStreamResponse
				if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
					continue
				}

				if len(streamResp.Choice) > 0 && streamResp.Choice[0].Delta.Content != "" {
					responseChan <- streamResp.Choice[0].Delta.Content
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading stream response: %w", err)
		}
	}()

	return responseChan, errChan
}
