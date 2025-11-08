package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderapi "github.com/t4RG3T21/GoBigTech/services/order/api" // сгенерированный OpenAPI
	"github.com/t4RG3T21/GoBigTech/services/order/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/service"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

func main() {
	// 1. Инициализация зависимостей
	orderRepo := repository.NewInMemoryOrderRepo()

	// 2. Подключение к внешним сервисам
	connInv, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to inventory: %v", err)
	}
	defer connInv.Close()

	connPay, err := grpc.NewClient("127.0.0.1:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to payment: %v", err)
	}
	defer connPay.Close()

	// 3. Создание адаптеров для gRPC клиентов
	invAdapter := &service.InventoryClientAdapter{
		Client: inventorypb.NewInventoryServiceClient(connInv),
	}
	payAdapter := &service.PaymentClientAdapter{
		Client: paymentpb.NewPaymentServiceClient(connPay),
	}

	// 4. Создание сервиса (бизнес-логики)
	orderService := service.NewOrderService(orderRepo, invAdapter, payAdapter)

	// 5. Создание HTTP обработчиков
	orderHandler := api.NewOrderHandler(orderService)

	// 6. Настройка маршрутов с использованием OpenAPI генерации
	r := chi.NewRouter()
	orderapi.HandlerFromMux(orderHandler, r)

	log.Println("Order service (refactored) starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
