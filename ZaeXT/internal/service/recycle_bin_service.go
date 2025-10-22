package service

import (
	"ai-qa-backend/internal/configs"
	"ai-qa-backend/internal/model"
	"ai-qa-backend/internal/repository"
	"fmt"
	"time"
)

type RecycleBinService interface {
	List(userID uint) ([]*model.Conversation, error)
	Restore(convID, userID uint) error
	PermanentDelete(convID, userID uint) error
	CleanupExpired() (int64, error)
}

type recycleBinService struct {
	convRepo repository.ConversationRepository
}

func NewRecycleBinService(convRepo repository.ConversationRepository) RecycleBinService {
	return &recycleBinService{convRepo: convRepo}
}

func (s *recycleBinService) List(userID uint) ([]*model.Conversation, error) {
	return s.convRepo.ListDeletedByUserID(userID)
}

func (s *recycleBinService) Restore(convID, userID uint) error {
	return s.convRepo.RestoreByID(convID, userID)
}

func (s *recycleBinService) PermanentDelete(convID, userID uint) error {
	return s.convRepo.PermanentDeleteByID(convID, userID)
}

func (s *recycleBinService) CleanupExpired() (int64, error) {
	retentionDays := configs.Conf.RecycleBin.RetentionDays
	if retentionDays <= 0 {
		return 0, nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	deletedCount, err := s.convRepo.PermanentDeleteBefore(cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired conversations: %w", err)
	}

	return deletedCount, nil
}
