package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAvailableModels 获取所有可用的AI模型名称
func GetAvailableModels(c *gin.Context) {
	// 定义可用的模型列表
	models := []string{
		"qwen:7b",
		"deepseek-r1:8b",
		"deepseek-r1:14b",
		“deepseek-v3.2”
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"count":  len(models),
	})
}
