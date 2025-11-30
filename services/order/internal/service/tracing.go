package service

import (
	"context"

	platformtracing "github.com/t4RG3T21/GoBigTech/platform/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracerProvider = platformtracing.NewTracerProvider("order-service")

// withSpan создает span для операции и добавляет атрибуты
func withSpan(ctx context.Context, spanName string, attributes map[string]string, fn func(context.Context) error) error {
	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for key, value := range attributes {
		attrs = append(attrs, attribute.String(key, value))
	}
	ctx, span := tracerProvider.StartSpan(ctx, spanName, attrs...)
	defer span.End()

	return fn(ctx)
}

// createOrderSpan создает span для создания заказа
func createOrderSpan(ctx context.Context, userID string, itemsCount int) (context.Context, trace.Span) {
	attributes := []attribute.KeyValue{
		attribute.String("user.id", userID),
		attribute.Int("order.items.count", itemsCount),
	}

	return tracerProvider.StartSpan(ctx, "order.create", attributes...)
}

// inventoryCheckSpan создает span для проверки инвентаря
func inventoryCheckSpan(ctx context.Context, productID string, quantity int32) (context.Context, trace.Span) {
	attributes := []attribute.KeyValue{
		attribute.String("product.id", productID),
		attribute.Int("product.quantity", int(quantity)),
	}

	return tracerProvider.StartSpan(ctx, "inventory.check", attributes...)
}

// paymentProcessingSpan создает span для обработки платежа
func paymentProcessingSpan(ctx context.Context, orderID string, amount float64) (context.Context, trace.Span) {
	attributes := []attribute.KeyValue{
		attribute.String("order.id", orderID),
		attribute.Float64("payment.amount", amount),
	}

	return tracerProvider.StartSpan(ctx, "payment.process", attributes...)
}

// databaseSaveSpan создает span для сохранения заказа в БД
func databaseSaveSpan(ctx context.Context, orderID string) (context.Context, trace.Span) {
	ctx, span := tracerProvider.DatabaseSpan(ctx, "create", "postgresql", "orders")
	span.SetAttributes(attribute.String("order.id", orderID))
	return ctx, span
}

// kafkaSendSpan создает span для отправки сообщения в Kafka
func kafkaSendSpan(ctx context.Context, topic string, orderID string) (context.Context, trace.Span) {
	ctx, span := tracerProvider.MessagingSpan(ctx, "kafka", topic, "topic")
	span.SetAttributes(attribute.String("order.id", orderID))
	return ctx, span
}
