package main

import (
	"context"
	"log"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/service"
	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

func main() {
	// 1. Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI("mongodb://admin:password@localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Проверка подключения
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// 2. Инициализация репозитория
	collection := client.Database("inventory_service").Collection("stocks")
	inventoryRepo := repository.NewMongoInventoryRepo(collection)

	// 3. Создание начальных данных (ОБЯЗАТЕЛЬНО!)
	initData(ctx, collection)

	// 4. Создание сервиса
	inventoryService := service.NewInventoryService(inventoryRepo)

	// 5. Запуск gRPC сервера
	l, err := net.Listen("tcp4", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcSrv, inventoryService)

	log.Println("inventory gRPC (with MongoDB) listening on 127.0.0.1:50051")
	if err := grpcSrv.Serve(l); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func initData(ctx context.Context, collection *mongo.Collection) {
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
		log.Printf("Failed to insert test data: %v", err)
	} else {
		log.Printf("Inserted %d test stocks: p1 (100), p2 (50), p3 (200)", len(result.InsertedIDs))
	}
}
