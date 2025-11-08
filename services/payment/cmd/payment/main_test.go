package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

func TestServer_ProcessPayment_Success(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "order-123",
		UserId:  "user-456",
		Amount:  100.50,
		Method:  "card",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_order-123", resp.GetTransactionId())
}

func TestServer_ProcessPayment_DifferentOrderID(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "order-789",
		UserId:  "user-456",
		Amount:  250.75,
		Method:  "card",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_order-789", resp.GetTransactionId())
}

func TestServer_ProcessPayment_ZeroAmount(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "order-123",
		UserId:  "user-456",
		Amount:  0.0,
		Method:  "card",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_order-123", resp.GetTransactionId())
}

func TestServer_ProcessPayment_LargeAmount(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "order-999",
		UserId:  "user-111",
		Amount:  999999.99,
		Method:  "card",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_order-999", resp.GetTransactionId())
}

func TestServer_ProcessPayment_EmptyOrderID(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "",
		UserId:  "user-456",
		Amount:  100.50,
		Method:  "card",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_", resp.GetTransactionId())
}

func TestServer_ProcessPayment_EmptyUserID(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "order-123",
		UserId:  "",
		Amount:  100.50,
		Method:  "card",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_order-123", resp.GetTransactionId())
}

func TestServer_ProcessPayment_DifferentPaymentMethod(t *testing.T) {
	// Arrange
	s := &server{}
	ctx := context.Background()
	req := &paymentpb.ProcessPaymentRequest{
		OrderId: "order-555",
		UserId:  "user-777",
		Amount:  50.25,
		Method:  "paypal",
	}

	// Act
	resp, err := s.ProcessPayment(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.GetSuccess())
	assert.Equal(t, "tx_order-555", resp.GetTransactionId())
}
