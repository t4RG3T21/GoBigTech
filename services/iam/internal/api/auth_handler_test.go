package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/iam/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/service"
	iampb "github.com/t4RG3T21/GoBigTech/services/iam/v1"
)

// MockAuthService - мок для AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password, name string) (string, error) {
	args := m.Called(ctx, email, password, name)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password, username string) (string, string, error) {
	args := m.Called(ctx, email, password, username)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockAuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Убеждаемся, что MockAuthService реализует интерфейс
var _ service.AuthServiceInterface = (*MockAuthService)(nil)

func TestAuthHandler_Register_Success(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "Test User",
	}

	mockService.On("Register", ctx, req.GetEmail(), req.GetPassword(), req.GetUsername()).Return("user-123", nil)

	// Act
	resp, err := handler.Register(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "user-123", resp.UserId)
	assert.Contains(t, resp.Message, "successfully")
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_Failure(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Username: "Test User",
	}

	mockService.On("Register", ctx, req.GetEmail(), req.GetPassword(), req.GetUsername()).Return("", errors.New("user with this email already exists"))

	// Act
	resp, err := handler.Register(ctx, req)

	// Assert
	assert.NoError(t, err) // gRPC handler не возвращает ошибку, а устанавливает Success=false
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "already exists")
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockService.On("Login", ctx, req.GetEmail(), req.GetPassword(), req.GetUsername()).Return("session-123", "user-123", nil)

	// Act
	resp, err := handler.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "session-123", resp.SessionId)
	assert.Equal(t, "user-123", resp.UserId)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Failure(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockService.On("Login", ctx, req.GetEmail(), req.GetPassword(), req.GetUsername()).Return("", "", errors.New("invalid email or password"))

	// Act
	resp, err := handler.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "invalid email or password")
	mockService.AssertExpectations(t)
}

func TestAuthHandler_ValidateSession_Success(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.ValidateSessionRequest{
		SessionId: "session-123",
	}

	session := &models.Session{
		ID:     "session-123",
		UserID: "user-123",
		Email:  "test@example.com",
	}

	mockService.On("ValidateSession", ctx, req.GetSessionId()).Return(session, nil)

	// Act
	resp, err := handler.ValidateSession(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Valid)
	assert.Equal(t, "user-123", resp.UserId)
	assert.Equal(t, "test@example.com", resp.Email)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_ValidateSession_Invalid(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.ValidateSessionRequest{
		SessionId: "invalid-session",
	}

	mockService.On("ValidateSession", ctx, req.GetSessionId()).Return(nil, errors.New("invalid or expired session"))

	// Act
	resp, err := handler.ValidateSession(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Valid)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetUser_Success(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.GetUserRequest{
		UserId: "user-123",
	}

	user := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	mockService.On("GetUser", ctx, req.GetUserId()).Return(user, nil)

	// Act
	resp, err := handler.GetUser(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user-123", resp.UserId)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "Test User", resp.Name)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_GetUser_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockAuthService)
	logger := zap.NewNop()
	handler := NewAuthHandler(mockService, logger)

	ctx := context.Background()
	req := &iampb.GetUserRequest{
		UserId: "nonexistent-user",
	}

	mockService.On("GetUser", ctx, req.GetUserId()).Return(nil, errors.New("user not found"))

	// Act
	resp, err := handler.GetUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockService.AssertExpectations(t)
}
