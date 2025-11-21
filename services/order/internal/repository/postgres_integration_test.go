//go:build integration
// +build integration

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
)

// PostgresIntegrationTestSuite - интеграционные тесты для PostgresOrderRepo
// Требует запущенную PostgreSQL БД
type PostgresIntegrationTestSuite struct {
	suite.Suite
	db   *pgxpool.Pool
	repo *repository.PostgresOrderRepo
	ctx  context.Context
}

func (suite *PostgresIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Получаем URL БД из переменной окружения или используем дефолтный
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/order_service?sslmode=disable"
	}

	// Подключаемся к тестовой БД
	db, err := pgxpool.New(suite.ctx, dbURL)
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v\nMake sure PostgreSQL is running and TEST_DATABASE_URL is set correctly", err)
	}

	// Проверяем подключение
	if err := db.Ping(suite.ctx); err != nil {
		db.Close()
		suite.T().Fatalf("Failed to ping test database: %v", err)
	}

	suite.db = db
	suite.repo = repository.NewPostgresOrderRepo(db)

	// Убеждаемся, что таблицы существуют (применяем миграции если нужно)
	suite.ensureTablesExist()
}

func (suite *PostgresIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *PostgresIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом для изоляции
	_, err := suite.db.Exec(suite.ctx, "DELETE FROM order_items")
	require.NoError(suite.T(), err, "Failed to clean order_items table")

	_, err = suite.db.Exec(suite.ctx, "DELETE FROM orders")
	require.NoError(suite.T(), err, "Failed to clean orders table")
}

// ensureTablesExist проверяет существование таблиц и создает их при необходимости
func (suite *PostgresIntegrationTestSuite) ensureTablesExist() {
	ctx, cancel := context.WithTimeout(suite.ctx, 5*time.Second)
	defer cancel()

	// Проверяем существование таблицы orders
	var ordersExists bool
	err := suite.db.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'orders')").
		Scan(&ordersExists)
	if err != nil {
		suite.T().Fatalf("Failed to check if orders table exists: %v", err)
	}

	if !ordersExists {
		suite.T().Fatalf("Table 'orders' does not exist. Please run migrations first.")
	}

	// Проверяем существование таблицы order_items
	var orderItemsExists bool
	err = suite.db.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'order_items')").
		Scan(&orderItemsExists)
	if err != nil {
		suite.T().Fatalf("Failed to check if order_items table exists: %v", err)
	}

	if !orderItemsExists {
		suite.T().Fatalf("Table 'order_items' does not exist. Please run migrations first.")
	}
}

func TestPostgresIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresIntegrationTestSuite))
}

func (suite *PostgresIntegrationTestSuite) TestCreateOrder_Success() {
	// Arrange
	order := &models.Order{
		ID:     "integration-test-order-1",
		UserID: "user-integration-1",
		Status: "created",
		Total:  150.0,
		Items: []models.OrderItem{
			{ProductID: "prod-1", Quantity: 2, Price: 50.0},
			{ProductID: "prod-2", Quantity: 1, Price: 50.0},
		},
	}

	// Act
	err := suite.repo.Create(suite.ctx, order)

	// Assert
	require.NoError(suite.T(), err, "Failed to create order")

	// Проверяем, что заказ действительно сохранен в БД
	var count int
	err = suite.db.QueryRow(suite.ctx, "SELECT COUNT(*) FROM orders WHERE id = $1", order.ID).Scan(&count)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count, "Order should be saved in database")

	// Проверяем количество элементов заказа
	err = suite.db.QueryRow(suite.ctx, "SELECT COUNT(*) FROM order_items WHERE order_id = $1", order.ID).Scan(&count)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count, "Order should have 2 items")
}

func (suite *PostgresIntegrationTestSuite) TestGetOrderByID_Success() {
	// Arrange
	order := &models.Order{
		ID:     "integration-test-order-2",
		UserID: "user-integration-2",
		Status: "paid",
		Total:  300.0,
		Items: []models.OrderItem{
			{ProductID: "prod-3", Quantity: 3, Price: 100.0},
		},
	}

	// Создаем заказ
	err := suite.repo.Create(suite.ctx, order)
	require.NoError(suite.T(), err)

	// Act
	retrievedOrder, err := suite.repo.GetByID(suite.ctx, order.ID)

	// Assert
	require.NoError(suite.T(), err, "Failed to get order")
	require.NotNil(suite.T(), retrievedOrder, "Retrieved order should not be nil")

	// Проверяем корректность всех полей
	assert.Equal(suite.T(), order.ID, retrievedOrder.ID, "Order ID should match")
	assert.Equal(suite.T(), order.UserID, retrievedOrder.UserID, "User ID should match")
	assert.Equal(suite.T(), order.Status, retrievedOrder.Status, "Status should match")
	assert.Equal(suite.T(), order.Total, retrievedOrder.Total, "Total should match")

	// Проверяем элементы заказа
	require.Len(suite.T(), retrievedOrder.Items, 1, "Order should have 1 item")
	assert.Equal(suite.T(), order.Items[0].ProductID, retrievedOrder.Items[0].ProductID, "Product ID should match")
	assert.Equal(suite.T(), order.Items[0].Quantity, retrievedOrder.Items[0].Quantity, "Quantity should match")
	assert.Equal(suite.T(), order.Items[0].Price, retrievedOrder.Items[0].Price, "Price should match")
}

func (suite *PostgresIntegrationTestSuite) TestGetOrderByID_NotFound() {
	// Act
	order, err := suite.repo.GetByID(suite.ctx, "non-existent-order-id")

	// Assert
	assert.Error(suite.T(), err, "Should return error for non-existent order")
	assert.Nil(suite.T(), order, "Order should be nil when not found")
	assert.Contains(suite.T(), err.Error(), "order not found", "Error message should indicate order not found")
}

func (suite *PostgresIntegrationTestSuite) TestCreateOrder_WithMultipleItems() {
	// Arrange
	order := &models.Order{
		ID:     "integration-test-order-3",
		UserID: "user-integration-3",
		Status: "created",
		Total:  500.0,
		Items: []models.OrderItem{
			{ProductID: "prod-1", Quantity: 1, Price: 100.0},
			{ProductID: "prod-2", Quantity: 2, Price: 150.0},
			{ProductID: "prod-3", Quantity: 1, Price: 100.0},
		},
	}

	// Act
	err := suite.repo.Create(suite.ctx, order)

	// Assert
	require.NoError(suite.T(), err)

	// Получаем заказ обратно и проверяем все элементы
	retrievedOrder, err := suite.repo.GetByID(suite.ctx, order.ID)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedOrder)

	assert.Equal(suite.T(), order.ID, retrievedOrder.ID)
	assert.Len(suite.T(), retrievedOrder.Items, 3, "Order should have 3 items")

	// Проверяем каждый элемент
	for i, expectedItem := range order.Items {
		assert.Equal(suite.T(), expectedItem.ProductID, retrievedOrder.Items[i].ProductID,
			"Product ID should match for item %d", i)
		assert.Equal(suite.T(), expectedItem.Quantity, retrievedOrder.Items[i].Quantity,
			"Quantity should match for item %d", i)
		assert.Equal(suite.T(), expectedItem.Price, retrievedOrder.Items[i].Price,
			"Price should match for item %d", i)
	}
}

func (suite *PostgresIntegrationTestSuite) TestCreateOrder_DataIntegrity() {
	// Arrange
	order := &models.Order{
		ID:     "integration-test-order-4",
		UserID: "user-integration-4",
		Status: "paid",
		Total:  250.75,
		Items: []models.OrderItem{
			{ProductID: "prod-expensive", Quantity: 1, Price: 250.75},
		},
	}

	// Act
	err := suite.repo.Create(suite.ctx, order)
	require.NoError(suite.T(), err)

	// Проверяем данные напрямую в БД
	var dbOrderID, dbUserID, dbStatus string
	var dbTotal float64
	err = suite.db.QueryRow(suite.ctx,
		"SELECT id, user_id, status, total FROM orders WHERE id = $1", order.ID).
		Scan(&dbOrderID, &dbUserID, &dbStatus, &dbTotal)
	require.NoError(suite.T(), err)

	// Assert
	assert.Equal(suite.T(), order.ID, dbOrderID, "Order ID in DB should match")
	assert.Equal(suite.T(), order.UserID, dbUserID, "User ID in DB should match")
	assert.Equal(suite.T(), order.Status, dbStatus, "Status in DB should match")
	assert.InDelta(suite.T(), order.Total, dbTotal, 0.01, "Total in DB should match (with float precision)")

	// Проверяем элементы заказа в БД
	var dbProductID string
	var dbQuantity int
	var dbPrice float64
	err = suite.db.QueryRow(suite.ctx,
		"SELECT product_id, quantity, price FROM order_items WHERE order_id = $1", order.ID).
		Scan(&dbProductID, &dbQuantity, &dbPrice)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), order.Items[0].ProductID, dbProductID, "Product ID in DB should match")
	assert.Equal(suite.T(), order.Items[0].Quantity, dbQuantity, "Quantity in DB should match")
	assert.InDelta(suite.T(), order.Items[0].Price, dbPrice, 0.01, "Price in DB should match")
}

func (suite *PostgresIntegrationTestSuite) TestCreateOrder_TransactionRollback() {
	// Arrange - создаем заказ с валидными данными
	order1 := &models.Order{
		ID:     "integration-test-order-5",
		UserID: "user-integration-5",
		Status: "created",
		Total:  100.0,
		Items: []models.OrderItem{
			{ProductID: "prod-1", Quantity: 1, Price: 100.0},
		},
	}

	// Act - создаем первый заказ
	err := suite.repo.Create(suite.ctx, order1)
	require.NoError(suite.T(), err)

	// Проверяем, что заказ создан
	retrievedOrder, err := suite.repo.GetByID(suite.ctx, order1.ID)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedOrder)

	// Создаем второй заказ с тем же ID (должна быть ошибка из-за уникальности)
	order2 := &models.Order{
		ID:     "integration-test-order-5", // Дублирующий ID
		UserID: "user-integration-6",
		Status: "created",
		Total:  200.0,
		Items: []models.OrderItem{
			{ProductID: "prod-2", Quantity: 2, Price: 100.0},
		},
	}

	err = suite.repo.Create(suite.ctx, order2)

	// Assert - должна быть ошибка
	assert.Error(suite.T(), err, "Should fail when creating order with duplicate ID")

	// Проверяем, что первый заказ не был изменен (транзакция откатилась)
	retrievedOrder, err = suite.repo.GetByID(suite.ctx, order1.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), order1.UserID, retrievedOrder.UserID, "First order should remain unchanged")
}

func (suite *PostgresIntegrationTestSuite) TestGetOrderByID_EmptyItems() {
	// Arrange - создаем заказ без элементов (если это разрешено)
	// В реальности заказ должен иметь хотя бы один элемент, но тестируем граничный случай
	order := &models.Order{
		ID:     "integration-test-order-6",
		UserID: "user-integration-6",
		Status: "created",
		Total:  0.0,
		Items:  []models.OrderItem{}, // Пустой список элементов
	}

	// Act
	err := suite.repo.Create(suite.ctx, order)
	require.NoError(suite.T(), err)

	retrievedOrder, err := suite.repo.GetByID(suite.ctx, order.ID)

	// Assert
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedOrder)
	assert.Equal(suite.T(), order.ID, retrievedOrder.ID)
	assert.Empty(suite.T(), retrievedOrder.Items, "Order with no items should have empty items list")
}
