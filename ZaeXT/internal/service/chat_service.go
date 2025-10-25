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
	ChatStream(req volcengine.ChatRequest, userTier, modelID string, enableThinking bool) (<-chan []byte, <-chan error)
	GetAvailableModelsForTier(userTier string) []volcengine.AvailableModel
}

type ChatService interface {
	CreateConversation(userID uint, isTemporary bool, categoryID *uint) (*model.Conversation, error)
	GetConversation(convID, userID uint) (*model.Conversation, error)
	ListConversations(userID uint) ([]*model.Conversation, error)
	ProcessUserMessage(convID, userID uint, userTier, message, modelID string, enableThinking bool) (<-chan []byte, <-chan error)
	ListAvailableModels(userTier string) []volcengine.AvailableModel
	UpdateConversationTitle(convID, userID uint, title string) error
	DeleteConversation(convID, userID uint) error
	AutoClassify(convID, userID uint) error
	GetMessagesByConversationID(convID, userID uint) ([]*model.Message, error)
	UpdateConversationCategory(convID, userID uint, newCategoryID *uint) error
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
	userContent := fmt.Sprintf("=== 分类列表 ===\n%s\n\n=== 对话内容 ===\n%s", string(categoriesJSON), conversationContext.String())

	req := volcengine.ChatRequest{
		SystemPrompt: systemPrompt,
		Messages:     []*model.Message{{Role: "user", Content: userContent}},
	}
	freeModels := s.aiAdapter.GetAvailableModelsForTier("free")
	if len(freeModels) == 0 {
		return errors.New("no 'free' models configured for auto-classification")
	}
	modelID := freeModels[0].ID

	resChan, errChan := s.aiAdapter.ChatStream(req, "free", modelID, false)

	var jsonAccumulator strings.Builder
	for chunk := range resChan {
		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(chunk, &streamResp); err == nil {
			if len(streamResp.Choices) > 0 {
				jsonAccumulator.WriteString(streamResp.Choices[0].Delta.Content)
			}
		}
	}

	if err := <-errChan; err != nil {
		return fmt.Errorf("ai call failed: %w", err)
	}

	responseStr := jsonAccumulator.String()
	var result struct {
		CategoryID uint `json:"category_id"`
	}
	if err := json.Unmarshal([]byte(responseStr), &result); err != nil {
		return fmt.Errorf("failed to parse ai response json: %w (raw response: %s)", err, responseStr)
	}

	if _, ok := userCategoryMap[result.CategoryID]; !ok {
		return errors.New("ai returned an invalid or unauthorized category id")
	}

	conv.CategoryID = &result.CategoryID
	return s.convRepo.Update(conv)
}

func (s *chatService) UpdateConversationCategory(convID, userID uint, newCategoryID *uint) error {
	conv, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		return errors.New("conversation not found or permission denied")
	}

	if newCategoryID != nil {
		_, err := s.categoryRepo.GetByID(*newCategoryID, userID)
		if err != nil {
			return errors.New("target category not found or permission denied")
		}
	}

	conv.CategoryID = newCategoryID
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

func (s *chatService) GetMessagesByConversationID(convID, userID uint) ([]*model.Message, error) {
	_, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		return nil, errors.New("conversation not found or permission denied")
	}
	return s.msgRepo.GetByConversationID(convID)
}

func (s *chatService) ProcessUserMessage(convID, userID uint, userTier, message, modelID string, enableThinking bool) (<-chan []byte, <-chan error) {
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

	handlerResponseChan := make(chan []byte)
	handlerErrChan := make(chan error, 1)

	go func() {
		defer close(handlerResponseChan)
		defer close(handlerErrChan)

		systemPrompt := fmt.Sprintf("这是关于 '%s' 的对话。请记住以下用户信息：%s", conv.Title, user.MemoryInfo)
		aiReq := volcengine.ChatRequest{SystemPrompt: systemPrompt, Messages: history}
		adapterResponseChan, adapterErrChan := s.aiAdapter.ChatStream(aiReq, userTier, modelID, enableThinking)

		var dbContentAccumulator strings.Builder
		var streamErr error

		for {
			select {
			case chunk, ok := <-adapterResponseChan:
				if !ok {
					adapterResponseChan = nil
				} else {
					handlerResponseChan <- chunk
					var streamResp struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}
					if err := json.Unmarshal(chunk, &streamResp); err == nil {
						if len(streamResp.Choices) > 0 {
							dbContentAccumulator.WriteString(streamResp.Choices[0].Delta.Content)
						}
					}
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
		if streamErr == nil && dbContentAccumulator.Len() > 0 {
			assistantMsg := &model.Message{
				ConversationID: conv.ID,
				Role:           "assistant",
				Content:        dbContentAccumulator.String(),
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
		return errors.New("conversation not found or permission denied")
	}
	if strings.TrimSpace(title) == "" {
		history, err := s.msgRepo.GetByConversationID(convID)
		if err != nil || len(history) == 0 {
			conv.Title = "New Chat"
			conv.IsTitleUserModified = false
			return s.convRepo.Update(conv)
		}

		conv.IsTitleUserModified = false
		if err := s.convRepo.Update(conv); err != nil {
			return err
		}

		go s.autoGenerateTitle(conv, history)

		return nil

	} else {
		conv.Title = title
		conv.IsTitleUserModified = true
		return s.convRepo.Update(conv)
	}
}

func (s *chatService) DeleteConversation(convID, userID uint) error {
	_, err := s.convRepo.GetByID(convID, userID)
	if err != nil {
		return errors.New("permission denied or conversation not found")
	}

	return s.convRepo.DeleteByID(convID, userID)
}

func (s *chatService) autoGenerateTitle(conv *model.Conversation, history []*model.Message) {
	log.Printf("INFO: Starting auto title generation for conv %d", conv.ID)
	if len(history) == 0 {
		return
	}
	titlePrompt := "你是一个对话标题生成助手。根据用户和助手的对话内容，生成一个简短、精确、不超过10个字的摘要作为标题。你的回答必须只包含标题本身，不要任何额外的解释、引言或标点符号。"
	messagesForTitle := history
	finalInstruction := &model.Message{
		Role:    "user",
		Content: "根据以上对话，生成一个简洁的标题。",
	}
	messagesForTitle = append(messagesForTitle, finalInstruction)
	req := volcengine.ChatRequest{
		SystemPrompt: titlePrompt,
		Messages:     messagesForTitle,
	}

	freeModels := s.aiAdapter.GetAvailableModelsForTier("free")
	if len(freeModels) == 0 {
		log.Printf("ERROR: No 'free' tier models available for auto title generation.")
		return
	}
	modelID := freeModels[0].ID
	resChan, errChan := s.aiAdapter.ChatStream(req, "free", modelID, false)
	var titleAccumulator strings.Builder
	for chunk := range resChan {
		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(chunk, &streamResp); err == nil {
			if len(streamResp.Choices) > 0 {
				titleAccumulator.WriteString(streamResp.Choices[0].Delta.Content)
			}
		}
	}
	if err := <-errChan; err != nil {
		log.Printf("ERROR: AI call for auto title generation failed for conv %d: %v", conv.ID, err)
		return
	}
	title := strings.Trim(titleAccumulator.String(), "\"“” \n\r")
	if title != "" {
		latestConv, err := s.convRepo.GetByID(conv.ID, conv.UserID)
		if err == nil && !latestConv.IsTitleUserModified {
			latestConv.Title = title
			if err := s.convRepo.Update(latestConv); err != nil {
				log.Printf("ERROR: Failed to update auto-generated title for conv %d: %v", conv.ID, err)
			} else {
				log.Printf("INFO: Auto-generated title '%s' for conv %d", title, conv.ID)
			}
		}
	} else {
		log.Printf("INFO: Auto-generated title is empty for conv %d", conv.ID)
	}
}
