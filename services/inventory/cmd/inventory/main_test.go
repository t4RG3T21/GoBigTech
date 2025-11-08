package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

func TestServer_GetStock_Success(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.GetStockRequest{
		ProductId: "prod1",
	}

	// Act
	resp, err := s.GetStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "prod1", resp.GetProductId())
	assert.Equal(t, int32(42), resp.GetAvailable())
}

func TestServer_GetStock_DifferentProduct(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.GetStockRequest{
		ProductId: "prod2",
	}

	// Act
	resp, err := s.GetStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "prod2", resp.GetProductId())
	assert.Equal(t, int32(42), resp.GetAvailable())
}

func TestServer_ReserveStock_Success(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  10,
	}

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
}

func TestServer_ReserveStock_QuantityExceedsAvailable(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  50, // Больше чем доступно (42)
	}

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.GetSuccess())
}

func TestServer_ReserveStock_ExactQuantity(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  42, // Точно доступное количество
	}

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
}

func TestServer_ReserveStock_ZeroQuantity(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "prod1",
		Quantity:  0,
	}

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess()) // 0 <= 42, должно быть успешно
}

func TestServer_ReserveStock_EmptyProductID(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &inventorypb.ReserveStockRequest{
		ProductId: "",
		Quantity:  10,
	}

	// Act
	resp, err := s.ReserveStock(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
}
