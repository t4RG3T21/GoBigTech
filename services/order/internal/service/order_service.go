package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
)

// Интерфейсы внешних сервисов (Dependency Inversion)
type InventoryClient interface {
	ReserveStock(ctx context.Context, productID string, quantity int32) (bool, error)
}

type PaymentClient interface {
	ProcessPayment(ctx context.Context, orderID, userID string, amount float64) (string, error)
}

// OrderServiceInterface - интерфейс сервиса заказов (для тестирования и зависимостей)
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, userID string, items []models.OrderItem) (*models.Order, error)
	GetOrderByID(ctx context.Context, id string) (*models.Order, error)
}

// OrderService - содержит бизнес-логику
type OrderService struct {
	repo        repository.OrderRepository
	inventory   InventoryClient
	payment     PaymentClient
	kafka       *KafkaProducer
	logger      *zap.Logger
	metrics     *OrderMetrics
	promMetrics *PrometheusMetrics
}

// Убеждаемся, что OrderService реализует интерфейс
var _ OrderServiceInterface = (*OrderService)(nil)

func NewOrderService(repo repository.OrderRepository, inv InventoryClient, pay PaymentClient, kafka *KafkaProducer, logger *zap.Logger, metrics *OrderMetrics, promMetrics *PrometheusMetrics) *OrderService {
	return &OrderService{
		repo:        repo,
		inventory:   inv,
		payment:     pay,
		kafka:       kafka,
		logger:      logger,
		metrics:     metrics,
		promMetrics: promMetrics,
	}
}

// CreateOrder - use case создания заказа
func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []models.OrderItem) (*models.Order, error) {
	startTime := time.Now()

	// Создаем span для всей операции создания заказа
	ctx, span := createOrderSpan(ctx, userID, len(items))
	defer span.End()

	// 1. Валидация входных данных
	if userID == "" {
		span.SetAttributes(attribute.Bool("validation.error", true))
		duration := time.Since(startTime)
		if s.metrics != nil {
			s.metrics.RecordOrderFailure("empty_user_id", duration)
		}
		if s.promMetrics != nil {
			s.promMetrics.RecordOrderFailure("empty_user_id")
		}
		return nil, errors.New("user ID cannot be empty")
	}
	if len(items) == 0 {
		span.SetAttributes(attribute.Bool("validation.error", true))
		duration := time.Since(startTime)
		if s.metrics != nil {
			s.metrics.RecordOrderFailure("empty_items", duration)
		}
		if s.promMetrics != nil {
			s.promMetrics.RecordOrderFailure("empty_items")
		}
		return nil, errors.New("order must contain at least one item")
	}

	span.SetAttributes(attribute.Bool("validation.success", true))

	// 2. Проверка доступности товаров
	for _, item := range items {
		// Создаем span для проверки каждого товара
		invCtx, invSpan := inventoryCheckSpan(ctx, item.ProductID, int32(item.Quantity))

		available, err := s.inventory.ReserveStock(invCtx, item.ProductID, int32(item.Quantity))

		if err != nil {
			invSpan.SetAttributes(attribute.Bool("inventory.error", true))
			invSpan.End()
			duration := time.Since(startTime)
			if s.metrics != nil {
				s.metrics.RecordOrderFailure("inventory_check_failed", duration)
			}
			if s.promMetrics != nil {
				s.promMetrics.RecordOrderFailure("inventory_check_failed")
			}
			return nil, fmt.Errorf("inventory check failed: %w", err)
		}
		if !available {
			invSpan.SetAttributes(attribute.Bool("inventory.available", false))
			invSpan.End()
			duration := time.Since(startTime)
			if s.metrics != nil {
				s.metrics.RecordOrderFailure("product_not_available", duration)
			}
			if s.promMetrics != nil {
				s.promMetrics.RecordOrderFailure("product_not_available")
			}
			return nil, fmt.Errorf("product %s is not available", item.ProductID)
		}

		invSpan.SetAttributes(attribute.Bool("inventory.available", true))
		invSpan.End()
	}

	// 3. Создание заказа
	order := &models.Order{
		ID:     generateOrderID(),
		UserID: userID,
		Status: "created",
		Items:  items,
	}
	order.CalculateTotal()

	span.SetAttributes(attribute.String("order.id", order.ID))

	// 4. Обработка платежа
	paymentCtx, paymentSpan := paymentProcessingSpan(ctx, order.ID, order.Total)

	transactionID, err := s.payment.ProcessPayment(paymentCtx, order.ID, userID, order.Total)

	if err != nil {
		paymentSpan.SetAttributes(attribute.Bool("payment.error", true))
		paymentSpan.End()
		duration := time.Since(startTime)
		if s.metrics != nil {
			s.metrics.RecordOrderFailure("payment_processing_failed", duration)
		}
		if s.promMetrics != nil {
			s.promMetrics.RecordOrderFailure("payment_processing_failed")
		}
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	paymentSpan.SetAttributes(
		attribute.Bool("payment.success", true),
		attribute.String("payment.transaction_id", transactionID),
	)
	paymentSpan.End()

	// 5. Сохранение заказа
	order.Status = "paid"
	dbCtx, dbSpan := databaseSaveSpan(ctx, order.ID)
	if err := s.repo.Create(dbCtx, order); err != nil {
		dbSpan.SetAttributes(attribute.Bool("database.error", true))
		dbSpan.RecordError(err)
		dbSpan.End()
		span.SetAttributes(attribute.Bool("database.error", true))
		duration := time.Since(startTime)
		if s.metrics != nil {
			s.metrics.RecordOrderFailure("database_save_failed", duration)
		}
		if s.promMetrics != nil {
			s.promMetrics.RecordOrderFailure("database_save_failed")
		}
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	dbSpan.SetAttributes(attribute.Bool("database.success", true))
	dbSpan.End()
	span.SetAttributes(attribute.Bool("database.success", true))

	// 6. Отправка события оплаты в Kafka
	if s.kafka != nil {
		paymentEvent := &PaymentEvent{
			OrderID:       order.ID,
			UserID:        userID,
			Amount:        order.Total,
			TransactionID: transactionID,
			Timestamp:     time.Now(),
		}

		kafkaCtx, kafkaSpan := kafkaSendSpan(ctx, "orders.payment", order.ID)
		if err := s.kafka.SendPaymentEvent(kafkaCtx, paymentEvent); err != nil {
			kafkaSpan.SetAttributes(attribute.Bool("kafka.error", true))
			kafkaSpan.RecordError(err)
			kafkaSpan.End()
			s.logger.Error("Failed to send payment event to Kafka",
				zap.Error(err),
				zap.String("order_id", order.ID),
			)
			span.SetAttributes(attribute.Bool("kafka.error", true))
			// Не возвращаем ошибку, так как заказ уже создан и оплачен
		} else {
			kafkaSpan.SetAttributes(attribute.Bool("kafka.success", true))
			kafkaSpan.End()
			span.SetAttributes(attribute.Bool("kafka.success", true))
			// Записываем метрику отправки Kafka сообщения
			if s.metrics != nil {
				s.metrics.RecordKafkaMessage("orders.payment")
			}
		}
	}

	span.SetAttributes(attribute.String("order.status", "completed"))

	// ЗАПИСЫВАЕМ МЕТРИКИ УСПЕШНОГО ЗАКАЗА
	duration := time.Since(startTime)

	// Используем OpenTelemetry метрики
	if s.metrics != nil {
		s.metrics.RecordOrderSuccess(order.Total, duration)
	}

	// Используем Prometheus метрики
	if s.promMetrics != nil {
		s.promMetrics.RecordOrderSuccess(order.Total, duration.Seconds())
	}

	return order, nil
}

// GetOrderByID - use case получения заказа по ID
func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	ctx, span := tracerProvider.StartSpan(ctx, "order.get_by_id")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", id))

	if id == "" {
		span.SetAttributes(attribute.Bool("validation.error", true))
		return nil, errors.New("order ID cannot be empty")
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.SetAttributes(attribute.Bool("database.error", true))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	span.SetAttributes(
		attribute.Bool("database.success", true),
		attribute.String("order.status", order.Status),
	)
	return order, nil
}

func generateOrderID() string {
	return fmt.Sprintf("order-%d", time.Now().UnixNano())
}
