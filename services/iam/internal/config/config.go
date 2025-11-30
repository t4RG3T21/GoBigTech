package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит конфигурацию IAM Service
type Config struct {
	// GRPCPort - порт для gRPC сервера
	GRPCPort string

	// HTTPPort - порт для HTTP Gateway сервера
	HTTPPort string

	// DatabaseURL - URL подключения к PostgreSQL
	DatabaseURL string

	// RedisAddress - адрес Redis сервера
	RedisAddress string

	// RedisPassword - пароль для Redis
	RedisPassword string

	// RedisDB - номер базы данных Redis
	RedisDB int

	// SessionTTL - время жизни сессии
	SessionTTL time.Duration

	// LogLevel - уровень логирования (debug, info, warn, error)
	LogLevel string
}

// Load загружает конфигурацию из переменных окружения с разумными значениями по умолчанию
func Load() *Config {
	return &Config{
		GRPCPort:      getEnv("GRPC_PORT", "50053"),
		HTTPPort:      getEnv("HTTP_PORT", "8082"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/iam_service?sslmode=disable"),
		RedisAddress:  getEnv("REDIS_ADDRESS", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "redis_password"),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		SessionTTL:    getEnvAsDuration("SESSION_TTL", 24*time.Hour),
		LogLevel:      getEnv("LOG_LEVEL", "debug"),
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt возвращает значение переменной окружения как int или значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsDuration возвращает значение переменной окружения как time.Duration или значение по умолчанию
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.GRPCPort == "" {
		return fmt.Errorf("GRPC_PORT cannot be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL cannot be empty")
	}
	if c.RedisAddress == "" {
		return fmt.Errorf("REDIS_ADDRESS cannot be empty")
	}
	return nil
}
