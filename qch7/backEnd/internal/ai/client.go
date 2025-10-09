package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"backEnd/internal/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type ChoiceDelta struct {
	Content string `json:"content"`
}

type StreamChoice struct {
	Delta ChoiceDelta `json:"delta"`
}

type StreamChunk struct {
	Choices []StreamChoice `json:"choices"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

// Client is a simple OpenAI-compatible HTTP client targeting DashScope compatible endpoint.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
}

func NewClient() *Client {
	cfg := config.Get()
	return &Client{
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
		baseURL:    cfg.AIBaseURL,
		apiKey:     cfg.AIAPIKey,
		model:      cfg.AIModel,
		timeout:    cfg.RequestTimeout,
	}
}

func (c *Client) Model() string { return c.model }

// Chat sends a non-streaming chat completion request and returns the assistant's full reply content.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{Model: c.model, Messages: messages, Stream: false}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chat request failed: %s - %s", resp.Status, string(b))
	}
	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", nil
	}
	return out.Choices[0].Message.Content, nil
}

// ChatStream streams tokens and calls onData for each text delta.
func (c *Client) ChatStream(ctx context.Context, messages []Message, onData func(delta string) error) error {
	reqBody := ChatRequest{Model: c.model, Messages: messages, Stream: true}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chat stream failed: %s - %s", resp.Status, string(b))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			// SSE-like: lines start with "data: {json}"
			if bytes.HasPrefix(line, []byte("data:")) {
				payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
				if bytes.Equal(payload, []byte("[DONE]")) {
					break
				}
				var chunk StreamChunk
				if err := json.Unmarshal(payload, &chunk); err == nil {
					for _, choice := range chunk.Choices {
						if choice.Delta.Content != "" {
							if err := onData(choice.Delta.Content); err != nil {
								return err
							}
						}
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}
