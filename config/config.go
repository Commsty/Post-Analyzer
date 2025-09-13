package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"app"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"-"`
		SSLMode  string `yaml:"ssl_mode"`
	} `yaml:"database"`

	API struct {
		Telegram struct {
			BotToken    string `yaml:"-"`
			AppID       int    `yaml:"-"`
			AppHash     string `yaml:"-"`
			SessionPath string `yaml:"-"`

			Phone    string `yaml:"-"`
			Password string `yaml:"-"`
		} `yaml:"-"`
		OpenRouter struct {
			APIKey string `yaml:"-"`
		} `yaml:"-"`
	} `yaml:"-"`
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	cfg.API.Telegram.BotToken = os.Getenv("TG_BOT_TOKEN")
	cfg.API.Telegram.AppID, err = strconv.Atoi(os.Getenv("TG_APP_ID"))
	if err != nil {
		return nil, err
	}
	cfg.API.Telegram.AppHash = os.Getenv("TG_APP_HASH")
	cfg.API.Telegram.SessionPath = os.Getenv("SESSION_PATH")
	cfg.API.Telegram.Phone = os.Getenv("AUTH_PHONE")
	cfg.API.Telegram.Password = os.Getenv("AUTH_PASSWORD")
	cfg.API.OpenRouter.APIKey = os.Getenv("OPENROUTER_API_KEY")

	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD не задан в переменных окружения")
	}
	if cfg.API.Telegram.BotToken == "" {
		return nil, fmt.Errorf("TG_BOT_TOKEN не задан в переменных окружения")
	}
	if cfg.API.Telegram.AppID == 0 {
		return nil, fmt.Errorf("TG_APP_ID не задан в переменных окружения")
	}
	if cfg.API.Telegram.AppHash == "" {
		return nil, fmt.Errorf("TG_APP_HASH не задан в переменных окружения")
	}
	if cfg.API.Telegram.SessionPath == "" {
		return nil, fmt.Errorf("SESSION_PATH не задан в переменных окружения")
	}
	if cfg.API.Telegram.Phone == "" {
		return nil, fmt.Errorf("AUTH_PHONE не задан в переменных окружения")
	}
	if cfg.API.Telegram.Password == "" {
		return nil, fmt.Errorf("AUTH_PASSWORD не задан в переменных окружения")
	}
	if cfg.API.OpenRouter.APIKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY не задан в переменных окружения")
	}

	return &cfg, nil
}
