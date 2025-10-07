package controllers

import (
	"ai-qa-system/dao"
	"ai-qa-system/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AskRequest struct {
	Question string `json:"question"`
}

func AskQuestion(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 获取用户信息
	user, err := dao.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 用户的 VIP 等级（例如 "VIP0", "VIP1", "VIP2"）
	role := user.VipLevel

	// 遍历所有可能的模型，找到用户有权限用的模型
	models := []string{"doubao-seed-1.6-vision", "deepseek-v3.1", "kimi-k2"}
	var chosenModel string
	for _, m := range models {
		ok, _ := utils.Enforcer.Enforce(strconv.Itoa(role), m, "use")
		if ok {
			chosenModel = m
			break
		}
	}

	if chosenModel == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无可用模型"})
		return
	}

	// 调用 AI
	answer, err := utils.CallAI(chosenModel, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 调用失败"})
		return
	}

	// 保存问答记录
	err = dao.SaveQARecord(userID, req.Question, answer, chosenModel, user.QuestionCount)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 更新提问次数 & VIP等级
	_ = dao.UpdateUserQuestionCountAndVIP(userID)

	c.JSON(http.StatusOK, gin.H{
		"question":       req.Question,
		"answer":         answer,
		"model":          chosenModel,
		"question_count": user.QuestionCount + 1, // 返回更新后的次数
	})
}

func GetHistory(c *gin.Context) {
	userID := c.GetInt64("userID")

	records, err := dao.GetQARecordsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
	})
}

func DeleteQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	err = dao.DeleteQARecord(id)
	if err != nil {
		if err.Error() == "记录不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func ClearHistory(c *gin.Context) {
	userID := c.GetInt64("userID")
	if err := dao.ClearQARecords(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "清空成功"})
}
