package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

type server struct {
	paymentpb.UnimplementedPaymentServiceServer
}

func (s *server) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	log.Printf("ProcessPayment called: order=%s, user=%s, amount=%.2f",
		req.GetOrderId(), req.GetUserId(), req.GetAmount())

	return &paymentpb.ProcessPaymentResponse{
		Success:       true,
		TransactionId: "tx_" + req.GetOrderId(),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(s, &server{})

	log.Printf("Payment gRPC server listening on 127.0.0.1:50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
