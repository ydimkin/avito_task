package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

// Config конфигурация сервиса-эмулятора заведения.
type Config struct {
	Port           string
	CoreServiceURL string
	RestaurantID   string
	LogLevel       string
	LogFormat      string
}

// LoadConfig загружает конфигурацию из .env и переменных окружения через Viper.
func LoadConfig() (*Config, error) {
	vip := viper.New()

	vip.SetConfigFile(".env")
	vip.SetConfigType("env")

	if err := vip.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) && !os.IsNotExist(err) {
			return nil, err
		}
	}

	vip.AutomaticEnv()

	vip.SetDefault("PORT", "8081")
	vip.SetDefault("CORE_SERVICE_URL", "http://localhost:8080")
	vip.SetDefault("RESTAURANT_ID", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	vip.SetDefault("LOG_LEVEL", "info")
	vip.SetDefault("LOG_FORMAT", "json")

	return &Config{
		Port:           vip.GetString("PORT"),
		CoreServiceURL: vip.GetString("CORE_SERVICE_URL"),
		RestaurantID:   vip.GetString("RESTAURANT_ID"),
		LogLevel:       vip.GetString("LOG_LEVEL"),
		LogFormat:      vip.GetString("LOG_FORMAT"),
	}, nil
}
