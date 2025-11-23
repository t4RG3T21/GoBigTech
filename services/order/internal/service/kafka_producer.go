package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
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

// KafkaProducer отправляет события в Kafka
type KafkaProducer struct {
	writer *kafka.Writer
	logger *zap.Logger
	topic  string
}

// NewKafkaProducer создает новый Kafka producer
func NewKafkaProducer(bootstrapServers, topic string, logger *zap.Logger) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(bootstrapServers),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		WriteTimeout: 10 * time.Second,
	}

	return &KafkaProducer{
		writer: writer,
		logger: logger,
		topic:  topic,
	}
}

// SendPaymentEvent отправляет событие оплаты в Kafka
func (p *KafkaProducer) SendPaymentEvent(ctx context.Context, event *PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal payment event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.OrderID),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to send payment event to Kafka",
			zap.Error(err),
			zap.String("order_id", event.OrderID),
			zap.String("topic", p.topic),
		)
		return fmt.Errorf("failed to send payment event: %w", err)
	}

	p.logger.Info("Payment event sent to Kafka",
		zap.String("order_id", event.OrderID),
		zap.String("topic", p.topic),
	)

	return nil
}

// Close закрывает соединение с Kafka
func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

