package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config содержит конфигурацию Notification Service
type Config struct {
	// KafkaBootstrapServers - адреса Kafka брокеров (через запятую)
	KafkaBootstrapServers string

	// PaymentTopic - топик для получения событий оплаты
	PaymentTopic string

	// AssemblyTopic - топик для получения событий сборки
	AssemblyTopic string

	// ConsumerGroupID - ID группы потребителей Kafka
	ConsumerGroupID string

	// TelegramBotToken - токен Telegram бота (получить от @BotFather)
	TelegramBotToken string

	// TelegramChatID - ID чата для отправки уведомлений
	TelegramChatID int64

	// LogLevel - уровень логирования (debug, info, warn, error)
	LogLevel string
}

// Load загружает конфигурацию из переменных окружения с разумными значениями по умолчанию
func Load() *Config {
	return &Config{
		KafkaBootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		PaymentTopic:          getEnv("KAFKA_PAYMENT_TOPIC", "orders.payment"),
		AssemblyTopic:         getEnv("KAFKA_ASSEMBLY_TOPIC", "orders.assembly"),
		ConsumerGroupID:       getEnv("KAFKA_CONSUMER_GROUP_ID", "notification-service"),
		TelegramBotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:        getEnvAsInt64("TELEGRAM_CHAT_ID", 0),
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

// getEnvAsInt64 возвращает значение переменной окружения как int64 или значение по умолчанию
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
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
	if c.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN cannot be empty (get it from @BotFather)")
	}
	if c.TelegramChatID == 0 {
		return fmt.Errorf("TELEGRAM_CHAT_ID cannot be empty or zero")
	}
	return nil
}
