package service

import (
	"encoding/json"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

// NotificationService обрабатывает отправку уведомлений в Telegram
type NotificationService struct {
	bot    *tgbotapi.BotAPI
	chatID int64
	logger *zap.Logger
}

// NewNotificationService создает новый сервис уведомлений
func NewNotificationService(botToken string, chatID int64, logger *zap.Logger) (*NotificationService, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	logger.Info("Telegram bot authorized",
		zap.String("bot_username", bot.Self.UserName),
		zap.Int64("chat_id", chatID),
	)

	return &NotificationService{
		bot:    bot,
		chatID: chatID,
		logger: logger,
	}, nil
}

// SendPaymentNotification отправляет уведомление о платеже
func (s *NotificationService) SendPaymentNotification(event *PaymentEvent) error {
	message := fmt.Sprintf(
		"💳 *Платеж получен*\n\n"+
			"Заказ: `%s`\n"+
			"Пользователь: `%s`\n"+
			"Сумма: *%.2f*\n"+
			"Транзакция: `%s`\n"+
			"Время: %s",
		event.OrderID,
		event.UserID,
		event.Amount,
		event.TransactionID,
		event.Timestamp.Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(s.chatID, message)
	msg.ParseMode = "Markdown"

	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("Failed to send payment notification",
			zap.Error(err),
			zap.String("order_id", event.OrderID),
		)
		return fmt.Errorf("failed to send payment notification: %w", err)
	}

	s.logger.Info("Payment notification sent",
		zap.String("order_id", event.OrderID),
		zap.String("user_id", event.UserID),
	)

	return nil
}

// SendAssemblyNotification отправляет уведомление о сборке
func (s *NotificationService) SendAssemblyNotification(event *AssemblyEvent) error {
	message := fmt.Sprintf(
		"📦 *Сборка завершена*\n\n"+
			"Заказ: `%s`\n"+
			"Пользователь: `%s`\n"+
			"Статус: *%s*\n"+
			"Время сборки: %s",
		event.OrderID,
		event.UserID,
		event.Status,
		event.AssembledAt.Format("2006-01-02 15:04:05"),
	)

	msg := tgbotapi.NewMessage(s.chatID, message)
	msg.ParseMode = "Markdown"

	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("Failed to send assembly notification",
			zap.Error(err),
			zap.String("order_id", event.OrderID),
		)
		return fmt.Errorf("failed to send assembly notification: %w", err)
	}

	s.logger.Info("Assembly notification sent",
		zap.String("order_id", event.OrderID),
		zap.String("user_id", event.UserID),
		zap.String("status", event.Status),
	)

	return nil
}

// ParsePaymentEvent парсит JSON сообщение в PaymentEvent
func ParsePaymentEvent(data []byte) (*PaymentEvent, error) {
	var event PaymentEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse payment event: %w", err)
	}
	return &event, nil
}

// ParseAssemblyEvent парсит JSON сообщение в AssemblyEvent
func ParseAssemblyEvent(data []byte) (*AssemblyEvent, error) {
	var event AssemblyEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse assembly event: %w", err)
	}
	return &event, nil
}
