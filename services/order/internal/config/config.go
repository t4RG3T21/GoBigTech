package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит конфигурацию Order Service
type Config struct {
	// HTTPPort - порт для HTTP сервера
	HTTPPort string

	// DatabaseURL - URL подключения к PostgreSQL
	DatabaseURL string

	// InventoryGRPCAddress - адрес gRPC сервера Inventory Service
	InventoryGRPCAddress string

	// PaymentGRPCAddress - адрес gRPC сервера Payment Service
	PaymentGRPCAddress string

	// KafkaBootstrapServers - адреса Kafka брокеров (через запятую)
	KafkaBootstrapServers string

	// PaymentTopic - топик для отправки событий оплаты
	PaymentTopic string

	// AssemblyTopic - топик для получения событий сборки
	AssemblyTopic string

	// ConsumerGroupID - ID группы потребителей Kafka
	ConsumerGroupID string

	// LogLevel - уровень логирования (debug, info, warn, error)
	LogLevel string
}

// Load загружает конфигурацию из переменных окружения с разумными значениями по умолчанию
func Load() *Config {
	return &Config{
		HTTPPort:              getEnv("HTTP_PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/order_service?sslmode=disable"),
		InventoryGRPCAddress:  getEnv("INVENTORY_GRPC_ADDRESS", "localhost:50051"),
		PaymentGRPCAddress:    getEnv("PAYMENT_GRPC_ADDRESS", "localhost:50052"),
		KafkaBootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		PaymentTopic:          getEnv("KAFKA_PAYMENT_TOPIC", "orders.payment"),
		AssemblyTopic:         getEnv("KAFKA_ASSEMBLY_TOPIC", "orders.assembly"),
		ConsumerGroupID:       getEnv("KAFKA_CONSUMER_GROUP_ID", "order-service"),
		LogLevel:              getEnv("LOG_LEVEL", "debug"),
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
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT cannot be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL cannot be empty")
	}
	if c.InventoryGRPCAddress == "" {
		return fmt.Errorf("INVENTORY_GRPC_ADDRESS cannot be empty")
	}
	if c.PaymentGRPCAddress == "" {
		return fmt.Errorf("PAYMENT_GRPC_ADDRESS cannot be empty")
	}
	if c.KafkaBootstrapServers == "" {
		return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS cannot be empty")
	}
	if c.PaymentTopic == "" {
		return fmt.Errorf("KAFKA_PAYMENT_TOPIC cannot be empty")
	}
	if c.AssemblyTopic == "" {
		return fmt.Errorf("KAFKA_ASSEMBLY_TOPIC cannot be empty")
	}
	if c.ConsumerGroupID == "" {
		return fmt.Errorf("KAFKA_CONSUMER_GROUP_ID cannot be empty")
	}
	return nil
}
