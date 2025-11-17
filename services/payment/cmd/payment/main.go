package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	paymentservice "github.com/t4RG3T21/GoBigTech/services/payment/internal/service"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(s, paymentservice.NewPaymentService())

	log.Printf("Payment gRPC server listening on 127.0.0.1:50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
