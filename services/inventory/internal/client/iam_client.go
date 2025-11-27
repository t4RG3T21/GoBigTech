package client

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	iampb "github.com/t4RG3T21/GoBigTech/services/iam/v1"
)

// IAMClient интерфейс для работы с IAM Service
type IAMClient interface {
	ValidateSession(ctx context.Context, sessionID string) (*iampb.ValidateSessionResponse, error)
	Close() error
}

// iamClient реализует IAMClient
type iamClient struct {
	client iampb.IAMServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

// NewIAMClient создает новый клиент для IAM Service
func NewIAMClient(address string, logger *zap.Logger) (IAMClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := iampb.NewIAMServiceClient(conn)

	return &iamClient{
		client: client,
		conn:   conn,
		logger: logger,
	}, nil
}

// ValidateSession проверяет валидность сессии через IAM Service
func (c *iamClient) ValidateSession(ctx context.Context, sessionID string) (*iampb.ValidateSessionResponse, error) {
	req := &iampb.ValidateSessionRequest{
		SessionId: sessionID,
	}

	resp, err := c.client.ValidateSession(ctx, req)
	if err != nil {
		c.logger.Warn("Failed to validate session",
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		return nil, err
	}

	return resp, nil
}

// Close закрывает соединение с IAM Service
func (c *iamClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
