package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/models"
)

type MongoInventoryRepoTestSuite struct {
	suite.Suite
	client     *mongo.Client
	collection *mongo.Collection
	repo       *MongoInventoryRepo
	ctx        context.Context
}

func (suite *MongoInventoryRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Подключаемся к тестовой MongoDB
	client, err := mongo.Connect(suite.ctx, options.Client().
		ApplyURI("mongodb://admin:password@localhost:27017"))
	if err != nil {
		suite.T().Fatalf("Failed to connect to MongoDB: %v", err)
	}
	suite.client = client
	suite.collection = client.Database("test_inventory").Collection("stocks")
	suite.repo = NewMongoInventoryRepo(suite.collection)
}

func (suite *MongoInventoryRepoTestSuite) TearDownSuite() {
	if suite.client != nil {
		suite.client.Disconnect(suite.ctx)
	}
}

func (suite *MongoInventoryRepoTestSuite) SetupTest() {
	// Очищаем коллекцию перед каждым тестом
	suite.collection.Drop(suite.ctx)

	// Создаем тестовые данные
	stocks := []interface{}{
		models.Stock{ProductID: "p1", Quantity: 100, Reserved: 0},
		models.Stock{ProductID: "p2", Quantity: 50, Reserved: 0},
	}
	_, err := suite.collection.InsertMany(suite.ctx, stocks)
	if err != nil {
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}
}

func TestMongoInventoryRepoTestSuite(t *testing.T) {
	suite.Run(t, new(MongoInventoryRepoTestSuite))
}

func (suite *MongoInventoryRepoTestSuite) TestGetStock_Success() {
	// Act
	stock, err := suite.repo.GetStock(suite.ctx, "p1")

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "p1", stock.ProductID)
	assert.Equal(suite.T(), int32(100), stock.Quantity)
}

func (suite *MongoInventoryRepoTestSuite) TestGetStock_NotFound() {
	// Act
	stock, err := suite.repo.GetStock(suite.ctx, "non-existent")

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), stock)
	assert.Contains(suite.T(), err.Error(), "stock not found")
}

func (suite *MongoInventoryRepoTestSuite) TestReserveStock_Success() {
	// Act
	err := suite.repo.ReserveStock(suite.ctx, "p1", 10)

	// Assert
	assert.NoError(suite.T(), err)

	// Проверяем что reserved увеличилось
	stock, _ := suite.repo.GetStock(suite.ctx, "p1")
	assert.Equal(suite.T(), int32(10), stock.Reserved)
}

func (suite *MongoInventoryRepoTestSuite) TestReserveStock_InsufficientStock() {
	// Act
	err := suite.repo.ReserveStock(suite.ctx, "p1", 150) // больше чем доступно

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "insufficient stock")
}

func (suite *MongoInventoryRepoTestSuite) TestCreateStock_Success() {
	// Arrange
	newStock := &models.Stock{
		ProductID: "p3",
		Quantity:  200,
		Reserved:  0,
	}

	// Act
	err := suite.repo.CreateStock(suite.ctx, newStock)

	// Assert
	assert.NoError(suite.T(), err)

	// Проверяем что создался
	stock, _ := suite.repo.GetStock(suite.ctx, "p3")
	assert.Equal(suite.T(), "p3", stock.ProductID)
}
