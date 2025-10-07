package dao

import (
	"ai-qa-system/models"
	"fmt"
	"strings"
)

func GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	if err := DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateUser(user *models.User) error {
	err := DB.Create(user).Error
	if err != nil {
		// 判断是否违反唯一约束（MySQL错误码 1062）
		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("用户名已存在")
		}
		return err
	}
	return nil
}

func UpdateUserQuestionCountAndVIP(userID int64) error {
	user, err := GetUserByID(userID)
	if err != nil {
		return err
	}

	user.QuestionCount++

	if user.QuestionCount >= 2 {
		user.VipLevel = 2
	} else if user.QuestionCount >= 1 {
		user.VipLevel = 1
	} else {
		user.VipLevel = 0
	}

	return DB.Save(&user).Error
}
