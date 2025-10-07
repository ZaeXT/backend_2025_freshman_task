package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"AIBackend/internal/models"
	"AIBackend/internal/provider"
)

type ChatService struct {
	DB       *gorm.DB
	LLM      provider.LLMProvider
	MaxTurns int // number of last messages to keep in context
}

func NewChatService(db *gorm.DB, llm provider.LLMProvider) *ChatService {
	return &ChatService{DB: db, LLM: llm, MaxTurns: 10}
}

// EnsureConversation ensures conversation exists (and belongs to user), creating if needed.
func (s *ChatService) EnsureConversation(userID uint, convID uint, model string, title string) (*models.Conversation, error) {
	if convID != 0 {
		var conv models.Conversation
		if err := s.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
			return nil, err
		}
		return &conv, nil
	}
	conv := &models.Conversation{UserID: userID, Title: title, Model: model}
	if conv.Title == "" {
		conv.Title = "New Chat"
	}
	if err := s.DB.Create(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

// ListConversations returns user's conversations.
func (s *ChatService) ListConversations(userID uint) ([]models.Conversation, error) {
	var convs []models.Conversation
	if err := s.DB.Where("user_id = ?", userID).Order("updated_at desc").Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

// GetMessages returns messages for a conversation if owned by user.
func (s *ChatService) GetMessages(userID, convID uint) ([]models.Message, error) {
	var conv models.Conversation
	if err := s.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		return nil, err
	}
	var msgs []models.Message
	if err := s.DB.Where("conversation_id = ?", convID).Order("id asc").Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

// SendMessage adds a user message, streams assistant reply via callback, and persists the assistant message.
func (s *ChatService) SendMessage(ctx context.Context, userID uint, convID uint, model string, userText string, stream func(chunk string) error) (uint, string, error) {
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return 0, "", errors.New("message content required")
	}
	conv, err := s.EnsureConversation(userID, convID, model, "")
	if err != nil {
		return 0, "", err
	}
	// Save user message
	um := &models.Message{ConversationID: conv.ID, Role: "user", Content: userText}
	if err := s.DB.Create(um).Error; err != nil {
		return 0, "", err
	}
	// Load recent messages for context
	var msgs []models.Message
	if err := s.DB.Where("conversation_id = ?", conv.ID).
		Order("id desc").
		Limit(s.MaxTurns * 2).
		Find(&msgs).Error; err != nil {
		return conv.ID, "", err
	}
	// reverse to chronological
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	llmMsgs := make([]provider.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		llmMsgs = append(llmMsgs, provider.ChatMessage{Role: m.Role, Content: m.Content})
	}
	// Stream assistant reply
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	ch, err := s.LLM.ChatCompletionStream(ctx, conv.Model, llmMsgs)
	if err != nil {
		return 0, "", err
	}
	assistantContent := strings.Builder{}
	for chunk := range ch {
		if chunk.Err != nil {
			return conv.ID, "", chunk.Err
		}
		if chunk.Content != "" {
			assistantContent.WriteString(chunk.Content)
			if stream != nil {
				if err := stream(chunk.Content); err != nil {
					return conv.ID, "", err
				}
			}
		}
		if chunk.Done {
			break
		}
	}
	// Save assistant message
	am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: assistantContent.String()}
	if err := s.DB.Create(am).Error; err != nil {
		return conv.ID, "", err
	}
	if err := s.DB.Model(conv).UpdateColumn("updated_at", time.Now()).Error; err != nil {
		return conv.ID, "", err
	}
	return conv.ID, am.Content, nil
}
