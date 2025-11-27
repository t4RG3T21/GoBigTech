package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/t4RG3T21/GoBigTech/services/iam/internal/models"
)

// MockUserRepository - мок для UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByName(ctx context.Context, name string) (*models.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockSessionRepository - мок для SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session, ttl time.Duration) error {
	args := m.Called(ctx, session, ttl)
	return args.Error(0)
}

func (m *MockSessionRepository) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	args := m.Called(ctx, sessionID, ttl)
	return args.Error(0)
}

func TestAuthService_Register_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	name := "Test User"

	// Пользователь не существует
	mockUserRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("user not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	userID, err := service.Register(ctx, email, password, name)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, userID)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	email := "existing@example.com"
	password := "password123"
	name := "Test User"

	existingUser := &models.User{
		ID:    "user-123",
		Email: email,
	}

	// Пользователь уже существует
	mockUserRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	// Act
	userID, err := service.Register(ctx, email, password, name)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "already exists")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmptyEmail(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()

	// Act
	userID, err := service.Register(ctx, "", "password123", "Test User")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestAuthService_Register_EmptyPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()

	// Act
	userID, err := service.Register(ctx, "test@example.com", "", "Test User")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"

	// Создаем пользователя с хешированным паролем
	hashedPassword, _ := hashPassword(password)
	user := &models.User{
		ID:       "user-123",
		Email:    email,
		Password: hashedPassword,
	}

	mockUserRepo.On("GetByEmail", ctx, email).Return(user, nil)
	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*models.Session"), 24*time.Hour).Return(nil)

	// Act
	sessionID, userID, err := service.Login(ctx, email, password, "")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)
	assert.Equal(t, "user-123", userID)
	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	email := "nonexistent@example.com"
	password := "password123"

	mockUserRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("user not found"))

	// Act
	sessionID, userID, err := service.Login(ctx, email, password, "")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, sessionID)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	email := "test@example.com"
	password := "wrongpassword"

	hashedPassword, _ := hashPassword("correctpassword")
	user := &models.User{
		ID:       "user-123",
		Email:    email,
		Password: hashedPassword,
	}

	mockUserRepo.On("GetByEmail", ctx, email).Return(user, nil)

	// Act
	sessionID, userID, err := service.Login(ctx, email, password, "")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, sessionID)
	assert.Empty(t, userID)
	assert.Contains(t, err.Error(), "invalid email or password")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_ValidateSession_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	sessionID := "session-123"

	session := &models.Session{
		ID:        sessionID,
		UserID:    "user-123",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	mockSessionRepo.On("Get", ctx, sessionID).Return(session, nil)

	// Act
	result, err := service.ValidateSession(ctx, sessionID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, sessionID, result.ID)
	assert.Equal(t, "user-123", result.UserID)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthService_ValidateSession_EmptySessionID(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()

	// Act
	result, err := service.ValidateSession(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "session ID cannot be empty")
}

func TestAuthService_ValidateSession_InvalidSession(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	sessionID := "invalid-session"

	mockSessionRepo.On("Get", ctx, sessionID).Return(nil, errors.New("session not found"))

	// Act
	result, err := service.ValidateSession(ctx, sessionID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid or expired session")
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthService_GetUser_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	userID := "user-123"

	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Act
	result, err := service.GetUser(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetUser_NotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()
	userID := "nonexistent-user"

	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Act
	result, err := service.GetUser(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetUser_EmptyUserID(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)
	logger := zap.NewNop()
	service := NewAuthService(mockUserRepo, mockSessionRepo, 24*time.Hour, logger)

	ctx := context.Background()

	// Act
	result, err := service.GetUser(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

// Вспомогательная функция для хеширования пароля в тестах
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}
