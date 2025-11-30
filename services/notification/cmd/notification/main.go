package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
	"github.com/t4RG3T21/GoBigTech/services/notification/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/notification/internal/di"
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
	logger, err := platformlogger.New("notification-service", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Устанавливаем логгер в контейнер
	container.SetLogger(logger)

	logger.Info("Starting Notification Service",
		zap.String("kafka_bootstrap_servers", cfg.KafkaBootstrapServers),
		zap.String("payment_topic", cfg.PaymentTopic),
		zap.String("assembly_topic", cfg.AssemblyTopic),
		zap.String("consumer_group_id", cfg.ConsumerGroupID),
		zap.Int64("telegram_chat_id", cfg.TelegramChatID),
		zap.String("log_level", cfg.LogLevel),
	)

	// 4. Инициализация сервиса уведомлений (проверяем подключение к Telegram)
	notificationService, err := container.Service()
	if err != nil {
		logger.Fatal("Failed to initialize notification service",
			zap.Error(err),
			zap.String("hint", "Check TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID environment variables"),
		)
	}
	logger.Info("Notification service initialized successfully")

	// 5. Инициализация Kafka handler
	kafkaHandler, err := container.KafkaHandler()
	if err != nil {
		logger.Fatal("Failed to initialize Kafka handler", zap.Error(err))
	}
	logger.Info("Kafka handler initialized successfully")

	// 6. Создаем HTTP роутер и настраиваем эндпоинты
	r := chi.NewRouter()

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Health check called")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"service":   "notification",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Alert endpoint
	alertHandler := api.NewAlertHandler(notificationService, logger.Logger)
	r.Post("/alert", alertHandler.HandleAlert)

	// 7. Настройка HTTP сервера
	serverAddress := ":" + cfg.HTTPPort
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ДОБАВИМ ЛОГ ДЛЯ ОТЛАДКИ
	logger.Info("Configured HTTP server",
		zap.String("address", serverAddress),
		zap.String("http_port", cfg.HTTPPort))

	// 8. Запуск HTTP сервера в отдельной горутине
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", zap.Error(err))
			serverErr <- err
		}
	}()

	// 9. Создаем контекст для обработки Kafka сообщений
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 10. Запускаем обработку Kafka сообщений в отдельной горутине
	handlerErr := make(chan error, 1)
	go func() {
		logger.Info("Starting Kafka message processing")
		if err := kafkaHandler.Start(ctx); err != nil {
			handlerErr <- err
		}
	}()

	// 11. Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Fatal("HTTP server error", zap.Error(err))
	case err := <-handlerErr:
		logger.Fatal("Kafka handler error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Создаем контекст с таймаутом для graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Останавливаем HTTP сервер
		logger.Info("Shutting down HTTP server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during HTTP server shutdown", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped gracefully")
		}

		// Отменяем контекст для остановки обработки Kafka сообщений
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

	logger.Info("Notification Service stopped")
}
