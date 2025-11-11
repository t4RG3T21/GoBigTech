package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/service"
	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

// MockInventoryRepository - мок для репозитория
type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockInventoryRepository) ReserveStock(ctx context.Context, productID string, quantity int32) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func TestServer_GetStock_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.GetStockRequest{
		ProductId: "prod1",
	}

	mockRepo.On("GetStock", ctx, "prod1").Return(&models.Stock{
		ProductID: "prod1",
		Quantity:  100,
		Reserved:  10,
	}, nil)

	// Act
	resp, err := s.GetStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "prod1", resp.GetProductId())
	assert.Equal(t, int32(90), resp.GetAvailable()) // 100 - 10
	mockRepo.AssertExpectations(t)
}

func TestServer_GetStock_DifferentProduct(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.GetStockRequest{
		ProductId: "prod2",
	}

	mockRepo.On("GetStock", ctx, "prod2").Return(&models.Stock{
		ProductID: "prod2",
		Quantity:  50,
		Reserved:  5,
	}, nil)

	// Act
	resp, err := s.GetStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "prod2", resp.GetProductId())
	assert.Equal(t, int32(45), resp.GetAvailable()) // 50 - 5
	mockRepo.AssertExpectations(t)
}

func TestServer_ReserveStock_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  10,
	}

	mockRepo.On("ReserveStock", ctx, "prod1", int32(10)).Return(nil)

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	mockRepo.AssertExpectations(t)
}

func TestServer_ReserveStock_QuantityExceedsAvailable(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  50,
	}

	mockRepo.On("ReserveStock", ctx, "prod1", int32(50)).Return(assert.AnError)

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.GetSuccess())
	mockRepo.AssertExpectations(t)
}

func TestServer_ReserveStock_ExactQuantity(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  42,
	}

	mockRepo.On("ReserveStock", ctx, "prod1", int32(42)).Return(nil)

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	mockRepo.AssertExpectations(t)
}

func TestServer_ReserveStock_ZeroQuantity(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  0,
	}

	mockRepo.On("ReserveStock", ctx, "prod1", int32(0)).Return(nil)

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	mockRepo.AssertExpectations(t)
}

func TestServer_ReserveStock_EmptyProductID(t *testing.T) {
	// Arrange
	mockRepo := new(MockInventoryRepository)
	s := service.NewInventoryService(mockRepo)
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "",
		Quantity:  10,
	}

	mockRepo.On("ReserveStock", ctx, "", int32(10)).Return(nil)

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	mockRepo.AssertExpectations(t)
}
