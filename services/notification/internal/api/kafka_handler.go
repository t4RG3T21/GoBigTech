package api

import (
	"context"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/notification/internal/service"
)

// KafkaHandler обрабатывает Kafka сообщения из обоих топиков
type KafkaHandler struct {
	paymentReader   *kafka.Reader
	assemblyReader  *kafka.Reader
	notificationSvc *service.NotificationService
	logger          *zap.Logger
	paymentTopic    string
	assemblyTopic   string
	wg              sync.WaitGroup
}

// NewKafkaHandler создает новый обработчик Kafka
func NewKafkaHandler(
	bootstrapServers string,
	consumerGroupID string,
	paymentTopic string,
	assemblyTopic string,
	notificationSvc *service.NotificationService,
	logger *zap.Logger,
) *KafkaHandler {
	// Создаем reader для подписки на топик orders.payment
	paymentReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{bootstrapServers},
		Topic:    paymentTopic,
		GroupID:  consumerGroupID + "-payment",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	// Создаем reader для подписки на топик orders.assembly
	assemblyReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{bootstrapServers},
		Topic:    assemblyTopic,
		GroupID:  consumerGroupID + "-assembly",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaHandler{
		paymentReader:   paymentReader,
		assemblyReader:  assemblyReader,
		notificationSvc: notificationSvc,
		logger:          logger,
		paymentTopic:    paymentTopic,
		assemblyTopic:   assemblyTopic,
	}
}

// Start начинает обработку сообщений из обоих топиков Kafka
func (h *KafkaHandler) Start(ctx context.Context) error {
	h.logger.Info("Starting Kafka consumers",
		zap.String("payment_topic", h.paymentTopic),
		zap.String("assembly_topic", h.assemblyTopic),
	)

	// Запускаем обработку сообщений из топика payment в отдельной горутине
	h.wg.Add(1)
	go h.processPaymentMessages(ctx)

	// Запускаем обработку сообщений из топика assembly в отдельной горутине
	h.wg.Add(1)
	go h.processAssemblyMessages(ctx)

	// Ждем завершения всех горутин
	h.wg.Wait()

	return nil
}

// processPaymentMessages обрабатывает сообщения из топика orders.payment
func (h *KafkaHandler) processPaymentMessages(ctx context.Context) {
	defer h.wg.Done()

	h.logger.Info("Starting payment messages consumer",
		zap.String("topic", h.paymentTopic),
	)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Stopping payment messages consumer due to context cancellation")
			return
		default:
			// Читаем сообщение с таймаутом
			msg, err := h.paymentReader.ReadMessage(ctx)
			if err != nil {
				h.logger.Error("Failed to read payment message from Kafka",
					zap.Error(err),
					zap.String("topic", h.paymentTopic),
				)
				// Продолжаем обработку, чтобы не останавливать сервис
				time.Sleep(1 * time.Second)
				continue
			}

			// Обрабатываем сообщение
			h.handlePaymentMessage(ctx, msg)
		}
	}
}

// processAssemblyMessages обрабатывает сообщения из топика orders.assembly
func (h *KafkaHandler) processAssemblyMessages(ctx context.Context) {
	defer h.wg.Done()

	h.logger.Info("Starting assembly messages consumer",
		zap.String("topic", h.assemblyTopic),
	)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Stopping assembly messages consumer due to context cancellation")
			return
		default:
			// Читаем сообщение с таймаутом
			msg, err := h.assemblyReader.ReadMessage(ctx)
			if err != nil {
				h.logger.Error("Failed to read assembly message from Kafka",
					zap.Error(err),
					zap.String("topic", h.assemblyTopic),
				)
				// Продолжаем обработку, чтобы не останавливать сервис
				time.Sleep(1 * time.Second)
				continue
			}

			// Обрабатываем сообщение
			h.handleAssemblyMessage(ctx, msg)
		}
	}
}

// handlePaymentMessage обрабатывает одно сообщение о платеже
func (h *KafkaHandler) handlePaymentMessage(ctx context.Context, msg kafka.Message) {
	h.logger.Info("Received payment event",
		zap.String("topic", h.paymentTopic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.ByteString("key", msg.Key),
		zap.ByteString("value", msg.Value),
	)

	// Парсим событие оплаты
	paymentEvent, err := service.ParsePaymentEvent(msg.Value)
	if err != nil {
		h.logger.Error("Failed to parse payment event",
			zap.Error(err),
			zap.ByteString("message", msg.Value),
		)
		return
	}

	// Отправляем уведомление в Telegram
	if err := h.notificationSvc.SendPaymentNotification(paymentEvent); err != nil {
		h.logger.Error("Failed to send payment notification",
			zap.Error(err),
			zap.String("order_id", paymentEvent.OrderID),
		)
		return
	}

	h.logger.Info("Payment notification processed successfully",
		zap.String("order_id", paymentEvent.OrderID),
	)
}

// handleAssemblyMessage обрабатывает одно сообщение о сборке
func (h *KafkaHandler) handleAssemblyMessage(ctx context.Context, msg kafka.Message) {
	h.logger.Info("Received assembly event",
		zap.String("topic", h.assemblyTopic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.ByteString("key", msg.Key),
		zap.ByteString("value", msg.Value),
	)

	// Парсим событие сборки
	assemblyEvent, err := service.ParseAssemblyEvent(msg.Value)
	if err != nil {
		h.logger.Error("Failed to parse assembly event",
			zap.Error(err),
			zap.ByteString("message", msg.Value),
		)
		return
	}

	// Отправляем уведомление в Telegram
	if err := h.notificationSvc.SendAssemblyNotification(assemblyEvent); err != nil {
		h.logger.Error("Failed to send assembly notification",
			zap.Error(err),
			zap.String("order_id", assemblyEvent.OrderID),
		)
		return
	}

	h.logger.Info("Assembly notification processed successfully",
		zap.String("order_id", assemblyEvent.OrderID),
	)
}

// Close закрывает соединения с Kafka
func (h *KafkaHandler) Close() error {
	var errs []error

	if err := h.paymentReader.Close(); err != nil {
		errs = append(errs, err)
		h.logger.Error("Failed to close payment Kafka reader", zap.Error(err))
	}

	if err := h.assemblyReader.Close(); err != nil {
		errs = append(errs, err)
		h.logger.Error("Failed to close assembly Kafka reader", zap.Error(err))
	}

	if len(errs) > 0 {
		return errs[0]
	}

	h.logger.Info("Kafka handler closed successfully")
	return nil
}
