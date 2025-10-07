package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Name      string `yaml:"name"`
	Charset   string `yaml:"charset"`
	ParseTime bool   `yaml:"parse_time"`
	Loc       string `yaml:"loc"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire,omitempty"`
}

type AIModel struct {
	Key      string `yaml:"key"`
	Endpoint string `yaml:"endpoint"` // 这里的 endpoint 实际就是 model id（ep-xxxx）
}

type AIConfig struct {
	BaseURL string `yaml:"base_url"`
}

type Config struct {
	Database      DBConfig           `yaml:"database"`
	JWT           JWTConfig          `yaml:"jwt"`
	AI            AIConfig           `yaml:"ai"`
	AIModels      map[string]AIModel `yaml:"ai_models"`
	VIPThresholds []int              `yaml:"vip_thresholds"`
}

var AppConfig Config

func LoadConfig(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("LoadConfig read file failed: %v", err)
	}
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		log.Fatalf("LoadConfig yaml unmarshal failed: %v", err)
	}
}
