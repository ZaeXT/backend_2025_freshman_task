package repository

import (
	"ai-qa-backend/internal/model"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *model.Message) error
	CreateBatch(messages []*model.Message) error
	GetByConversationID(convID uint) ([]*model.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) CreateBatch(messages []*model.Message) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, msg := range messages {
			if err := tx.Create(msg).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *messageRepository) GetByConversationID(convID uint) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.Where("conversation_id = ?", convID).Order("created_at asc").Find(&messages).Error
	return messages, err
}
