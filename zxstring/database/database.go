package database

import (
	"log"
	"webtest/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("userInfo.db"), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	// 自动迁移用户模型和聊天记录模型
	err = database.AutoMigrate(&models.User{}, &models.ChatRecord{})
	if err != nil {
		panic("Failed to migrate database!")
	}

	DB = database
	log.Println("Database connected!")
}

// SaveChatRecord 保存聊天记录到数据库
// 参数: 用户名, 模型名称, 角色, 时间(毫秒), 内容, 对话名
func SaveChatRecord(username, model, role string, time int64, content, cid string) error {
	chatRecord := models.ChatRecord{
		Username: username,
		Model:    model,
		Role:     role,
		Time:     time,
		Content:  content,
		Cid:      cid,
	}

	result := DB.Create(&chatRecord)
	return result.Error
}

// GetChatHistory 获取历史对话记录
// 参数: 用户名, 对话id, 记录数
// 返回: 匹配的聊天记录数组，按时间降序排列
func GetChatHistory(username, cid string, count int) ([]models.ChatRecord, error) {
	var chatRecords []models.ChatRecord

	// 筛选出用户名和对话id都匹配的记录，按时间字段降序取前count条数据
	result := DB.Where("username = ? AND cid = ?", username, cid).
		Order("time DESC").
		Limit(count).
		Find(&chatRecords)

	if result.Error != nil {
		return nil, result.Error
	}

	return chatRecords, nil
}

// GetUserConversationIds 获取用户所有不同的对话ID
// 参数: 用户名
// 返回: 不同的对话ID数组
func GetUserConversationIds(username string) ([]string, error) {
	var cids []string

	// 查询指定用户的所有不同对话ID
	result := DB.Model(&models.ChatRecord{}).
		Where("username = ?", username).
		Distinct("cid").
		Pluck("cid", &cids)

	if result.Error != nil {
		return nil, result.Error
	}

	return cids, nil
}
