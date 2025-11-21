package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит конфигурацию Inventory Service
type Config struct {
	// GRPCPort - порт для gRPC сервера
	GRPCPort string

	// MongoDBURL - URL подключения к MongoDB
	MongoDBURL string

	// DatabaseName - имя базы данных MongoDB
	DatabaseName string

	// CollectionName - имя коллекции для stocks
	CollectionName string

	// LogLevel - уровень логирования (debug, info, warn, error)
	LogLevel string
}

// Load загружает конфигурацию из переменных окружения с разумными значениями по умолчанию
func Load() *Config {
	return &Config{
		GRPCPort:       getEnv("GRPC_PORT", "50051"),
		MongoDBURL:     getEnv("MONGODB_URL", "mongodb://admin:password@localhost:27017"),
		DatabaseName:   getEnv("DATABASE_NAME", "inventory_service"),
		CollectionName: getEnv("COLLECTION_NAME", "stocks"),
		LogLevel:       getEnv("LOG_LEVEL", "debug"),
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
	if c.MongoDBURL == "" {
		return fmt.Errorf("MONGODB_URL cannot be empty")
	}
	if c.DatabaseName == "" {
		return fmt.Errorf("DATABASE_NAME cannot be empty")
	}
	if c.CollectionName == "" {
		return fmt.Errorf("COLLECTION_NAME cannot be empty")
	}
	return nil
}
