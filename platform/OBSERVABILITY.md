# Платформенные компоненты наблюдаемости

Платформенная библиотека предоставляет переиспользуемые компоненты для наблюдаемости во всех микросервисах.

## Компоненты

### 1. Logger (`platform/logger`)

Уже существующий компонент для логирования с использованием zap.

**Использование:**
```go
import platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"

logger, err := platformlogger.New("my-service", "info")
if err != nil {
    log.Fatal(err)
}
defer logger.Sync()

logger.Info("Service started")
```

### 2. Prometheus Metrics (`platform/metrics`)

Общие Prometheus метрики для HTTP запросов.

**Использование:**
```go
import platformmetrics "github.com/t4RG3T21/GoBigTech/platform/metrics"

// Создаем метрики для сервиса
httpMetrics := platformmetrics.NewHTTPMetrics("order")

// Записываем метрику запроса
httpMetrics.RecordRequest("POST", "/orders", "200", 0.123)
```

**Метрики:**
- `{service}_http_requests_total` - счетчик HTTP запросов (labels: method, endpoint, status)
- `{service}_http_request_duration_seconds` - гистограмма длительности запросов (labels: method, endpoint, status)

### 3. HTTP Metrics Middleware (`platform/httpmetrics`)

Middleware для автоматического сбора HTTP метрик.

**Использование:**
```go
import (
    platformmetrics "github.com/t4RG3T21/GoBigTech/platform/metrics"
    platformhttpmetrics "github.com/t4RG3T21/GoBigTech/platform/httpmetrics"
    "github.com/go-chi/chi/v5"
)

// Создаем метрики
httpMetrics := platformmetrics.NewHTTPMetrics("order")

// Создаем middleware
metricsMiddleware := platformhttpmetrics.MetricsMiddleware(httpMetrics)

// Применяем к роутеру
r := chi.NewRouter()
r.Use(metricsMiddleware)
```

### 4. OpenTelemetry Tracing (`platform/tracing`)

Утилиты для трассировки через OpenTelemetry.

#### Инициализация

**Использование:**
```go
import platformtracing "github.com/t4RG3T21/GoBigTech/platform/tracing"

// Создаем конфигурацию
cfg := platformtracing.DefaultConfig("order-service")
// Или кастомная конфигурация:
cfg := &platformtracing.Config{
    ServiceName:    "order-service",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    OTLPEndpoint:   "otel-collector:4317",
    Timeout:        10 * time.Second,
}

// Инициализируем OpenTelemetry
shutdown, err := platformtracing.InitOpenTelemetry(cfg)
if err != nil {
    log.Fatal(err)
}
defer shutdown(context.Background())
```

#### HTTP Tracing Middleware

**Использование:**
```go
import platformtracing "github.com/t4RG3T21/GoBigTech/platform/tracing"

// Создаем middleware для трассировки
tracingMiddleware := platformtracing.TracingMiddleware("order-service-http")

// Применяем к роутеру (должен быть первым)
r := chi.NewRouter()
r.Use(tracingMiddleware)
```

#### Утилиты для создания спанов

**Использование:**
```go
import platformtracing "github.com/t4RG3T21/GoBigTech/platform/tracing"

// Создаем провайдер трассировки
tracerProvider := platformtracing.NewTracerProvider("order-service")

// Простой span
ctx, span := tracerProvider.StartSpan(ctx, "operation.name", 
    attribute.String("key", "value"),
)
defer span.End()

// Span для операции с БД
ctx, span := tracerProvider.DatabaseSpan(ctx, "create", "postgresql", "orders")
defer span.End()

// Span для отправки сообщения
ctx, span := tracerProvider.MessagingSpan(ctx, "kafka", "orders.payment", "topic")
defer span.End()

// Span для gRPC клиентского вызова
ctx, span := tracerProvider.RPCClientSpan(ctx, "inventory-service", "ReserveStock")
defer span.End()

// Span для gRPC серверного вызова
ctx, span := tracerProvider.RPCServerSpan(ctx, "order-service", "CreateOrder")
defer span.End()
```

## Полный пример использования

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
    platformmetrics "github.com/t4RG3T21/GoBigTech/platform/metrics"
    platformhttpmetrics "github.com/t4RG3T21/GoBigTech/platform/httpmetrics"
    platformtracing "github.com/t4RG3T21/GoBigTech/platform/tracing"
)

func main() {
    // 1. Инициализация логгера
    logger, _ := platformlogger.New("my-service", "info")
    defer logger.Sync()

    // 2. Инициализация OpenTelemetry
    otelConfig := platformtracing.DefaultConfig("my-service")
    otelShutdown, _ := platformtracing.InitOpenTelemetry(otelConfig)
    defer otelShutdown(context.Background())

    // 3. Создание метрик
    httpMetrics := platformmetrics.NewHTTPMetrics("my")

    // 4. Создание middleware
    tracingMiddleware := platformtracing.TracingMiddleware("my-service-http")
    metricsMiddleware := platformhttpmetrics.MetricsMiddleware(httpMetrics)

    // 5. Настройка роутера
    r := chi.NewRouter()
    r.Use(tracingMiddleware)  // Первым для захвата всего запроса
    r.Use(metricsMiddleware)

    // 6. Endpoints
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    r.Get("/metrics", promhttp.Handler().ServeHTTP)

    // 7. Запуск сервера
    http.ListenAndServe(":8080", r)
}
```

## Переменные окружения

- `OTEL_EXPORTER_OTLP_ENDPOINT` - endpoint OpenTelemetry Collector (по умолчанию: `localhost:4317`)
- `ENVIRONMENT` - окружение (development, production, staging) (по умолчанию: `development`)

## Зависимости

Все необходимые зависимости уже включены в `platform/go.mod`:
- `github.com/prometheus/client_golang` - для Prometheus метрик
- `go.opentelemetry.io/otel/*` - для OpenTelemetry трассировки

## Миграция существующих сервисов

Для миграции существующего сервиса:

1. Замените локальные метрики на `platform/metrics`
2. Замените HTTP middleware на `platform/httpmetrics`
3. Замените инициализацию OpenTelemetry на `platform/tracing.InitOpenTelemetry`
4. Замените HTTP tracing middleware на `platform/tracing.TracingMiddleware`
5. Используйте `platform/tracing.NewTracerProvider` для создания спанов

