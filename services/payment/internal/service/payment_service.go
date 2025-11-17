package service

import (
	"context"
	"log"

	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

// PaymentService implements gRPC PaymentServiceServer without side effects so it can be unit tested.
type PaymentService struct {
	paymentpb.UnimplementedPaymentServiceServer
}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	log.Printf("ProcessPayment called: order=%s, user=%s, amount=%.2f",
		req.GetOrderId(), req.GetUserId(), req.GetAmount())

	return &paymentpb.ProcessPaymentResponse{
		Success:       true,
		TransactionId: "tx_" + req.GetOrderId(),
	}, nil
}
