package interceptors

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/client"
)

const (
	// SessionIDKey ключ для session_id в метаданных gRPC
	SessionIDKey = "session_id"
)

// AuthInterceptor проверяет аутентификацию для защищенных методов
type AuthInterceptor struct {
	iamClient client.IAMClient
	logger    *zap.Logger
}

// NewAuthInterceptor создает новый интерцептор аутентификации
func NewAuthInterceptor(iamClient client.IAMClient, logger *zap.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		iamClient: iamClient,
		logger:    logger,
	}
}

// Unary возвращает unary interceptor для проверки аутентификации
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Проверяем, требуется ли аутентификация для этого метода
		if !a.requiresAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Извлекаем session_id из метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			a.logger.Warn("Missing metadata in request", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// Получаем session_id из метаданных
		sessionIDs := md.Get(SessionIDKey)
		if len(sessionIDs) == 0 || sessionIDs[0] == "" {
			a.logger.Warn("Missing session_id in metadata", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "missing session_id")
		}

		sessionID := sessionIDs[0]

		// Валидируем сессию через IAM Service
		sessionResp, err := a.iamClient.ValidateSession(ctx, sessionID)
		if err != nil {
			a.logger.Warn("Session validation failed",
				zap.String("session_id", sessionID),
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Unauthenticated, "invalid or expired session")
		}

		if !sessionResp.Valid {
			a.logger.Warn("Session is not valid",
				zap.String("session_id", sessionID),
				zap.String("method", info.FullMethod),
			)
			return nil, status.Errorf(codes.Unauthenticated, "invalid or expired session")
		}

		// Добавляем информацию о пользователе в контекст для использования в обработчиках
		ctx = context.WithValue(ctx, "user_id", sessionResp.UserId)
		ctx = context.WithValue(ctx, "user_email", sessionResp.Email)

		a.logger.Debug("Session validated successfully",
			zap.String("session_id", sessionID),
			zap.String("user_id", sessionResp.UserId),
			zap.String("method", info.FullMethod),
		)

		// Вызываем следующий обработчик
		return handler(ctx, req)
	}
}

// requiresAuth проверяет, требуется ли аутентификация для метода
func (a *AuthInterceptor) requiresAuth(fullMethod string) bool {
	// Методы, требующие аутентификации
	protectedMethods := []string{
		"/inventory.v1.InventoryService/GetStock",
		"/inventory.v1.InventoryService/ReserveStock",
	}

	for _, method := range protectedMethods {
		if strings.HasSuffix(fullMethod, method) {
			return true
		}
	}

	return false
}
