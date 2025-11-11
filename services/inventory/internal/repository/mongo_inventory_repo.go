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
	// Используем транзакцию для обеспечения атомарности
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessionContext mongo.SessionContext) (interface{}, error) {
		// Сначала проверяем доступное количество
		var stock models.Stock
		err := r.collection.FindOne(sessionContext, bson.M{"product_id": productID}).Decode(&stock)
		if err != nil {
			return nil, fmt.Errorf("failed to find stock: %w", err)
		}

		available := stock.Quantity - stock.Reserved
		if available < quantity {
			return nil, errors.New("insufficient stock")
		}

		// Резервируем товар
		result, err := r.collection.UpdateOne(
			sessionContext,
			bson.M{"product_id": productID},
			bson.M{"$inc": bson.M{"reserved": quantity}},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reserve stock: %w", err)
		}

		if result.MatchedCount == 0 {
			return nil, errors.New("stock not found during reservation")
		}

		return result, nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	if result == nil {
		return errors.New("reservation failed")
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
