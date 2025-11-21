package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
	"github.com/t4RG3T21/GoBigTech/services/payment/internal/di"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
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
	logger, err := platformlogger.New("payment-service", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Payment Service",
		zap.String("grpc_port", cfg.GRPCPort),
		zap.String("log_level", cfg.LogLevel),
	)

	// 4. Инициализация зависимостей через DI контейнер
	paymentService, err := container.Service(context.Background())
	if err != nil {
		logger.Fatal("Failed to initialize payment service", zap.Error(err))
	}
	logger.Info("Dependencies initialized successfully")

	// 5. Запуск gRPC сервера
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("Failed to create listener", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentService)

	// 6. Запуск сервера в отдельной горутине
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("gRPC server starting", zap.String("address", listener.Addr().String()))
		if err := grpcServer.Serve(listener); err != nil {
			serverErr <- err
		}
	}()

	// 7. Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Fatal("Server error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Graceful shutdown gRPC сервера
		logger.Info("Shutting down gRPC server...")
		grpcServer.GracefulStop()

		// Закрываем ресурсы контейнера
		logger.Info("Closing container resources...")
		if err := container.Close(); err != nil {
			logger.Error("Error closing container resources", zap.Error(err))
		} else {
			logger.Info("Container resources closed successfully")
		}
	}

	logger.Info("Payment Service stopped")
}
