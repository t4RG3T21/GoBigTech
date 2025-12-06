# GoBigTech - Microservices Platform

Микросервисная платформа на Go с использованием Envoy API Gateway, gRPC, HTTP, Kafka и различных баз данных.

## 🏗️ Архитектура

Проект состоит из 6 микросервисов, объединенных через **Envoy API Gateway**:

```
┌─────────────────────────────────────────────────────────────┐
│                    Envoy API Gateway                        │
│  Port 80 (HTTP) | 9091 (gRPC) | 8084 (REST→gRPC Gateway)   │
│  - Authentication (Lua)                                    │
│  - Routing & Load Balancing                                 │
│  - REST→gRPC Transcoding                                    │
└─────────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
        ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ Order Service│  │  IAM Service │  │Notification  │
│   (HTTP)     │  │  (gRPC/HTTP) │  │   Service    │
│   :8080      │  │   :8082/50053│  │   (Kafka)    │
└──────────────┘  └──────────────┘  └──────────────┘
        │                 │
        │                 │
        ▼                 ▼
┌──────────────┐  ┌──────────────┐
│Inventory Svc │  │ Payment Svc  │
│   (gRPC)     │  │   (gRPC)     │
│   :50051     │  │   :50052     │
└──────────────┘  └──────────────┘
        │
        ▼
┌──────────────┐
│Assembly Svc  │
│   (Kafka)    │
└──────────────┘
```

### Core Services

- **Envoy API Gateway** - единая точка входа для всех HTTP запросов
  - Порт **80** - HTTP API (основной вход)
  - Порт **9091** - нативный gRPC
  - Порт **8084** - REST→gRPC Gateway (трансформация REST в gRPC)
  - Lua скрипты для аутентификации сессий
  - Маршрутизация, retry policies, timeouts

- **IAM Service** (gRPC + HTTP Gateway) - управление пользователями и аутентификацией
  - PostgreSQL для хранения пользователей
  - Redis для хранения сессий
  - Поддержка регистрации, входа по email/username, валидации сессий
  - HTTP Gateway на порту **8082** (через Envoy: `/api/iam/v1/iam/*`)
  - gRPC на порту **50053** (через Envoy: порт **9091**)

- **Inventory Service** (gRPC) - управление складом и запасами
  - MongoDB для хранения данных о товарах
  - Интеграция с IAM Service для аутентификации
  - Поддержка резервирования товаров
  - gRPC на порту **50051** (через Envoy: порт **9091** или REST→gRPC на **8084**)

- **Payment Service** (gRPC) - обработка платежей
  - Обработка платежных транзакций
  - gRPC на порту **50052** (через Envoy: порт **9091** или REST→gRPC на **8084**)

- **Order Service** (HTTP) - управление заказами
  - PostgreSQL для хранения заказов
  - Интеграция с Inventory и Payment сервисами через gRPC
  - Kafka producer для отправки событий
  - HTTP на порту **8080** (через Envoy: `/api/orders/*`)

### Event-Driven Services

- **Assembly Service** (Kafka Consumer) - сборка заказов
  - Потребляет события из топика `orders.payment`
  - Отправляет события в топик `orders.assembly`

- **Notification Service** (Kafka Consumer) - уведомления
  - Потребляет события из топиков `orders.payment` и `orders.assembly`
  - Отправляет уведомления через Telegram Bot API

## 🚀 Быстрый старт

### Предварительные требования

- **Docker** и **Docker Compose** (обязательно)
- Go 1.25.1 или выше (только для локальной разработки)
- protoc (для генерации proto файлов) или Docker

### 1. Запуск всей системы

```bash
docker-compose up -d
```

Это запустит:
- **PostgreSQL** - для Order и IAM сервисов
- **MongoDB** - для Inventory сервиса
- **Redis** - для IAM сервиса (сессии)
- **Kafka** - для event streaming
- **Kafka UI** - веб-интерфейс для Kafka (http://localhost:8081)
- **Envoy API Gateway** - единая точка входа (http://localhost:80)
- **Jaeger** - распределенная трассировка (http://localhost:16686)
- **Prometheus** - сбор метрик (http://localhost:9090)
- **Grafana** - визуализация метрик (http://localhost:3000, admin/admin)
- **Elasticsearch + Kibana** - логи (http://localhost:5601)
- **OpenTelemetry Collector** - сбор телеметрии
- Все микросервисы в Docker контейнерах

### 2. Проверка работоспособности

Дождитесь, пока все сервисы станут healthy:

```bash
docker ps
```

Все сервисы должны иметь статус `(healthy)` или `(health: starting)`.

### 3. Тестирование через Envoy

Используйте скрипт для автоматического тестирования:

```powershell
# Windows PowerShell
.\test-envoy.ps1

# С подробным выводом
.\test-envoy.ps1 -Verbose
```

Или тестируйте вручную через curl/Postman:

```bash
# Health check
curl http://localhost/health

# Регистрация пользователя
curl -X POST http://localhost/api/iam/v1/iam/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "username": "testuser"
  }'

# Вход пользователя
curl -X POST http://localhost/api/iam/v1/iam/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# Создание заказа (требуется токен из login)
curl -X POST http://localhost/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_SESSION_TOKEN" \
  -d '{
    "user_id": "user123",
    "items": [
      {"product_id": "prod-123", "quantity": 2},
      {"product_id": "prod-456", "quantity": 1}
    ]
  }'
```

## 📡 API через Envoy

### HTTP API (порт 80)

Все HTTP запросы идут через Envoy на порт **80**:

| Endpoint | Метод | Описание | Аутентификация |
|----------|-------|----------|----------------|
| `/health` | GET | Health check | Нет |
| `/api/iam/v1/iam/register` | POST | Регистрация пользователя | Нет |
| `/api/iam/v1/iam/login` | POST | Вход пользователя | Нет |
| `/api/iam/v1/iam/validate-session` | POST | Валидация сессии | Да |
| `/api/orders` | POST | Создание заказа | Да |
| `/api/orders/{id}` | GET | Получение заказа | Да |

### gRPC API (порт 9091)

Нативный gRPC через Envoy:

```bash
# Список сервисов
grpcurl -plaintext localhost:9091 list

# IAM Service
grpcurl -plaintext localhost:9091 list iam.v1.IAMService
grpcurl -plaintext -d '{"email":"user@example.com","password":"pass"}' \
  localhost:9091 iam.v1.IAMService/Login

# Inventory Service
grpcurl -plaintext localhost:9091 list inventory.v1.InventoryService
grpcurl -plaintext -H "session_id: TOKEN" -d '{"productId":"p1"}' \
  localhost:9091 inventory.v1.InventoryService/GetStock

# Payment Service
grpcurl -plaintext localhost:9091 list payment.v1.PaymentService
```

### REST→gRPC Gateway (порт 8084)

REST запросы автоматически преобразуются в gRPC:

```bash
# Inventory Service через REST
curl -X POST http://localhost:8084/inventory.v1.InventoryService/GetStock \
  -H "Content-Type: application/json" \
  -d '{"product_id": "prod-123"}'

# Payment Service через REST
curl -X POST http://localhost:8084/payment.v1.PaymentService/ProcessPayment \
  -H "Content-Type: application/json" \
  -d '{"order_id": "order-123", "amount": 100.50}'
```

## ⚙️ Конфигурация Envoy

Envoy настроен в файле `envoy/envoy.yaml`:

### Основные компоненты:

1. **HTTP Listener (порт 8080 → внешний 80)**
   - Маршрутизация HTTP запросов
   - Lua фильтр для аутентификации сессий
   - Retry policies и timeouts
   - Access logging в JSON формате

2. **gRPC Listener (порт 9090 → внешний 9091)**
   - Нативная поддержка gRPC
   - gRPC-Web поддержка
   - Маршрутизация gRPC вызовов

3. **REST→gRPC Gateway (порт 8084)**
   - `grpc_json_transcoder` для преобразования REST в gRPC
   - Использует proto descriptor файл `envoy/proto/inventory_payment.pb`

4. **Lua Authentication**
   - Скрипт `envoy/lua/verify_session.lua`
   - Проверяет токены через IAM Service
   - Добавляет заголовок `X-User-Id` для валидных запросов
   - Исключает публичные эндпоинты (login, register, health)

### Кластеры (upstream services):

- `order-service` → `order-service:8080`
- `iam-service` → `iam-service:8082`
- `iam-service-grpc` → `iam-service:50053`
- `inventory-service-grpc` → `inventory-service:50051`
- `payment-service-grpc` → `payment-service:50052`
- `notification-service` → `notification-service:8083`

Подробнее: [envoy/README.md](envoy/README.md)

## 📁 Структура проекта

```
.
├── api/
│   ├── proto/              # Protocol Buffers определения
│   │   ├── iam/v1/
│   │   ├── inventory/v1/
│   │   └── payment/v1/
│   └── openapi/            # OpenAPI спецификации
├── services/
│   ├── iam/                # IAM Service (gRPC + HTTP Gateway)
│   ├── inventory/          # Inventory Service (gRPC)
│   ├── payment/            # Payment Service (gRPC)
│   ├── order/              # Order Service (HTTP)
│   ├── assembly/           # Assembly Service (Kafka Consumer)
│   └── notification/        # Notification Service (Kafka Consumer)
├── envoy/
│   ├── envoy.yaml          # Конфигурация Envoy API Gateway
│   ├── lua/                # Lua скрипты для аутентификации
│   └── proto/              # Proto descriptor для REST→gRPC
├── platform/
│   ├── logger/             # Платформенная библиотека логирования
│   ├── metrics/            # Prometheus метрики
│   ├── tracing/            # OpenTelemetry трассировка
│   └── httpmetrics/        # HTTP метрики middleware
├── monitoring/             # Конфигурация мониторинга
│   ├── prometheus.yml
│   ├── otel-collector-config.yaml
│   └── grafana/
├── docker-compose.yml      # Docker Compose для всей системы
├── test-envoy.ps1          # Скрипт тестирования через Envoy
└── README.md              # Этот файл
```

## ⚙️ Конфигурация сервисов

Все сервисы используют переменные окружения с разумными значениями по умолчанию.

### IAM Service
- `GRPC_PORT` - порт gRPC сервера (по умолчанию: 50053)
- `HTTP_PORT` - порт HTTP Gateway (по умолчанию: 8082)
- `DATABASE_URL` - URL PostgreSQL
- `REDIS_ADDRESS` - адрес Redis
- `REDIS_PASSWORD` - пароль Redis
- `SESSION_TTL` - время жизни сессии (по умолчанию: 24h)

### Inventory Service
- `GRPC_PORT` - порт gRPC сервера (по умолчанию: 50051)
- `MONGODB_URL` - URL MongoDB
- `IAM_GRPC_ADDRESS` - адрес IAM Service (по умолчанию: iam-service:50053)

### Payment Service
- `GRPC_PORT` - порт gRPC сервера (по умолчанию: 50052)

### Order Service
- `HTTP_PORT` - порт HTTP сервера (по умолчанию: 8080)
- `DATABASE_URL` - URL PostgreSQL
- `INVENTORY_GRPC_ADDRESS` - адрес Inventory Service (по умолчанию: inventory-service:50051)
- `PAYMENT_GRPC_ADDRESS` - адрес Payment Service (по умолчанию: payment-service:50052)
- `KAFKA_BOOTSTRAP_SERVERS` - адреса Kafka брокеров (по умолчанию: kafka:29092)

### Notification Service
- `TELEGRAM_BOT_TOKEN` - токен Telegram бота
- `TELEGRAM_CHAT_ID` - ID чата для уведомлений
- `KAFKA_BOOTSTRAP_SERVERS` - адреса Kafka брокеров

### Assembly Service
- `KAFKA_BOOTSTRAP_SERVERS` - адреса Kafka брокеров

## 🧪 Тестирование

### Автоматическое тестирование через Envoy

```powershell
# Полный набор тестов
.\test-envoy.ps1

# С подробным выводом
.\test-envoy.ps1 -Verbose
```

Скрипт проверяет:
- ✅ Доступность всех сервисов через Envoy
- ✅ Аутентификацию (регистрация → логин → токен)
- ✅ Создание заказа с токеном
- ✅ Защиту эндпоинтов (401 без токена)
- ✅ gRPC маршрутизацию
- ✅ Ограничение прямого доступа к сервисам

### Unit тесты

```bash
# Все сервисы
go test ./services/...

# Конкретный сервис
go test ./services/order/...
```

### Интеграционные тесты

```bash
# Требуют запущенные БД
go test -tags=integration ./services/order/internal/repository
```

### E2E тесты

```bash
# Требуют запущенные БД и сервисы
go test -tags=e2e ./services/inventory/internal/e2e
```

### Покрытие кода

```bash
# Для конкретного сервиса
cd services/order
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 Мониторинг и Observability

Проект включает полный стек Observability:

### Метрики (Prometheus)
- **Prometheus**: http://localhost:9090 - сбор и хранение метрик
- **Grafana**: http://localhost:3000 - визуализация метрик (admin/admin)
- **Order Service метрики**: http://localhost:8080/metrics
- **OpenTelemetry Collector метрики**: http://localhost:8889/metrics

### Трассировка (Jaeger)
- **Jaeger UI**: http://localhost:16686 - просмотр трейсов запросов
- Трассировка настроена для всех HTTP и gRPC запросов
- Полная трассировка от HTTP входа через Envoy до вызовов Payment и Inventory сервисов

### Логи (Elasticsearch/Kibana)
- **Kibana**: http://localhost:5601 - просмотр и поиск логов
- **Elasticsearch**: http://localhost:9200 - хранение логов
- Логи собираются автоматически из всех Docker контейнеров через Filebeat
- Логи в JSON формате для удобного парсинга

### Алерты (Alertmanager)
- **Alertmanager**: http://localhost:9093 - управление алертами
- Алерты настроены для:
  - Высокой частоты заказов
  - Ошибок платежей
  - Медленной обработки заказов
  - Проблем с базой данных
- Уведомления отправляются в Telegram через Notification Service

### Другие инструменты
- **Kafka UI**: http://localhost:8081 - веб-интерфейс для управления Kafka
- **Envoy Admin**: http://localhost:9901 - административный интерфейс Envoy

Подробная документация:
- [monitoring/ELASTICSEARCH_SETUP.md](monitoring/ELASTICSEARCH_SETUP.md) - настройка Elasticsearch и Kibana
- [monitoring/ALERTING_SETUP.md](monitoring/ALERTING_SETUP.md) - настройка алертов
- [monitoring/GRAFANA_SETUP.md](monitoring/GRAFANA_SETUP.md) - настройка Grafana
- [platform/OBSERVABILITY.md](platform/OBSERVABILITY.md) - использование платформенных компонентов

## 🛠️ Разработка

### Генерация Proto файлов

Используйте Docker для генерации (рекомендуется на Windows):

```bash
# Генерация для всех сервисов
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.25-alpine sh -c \
  "apk add --no-cache protoc protobuf-dev git && \
   go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && \
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && \
   git clone --depth 1 https://github.com/googleapis/googleapis.git /tmp/googleapis && \
   protoc --go_out=services --go_opt=paths=source_relative \
          --go-grpc_out=services --go-grpc_opt=paths=source_relative \
          --proto_path=api/proto \
          --proto_path=/tmp/googleapis \
          api/proto/iam/v1/iam.proto \
          api/proto/inventory/v1/inventory.proto \
          api/proto/payment/v1/payment.proto"
```

### Генерация Proto Descriptor для Envoy

Для REST→gRPC трансформации:

```powershell
.\scripts\generate-proto-descriptor.ps1
```

Или вручную:

```bash
protoc --descriptor_set_out=envoy/proto/inventory_payment.pb \
       --include_imports \
       --include_source_info \
       api/proto/inventory/v1/inventory.proto \
       api/proto/payment/v1/payment.proto
```

### Использование grpcurl

Все gRPC сервисы поддерживают reflection API:

```bash
# Через Envoy
grpcurl -plaintext localhost:9091 list

# Напрямую к сервису (внутри Docker сети)
grpcurl -plaintext iam-service:50053 list
```

## 🐛 Troubleshooting

### Envoy не запускается

1. Проверьте логи:
   ```bash
   docker logs gobigtech_envoy
   ```

2. Проверьте конфигурацию:
   ```bash
   docker exec gobigtech_envoy cat /etc/envoy/envoy.yaml
   ```

3. Проверьте proto descriptor:
   ```bash
   ls -la envoy/proto/inventory_payment.pb
   ```

### Сервисы недоступны через Envoy

1. Проверьте, что сервисы запущены:
   ```bash
   docker ps | grep gobigtech
   ```

2. Проверьте маршрутизацию в Envoy:
   ```bash
   curl http://localhost:9901/config_dump | jq '.configs[2].dynamic_route_configs'
   ```

3. Проверьте логи Envoy:
   ```bash
   docker logs gobigtech_envoy --tail 50
   ```

### Ошибка 401 Unauthorized

1. Проверьте, что токен передается правильно:
   ```bash
   curl -v -H "Authorization: Bearer YOUR_TOKEN" http://localhost/api/orders
   ```

2. Проверьте логи IAM Service:
   ```bash
   docker logs gobigtech_iam_service --tail 50
   ```

3. Проверьте Lua скрипт:
   ```bash
   docker exec gobigtech_envoy cat /etc/envoy/lua/verify_session.lua
   ```

### gRPC не работает

1. Проверьте порты:
   ```bash
   docker ps | grep envoy
   # Должны быть: 80, 9091, 8084
   ```

2. Проверьте proto descriptor:
   ```bash
   docker exec gobigtech_envoy ls -la /etc/envoy/proto/
   ```

3. Тестируйте через grpcurl:
   ```bash
   grpcurl -plaintext localhost:9091 list
   ```

### База данных недоступна

1. Проверьте статус:
   ```bash
   docker ps | grep postgres
   docker ps | grep mongodb
   ```

2. Проверьте логи:
   ```bash
   docker logs gobigtech_postgres
   docker logs gobigtech_mongodb
   ```

3. Проверьте подключение:
   ```bash
   docker exec -it gobigtech_postgres psql -U postgres -c "SELECT 1"
   ```

### Kafka проблемы

1. Проверьте топики:
   ```bash
   docker exec gobigtech_kafka kafka-topics --list --bootstrap-server localhost:9092
   ```

2. Проверьте логи:
   ```bash
   docker logs gobigtech_kafka --tail 50
   ```

3. Используйте Kafka UI: http://localhost:8081

Подробнее: [envoy/TROUBLESHOOTING.md](envoy/TROUBLESHOOTING.md)

## 🛑 Остановка

```bash
# Остановить все контейнеры
docker-compose down

# Остановить с удалением данных
docker-compose down -v

# Остановить только сервисы (оставить БД)
docker-compose stop order-service iam-service inventory-service payment-service
```

## 📚 Дополнительная документация

- [TEST_ENVOY_README.md](TEST_ENVOY_README.md) - документация по тестированию через Envoy
- [QUICK_START_TEST.md](QUICK_START_TEST.md) - быстрый старт тестирования
- [envoy/README.md](envoy/README.md) - документация по Envoy конфигурации
- [envoy/TROUBLESHOOTING.md](envoy/TROUBLESHOOTING.md) - решение проблем с Envoy
- [README_KAFKA.md](README_KAFKA.md) - документация по Kafka
- [README_TESTING.md](README_TESTING.md) - документация по тестированию
- [services/order/README.md](services/order/README.md) - документация Order Service
- [services/notification/README.md](services/notification/README.md) - документация Notification Service

## 🔧 Технологии

### Основные
- **Go 1.25.1** - основной язык программирования
- **Envoy Proxy** - API Gateway и прокси-сервер
- **gRPC** - межсервисная коммуникация
- **Protocol Buffers** - сериализация данных
- **PostgreSQL** - реляционная БД
- **MongoDB** - документная БД
- **Redis** - кэш и хранение сессий
- **Kafka** - event streaming
- **Docker** - контейнеризация

### Observability
- **Prometheus** - сбор метрик
- **Grafana** - визуализация метрик
- **Jaeger** - распределенная трассировка
- **OpenTelemetry** - стандарт для телеметрии
- **Elasticsearch** - хранение логов
- **Kibana** - визуализация логов
- **Filebeat** - сбор логов
- **Alertmanager** - управление алертами

### API Gateway
- **Envoy Proxy** - высокопроизводительный прокси
- **Lua** - скрипты для аутентификации
- **gRPC JSON Transcoding** - REST→gRPC преобразование

## 📝 Лицензия

MIT

## 🤝 Вклад в проект

См. [CONTRIBUTING.md](CONTRIBUTING.md)
