package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// PaymentEvent представляет событие оплаты заказа
type PaymentEvent struct {
	OrderID       string    `json:"order_id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	TransactionID string    `json:"transaction_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// AssemblyEvent представляет событие сборки заказа
type AssemblyEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"`
	AssembledAt time.Time `json:"assembled_at"`
}

// AssemblyService обрабатывает логику сборки заказов
type AssemblyService struct {
	logger *zap.Logger
}

// NewAssemblyService создает новый сервис сборки
func NewAssemblyService(logger *zap.Logger) *AssemblyService {
	return &AssemblyService{
		logger: logger,
	}
}

// ProcessAssembly обрабатывает событие оплаты и создает событие сборки
// Имитирует процесс сборки заказа (ожидание 10 секунд)
func (s *AssemblyService) ProcessAssembly(ctx context.Context, paymentEvent *PaymentEvent) (*AssemblyEvent, error) {
	s.logger.Info("Processing assembly for order",
		zap.String("order_id", paymentEvent.OrderID),
		zap.String("user_id", paymentEvent.UserID),
		zap.Float64("amount", paymentEvent.Amount),
		zap.String("transaction_id", paymentEvent.TransactionID),
	)

	// Имитация процесса сборки - ожидание 10 секунд
	s.logger.Info("Starting assembly process", zap.String("order_id", paymentEvent.OrderID))

	select {
	case <-time.After(10 * time.Second):
		// Сборка завершена
		s.logger.Info("Assembly completed", zap.String("order_id", paymentEvent.OrderID))
	case <-ctx.Done():
		s.logger.Warn("Assembly cancelled due to context cancellation",
			zap.String("order_id", paymentEvent.OrderID),
			zap.Error(ctx.Err()),
		)
		return nil, ctx.Err()
	}

	// Создаем событие сборки
	assemblyEvent := &AssemblyEvent{
		OrderID:     paymentEvent.OrderID,
		UserID:      paymentEvent.UserID,
		Status:      "completed",
		AssembledAt: time.Now(),
	}

	return assemblyEvent, nil
}

// ParsePaymentEvent парсит JSON сообщение в PaymentEvent
func ParsePaymentEvent(data []byte) (*PaymentEvent, error) {
	var event PaymentEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse payment event: %w", err)
	}
	return &event, nil
}

// ToJSON преобразует AssemblyEvent в JSON
func (e *AssemblyEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
