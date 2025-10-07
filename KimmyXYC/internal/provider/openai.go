package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAIProvider implements LLMProvider using OpenAI-compatible Chat Completions API.
// It supports custom endpoint and token via environment variables:
//   OPENAI_API_KEY  - required to enable this provider
//   OPENAI_API_BASE - optional, defaults to https://api.openai.com
// The API path used is {BASE}/v1/chat/completions with stream=true.
// The "model" passed from caller is forwarded as-is.

type OpenAIProvider struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func NewOpenAIProviderFromEnv() *OpenAIProvider {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil
	}
	base := os.Getenv("OPENAI_API_BASE")
	if base == "" {
		base = "https://api.openai.com"
	}
	return &OpenAIProvider{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey:  key,
		Client:  &http.Client{Timeout: 90 * time.Second},
	}
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIStreamChunk struct {
	ID      string                    `json:"id"`
	Object  string                    `json:"object"`
	Created int64                     `json:"created"`
	Model   string                    `json:"model"`
	Choices []openAIStreamChunkChoice `json:"choices"`
}

type openAIStreamChunkChoice struct {
	Index int               `json:"index"`
	Delta openAIStreamDelta `json:"delta"`
	// finish_reason may be "stop" etc.
	FinishReason *string `json:"finish_reason"`
}

type openAIStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ChatCompletionStream implements streaming chat using OpenAI SSE.
func (p *OpenAIProvider) ChatCompletionStream(ctx context.Context, model string, messages []ChatMessage) (<-chan StreamChunk, error) {
	if p == nil || p.APIKey == "" {
		return nil, errors.New("openai provider not configured")
	}
	url := p.BaseURL + "/v1/chat/completions"

	reqPayload := openAIChatRequest{
		Model:  model,
		Stream: true,
	}
	for _, m := range messages {
		role := strings.ToLower(m.Role)
		if role == "assistant" || role == "user" || role == "system" {
			// ok
		} else {
			// map unknown roles to user to avoid API errors
			role = "user"
		}
		reqPayload.Messages = append(reqPayload.Messages, openAIChatMessage{Role: role, Content: m.Content})
	}
	buf, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, errors.New(strings.TrimSpace(string(b)))
	}

	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		r := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				select {
				case ch <- StreamChunk{Err: ctx.Err()}:
				default:
				}
				return
			default:
			}
			line, err := r.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					// end of stream
					ch <- StreamChunk{Done: true}
				}
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" || strings.HasPrefix(line, ":") { // comments/keepalive
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				return
			}
			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// send as raw text if JSON parse fails
				ch <- StreamChunk{Content: data}
				continue
			}
			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					ch <- StreamChunk{Content: choice.Delta.Content}
				}
				if choice.FinishReason != nil && *choice.FinishReason != "" {
					// when finish reason received, mark done soon
					// we won't break immediately because there could be other choices
				}
			}
		}
	}()
	return ch, nil
}
