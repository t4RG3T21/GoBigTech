package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

func TestPaymentService_ProcessPayment_Scenarios(t *testing.T) {
	tests := []struct {
		name string
		req  *paymentpb.ProcessPaymentRequest
	}{
		{
			name: "success",
			req: &paymentpb.ProcessPaymentRequest{
				OrderId: "order-123",
				UserId:  "user-456",
				Amount:  100.50,
				Method:  "card",
			},
		},
		{
			name: "zero amount",
			req: &paymentpb.ProcessPaymentRequest{
				OrderId: "order-321",
				UserId:  "user-888",
				Amount:  0,
				Method:  "card",
			},
		},
		{
			name: "empty user",
			req: &paymentpb.ProcessPaymentRequest{
				OrderId: "order-111",
				UserId:  "",
				Amount:  42.0,
				Method:  "paypal",
			},
		},
		{
			name: "empty order",
			req: &paymentpb.ProcessPaymentRequest{
				OrderId: "",
				UserId:  "user-xyz",
				Amount:  10,
				Method:  "card",
			},
		},
	}

	svc := NewPaymentService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.ProcessPayment(context.Background(), tt.req)
			assert.NoError(t, err)
			assert.True(t, resp.GetSuccess())
			assert.Equal(t, "tx_"+tt.req.GetOrderId(), resp.GetTransactionId())
		})
	}
}
