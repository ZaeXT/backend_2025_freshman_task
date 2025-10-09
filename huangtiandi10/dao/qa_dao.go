package dao

import (
	"ai-qa-system/models"
	"fmt"
)

func SaveQARecord(userID int64, question, answer, model string, questionCount int) error {
	qa := models.QARecord{
		UserID:        userID,
		Question:      question,
		Answer:        answer,
		ModelUsed:     model,
		QuestionCount: questionCount,
	}
	return DB.Create(&qa).Error
}

func GetQARecordsByUser(userID int64) ([]models.QARecord, error) {
	var records []models.QARecord
	if err := DB.Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func DeleteQARecord(id int64) error {
	result := DB.Delete(&models.QARecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("记录不存在")
	}
	return nil
}

func ClearQARecords(userID int64) error {
	return DB.Where("user_id = ?", userID).Delete(&models.QARecord{}).Error
}
