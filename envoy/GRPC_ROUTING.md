# gRPC Routing Configuration in Envoy

## Обзор

Envoy настроен для маршрутизации gRPC трафика с поддержкой REST→gRPC трансформации.

## Listeners

### 1. HTTP Listener (порт 8080)
- Обрабатывает HTTP запросы к HTTP сервисам
- Маршрутизация: `/api/orders`, `/api/iam`, `/api/notifications`

### 2. gRPC Listener (порт 9090, внешний 9091)
- Обрабатывает нативный gRPC трафик
- Поддержка gRPC-Web для браузеров
- Маршрутизация:
  - `/iam.v1.IAMService/*` → `iam-service:50053`
  - `/inventory.v1.InventoryService/*` → `inventory-service:50051`
  - `/payment.v1.PaymentService/*` → `payment-service:50052`

### 3. gRPC REST Gateway Listener (порт 8084)
- Обрабатывает REST запросы и трансформирует их в gRPC
- Использует `grpc_json_transcoder` для преобразования
- Маршрутизация:
  - `/inventory.v1.InventoryService/*` → `inventory-service:50051` (gRPC)
  - `/payment.v1.PaymentService/*` → `payment-service:50052` (gRPC)
- **Примечание**: Порт 8081 занят Kafka UI, поэтому используется 8084

## REST→gRPC Трансформация

### Конфигурация

Используется фильтр `envoy.filters.http.grpc_json_transcoder`:
- **Proto descriptor**: `/etc/envoy/proto/inventory_payment.pb`
- **Services**: 
  - `inventory.v1.InventoryService`
  - `payment.v1.PaymentService`

### Формат запросов

#### Inventory Service

**GetStock**:
```bash
# REST запрос
curl -X POST http://localhost:8084/inventory.v1.InventoryService/GetStock \
  -H "Content-Type: application/json" \
  -d '{"product_id": "prod-123"}'

# Эквивалентный gRPC вызов
grpcurl -plaintext localhost:9091 inventory.v1.InventoryService/GetStock \
  -d '{"product_id": "prod-123"}'
```

**ReserveStock**:
```bash
# REST запрос
curl -X POST http://localhost:8084/inventory.v1.InventoryService/ReserveStock \
  -H "Content-Type: application/json" \
  -d '{"product_id": "prod-123", "quantity": 5}'
```

#### Payment Service

**ProcessPayment**:
```bash
# REST запрос
curl -X POST http://localhost:8084/payment.v1.PaymentService/ProcessPayment \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "order-123",
    "user_id": "user-456",
    "amount": 99.99,
    "method": "credit_card"
  }'
```

## Proto Descriptor

Для работы `grpc_json_transcoder` необходим скомпилированный proto descriptor файл.

### Генерация descriptor

```bash
# Установка protoc и protoc-gen-grpc-gateway
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Генерация descriptor для inventory и payment
protoc \
  --descriptor_set_out=envoy/proto/inventory_payment.pb \
  --include_imports \
  --include_source_info \
  api/proto/inventory/v1/inventory.proto \
  api/proto/payment/v1/payment.proto
```

### Структура файлов

```
envoy/
├── envoy.yaml
├── proto/
│   └── inventory_payment.pb  # Proto descriptor для трансформации
└── lua/
```

## HTTP/2 Поддержка

Все gRPC кластеры настроены с `http2_protocol_options: {}` для поддержки HTTP/2, который требуется для gRPC.

## Использование

### Нативный gRPC (порт 9091)

```bash
# Inventory Service
grpcurl -plaintext localhost:9091 \
  inventory.v1.InventoryService/GetStock \
  -d '{"product_id": "prod-123"}'

# Payment Service
grpcurl -plaintext localhost:9091 \
  payment.v1.PaymentService/ProcessPayment \
  -d '{"order_id": "order-123", "user_id": "user-456", "amount": 99.99, "method": "credit_card"}'
```

### REST→gRPC (порт 8084)

```bash
# Inventory Service через REST
curl -X POST http://localhost:8084/inventory.v1.InventoryService/GetStock \
  -H "Content-Type: application/json" \
  -d '{"product_id": "prod-123"}'

# Payment Service через REST
curl -X POST http://localhost:8084/payment.v1.PaymentService/ProcessPayment \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "order-123",
    "user_id": "user-456",
    "amount": 99.99,
    "method": "credit_card"
  }'
```

## Troubleshooting

### Ошибка: "proto descriptor not found"

Убедитесь, что:
1. Файл `envoy/proto/inventory_payment.pb` существует
2. Volume правильно смонтирован в docker-compose.yml
3. Файл имеет правильные права доступа

### Ошибка: "service not found in proto descriptor"

Проверьте:
1. Proto descriptor включает нужные сервисы
2. Имена сервисов в конфигурации совпадают с proto файлами
3. Proto descriptor сгенерирован с `--include_imports`

### gRPC запросы не работают

Проверьте:
1. gRPC сервисы запущены и доступны
2. Порты правильно проброшены в docker-compose.yml
3. Кластеры правильно настроены в envoy.yaml
4. HTTP/2 включен для gRPC кластеров

