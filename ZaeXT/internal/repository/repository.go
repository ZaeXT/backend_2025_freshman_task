package repository

import "gorm.io/gorm"

type Repository struct {
	User         UserRepository
	Conversation ConversationRepository
	Message      MessageRepository
	Category     CategoryRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		User:         NewUserRepository(db),
		Conversation: NewConversationRepository(db),
		Message:      NewMessageRepository(db),
		Category:     NewCategoryRepository(db),
	}
}
