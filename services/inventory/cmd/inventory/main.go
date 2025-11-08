package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

// Server implementation
type server struct {
	inventorypb.UnimplementedInventoryServiceServer
}

func (s *server) GetStock(ctx context.Context, req *inventorypb.GetStockRequest) (*inventorypb.GetStockResponse, error) {
	log.Printf("GetStock called for product: %s", req.GetProductId())
	return &inventorypb.GetStockResponse{
		ProductId: req.GetProductId(),
		Available: 42,
	}, nil
}

func (s *server) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	log.Printf("ReserveStock called: product=%s, quantity=%d", req.GetProductId(), req.GetQuantity())
	success := req.GetQuantity() <= 42
	return &inventorypb.ReserveStockResponse{
		Success: success,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(s, &server{})

	log.Printf("Inventory gRPC server listening on 127.0.0.1:50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
