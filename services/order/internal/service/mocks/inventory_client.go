package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) ReserveStock(ctx context.Context, productID string, quantity int32) (bool, error) {
	args := m.Called(ctx, productID, quantity)
	return args.Bool(0), args.Error(1)
}

type MockPaymentClient struct {
	mock.Mock
}

func (m *MockPaymentClient) ProcessPayment(ctx context.Context, orderID, userID string, amount float64) (string, error) {
	args := m.Called(ctx, orderID, userID, amount)
	return args.String(0), args.Error(1)
}
