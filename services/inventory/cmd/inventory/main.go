package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/client"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/config"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/interceptors"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/service"
	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Инициализация логгера
	logger, err := platformlogger.New("inventory-service", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Inventory Service",
		zap.String("grpc_port", cfg.GRPCPort),
		zap.String("iam_grpc_address", cfg.IAMGRPCAddress),
		zap.String("log_level", cfg.LogLevel),
	)

	// 3. Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURL))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
	}()

	// Проверка подключения
	if err = mongoClient.Ping(ctx, nil); err != nil {
		logger.Fatal("Failed to ping MongoDB", zap.Error(err))
	}

	// 4. Инициализация репозитория
	collection := mongoClient.Database(cfg.DatabaseName).Collection(cfg.CollectionName)
	inventoryRepo := repository.NewMongoInventoryRepo(collection)

	// 5. Создание начальных данных (ОБЯЗАТЕЛЬНО!)
	initData(ctx, collection, logger)

	// 6. Создание клиента IAM Service
	iamClient, err := client.NewIAMClient(cfg.IAMGRPCAddress, logger.Logger)
	if err != nil {
		logger.Fatal("Failed to create IAM client", zap.Error(err))
	}
	defer func() {
		if err := iamClient.Close(); err != nil {
			logger.Error("Failed to close IAM client", zap.Error(err))
		}
	}()

	// 7. Создание интерцептора аутентификации
	authInterceptor := interceptors.NewAuthInterceptor(iamClient, logger.Logger)

	// 8. Создание сервиса
	inventoryService := service.NewInventoryService(inventoryRepo)

	// 9. Настройка gRPC сервера с интерцептором
	listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("Failed to create listener", zap.Error(err))
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)
	inventorypb.RegisterInventoryServiceServer(grpcSrv, inventoryService)

	// Регистрация gRPC reflection для работы с grpcurl
	reflection.Register(grpcSrv)

	// 10. Запуск gRPC сервера в отдельной горутине
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("gRPC server starting", zap.String("address", listener.Addr().String()))
		if err := grpcSrv.Serve(listener); err != nil {
			serverErr <- err
		}
	}()

	// 11. Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Fatal("gRPC server error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Останавливаем gRPC сервер
		logger.Info("Shutting down gRPC server...")
		grpcSrv.GracefulStop()

		// Закрываем клиент IAM
		if err := iamClient.Close(); err != nil {
			logger.Error("Error closing IAM client", zap.Error(err))
		}
	}

	logger.Info("Inventory Service stopped")
}

func initData(ctx context.Context, collection *mongo.Collection, logger *platformlogger.Logger) {
	// Создаем тестовые данные
	stocks := []interface{}{
		bson.M{"product_id": "p1", "quantity": int32(100), "reserved": int32(0)},
		bson.M{"product_id": "p2", "quantity": int32(50), "reserved": int32(0)},
		bson.M{"product_id": "p3", "quantity": int32(200), "reserved": int32(0)},
	}

	// Очищаем и создаем заново (для демо)
	collection.Drop(ctx)

	result, err := collection.InsertMany(ctx, stocks)
	if err != nil {
		logger.Warn("Failed to insert test data", zap.Error(err))
	} else {
		logger.Info("Inserted test stocks", zap.Int("count", len(result.InsertedIDs)))
	}
}
