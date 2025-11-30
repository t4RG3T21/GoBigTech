package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config содержит конфигурацию для инициализации OpenTelemetry
type Config struct {
	// ServiceName - имя сервиса (обязательно)
	ServiceName string

	// ServiceVersion - версия сервиса (по умолчанию: "1.0.0")
	ServiceVersion string

	// Environment - окружение (development, production, staging)
	Environment string

	// OTLPEndpoint - endpoint OpenTelemetry Collector (по умолчанию: "localhost:4317")
	OTLPEndpoint string

	// Timeout - таймаут для экспортеров (по умолчанию: 10s)
	Timeout time.Duration
}

// DefaultConfig создает конфигурацию с разумными значениями по умолчанию
func DefaultConfig(serviceName string) *Config {
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4317"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return &Config{
		ServiceName:    serviceName,
		ServiceVersion: "1.0.0",
		Environment:    env,
		OTLPEndpoint:   otlpEndpoint,
		Timeout:        10 * time.Second,
	}
}

// InitOpenTelemetry инициализирует OpenTelemetry tracer и metrics
// Возвращает функцию для graceful shutdown
func InitOpenTelemetry(cfg *Config) (func(context.Context) error, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	ctx := context.Background()

	// Устанавливаем значения по умолчанию
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "1.0.0"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.OTLPEndpoint == "" {
		cfg.OTLPEndpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		if cfg.OTLPEndpoint == "" {
			cfg.OTLPEndpoint = "localhost:4317"
		}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	// Создаем ресурс с информацией о сервисе
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Создаем экспортер для трейсов
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(cfg.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Создаем Tracer Provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Устанавливаем глобальный Tracer Provider
	otel.SetTracerProvider(tracerProvider)

	// Создаем экспортер для метрик
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithTimeout(cfg.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Создаем Meter Provider
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(10*time.Second))),
	)

	// Устанавливаем глобальный Meter Provider
	otel.SetMeterProvider(meterProvider)

	// Устанавливаем propagator для распространения трассировки между сервисами
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Возвращаем функцию для shutdown обоих провайдеров
	return func(ctx context.Context) error {
		var errs []error
		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown meter provider: %w", err))
		}
		if len(errs) > 0 {
			return fmt.Errorf("errors during shutdown: %v", errs)
		}
		return nil
	}, nil
}
