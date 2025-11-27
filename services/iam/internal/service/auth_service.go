package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/t4RG3T21/GoBigTech/services/iam/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/repository"
)

// AuthServiceInterface определяет интерфейс для сервиса аутентификации
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, name string) (string, error)
	Login(ctx context.Context, email, password, username string) (string, string, error)
	ValidateSession(ctx context.Context, sessionID string) (*models.Session, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
}

// AuthService содержит бизнес-логику аутентификации
type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	sessionTTL  time.Duration
	logger      *zap.Logger
}

// Убеждаемся, что AuthService реализует интерфейс
var _ AuthServiceInterface = (*AuthService)(nil)

// NewAuthService создает новый сервис аутентификации
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	sessionTTL time.Duration,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		sessionTTL:  sessionTTL,
		logger:      logger,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, email, password, name string) (string, error) {
	// Валидация входных данных
	if email == "" {
		return "", errors.New("email cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}
	if name == "" {
		return "", errors.New("name cannot be empty")
	}

	// Проверяем, существует ли пользователь с таким email
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return "", errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("User registered", zap.String("user_id", user.ID), zap.String("email", email))
	return user.ID, nil
}

// Login выполняет вход пользователя и создает сессию
// Поддерживает вход как по email, так и по username
func (s *AuthService) Login(ctx context.Context, email, password, username string) (string, string, error) {
	// Валидация входных данных
	if password == "" {
		return "", "", errors.New("password cannot be empty")
	}

	var user *models.User
	var err error

	// Пытаемся найти пользователя по email или username
	if email != "" {
		user, err = s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return "", "", errors.New("invalid email or password")
		}
	} else if username != "" {
		user, err = s.userRepo.GetByName(ctx, username)
		if err != nil {
			return "", "", errors.New("invalid username or password")
		}
	} else {
		return "", "", errors.New("email or username must be provided")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Создаем сессию
	sessionID := uuid.New().String()
	session := &models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}

	if err := s.sessionRepo.Create(ctx, session, s.sessionTTL); err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	loginIdentifier := email
	if loginIdentifier == "" {
		loginIdentifier = username
	}
	s.logger.Info("User logged in", zap.String("user_id", user.ID), zap.String("identifier", loginIdentifier))
	return sessionID, user.ID, nil
}

// ValidateSession проверяет валидность сессии
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID cannot be empty")
	}

	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, errors.New("invalid or expired session")
	}

	return session, nil
}

// GetUser получает информацию о пользователе
func (s *AuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
