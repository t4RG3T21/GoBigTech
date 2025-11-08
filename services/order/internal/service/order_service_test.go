package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	repomocks "github.com/t4RG3T21/GoBigTech/services/order/internal/repository/mocks"
	servicemocks "github.com/t4RG3T21/GoBigTech/services/order/internal/service/mocks"
)

func TestOrderService_CreateOrder_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Создаем моки
	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	// Создаем сервис с моками
	service := NewOrderService(repoMock, invMock, payMock)

	// Тестовые данные
	userID := "user123"
	items := []models.OrderItem{
		{ProductID: "prod1", Quantity: 2, Price: 100.0},
		{ProductID: "prod2", Quantity: 1, Price: 50.0},
	}

	// Настраиваем ожидания моков
	invMock.On("ReserveStock", ctx, "prod1", int32(2)).Return(true, nil)
	invMock.On("ReserveStock", ctx, "prod2", int32(1)).Return(true, nil)
	payMock.On("ProcessPayment", ctx, mock.AnythingOfType("string"), userID, 250.0).Return("tx_12345", nil)
	repoMock.On("Create", ctx, mock.AnythingOfType("*models.Order")).Return(nil)

	// Act
	order, err := service.CreateOrder(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, "paid", order.Status)
	assert.Equal(t, 250.0, order.Total)
	assert.Len(t, order.Items, 2)

	// Проверяем что все моки были вызваны
	invMock.AssertExpectations(t)
	payMock.AssertExpectations(t)
	repoMock.AssertExpectations(t)
}

func TestOrderService_CreateOrder_InventoryUnavailable(t *testing.T) {
	// Arrange
	ctx := context.Background()

	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repoMock, invMock, payMock)

	userID := "user123"
	items := []models.OrderItem{
		{ProductID: "prod1", Quantity: 10, Price: 100.0},
	}

	// Настраиваем что товар недоступен
	invMock.On("ReserveStock", ctx, "prod1", int32(10)).Return(false, nil)

	// Act
	order, err := service.CreateOrder(ctx, userID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "not available")

	// Проверяем что оплата НЕ вызывалась
	payMock.AssertNotCalled(t, "ProcessPayment")
	repoMock.AssertNotCalled(t, "Create")
}

func TestOrderService_CreateOrder_PaymentFailed(t *testing.T) {
	// Arrange
	ctx := context.Background()

	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repoMock, invMock, payMock)

	userID := "user123"
	items := []models.OrderItem{
		{ProductID: "prod1", Quantity: 2, Price: 100.0},
	}

	invMock.On("ReserveStock", ctx, "prod1", int32(2)).Return(true, nil)
	payMock.On("ProcessPayment", ctx, mock.AnythingOfType("string"), userID, 200.0).Return("", errors.New("insufficient funds"))

	// Act
	order, err := service.CreateOrder(ctx, userID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "payment processing failed")

	// Проверяем что репозиторий НЕ вызывался
	repoMock.AssertNotCalled(t, "Create")
}

func TestOrderService_CreateOrder_EmptyItems(t *testing.T) {
	// Arrange
	ctx := context.Background()

	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repoMock, invMock, payMock)

	// Act
	order, err := service.CreateOrder(ctx, "user123", []models.OrderItem{})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "must contain at least one item")

	// Проверяем что НИЧЕГО не вызывалось
	invMock.AssertNotCalled(t, "ReserveStock")
	payMock.AssertNotCalled(t, "ProcessPayment")
	repoMock.AssertNotCalled(t, "Create")
}

func TestOrderService_GetOrderByID_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repoMock, invMock, payMock)

	orderID := "order-123"
	expectedOrder := &models.Order{
		ID:     orderID,
		UserID: "user123",
		Status: "paid",
		Items: []models.OrderItem{
			{ProductID: "prod1", Quantity: 2, Price: 100.0},
		},
		Total: 200.0,
	}

	repoMock.On("GetByID", ctx, orderID).Return(expectedOrder, nil)

	// Act
	order, err := service.GetOrderByID(ctx, orderID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, orderID, order.ID)
	assert.Equal(t, "user123", order.UserID)
	repoMock.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()

	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repoMock, invMock, payMock)

	orderID := "non-existent"
	repoMock.On("GetByID", ctx, orderID).Return(nil, errors.New("order not found"))

	// Act
	order, err := service.GetOrderByID(ctx, orderID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to get order")
	repoMock.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_EmptyID(t *testing.T) {
	// Arrange
	ctx := context.Background()

	repoMock := new(repomocks.MockOrderRepository)
	invMock := new(servicemocks.MockInventoryClient)
	payMock := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repoMock, invMock, payMock)

	// Act
	order, err := service.GetOrderByID(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "order ID cannot be empty")
	repoMock.AssertNotCalled(t, "GetByID")
}
