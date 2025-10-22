package configs

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var Conf *Config

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	VolcEngine VolcEngineConfig `mapstructure:"volcengine"`
	Log        LogConfig        `mapstructure:"log"`
	RecycleBin RecycleBinConfig `mapstructure:"recycle_bin"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type ModelInfo struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	Tier string `mapstructure:"tier"`
}

type VolcEngineConfig struct {
	APIKey          string      `mapstructure:"api_key"`
	BaseURL         string      `mapstructure:"base_url"`
	AvailableModels []ModelInfo `mapstructure:"available_models"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RecycleBinConfig struct {
	RetentionDays int `mapstructure:"retention_days"`
}

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs/")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found: %w", err)
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if Conf.JWT.Secret == "" || Conf.VolcEngine.APIKey == "" {
		return fmt.Errorf("JWT secret 或 VolcEngine API key 未设置")
	}
	return nil
}
