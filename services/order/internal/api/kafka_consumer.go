package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
)

// AssemblyEvent представляет событие сборки заказа
type AssemblyEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"`
	AssembledAt time.Time `json:"assembled_at"`
}

// KafkaConsumer обрабатывает сообщения из Kafka
type KafkaConsumer struct {
	reader     *kafka.Reader
	orderRepo  repository.OrderRepository
	logger     *zap.Logger
	topic      string
}

// NewKafkaConsumer создает новый Kafka consumer для топика orders.assembly
func NewKafkaConsumer(
	bootstrapServers string,
	consumerGroupID string,
	topic string,
	orderRepo repository.OrderRepository,
	logger *zap.Logger,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{bootstrapServers},
		Topic:    topic,
		GroupID:  consumerGroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader:    reader,
		orderRepo: orderRepo,
		logger:    logger,
		topic:     topic,
	}
}

// Start начинает обработку сообщений из Kafka
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer",
		zap.String("topic", c.topic),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping Kafka consumer due to context cancellation")
			return ctx.Err()
		default:
			// Читаем сообщение с таймаутом
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read message from Kafka",
					zap.Error(err),
					zap.String("topic", c.topic),
				)
				// Продолжаем обработку, чтобы не останавливать сервис
				time.Sleep(1 * time.Second)
				continue
			}

			// Обрабатываем сообщение
			c.handleMessage(ctx, msg)
		}
	}
}

// handleMessage обрабатывает одно сообщение о сборке
func (c *KafkaConsumer) handleMessage(ctx context.Context, msg kafka.Message) {
	c.logger.Info("Received assembly event",
		zap.String("topic", c.topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.ByteString("key", msg.Key),
		zap.ByteString("value", msg.Value),
	)

	// Парсим событие сборки
	var assemblyEvent AssemblyEvent
	if err := json.Unmarshal(msg.Value, &assemblyEvent); err != nil {
		c.logger.Error("Failed to parse assembly event",
			zap.Error(err),
			zap.ByteString("message", msg.Value),
		)
		return
	}

	// Обновляем статус заказа в базе данных
	order, err := c.orderRepo.GetByID(ctx, assemblyEvent.OrderID)
	if err != nil {
		c.logger.Error("Failed to get order",
			zap.Error(err),
			zap.String("order_id", assemblyEvent.OrderID),
		)
		return
	}

	// Обновляем статус заказа
	if assemblyEvent.Status == "completed" {
		order.Status = "assembled"
		if err := c.orderRepo.Update(ctx, order); err != nil {
			c.logger.Error("Failed to update order status",
				zap.Error(err),
				zap.String("order_id", assemblyEvent.OrderID),
			)
			return
		}

		c.logger.Info("Order status updated to assembled",
			zap.String("order_id", assemblyEvent.OrderID),
		)
	}
}

// Close закрывает соединение с Kafka
func (c *KafkaConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

