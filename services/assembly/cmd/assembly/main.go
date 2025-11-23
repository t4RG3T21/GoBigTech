package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
	"github.com/t4RG3T21/GoBigTech/services/assembly/internal/di"
)

func main() {
	// 1. Создание DI контейнера
	container := di.NewContainer()

	// 2. Получение конфигурации
	cfg := container.Config()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// 3. Инициализация логгера с уровнем из конфигурации
	logger, err := platformlogger.New("assembly-service", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Устанавливаем логгер в контейнер
	container.SetLogger(logger)

	logger.Info("Starting Assembly Service",
		zap.String("kafka_bootstrap_servers", cfg.KafkaBootstrapServers),
		zap.String("payment_topic", cfg.PaymentTopic),
		zap.String("assembly_topic", cfg.AssemblyTopic),
		zap.String("consumer_group_id", cfg.ConsumerGroupID),
		zap.String("log_level", cfg.LogLevel),
	)

	// 4. Инициализация Kafka handler
	kafkaHandler := container.KafkaHandler()
	logger.Info("Kafka handler initialized successfully")

	// 5. Создаем контекст для обработки сообщений
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 6. Запускаем обработку сообщений в отдельной горутине
	handlerErr := make(chan error, 1)
	go func() {
		logger.Info("Starting Kafka message processing")
		if err := kafkaHandler.Start(ctx); err != nil {
			handlerErr <- err
		}
	}()

	// 7. Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-handlerErr:
		logger.Fatal("Kafka handler error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Отменяем контекст для остановки обработки сообщений
		logger.Info("Stopping Kafka message processing...")
		cancel()

		// Даем время на завершение текущих операций
		time.Sleep(2 * time.Second)

		// Закрываем ресурсы контейнера
		logger.Info("Closing container resources...")
		if err := container.Close(); err != nil {
			logger.Error("Error closing container resources", zap.Error(err))
		} else {
			logger.Info("Container resources closed successfully")
		}
	}

	logger.Info("Assembly Service stopped")
}
