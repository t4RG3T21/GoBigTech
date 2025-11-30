# Трассировка Order Service через OpenTelemetry

Order Service настроен для полной трассировки запросов через OpenTelemetry с экспортом в Jaeger.

## Архитектура трассировки

```
HTTP Request → TracingMiddleware → OrderHandler → OrderService
                                                      ↓
                                    Inventory Service (gRPC с трассировкой)
                                                      ↓
                                    Payment Service (gRPC с трассировкой)
                                                      ↓
                                    Database Save (span)
                                                      ↓
                                    Kafka Send (span)
```

## Компоненты трассировки

### 1. HTTP Middleware (`internal/api/tracing_middleware.go`)
- Автоматически создает span для каждого HTTP запроса
- Извлекает контекст трассировки из HTTP заголовков
- Добавляет атрибуты: метод, путь, статус код, user agent

### 2. gRPC Client Interceptors (`internal/di/container.go`)
- Автоматически передает контекст трассировки через gRPC вызовы
- Использует `otelgrpc.NewClientHandler()` для инструментации клиентов
- Обеспечивает сквозную трассировку между микросервисами

### 3. Бизнес-логика Spans (`internal/service/order_service.go`)
- `order.create` - основной span для создания заказа
- `inventory.check` - проверка доступности товаров
- `payment.process` - обработка платежа
- `database.save_order` - сохранение заказа в БД
- `kafka.send_message` - отправка события в Kafka

## Конфигурация

### Переменные окружения

- `OTEL_EXPORTER_OTLP_ENDPOINT` - endpoint OpenTelemetry Collector (по умолчанию: `localhost:4317`)
  - Для локального запуска: `localhost:4317`
  - Для Docker: `otel-collector:4317`

### OpenTelemetry Collector

Трейсы экспортируются в Jaeger через OpenTelemetry Collector:
- Collector принимает OTLP на порту `4317` (gRPC)
- Collector экспортирует трейсы в Jaeger на порту `14250` (gRPC)

### Jaeger UI

Jaeger UI доступен по адресу: `http://localhost:16686`

## Использование

### Локальный запуск

1. Запустите OpenTelemetry Collector и Jaeger:
   ```bash
   docker-compose up -d otel-collector jaeger
   ```

2. Запустите Order Service:
   ```bash
   cd services/order
   go run cmd/order/main.go
   ```

3. Откройте Jaeger UI: `http://localhost:16686`

4. Выполните запрос на создание заказа:
   ```bash
   curl -X POST http://localhost:8080/orders \
     -H "Content-Type: application/json" \
     -d '{"user_id": "user123", "items": [{"product_id": "prod1", "quantity": 2}]}'
   ```

5. В Jaeger UI найдите трейсы для сервиса `order-service`

### Просмотр трейсов в Jaeger

1. Откройте Jaeger UI: `http://localhost:16686`
2. Выберите сервис: `order-service`
3. Нажмите "Find Traces"
4. Вы увидите полный путь запроса:
   - HTTP Request
   - Order Create
   - Inventory Check (для каждого товара)
   - Payment Process
   - Database Save
   - Kafka Send

## Структура трейса

Каждый трейс содержит следующие спаны:

1. **http.request** (root span)
   - Атрибуты: HTTP метод, путь, статус код
   - Длительность: весь HTTP запрос

2. **order.create** (child of http.request)
   - Атрибуты: user ID, количество товаров, order ID
   - Длительность: весь процесс создания заказа

3. **inventory.check** (child of order.create, для каждого товара)
   - Атрибуты: product ID, quantity
   - Длительность: проверка доступности товара

4. **payment.process** (child of order.create)
   - Атрибуты: order ID, amount, transaction ID
   - Длительность: обработка платежа

5. **database.save_order** (child of order.create)
   - Атрибуты: order ID, операция, система БД
   - Длительность: сохранение в PostgreSQL

6. **kafka.send_message** (child of order.create)
   - Атрибуты: topic, order ID
   - Длительность: отправка сообщения в Kafka

## Передача контекста между сервисами

Контекст трассировки автоматически передается:
- Через HTTP заголовки (TraceContext propagation)
- Через gRPC metadata (автоматически через otelgrpc)

Это обеспечивает сквозную трассировку через все микросервисы.

## Отладка

Если трейсы не появляются в Jaeger:

1. Проверьте, что OpenTelemetry Collector запущен:
   ```bash
   docker ps | grep otel-collector
   ```

2. Проверьте логи Collector:
   ```bash
   docker logs gobigtech_otel_collector
   ```

3. Проверьте, что Jaeger запущен:
   ```bash
   docker ps | grep jaeger
   ```

4. Проверьте endpoint в Order Service:
   - Убедитесь, что `OTEL_EXPORTER_OTLP_ENDPOINT` правильный
   - Для локального запуска: `localhost:4317`
   - Для Docker: `otel-collector:4317`

5. Проверьте логи Order Service на наличие ошибок OpenTelemetry

