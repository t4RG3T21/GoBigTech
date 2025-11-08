package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	api "github.com/t4RG3T21/GoBigTech/services/order/api"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/service"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID string, items []models.OrderItem) (*models.Order, error) {
	args := m.Called(ctx, userID, items)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, id string) (*models.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

// Убеждаемся, что MockOrderService реализует интерфейс
var _ service.OrderServiceInterface = (*MockOrderService)(nil)

func TestOrderHandler_PostOrders_Success(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	// Тестовый запрос
	requestBody := `{
        "user_id": "user123",
        "items": [
            {"product_id": "prod1", "quantity": 2}
        ]
    }`

	req, err := http.NewRequest("POST", "/orders", bytes.NewBufferString(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Ожидаемый вызов сервиса
	expectedItems := []models.OrderItem{
		{ProductID: "prod1", Quantity: 2, Price: 100.0},
	}
	expectedOrder := &models.Order{
		ID:     "order-123",
		UserID: "user123",
		Status: "paid",
		Items:  expectedItems,
		Total:  200.0,
	}

	mockService.On("CreateOrder", req.Context(), "user123", expectedItems).Return(expectedOrder, nil)

	// Act
	handler.PostOrders(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response api.Order
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, "order-123", *response.Id)
	assert.NotNil(t, response.UserId)
	assert.Equal(t, "user123", *response.UserId)
	assert.NotNil(t, response.Status)
	assert.Equal(t, "paid", *response.Status)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_PostOrders_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	// Невалидный JSON
	req, err := http.NewRequest("POST", "/orders", bytes.NewBufferString("invalid json"))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	handler.PostOrders(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "CreateOrder")
}

func TestOrderHandler_PostOrders_ServiceError(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	requestBody := `{
        "user_id": "user123", 
        "items": [
            {"product_id": "prod1", "quantity": 2}
        ]
    }`

	req, err := http.NewRequest("POST", "/orders", bytes.NewBufferString(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Сервис возвращает ошибку
	expectedItems := []models.OrderItem{
		{ProductID: "prod1", Quantity: 2, Price: 100.0},
	}
	mockService.On("CreateOrder", req.Context(), "user123", expectedItems).Return(nil, errors.New("inventory error"))

	// Act
	handler.PostOrders(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrdersId_Success(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	orderID := "order-123"
	expectedOrder := &models.Order{
		ID:     orderID,
		UserID: "user123",
		Status: "paid",
		Items: []models.OrderItem{
			{ProductID: "prod1", Quantity: 2, Price: 100.0},
		},
		Total: 200.0,
	}

	req, err := http.NewRequest("GET", "/orders/"+orderID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	mockService.On("GetOrderByID", req.Context(), orderID).Return(expectedOrder, nil)

	// Act
	handler.GetOrdersId(rr, req, orderID)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response api.Order
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, orderID, *response.Id)
	assert.NotNil(t, response.UserId)
	assert.Equal(t, "user123", *response.UserId)

	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrdersId_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	orderID := "non-existent"
	req, err := http.NewRequest("GET", "/orders/"+orderID, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	mockService.On("GetOrderByID", req.Context(), orderID).Return(nil, errors.New("order not found"))

	// Act
	handler.GetOrdersId(rr, req, orderID)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_PostOrders_EmptyUserID(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	requestBody := `{
        "user_id": "",
        "items": [
            {"product_id": "prod1", "quantity": 2}
        ]
    }`

	req, err := http.NewRequest("POST", "/orders", bytes.NewBufferString(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	handler.PostOrders(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "CreateOrder")
}

func TestOrderHandler_PostOrders_EmptyItems(t *testing.T) {
	// Arrange
	mockService := new(MockOrderService)
	handler := NewOrderHandler(mockService)

	requestBody := `{
        "user_id": "user123",
        "items": []
    }`

	req, err := http.NewRequest("POST", "/orders", bytes.NewBufferString(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Act
	handler.PostOrders(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "CreateOrder")
}
