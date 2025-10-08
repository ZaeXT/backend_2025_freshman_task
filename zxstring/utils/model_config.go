package utils

import (
	"encoding/json"
	"os"
)

// ModelConfig 代表models_config.json中的单个模型配置
type ModelConfig struct {
	ModelFullName string `json:"model_full_name"`
	APIKeyStore   string `json:"API_KEY_STORE"`
	BaseURL       string `json:"base_url"`
	Permission    int    `json:"permission"`
}

// ModelsConfig 代表整个models_config.json文件的结构
type ModelsConfig map[string]ModelConfig

// LoadModelsConfig 从文件加载模型配置
func LoadModelsConfig(filename string) (ModelsConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config ModelsConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetModelsByPermission 根据权限获取模型键列表
func (config ModelsConfig) GetModelsByPermission(permission int) []string {
	var models []string
	for key, modelConfig := range config {
		if modelConfig.Permission <= permission {
			models = append(models, key)
		}
	}
	return models
}
