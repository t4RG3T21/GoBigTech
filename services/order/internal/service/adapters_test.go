package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

type fakeInventoryClient struct {
	lastRequest *inventorypb.ReserveStockRequest
	response    *inventorypb.ReserveStockResponse
	err         error
}

func (f *fakeInventoryClient) GetStock(ctx context.Context, in *inventorypb.GetStockRequest, opts ...grpc.CallOption) (*inventorypb.GetStockResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeInventoryClient) ReserveStock(ctx context.Context, in *inventorypb.ReserveStockRequest, opts ...grpc.CallOption) (*inventorypb.ReserveStockResponse, error) {
	f.lastRequest = in
	return f.response, f.err
}

type fakePaymentClient struct {
	lastRequest *paymentpb.ProcessPaymentRequest
	response    *paymentpb.ProcessPaymentResponse
	err         error
}

func (f *fakePaymentClient) ProcessPayment(ctx context.Context, in *paymentpb.ProcessPaymentRequest, opts ...grpc.CallOption) (*paymentpb.ProcessPaymentResponse, error) {
	f.lastRequest = in
	return f.response, f.err
}

func TestInventoryClientAdapter_ReserveStockSuccess(t *testing.T) {
	client := &fakeInventoryClient{
		response: &inventorypb.ReserveStockResponse{Success: true},
	}

	adapter := &InventoryClientAdapter{Client: client}

	ok, err := adapter.ReserveStock(context.Background(), "prod-1", 5)

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "prod-1", client.lastRequest.GetProductId())
	assert.Equal(t, int32(5), client.lastRequest.GetQuantity())
}

func TestInventoryClientAdapter_ReserveStockError(t *testing.T) {
	client := &fakeInventoryClient{
		err: errors.New("inventory down"),
	}

	adapter := &InventoryClientAdapter{Client: client}

	ok, err := adapter.ReserveStock(context.Background(), "prod-2", 3)

	assert.Error(t, err)
	assert.False(t, ok)
	assert.Equal(t, "prod-2", client.lastRequest.GetProductId())
	assert.Equal(t, int32(3), client.lastRequest.GetQuantity())
}

func TestPaymentClientAdapter_ProcessPaymentSuccess(t *testing.T) {
	client := &fakePaymentClient{
		response: &paymentpb.ProcessPaymentResponse{
			Success:       true,
			TransactionId: "tx_1",
		},
	}

	adapter := &PaymentClientAdapter{Client: client}

	txID, err := adapter.ProcessPayment(context.Background(), "order-1", "user-1", 99.5)

	assert.NoError(t, err)
	assert.Equal(t, "tx_1", txID)
	assert.Equal(t, "order-1", client.lastRequest.GetOrderId())
	assert.Equal(t, "user-1", client.lastRequest.GetUserId())
	assert.Equal(t, 99.5, client.lastRequest.GetAmount())
}

func TestPaymentClientAdapter_ProcessPaymentError(t *testing.T) {
	client := &fakePaymentClient{
		err: errors.New("payment failed"),
	}

	adapter := &PaymentClientAdapter{Client: client}

	txID, err := adapter.ProcessPayment(context.Background(), "order-2", "user-5", 10.0)

	assert.Error(t, err)
	assert.Equal(t, "", txID)
	assert.Equal(t, "order-2", client.lastRequest.GetOrderId())
	assert.Equal(t, "user-5", client.lastRequest.GetUserId())
}
