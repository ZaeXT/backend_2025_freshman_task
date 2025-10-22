package service

import (
	"ai-qa-backend/internal/adapter/volcengine"
	"ai-qa-backend/internal/model"
	"ai-qa-backend/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type AIAdapter interface {
	ChatStream(req volcengine.ChatRequest, userTier, modelID string, enableThinking bool) (<-chan string, <-chan error)
	GetAvailableModelsForTier(userTier string) []volcengine.AvailableModel
}

type ChatService interface {
	CreateConversation(userID uint, isTemporary bool, categoryID *uint) (*model.Conversation, error)
	GetConversation(convID, userID uint) (*model.Conversation, error)
	ListConversations(userID uint) ([]*model.Conversation, error)
	ProcessUserMessage(convID, userID uint, userTier, message, modelID string, enableThinking bool) (<-chan string, <-chan error)
	ListAvailableModels(userTier string) []volcengine.AvailableModel
	UpdateConversationTitle(convID, userID uint, title string) error
	DeleteConversation(convID, userID uint) error
	AutoClassify(convID, userID uint) error
}

type chatService struct {
	convRepo     repository.ConversationRepository
	msgRepo      repository.MessageRepository
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	aiAdapter    AIAdapter
}

func NewChatService(
	convRepo repository.ConversationRepository,
	msgRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
	aiAdapter AIAdapter,
) ChatService {
	return &chatService{
		convRepo:     convRepo,
		msgRepo:      msgRepo,
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		aiAdapter:    aiAdapter,
	}
}

func (s *chatService) AutoClassify(convID, userID uint) error {
	conv, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		return errors.New("conversation not found or permission denied")
	}

	messages, err := s.msgRepo.GetByConversationID(convID)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return errors.New("cannot classify an empty conversation")
	}

	categories, err := s.categoryRepo.ListByUserID(userID)
	if err != nil {
		return err
	}
	if len(categories) == 0 {
		return errors.New("no categories available for classification")
	}

	type categoryOption struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	categoryOptions := make([]categoryOption, 0)
	userCategoryMap := make(map[uint]bool)
	for _, cat := range categories {
		categoryOptions = append(categoryOptions, categoryOption{ID: cat.ID, Name: cat.Name})
		userCategoryMap[cat.ID] = true
	}
	categoriesJSON, _ := json.Marshal(categoryOptions)

	var conversationContext strings.Builder
	for _, msg := range messages {
		conversationContext.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	systemPrompt := `你是一个对话分类助手。你的任务是根据下面提供的对话内容，从给定的分类列表中选择一个最匹配的分类。
你必须遵循以下规则：
1. 仔细阅读对话内容和分类列表。
2. 你的回答必须是一个JSON对象。
3. JSON对象中必须只包含一个键 "category_id"，其值是你选择的分类的数字ID。
例如: {"category_id": 123}`
	userPrompt := fmt.Sprintf("=== 分类列表 ===\n%s\n\n=== 对话内容 ===\n%s", string(categoriesJSON), conversationContext.String())

	req := volcengine.ChatRequest{
		SystemPrompt: systemPrompt,
		Messages: []*model.Message{
			{Role: "user", Content: userPrompt},
		},
	}
	freeModels := s.aiAdapter.GetAvailableModelsForTier("free")
	modelID := freeModels[0].ID
	resChan, errChan := s.aiAdapter.ChatStream(req, "free", modelID, false)

	var fullResponse strings.Builder
	for chunk := range resChan {
		fullResponse.WriteString(chunk)
	}
	if err := <-errChan; err != nil {
		return fmt.Errorf("ai call failed: %w", err)
	}

	responseStr := fullResponse.String()
	jsonStart := strings.Index(responseStr, "{")
	jsonEnd := strings.LastIndex(responseStr, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return errors.New("ai response was not a valid json")
	}
	jsonStr := responseStr[jsonStart : jsonEnd+1]

	var result struct {
		CategoryID uint `json:"category_id"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return fmt.Errorf("failed to parse ai response json: %w", err)
	}

	if _, ok := userCategoryMap[result.CategoryID]; !ok {
		return errors.New("ai returned an invalid or unauthorized category id")
	}

	conv.CategoryID = &result.CategoryID
	return s.convRepo.Update(conv)
}

func (s *chatService) ListAvailableModels(userTier string) []volcengine.AvailableModel {
	return s.aiAdapter.GetAvailableModelsForTier(userTier)
}

func (s *chatService) CreateConversation(userID uint, isTemporary bool, categoryID *uint) (*model.Conversation, error) {
	conv := &model.Conversation{
		UserID:      userID,
		IsTemporary: isTemporary,
		CategoryID:  categoryID,
	}
	err := s.convRepo.Create(conv)
	return conv, err
}

func (s *chatService) GetConversation(convID, userID uint) (*model.Conversation, error) {
	conv, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		return nil, err
	}
	if conv.UserID != userID {
		return nil, errors.New("permission denied")
	}
	conv.Messages, err = s.msgRepo.GetByConversationID(convID)
	return conv, err
}

func (s *chatService) ListConversations(userID uint) ([]*model.Conversation, error) {
	return s.convRepo.ListByUserID(userID)
}

func (s *chatService) ProcessUserMessage(convID, userID uint, userTier, message, modelID string, enableThinking bool) (<-chan string, <-chan error) {
	conv, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		errChan := make(chan error, 1)
		errChan <- errors.New("conversation not found or permission denied")
		close(errChan)
		return nil, errChan
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		errChan := make(chan error, 1)
		errChan <- err
		close(errChan)
		return nil, errChan
	}
	userMsg := &model.Message{ConversationID: convID, Role: "user", Content: message}
	if err := s.msgRepo.Create(userMsg); err != nil {
		errChan := make(chan error, 1)
		errChan <- err
		close(errChan)
		return nil, errChan
	}
	history, err := s.msgRepo.GetByConversationID(convID)
	if err != nil {
		errChan := make(chan error, 1)
		errChan <- err
		close(errChan)
		return nil, errChan
	}
	if modelID == "" {
		availableModels := s.aiAdapter.GetAvailableModelsForTier(userTier)
		if len(availableModels) == 0 {
			errChan := make(chan error, 1)
			errChan <- errors.New("no available models for your tier")
			close(errChan)
			return nil, errChan
		}
		modelID = availableModels[0].ID
	}

	handlerResponseChan := make(chan string)
	handlerErrChan := make(chan error, 1)

	go func() {
		defer close(handlerResponseChan)
		defer close(handlerErrChan)

		systemPrompt := fmt.Sprintf("这是关于 '%s' 的对话。请记住以下用户信息：%s", conv.Title, user.MemoryInfo)
		aiReq := volcengine.ChatRequest{SystemPrompt: systemPrompt, Messages: history}
		adapterResponseChan, adapterErrChan := s.aiAdapter.ChatStream(aiReq, userTier, modelID, enableThinking)

		var fullResponse strings.Builder
		var streamErr error

		for {
			select {
			case chunk, ok := <-adapterResponseChan:
				if !ok {
					adapterResponseChan = nil
				} else {
					handlerResponseChan <- chunk
					fullResponse.WriteString(chunk)
				}
			case err, ok := <-adapterErrChan:
				if !ok {
					adapterErrChan = nil
				} else if err != nil {
					handlerErrChan <- err
					streamErr = err
				}
			}
			if adapterResponseChan == nil && adapterErrChan == nil {
				break
			}
		}
		if streamErr == nil && fullResponse.Len() > 0 {
			assistantMsg := &model.Message{
				ConversationID: conv.ID,
				Role:           "assistant",
				Content:        fullResponse.String(),
			}
			if err := s.msgRepo.Create(assistantMsg); err != nil {
				log.Printf("ERROR: Failed to save assistant message for conv %d: %v", conv.ID, err)
				handlerErrChan <- err
			}

			if !conv.IsTitleUserModified {
				log.Printf("INFO: Triggering auto title generation for conv %d", conv.ID)
				fullHistory := append(history, assistantMsg)
				go s.autoGenerateTitle(conv, fullHistory)
			}
		}

	}()
	return handlerResponseChan, handlerErrChan
}

func (s *chatService) UpdateConversationTitle(convID, userID uint, title string) error {
	conv, err := s.GetConversation(convID, userID)
	if err != nil {
		return err
	}
	conv.Title = title
	conv.IsTitleUserModified = true
	return s.convRepo.Update(conv)
}

func (s *chatService) DeleteConversation(convID, userID uint) error {
	_, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		return errors.New("permission denied or conversation not found")
	}

	return s.convRepo.DeleteByID(convID, userID)
}

func (s *chatService) autoGenerateTitle(conv *model.Conversation, history []*model.Message) {
	var context strings.Builder
	for _, msg := range history {
		context.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	titlePrompt := "根据以上对话，生成一个不超过10个字的简洁标题。只返回标题本身，不要任何多余的文字或标点。"

	req := volcengine.ChatRequest{
		SystemPrompt: titlePrompt,
		Messages: []*model.Message{
			{Role: "user", Content: context.String()},
		},
	}

	freeModels := s.aiAdapter.GetAvailableModelsForTier("free")
	modelID := freeModels[0].ID
	resChan, errChan := s.aiAdapter.ChatStream(req, "free", modelID, false)

	select {
	case title := <-resChan:
		if title != "" {
			conv.Title = strings.Trim(title, "\"“” ")
			_ = s.convRepo.Update(conv)
		}
	case <-errChan:

	}
}
