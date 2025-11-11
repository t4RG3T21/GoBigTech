package repository_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
)

type PostgresOrderRepoTestSuite struct {
	suite.Suite
	db   *pgxpool.Pool
	repo *repository.PostgresOrderRepo
	ctx  context.Context
}

func (suite *PostgresOrderRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Подключаемся к тестовой БД (должна быть запущена через Docker Compose)
	dbURL := "postgres://postgres:password@localhost:5432/order_service?sslmode=disable"
	db, err := pgxpool.New(suite.ctx, dbURL)
	if err != nil {
		suite.T().Fatalf("Failed to connect to test database: %v", err)
	}
	suite.db = db
	suite.repo = repository.NewPostgresOrderRepo(db)
}

func (suite *PostgresOrderRepoTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *PostgresOrderRepoTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	suite.db.Exec(suite.ctx, "DELETE FROM order_items")
	suite.db.Exec(suite.ctx, "DELETE FROM orders")
}

func TestPostgresOrderRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresOrderRepoTestSuite))
}

func (suite *PostgresOrderRepoTestSuite) TestCreateAndGetOrder() {
	// Arrange
	order := &models.Order{
		ID:     "test-order-1",
		UserID: "user123",
		Status: "created",
		Total:  250.0,
		Items: []models.OrderItem{
			{ProductID: "p1", Quantity: 2, Price: 100.0},
			{ProductID: "p2", Quantity: 1, Price: 50.0},
		},
	}

	// Act
	err := suite.repo.Create(suite.ctx, order)
	assert.NoError(suite.T(), err)

	retrievedOrder, err := suite.repo.GetByID(suite.ctx, "test-order-1")

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), order.ID, retrievedOrder.ID)
	assert.Equal(suite.T(), order.UserID, retrievedOrder.UserID)
	assert.Equal(suite.T(), order.Status, retrievedOrder.Status)
	assert.Equal(suite.T(), order.Total, retrievedOrder.Total)
	assert.Len(suite.T(), retrievedOrder.Items, 2)
}

func (suite *PostgresOrderRepoTestSuite) TestGetOrder_NotFound() {
	// Act
	order, err := suite.repo.GetByID(suite.ctx, "non-existent-order")

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), order)
	assert.Contains(suite.T(), err.Error(), "order not found")
}
