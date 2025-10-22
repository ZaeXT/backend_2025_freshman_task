package service

import (
	"ai-qa-backend/internal/repository"
)

type Service struct {
	User       UserService
	Category   CategoryService
	Chat       ChatService
	RecycleBin RecycleBinService
}

func NewService(repo *repository.Repository, aiAdapter AIAdapter) *Service {
	userService := NewUserService(repo.User, repo.Category)
	return &Service{
		User:       userService,
		Category:   NewCategoryService(repo.Category),
		Chat:       NewChatService(repo.Conversation, repo.Message, repo.User, repo.Category, aiAdapter),
		RecycleBin: NewRecycleBinService(repo.Conversation),
	}
}
