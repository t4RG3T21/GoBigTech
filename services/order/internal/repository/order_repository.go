package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
)

// Интерфейс репозитория - абстракция над хранилищем
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
}

// InMemory реализация (позже заменим на PostgreSQL)
type InMemoryOrderRepo struct {
	orders map[string]*models.Order
	mutex  sync.RWMutex
}

func NewInMemoryOrderRepo() *InMemoryOrderRepo {
	return &InMemoryOrderRepo{
		orders: make(map[string]*models.Order),
	}
}

func (r *InMemoryOrderRepo) Create(ctx context.Context, order *models.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepo) GetByID(ctx context.Context, id string) (*models.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

func (r *InMemoryOrderRepo) Update(ctx context.Context, order *models.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return errors.New("order not found")
	}

	r.orders[order.ID] = order
	return nil
}
