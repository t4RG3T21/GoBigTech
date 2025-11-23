package config

import (
	"fmt"
	"os"
)

// Config содержит конфигурацию Assembly Service
type Config struct {
	// KafkaBootstrapServers - адреса Kafka брокеров (через запятую)
	KafkaBootstrapServers string

	// PaymentTopic - топик для получения событий оплаты
	PaymentTopic string

	// AssemblyTopic - топик для отправки событий сборки
	AssemblyTopic string

	// ConsumerGroupID - ID группы потребителей Kafka
	ConsumerGroupID string

	// LogLevel - уровень логирования (debug, info, warn, error)
	LogLevel string
}

// Load загружает конфигурацию из переменных окружения с разумными значениями по умолчанию
func Load() *Config {
	return &Config{
		KafkaBootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		PaymentTopic:          getEnv("KAFKA_PAYMENT_TOPIC", "orders.payment"),
		AssemblyTopic:         getEnv("KAFKA_ASSEMBLY_TOPIC", "orders.assembly"),
		ConsumerGroupID:       getEnv("KAFKA_CONSUMER_GROUP_ID", "assembly-service"),
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

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
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
