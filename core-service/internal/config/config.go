package config

import (
	"errors"
	"fmt"
	"os"

	"core-service/internal/domain"
	"github.com/spf13/viper"
)

// Config содержит параметры конфигурации ядра сервиса
type Config struct {
	Port              string
	DBHost            string
	DBName            string
	DBUser            string
	DBPassword        string
	DBPort            string
	DBSSLMode         string
	WebhookTimeoutSec int
	LogLevel          string
	LogFormat         string
}

// LoadConfig загружает конфигурацию из переменных окружения и .env файла
func LoadConfig() (*Config, error) {

	vip := viper.New()

	// Просто на всякий случай, если .env файл есть
	vip.SetConfigFile(".env")
	vip.SetConfigType("env")

	if err := vip.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) && !os.IsNotExist(err) {
			return nil, domain.ErrNotFoundConfig
		}
	}

	// Если файла не было
	vip.AutomaticEnv()

	// Значения по умолчанию, если не были переданы значения в Docker-compose или .env
	vip.SetDefault("PORT", "8080")
	vip.SetDefault("DB_HOST", "localhost")
	vip.SetDefault("DB_PORT", "5432")
	vip.SetDefault("DB_NAME", "avito_kitchen")
	vip.SetDefault("DB_USER", "postgres")
	vip.SetDefault("DB_PASSWORD", "postgres")
	vip.SetDefault("DB_SSLMODE", "disable")
	vip.SetDefault("WEBHOOK_TIMEOUT_SEC", 5)
	vip.SetDefault("LOG_LEVEL", "info")
	vip.SetDefault("LOG_FORMAT", "json")

	return &Config{
		Port:              vip.GetString("PORT"),
		DBHost:            vip.GetString("DB_HOST"),
		DBName:            vip.GetString("DB_NAME"),
		DBUser:            vip.GetString("DB_USER"),
		DBPassword:        vip.GetString("DB_PASSWORD"),
		DBPort:            vip.GetString("DB_PORT"),
		DBSSLMode:         vip.GetString("DB_SSLMODE"),
		WebhookTimeoutSec: vip.GetInt("WEBHOOK_TIMEOUT_SEC"),
		LogLevel:          vip.GetString("LOG_LEVEL"),
		LogFormat:         vip.GetString("LOG_FORMAT"),
	}, nil
}

// DSN возвращает строку подключения к базе данных PostgreSQL.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}
