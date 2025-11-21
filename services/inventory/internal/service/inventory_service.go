package service

import (
	"context"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/models"
	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

// InventoryRepository - интерфейс репозитория
type InventoryRepository interface {
	GetStock(ctx context.Context, productID string) (*models.Stock, error)
	ReserveStock(ctx context.Context, productID string, quantity int32) error
}

// InventoryService - сервис для работы с инвентарем
type InventoryService struct {
	inventorypb.UnimplementedInventoryServiceServer
	repo InventoryRepository
}

// NewInventoryService создает новый сервис инвентаря
func NewInventoryService(repo InventoryRepository) *InventoryService {
	return &InventoryService{
		repo: repo,
	}
}

// GetStock - получение информации о товаре
func (s *InventoryService) GetStock(ctx context.Context, req *inventorypb.GetStockRequest) (*inventorypb.GetStockResponse, error) {
	stock, err := s.repo.GetStock(ctx, req.GetProductId())
	if err != nil {
		return &inventorypb.GetStockResponse{
			ProductId: req.GetProductId(),
			Available: 0,
		}, nil
	}

	available := stock.Quantity - stock.Reserved
	return &inventorypb.GetStockResponse{
		ProductId: stock.ProductID,
		Available: available,
	}, nil
}

// ReserveStock - резервирование товара
func (s *InventoryService) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	err := s.repo.ReserveStock(ctx, req.GetProductId(), req.GetQuantity())
	if err != nil {
		// Возвращаем false при любой ошибке (товар не найден, недостаточно товара и т.д.)
		return &inventorypb.ReserveStockResponse{
			Success: false,
		}, nil
	}

	return &inventorypb.ReserveStockResponse{
		Success: true,
	}, nil
}
