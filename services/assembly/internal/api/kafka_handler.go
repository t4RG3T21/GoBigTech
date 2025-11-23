package api

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/assembly/internal/service"
)

// KafkaHandler обрабатывает Kafka сообщения
type KafkaHandler struct {
	reader        *kafka.Reader
	writer        *kafka.Writer
	assemblySvc   *service.AssemblyService
	logger        *zap.Logger
	paymentTopic  string
	assemblyTopic string
}

// NewKafkaHandler создает новый обработчик Kafka
func NewKafkaHandler(
	bootstrapServers string,
	consumerGroupID string,
	paymentTopic string,
	assemblyTopic string,
	assemblySvc *service.AssemblyService,
	logger *zap.Logger,
) *KafkaHandler {
	// Создаем reader для подписки на топик orders.payment
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{bootstrapServers},
		Topic:    paymentTopic,
		GroupID:  consumerGroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	// Создаем writer для отправки сообщений в топик orders.assembly
	writer := &kafka.Writer{
		Addr:         kafka.TCP(bootstrapServers),
		Topic:        assemblyTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		WriteTimeout: 10 * time.Second,
	}

	return &KafkaHandler{
		reader:        reader,
		writer:        writer,
		assemblySvc:   assemblySvc,
		logger:        logger,
		paymentTopic:  paymentTopic,
		assemblyTopic: assemblyTopic,
	}
}

// Start начинает обработку сообщений из Kafka
func (h *KafkaHandler) Start(ctx context.Context) error {
	h.logger.Info("Starting Kafka consumer",
		zap.String("topic", h.paymentTopic),
		zap.String("group_id", h.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Stopping Kafka consumer due to context cancellation")
			return ctx.Err()
		default:
			// Читаем сообщение с таймаутом
			msg, err := h.reader.ReadMessage(ctx)
			if err != nil {
				h.logger.Error("Failed to read message from Kafka",
					zap.Error(err),
					zap.String("topic", h.paymentTopic),
				)
				// Продолжаем обработку, чтобы не останавливать сервис
				time.Sleep(1 * time.Second)
				continue
			}

			// Обрабатываем сообщение в отдельной горутине
			go h.handleMessage(ctx, msg)
		}
	}
}

// handleMessage обрабатывает одно сообщение
func (h *KafkaHandler) handleMessage(ctx context.Context, msg kafka.Message) {
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

	// Обрабатываем сборку заказа
	assemblyEvent, err := h.assemblySvc.ProcessAssembly(ctx, paymentEvent)
	if err != nil {
		h.logger.Error("Failed to process assembly",
			zap.Error(err),
			zap.String("order_id", paymentEvent.OrderID),
		)
		return
	}

	// Преобразуем событие сборки в JSON
	assemblyData, err := assemblyEvent.ToJSON()
	if err != nil {
		h.logger.Error("Failed to marshal assembly event",
			zap.Error(err),
			zap.String("order_id", assemblyEvent.OrderID),
		)
		return
	}

	// Отправляем событие сборки в Kafka
	kafkaMsg := kafka.Message{
		Key:   []byte(assemblyEvent.OrderID),
		Value: assemblyData,
		Time:  time.Now(),
	}

	if err := h.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		h.logger.Error("Failed to write assembly event to Kafka",
			zap.Error(err),
			zap.String("topic", h.assemblyTopic),
			zap.String("order_id", assemblyEvent.OrderID),
		)
		return
	}

	h.logger.Info("Assembly event sent successfully",
		zap.String("order_id", assemblyEvent.OrderID),
		zap.String("topic", h.assemblyTopic),
	)
}

// Close закрывает соединения с Kafka
func (h *KafkaHandler) Close() error {
	var errs []error

	if err := h.reader.Close(); err != nil {
		errs = append(errs, err)
		h.logger.Error("Failed to close Kafka reader", zap.Error(err))
	}

	if err := h.writer.Close(); err != nil {
		errs = append(errs, err)
		h.logger.Error("Failed to close Kafka writer", zap.Error(err))
	}

	if len(errs) > 0 {
		return errs[0]
	}

	h.logger.Info("Kafka handler closed successfully")
	return nil
}
