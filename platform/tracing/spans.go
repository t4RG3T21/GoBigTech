package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider предоставляет утилиты для создания спанов
type TracerProvider struct {
	tracer trace.Tracer
}

// NewTracerProvider создает новый провайдер трассировки для указанного сервиса
func NewTracerProvider(serviceName string) *TracerProvider {
	return &TracerProvider{
		tracer: otel.Tracer(serviceName),
	}
}

// StartSpan создает новый span с указанным именем и атрибутами
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// StartSpanWithKind создает новый span с указанным именем, kind и атрибутами
func (tp *TracerProvider) StartSpanWithKind(ctx context.Context, name string, kind trace.SpanKind, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, trace.WithAttributes(attrs...), trace.WithSpanKind(kind))
}

// DatabaseSpan создает span для операции с базой данных
func (tp *TracerProvider) DatabaseSpan(ctx context.Context, operation, system, table string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.system", system),
	}
	if table != "" {
		attrs = append(attrs, attribute.String("db.sql.table", table))
	}
	return tp.StartSpan(ctx, "db."+operation, attrs...)
}

// MessagingSpan создает span для отправки сообщения в messaging систему
func (tp *TracerProvider) MessagingSpan(ctx context.Context, system, destination, destinationKind string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("messaging.system", system),
		attribute.String("messaging.destination", destination),
		attribute.String("messaging.destination_kind", destinationKind),
	}
	return tp.StartSpan(ctx, "messaging.send", attrs...)
}

// RPCClientSpan создает span для gRPC клиентского вызова
func (tp *TracerProvider) RPCClientSpan(ctx context.Context, service, method string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		semconv.RPCService(service),
		semconv.RPCMethod(method),
	}
	return tp.StartSpanWithKind(ctx, "rpc."+method, trace.SpanKindClient, attrs...)
}

// RPCServerSpan создает span для gRPC серверного вызова
func (tp *TracerProvider) RPCServerSpan(ctx context.Context, service, method string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		semconv.RPCService(service),
		semconv.RPCMethod(method),
	}
	return tp.StartSpanWithKind(ctx, "rpc."+method, trace.SpanKindServer, attrs...)
}
