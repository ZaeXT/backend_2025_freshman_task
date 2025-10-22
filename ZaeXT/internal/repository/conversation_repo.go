package repository

import (
	"ai-qa-backend/internal/model"
	"time"

	"gorm.io/gorm"
)

type ConversationRepository interface {
	Create(conv *model.Conversation) error
	GetByID(id, userID uint) (*model.Conversation, error)
	ListByUserID(userID uint) ([]*model.Conversation, error)
	Update(conv *model.Conversation) error
	DeleteByID(id, userID uint) error
	ListDeletedByUserID(userID uint) ([]*model.Conversation, error)
	RestoreByID(id, userID uint) error
	PermanentDeleteByID(id, userID uint) error
	PermanentDeleteBefore(cutoff time.Time) (int64, error)
}

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(conv *model.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *conversationRepository) GetByID(id, userID uint) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.Where("id = ? AND user_id = ?", id, userID).Preload("Messages").First(&conv).Error
	return &conv, err
}

func (r *conversationRepository) ListByUserID(userID uint) ([]*model.Conversation, error) {
	var conversations []*model.Conversation
	err := r.db.Where("user_id = ?", userID).Order("updated_at desc").Find(&conversations).Error
	return conversations, err
}

func (r *conversationRepository) Update(conv *model.Conversation) error {
	return r.db.Save(conv).Error
}

func (r *conversationRepository) DeleteByID(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Conversation{}).Error
}

func (r *conversationRepository) ListDeletedByUserID(userID uint) ([]*model.Conversation, error) {
	var conversations []*model.Conversation
	err := r.db.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", userID).Find(&conversations).Error
	return conversations, err
}

func (r *conversationRepository) RestoreByID(id, userID uint) error {
	return r.db.Unscoped().Model(&model.Conversation{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("deleted_at", nil).Error
}

func (r *conversationRepository) PermanentDeleteByID(id, userID uint) error {
	tx := r.db.Begin()
	if err := tx.Where("conversation_id = ?", id).Delete(&model.Message{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Where("id = ? AND user_id = ?", id, userID).Delete(&model.Conversation{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *conversationRepository) PermanentDeleteBefore(cutoff time.Time) (int64, error) {
	var conversationsToDelete []model.Conversation
	var idsToDelete []uint

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
			Find(&conversationsToDelete).Error; err != nil {
			return err
		}

		if len(conversationsToDelete) == 0 {
			return nil
		}

		for _, conv := range conversationsToDelete {
			idsToDelete = append(idsToDelete, conv.ID)
		}

		if err := tx.Where("conversation_id IN ?", idsToDelete).Delete(&model.Message{}).Error; err != nil {
			return err
		}

		result := tx.Unscoped().Where("id IN ?", idsToDelete).Delete(&model.Conversation{})
		if result.Error != nil {
			return result.Error
		}
		return nil
	})

	return int64(len(idsToDelete)), err
}
