package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/models"
	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

type mockRepo struct {
	getStockFn     func(ctx context.Context, productID string) (*models.Stock, error)
	reserveStockFn func(ctx context.Context, productID string, quantity int32) error
}

func (m *mockRepo) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	if m.getStockFn == nil {
		return nil, nil
	}
	return m.getStockFn(ctx, productID)
}

func (m *mockRepo) ReserveStock(ctx context.Context, productID string, quantity int32) error {
	if m.reserveStockFn == nil {
		return nil
	}
	return m.reserveStockFn(ctx, productID, quantity)
}

func TestInventoryService_GetStock_Success(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{
		getStockFn: func(ctx context.Context, productID string) (*models.Stock, error) {
			return &models.Stock{
				ProductID: productID,
				Quantity:  100,
				Reserved:  25,
			}, nil
		},
	}

	svc := NewInventoryService(repo)

	resp, err := svc.GetStock(context.Background(), &inventorypb.GetStockRequest{
		ProductId: "prod-123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "prod-123", resp.GetProductId())
	assert.Equal(t, int32(75), resp.GetAvailable())
}

func TestInventoryService_GetStock_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{
		getStockFn: func(ctx context.Context, productID string) (*models.Stock, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewInventoryService(repo)

	resp, err := svc.GetStock(context.Background(), &inventorypb.GetStockRequest{
		ProductId: "prod-missing",
	})

	assert.NoError(t, err)
	assert.Equal(t, "prod-missing", resp.GetProductId())
	assert.Equal(t, int32(0), resp.GetAvailable())
}

func TestInventoryService_ReserveStock_Success(t *testing.T) {
	t.Parallel()

	reserveCalled := false
	repo := &mockRepo{
		reserveStockFn: func(ctx context.Context, productID string, quantity int32) error {
			reserveCalled = true
			assert.Equal(t, "prod-123", productID)
			assert.Equal(t, int32(5), quantity)
			return nil
		},
	}

	svc := NewInventoryService(repo)

	resp, err := svc.ReserveStock(context.Background(), &inventorypb.ReserveStockRequest{
		ProductId: "prod-123",
		Quantity:  5,
	})

	assert.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.True(t, reserveCalled)
}

func TestInventoryService_ReserveStock_Error(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{
		reserveStockFn: func(ctx context.Context, productID string, quantity int32) error {
			return errors.New("insufficient")
		},
	}

	svc := NewInventoryService(repo)

	resp, err := svc.ReserveStock(context.Background(), &inventorypb.ReserveStockRequest{
		ProductId: "prod-123",
		Quantity:  50,
	})

	assert.NoError(t, err)
	assert.False(t, resp.GetSuccess())
}
