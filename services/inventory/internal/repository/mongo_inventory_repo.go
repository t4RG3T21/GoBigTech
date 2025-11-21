package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/models"
)

type MongoInventoryRepo struct {
	collection *mongo.Collection
}

func NewMongoInventoryRepo(collection *mongo.Collection) *MongoInventoryRepo {
	return &MongoInventoryRepo{collection: collection}
}

// Убеждаемся, что MongoInventoryRepo реализует интерфейс service.InventoryRepository
var _ interface {
	GetStock(ctx context.Context, productID string) (*models.Stock, error)
	ReserveStock(ctx context.Context, productID string, quantity int32) error
} = (*MongoInventoryRepo)(nil)

func (r *MongoInventoryRepo) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	var stock models.Stock
	err := r.collection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&stock)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("stock not found")
		}
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}
	return &stock, nil
}

func (r *MongoInventoryRepo) ReserveStock(ctx context.Context, productID string, quantity int32) error {
	// Сначала проверяем доступное количество
	var stock models.Stock
	err := r.collection.FindOne(ctx, bson.M{"product_id": productID}).Decode(&stock)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("product %s not found", productID)
		}
		return fmt.Errorf("failed to find stock: %w", err)
	}

	available := stock.Quantity - stock.Reserved
	if available < quantity {
		return fmt.Errorf("insufficient stock: available %d, requested %d", available, quantity)
	}

	// Резервируем товар (увеличиваем reserved на quantity)
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"product_id": productID},
		bson.M{"$inc": bson.M{"reserved": quantity}},
	)
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product %s not found during reservation", productID)
	}

	return nil
}

func (r *MongoInventoryRepo) CreateStock(ctx context.Context, stock *models.Stock) error {
	_, err := r.collection.InsertOne(ctx, stock)
	if err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}
	return nil
}
