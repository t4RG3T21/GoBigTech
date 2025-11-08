package service

import (
	"context"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

// InventoryClientAdapter - адаптирует gRPC клиент под наш интерфейс
type InventoryClientAdapter struct {
	Client inventorypb.InventoryServiceClient
}

func (a *InventoryClientAdapter) ReserveStock(ctx context.Context, productID string, quantity int32) (bool, error) {
	resp, err := a.Client.ReserveStock(ctx, &inventorypb.ReserveStockRequest{
		ProductId: productID,
		Quantity:  quantity,
	})
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

// PaymentClientAdapter - адаптирует gRPC клиент под наш интерфейс
type PaymentClientAdapter struct {
	Client paymentpb.PaymentServiceClient
}

func (a *PaymentClientAdapter) ProcessPayment(ctx context.Context, orderID, userID string, amount float64) (string, error) {
	resp, err := a.Client.ProcessPayment(ctx, &paymentpb.ProcessPaymentRequest{
		OrderId: orderID,
		UserId:  userID,
		Amount:  amount,
		Method:  "card",
	})
	if err != nil {
		return "", err
	}
	return resp.TransactionId, nil
}
