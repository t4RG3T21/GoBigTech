package service

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("order-service-otel") // Изменим имя meter чтобы избежать конфликтов

type OrderMetrics struct {
	// OpenTelemetry метрики (используем другие имена)
	ordersCreatedOTel    metric.Int64Counter
	ordersFailedOTel     metric.Int64Counter
	orderTotalAmountOTel metric.Float64Counter
	orderDurationOTel    metric.Float64Histogram
	activeOrdersOTel     metric.Int64UpDownCounter

	// Метрики для HTTP запросов
	httpRequestsOTel        metric.Int64Counter
	httpRequestDurationOTel metric.Float64Histogram

	// Метрики для интеграций
	grpcCallsOTel        metric.Int64Counter
	grpcCallDurationOTel metric.Float64Histogram
	kafkaMessagesOTel    metric.Int64Counter
}

func NewOrderMetrics() (*OrderMetrics, error) {
	ordersCreated, err := meter.Int64Counter("order.otel.orders_created",
		metric.WithDescription("Total number of orders created (OTel)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	ordersFailed, err := meter.Int64Counter("order.otel.orders_failed",
		metric.WithDescription("Total number of failed orders (OTel)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	orderTotalAmount, err := meter.Float64Counter("order.otel.total_revenue",
		metric.WithDescription("Total revenue from all orders (OTel)"),
		metric.WithUnit("USD"),
	)
	if err != nil {
		return nil, err
	}

	orderDuration, err := meter.Float64Histogram("order.otel.processing_duration",
		metric.WithDescription("Time taken to process orders (OTel)"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	activeOrders, err := meter.Int64UpDownCounter("order.otel.active_orders",
		metric.WithDescription("Number of currently active orders (OTel)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	httpRequests, err := meter.Int64Counter("order.otel.http_requests",
		metric.WithDescription("Total HTTP requests (OTel)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestDuration, err := meter.Float64Histogram("order.otel.http_request_duration",
		metric.WithDescription("HTTP request duration (OTel)"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	grpcCalls, err := meter.Int64Counter("order.otel.grpc_calls",
		metric.WithDescription("Total gRPC calls to other services (OTel)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	grpcCallDuration, err := meter.Float64Histogram("order.otel.grpc_call_duration",
		metric.WithDescription("gRPC call duration (OTel)"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	kafkaMessages, err := meter.Int64Counter("order.otel.kafka_messages",
		metric.WithDescription("Total Kafka messages sent (OTel)"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	return &OrderMetrics{
		ordersCreatedOTel:       ordersCreated,
		ordersFailedOTel:        ordersFailed,
		orderTotalAmountOTel:    orderTotalAmount,
		orderDurationOTel:       orderDuration,
		activeOrdersOTel:        activeOrders,
		httpRequestsOTel:        httpRequests,
		httpRequestDurationOTel: httpRequestDuration,
		grpcCallsOTel:           grpcCalls,
		grpcCallDurationOTel:    grpcCallDuration,
		kafkaMessagesOTel:       kafkaMessages,
	}, nil
}

// RecordOrderSuccess записывает метрики успешного заказа
func (m *OrderMetrics) RecordOrderSuccess(amount float64, duration time.Duration) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("status", "success"),
	}

	ctx := context.Background()
	m.ordersCreatedOTel.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.orderTotalAmountOTel.Add(ctx, amount, metric.WithAttributes(attrs...))
	m.orderDurationOTel.Record(ctx, duration.Seconds()*1000, metric.WithAttributes(attrs...))
	m.activeOrdersOTel.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordOrderFailure записывает метрики неудачного заказа
func (m *OrderMetrics) RecordOrderFailure(reason string, duration time.Duration) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("status", "failed"),
		attribute.String("reason", reason),
	}

	ctx := context.Background()
	m.ordersFailedOTel.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.orderDurationOTel.Record(ctx, duration.Seconds()*1000, metric.WithAttributes(attrs...))
}

// RecordHTTPRequest записывает метрики HTTP запросов
func (m *OrderMetrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", status),
	}

	ctx := context.Background()
	m.httpRequestsOTel.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.httpRequestDurationOTel.Record(ctx, duration.Seconds()*1000, metric.WithAttributes(attrs...))
}

// RecordGRPCCall записывает метрики gRPC вызовов
func (m *OrderMetrics) RecordGRPCCall(service, method, status string, duration time.Duration) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("service", service),
		attribute.String("method", method),
		attribute.String("status", status),
	}

	ctx := context.Background()
	m.grpcCallsOTel.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.grpcCallDurationOTel.Record(ctx, duration.Seconds()*1000, metric.WithAttributes(attrs...))
}

// RecordKafkaMessage записывает метрики отправки Kafka сообщений
func (m *OrderMetrics) RecordKafkaMessage(topic string) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("topic", topic),
	}

	ctx := context.Background()
	m.kafkaMessagesOTel.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// CompleteOrder отмечает завершение активного заказа
func (m *OrderMetrics) CompleteOrder() {
	if m == nil {
		return
	}

	ctx := context.Background()
	m.activeOrdersOTel.Add(ctx, -1)
}
