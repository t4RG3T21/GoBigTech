package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	repomocks "github.com/t4RG3T21/GoBigTech/services/order/internal/repository/mocks"
	servicemocks "github.com/t4RG3T21/GoBigTech/services/order/internal/service/mocks"
)

func TestOrderService_CreateOrder_EmptyUserID(t *testing.T) {
	repo := new(repomocks.MockOrderRepository)
	inv := new(servicemocks.MockInventoryClient)
	pay := new(servicemocks.MockPaymentClient)

	service := NewOrderService(repo, inv, pay, nil, zap.NewNop())
	order, err := service.CreateOrder(context.Background(), "", []models.OrderItem{
		{ProductID: "prod", Quantity: 1, Price: 10},
	})

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
	inv.AssertExpectations(t)
	pay.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_InventoryError(t *testing.T) {
	ctx := context.Background()
	repo := new(repomocks.MockOrderRepository)
	inv := new(servicemocks.MockInventoryClient)
	pay := new(servicemocks.MockPaymentClient)
	service := NewOrderService(repo, inv, pay, nil, zap.NewNop())

	items := []models.OrderItem{{ProductID: "prod", Quantity: 1, Price: 10}}
	inv.On("ReserveStock", ctx, "prod", int32(1)).Return(false, errors.New("rpc down"))

	order, err := service.CreateOrder(ctx, "user", items)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "inventory check failed")
	pay.AssertNotCalled(t, "ProcessPayment")
	repo.AssertNotCalled(t, "Create")
}

func TestOrderService_CreateOrder_SaveOrderFails(t *testing.T) {
	ctx := context.Background()
	repo := new(repomocks.MockOrderRepository)
	inv := new(servicemocks.MockInventoryClient)
	pay := new(servicemocks.MockPaymentClient)
	service := NewOrderService(repo, inv, pay, nil, zap.NewNop())

	items := []models.OrderItem{{ProductID: "prod", Quantity: 1, Price: 10}}
	inv.On("ReserveStock", ctx, "prod", int32(1)).Return(true, nil)
	pay.On("ProcessPayment", ctx, mock.AnythingOfType("string"), "user", 10.0).Return("tx-1", nil)
	repo.On("Create", ctx, mock.AnythingOfType("*models.Order")).Return(errors.New("db down"))

	order, err := service.CreateOrder(ctx, "user", items)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to save order")
}

func TestGenerateOrderID(t *testing.T) {
	id := generateOrderID()
	assert.True(t, strings.HasPrefix(id, "order-"))
	assert.NotEmpty(t, id)

	time.Sleep(time.Nanosecond)
	other := generateOrderID()
	assert.NotEqual(t, id, other)
}
