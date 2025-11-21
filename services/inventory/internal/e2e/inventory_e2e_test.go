//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/service"
	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
)

// InventoryE2ETestSuite - E2E тесты для Inventory Service
// Требует запущенную MongoDB и тестирует через реальный gRPC сервер
type InventoryE2ETestSuite struct {
	suite.Suite
	server      *grpc.Server
	client      inventorypb.InventoryServiceClient
	conn        *grpc.ClientConn
	mongoClient *mongo.Client
	collection  *mongo.Collection
	ctx         context.Context
	listener    net.Listener
}

func (suite *InventoryE2ETestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Получаем URL MongoDB из переменной окружения или используем дефолтный
	mongoURL := os.Getenv("TEST_MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:password@localhost:27017"
	}

	// Подключаемся к MongoDB
	client, err := mongo.Connect(suite.ctx, options.Client().ApplyURI(mongoURL))
	require.NoError(suite.T(), err, "Failed to connect to MongoDB")
	suite.mongoClient = client

	// Проверяем подключение
	err = client.Ping(suite.ctx, nil)
	require.NoError(suite.T(), err, "Failed to ping MongoDB")

	// Получаем коллекцию
	suite.collection = client.Database("test_inventory_e2e").Collection("stocks")

	// Создаем репозиторий напрямую для теста
	repo := repository.NewMongoInventoryRepo(suite.collection)

	// Создаем сервис
	service := service.NewInventoryService(repo)

	// Запускаем gRPC сервер на случайном порту
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(suite.T(), err, "Failed to create listener")
	suite.listener = listener

	suite.server = grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(suite.server, service)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := suite.server.Serve(listener); err != nil {
			suite.T().Logf("gRPC server error: %v", err)
		}
	}()

	// Даем серверу время на запуск
	time.Sleep(100 * time.Millisecond)

	// Подключаемся к серверу как клиент
	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(suite.T(), err, "Failed to connect to gRPC server")
	suite.conn = conn
	suite.client = inventorypb.NewInventoryServiceClient(conn)
}

func (suite *InventoryE2ETestSuite) TearDownSuite() {
	if suite.conn != nil {
		suite.conn.Close()
	}
	if suite.server != nil {
		suite.server.GracefulStop()
	}
	if suite.mongoClient != nil {
		suite.mongoClient.Disconnect(suite.ctx)
	}
}

func (suite *InventoryE2ETestSuite) SetupTest() {
	// Очищаем коллекцию перед каждым тестом
	_, err := suite.collection.DeleteMany(suite.ctx, bson.M{})
	require.NoError(suite.T(), err, "Failed to clean collection")

	// Создаем тестовые данные
	testStocks := []interface{}{
		models.Stock{ProductID: "prod-1", Quantity: 100, Reserved: 0},
		models.Stock{ProductID: "prod-2", Quantity: 50, Reserved: 10},
		models.Stock{ProductID: "prod-3", Quantity: 200, Reserved: 50},
	}
	_, err = suite.collection.InsertMany(suite.ctx, testStocks)
	require.NoError(suite.T(), err, "Failed to insert test data")
}

func TestInventoryE2ETestSuite(t *testing.T) {
	suite.Run(t, new(InventoryE2ETestSuite))
}

func (suite *InventoryE2ETestSuite) TestGetStock_Success() {
	// Act
	resp, err := suite.client.GetStock(suite.ctx, &inventorypb.GetStockRequest{
		ProductId: "prod-1",
	})

	// Assert
	require.NoError(suite.T(), err, "GetStock should not return error")
	require.NotNil(suite.T(), resp, "Response should not be nil")
	assert.Equal(suite.T(), "prod-1", resp.GetProductId(), "Product ID should match")
	assert.Equal(suite.T(), int32(100), resp.GetAvailable(), "Available should be 100 (100 - 0)")
}

func (suite *InventoryE2ETestSuite) TestGetStock_WithReserved() {
	// Act
	resp, err := suite.client.GetStock(suite.ctx, &inventorypb.GetStockRequest{
		ProductId: "prod-2",
	})

	// Assert
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "prod-2", resp.GetProductId())
	assert.Equal(suite.T(), int32(40), resp.GetAvailable(), "Available should be 40 (50 - 10)")
}

func (suite *InventoryE2ETestSuite) TestGetStock_NotFound() {
	// Act
	resp, err := suite.client.GetStock(suite.ctx, &inventorypb.GetStockRequest{
		ProductId: "non-existent",
	})

	// Assert
	require.NoError(suite.T(), err, "GetStock should not return error for non-existent product")
	assert.Equal(suite.T(), "non-existent", resp.GetProductId())
	assert.Equal(suite.T(), int32(0), resp.GetAvailable(), "Available should be 0 for non-existent product")
}

func (suite *InventoryE2ETestSuite) TestReserveStock_Success() {
	// Act
	resp, err := suite.client.ReserveStock(suite.ctx, &inventorypb.ReserveStockRequest{
		ProductId: "prod-1",
		Quantity:  10,
	})

	// Assert
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), resp)
	assert.True(suite.T(), resp.GetSuccess(), "ReserveStock should succeed")

	// Проверяем, что reserved увеличилось в БД
	var stock models.Stock
	err = suite.collection.FindOne(suite.ctx, bson.M{"product_id": "prod-1"}).Decode(&stock)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(10), stock.Reserved, "Reserved should be 10")

	// Проверяем доступное количество
	getResp, err := suite.client.GetStock(suite.ctx, &inventorypb.GetStockRequest{
		ProductId: "prod-1",
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(90), getResp.GetAvailable(), "Available should be 90 (100 - 10)")
}

func (suite *InventoryE2ETestSuite) TestReserveStock_InsufficientStock() {
	// Act - пытаемся зарезервировать больше, чем доступно
	resp, err := suite.client.ReserveStock(suite.ctx, &inventorypb.ReserveStockRequest{
		ProductId: "prod-2",
		Quantity:  50, // Доступно только 40 (50 - 10)
	})

	// Assert
	require.NoError(suite.T(), err)
	assert.False(suite.T(), resp.GetSuccess(), "ReserveStock should fail for insufficient stock")

	// Проверяем, что reserved не изменилось
	var stock models.Stock
	err = suite.collection.FindOne(suite.ctx, bson.M{"product_id": "prod-2"}).Decode(&stock)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(10), stock.Reserved, "Reserved should remain 10")
}

func (suite *InventoryE2ETestSuite) TestReserveStock_ExactAmount() {
	// Act - резервируем точное доступное количество
	resp, err := suite.client.ReserveStock(suite.ctx, &inventorypb.ReserveStockRequest{
		ProductId: "prod-2",
		Quantity:  40, // Точное доступное количество
	})

	// Assert
	require.NoError(suite.T(), err)
	assert.True(suite.T(), resp.GetSuccess(), "ReserveStock should succeed for exact amount")

	// Проверяем, что все зарезервировано
	var stock models.Stock
	err = suite.collection.FindOne(suite.ctx, bson.M{"product_id": "prod-2"}).Decode(&stock)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(50), stock.Reserved, "Reserved should be 50 (10 + 40)")

	// Проверяем доступное количество
	getResp, err := suite.client.GetStock(suite.ctx, &inventorypb.GetStockRequest{
		ProductId: "prod-2",
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(0), getResp.GetAvailable(), "Available should be 0")
}

func (suite *InventoryE2ETestSuite) TestReserveStock_NotFound() {
	// Act - пытаемся зарезервировать несуществующий товар
	resp, err := suite.client.ReserveStock(suite.ctx, &inventorypb.ReserveStockRequest{
		ProductId: "non-existent",
		Quantity:  10,
	})

	// Assert
	require.NoError(suite.T(), err)
	assert.False(suite.T(), resp.GetSuccess(), "ReserveStock should fail for non-existent product")
}

func (suite *InventoryE2ETestSuite) TestGetStock_AfterReserve() {
	// Arrange - резервируем товар
	_, err := suite.client.ReserveStock(suite.ctx, &inventorypb.ReserveStockRequest{
		ProductId: "prod-3",
		Quantity:  30,
	})
	require.NoError(suite.T(), err)

	// Act - получаем остаток
	resp, err := suite.client.GetStock(suite.ctx, &inventorypb.GetStockRequest{
		ProductId: "prod-3",
	})

	// Assert
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "prod-3", resp.GetProductId())
	assert.Equal(suite.T(), int32(120), resp.GetAvailable(), "Available should be 120 (200 - 50 - 30)")
}
