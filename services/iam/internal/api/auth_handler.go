package api

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/t4RG3T21/GoBigTech/services/iam/internal/service"
	iampb "github.com/t4RG3T21/GoBigTech/services/iam/v1"
)

// AuthHandler реализует gRPC обработчик для IAM Service
type AuthHandler struct {
	iampb.UnimplementedIAMServiceServer
	authService service.AuthServiceInterface
	logger      *zap.Logger
}

// NewAuthHandler создает новый gRPC обработчик
func NewAuthHandler(authService service.AuthServiceInterface, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register регистрирует нового пользователя
func (h *AuthHandler) Register(ctx context.Context, req *iampb.RegisterRequest) (*iampb.RegisterResponse, error) {
	userID, err := h.authService.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetUsername())
	if err != nil {
		h.logger.Warn("Registration failed", zap.String("email", req.GetEmail()), zap.Error(err))
		return &iampb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	h.logger.Info("User registered successfully", zap.String("user_id", userID), zap.String("email", req.GetEmail()))
	return &iampb.RegisterResponse{
		Success: true,
		UserId:  userID,
		Message: "User registered successfully",
	}, nil
}

// Login выполняет вход пользователя
func (h *AuthHandler) Login(ctx context.Context, req *iampb.LoginRequest) (*iampb.LoginResponse, error) {
	sessionID, userID, err := h.authService.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetUsername())
	if err != nil {
		identifier := req.GetEmail()
		if identifier == "" {
			identifier = req.GetUsername()
		}
		h.logger.Warn("Login failed", zap.String("identifier", identifier), zap.Error(err))
		return &iampb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	identifier := req.GetEmail()
	if identifier == "" {
		identifier = req.GetUsername()
	}
	h.logger.Info("User logged in successfully", zap.String("user_id", userID), zap.String("identifier", identifier))
	return &iampb.LoginResponse{
		Success:   true,
		SessionId: sessionID,
		UserId:    userID,
		Message:   "Login successful",
	}, nil
}

// ValidateSession проверяет валидность сессии
func (h *AuthHandler) ValidateSession(ctx context.Context, req *iampb.ValidateSessionRequest) (*iampb.ValidateSessionResponse, error) {
	session, err := h.authService.ValidateSession(ctx, req.GetSessionId())
	if err != nil {
		h.logger.Warn("Session validation failed", zap.String("session_id", req.GetSessionId()), zap.Error(err))
		return &iampb.ValidateSessionResponse{
			Valid: false,
		}, nil
	}

	return &iampb.ValidateSessionResponse{
		Valid:  true,
		UserId: session.UserID,
		Email:  session.Email,
	}, nil
}

// GetUser получает информацию о пользователе
func (h *AuthHandler) GetUser(ctx context.Context, req *iampb.GetUserRequest) (*iampb.GetUserResponse, error) {
	user, err := h.authService.GetUser(ctx, req.GetUserId())
	if err != nil {
		h.logger.Warn("Get user failed", zap.String("user_id", req.GetUserId()), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &iampb.GetUserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Unix(),
	}, nil
}
