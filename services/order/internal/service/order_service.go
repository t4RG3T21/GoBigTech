package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	repo      repository.OrderRepository
	inventory InventoryClient
	payment   PaymentClient
}

// Убеждаемся, что OrderService реализует интерфейс
var _ OrderServiceInterface = (*OrderService)(nil)

func NewOrderService(repo repository.OrderRepository, inv InventoryClient, pay PaymentClient) *OrderService {
	return &OrderService{
		repo:      repo,
		inventory: inv,
		payment:   pay,
	}
}

// CreateOrder - use case создания заказа
func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []models.OrderItem) (*models.Order, error) {
	// 1. Валидация входных данных
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if len(items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	// 2. Проверка доступности товаров
	for _, item := range items {
		available, err := s.inventory.ReserveStock(ctx, item.ProductID, int32(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("inventory check failed: %w", err)
		}
		if !available {
			return nil, fmt.Errorf("product %s is not available", item.ProductID)
		}
	}

	// 3. Создание заказа
	order := &models.Order{
		ID:     generateOrderID(), // В реальности UUID
		UserID: userID,
		Status: "created",
		Items:  items,
	}
	order.CalculateTotal() // доменная логика

	// 4. Обработка платежа
	_, err := s.payment.ProcessPayment(ctx, order.ID, userID, order.Total)
	if err != nil {
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// 5. Сохранение заказа
	order.Status = "paid"
	if err := s.repo.Create(ctx, order); err != nil {
		// В реальности нужна компенсирующая транзакция
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return order, nil
}

// GetOrderByID - use case получения заказа по ID
func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	if id == "" {
		return nil, errors.New("order ID cannot be empty")
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

func generateOrderID() string {
	// Временная реализация
	return fmt.Sprintf("order-%d", time.Now().UnixNano())
}
