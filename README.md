# GoBigTech - Microservices Project

Микросервисная архитектура на Go с использованием gRPC, HTTP, Kafka и различных баз данных.

## 🏗️ Архитектура

Проект состоит из 6 микросервисов:

### Core Services

- **IAM Service** (gRPC) - управление пользователями и аутентификацией, порт **50053**
  - PostgreSQL для хранения пользователей
  - Redis для хранения сессий
  - Поддержка регистрации, входа по email/username, валидации сессий

- **Inventory Service** (gRPC) - управление складом и запасами, порт **50051**
  - MongoDB для хранения данных о товарах
  - Интеграция с IAM Service для аутентификации
  - Поддержка резервирования товаров

- **Payment Service** (gRPC) - обработка платежей, порт **50052**
  - Обработка платежных транзакций

- **Order Service** (HTTP) - управление заказами, порт **8080**
  - PostgreSQL для хранения заказов
  - Интеграция с Inventory и Payment сервисами
  - Kafka producer для отправки событий

### Event-Driven Services

- **Assembly Service** (Kafka Consumer) - сборка заказов
  - Потребляет события из топика `orders.payment`
  - Отправляет события в топик `orders.assembly`

- **Notification Service** (Kafka Consumer) - уведомления
  - Потребляет события из топиков `orders.payment` и `orders.assembly`
  - Отправляет уведомления через Telegram Bot API

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.25.1 или выше
- Docker и Docker Compose
- protoc (для генерации proto файлов) или Docker

### 1. Запуск зависимостей

```bash
docker-compose up -d
```

Это запустит:
- **PostgreSQL** на порту 5432 (для Order и IAM сервисов)
- **MongoDB** на порту 27017 (для Inventory сервиса)
- **Redis** на порту 6379 (для IAM сервиса)
- **Kafka** на порту 9092
- **Kafka UI** на порту 8081

### 2. Генерация Proto файлов

Для генерации proto файлов используйте Docker (рекомендуется на Windows):

```bash
# Генерация для IAM Service
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.23-alpine sh -c "apk add --no-cache protoc protobuf-dev && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --proto_path=api/proto api/proto/iam/v1/iam.proto"

# Генерация для Inventory Service
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.23-alpine sh -c "apk add --no-cache protoc protobuf-dev && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --proto_path=api/proto api/proto/inventory/v1/inventory.proto"

# Генерация для Payment Service
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.23-alpine sh -c "apk add --no-cache protoc protobuf-dev && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --proto_path=api/proto api/proto/payment/v1/payment.proto"
```

Или используйте скрипт (см. `tools/generate-proto.ps1`).

### 3. Запуск сервисов

В отдельных терминалах запустите сервисы в следующем порядке:

**1. IAM Service (обязательно первым):**
```bash
cd services/iam/cmd/iam
go run main.go
```

**2. Inventory Service:**
```bash
cd services/inventory/cmd/inventory
go run main.go
```

**3. Payment Service:**
```bash
cd services/payment/cmd/payment
go run main.go
```

**4. Order Service:**
```bash
cd services/order/cmd/order
go run main.go
```

**5. Assembly Service:**
```bash
cd services/assembly/cmd/assembly
go run main.go
```

**6. Notification Service:**
```bash
cd services/notification/cmd/notification
go run main.go
```

### 4. Проверка работы

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Регистрация пользователя (IAM Service)
```bash
grpcurl -plaintext -d '{
  "email": "user@example.com",
  "password": "password123",
  "username": "testuser"
}' localhost:50053 iam.v1.IAMService/Register
```

#### Вход пользователя (IAM Service)
```bash
# По email
grpcurl -plaintext -d '{
  "email": "user@example.com",
  "password": "password123"
}' localhost:50053 iam.v1.IAMService/Login

# По username
grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "password123"
}' localhost:50053 iam.v1.IAMService/Login
```

#### Получение информации о товаре (Inventory Service)
```bash
# Требуется валидный session_id из IAM Service
grpcurl -plaintext -H "session_id: YOUR_SESSION_ID" -d '{
  "productId": "p1"
}' localhost:50051 inventory.v1.InventoryService/GetStock
```

#### Создание заказа (Order Service)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user123",
    "items": [
      {"productId": "p1", "quantity": 2},
      {"productId": "p2", "quantity": 1}
    ]
  }'
```

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
│   ├── iam/                # IAM Service (gRPC, порт 50053)
│   ├── inventory/           # Inventory Service (gRPC, порт 50051)
│   ├── payment/             # Payment Service (gRPC, порт 50052)
│   ├── order/               # Order Service (HTTP, порт 8080)
│   ├── assembly/           # Assembly Service (Kafka Consumer)
│   └── notification/        # Notification Service (Kafka Consumer)
├── platform/
│   └── logger/              # Платформенная библиотека логирования
├── tools/
│   ├── buf.gen.yaml         # Конфигурация для buf
│   └── generate-proto.ps1   # Скрипт генерации proto файлов
├── docker-compose.yml       # Docker Compose для зависимостей
├── go.work                  # Go workspace файл
└── README.md               # Этот файл
```

## ⚙️ Конфигурация

Все сервисы используют переменные окружения с разумными значениями по умолчанию.

### IAM Service
- `GRPC_PORT` - порт gRPC сервера (по умолчанию: 50053)
- `DATABASE_URL` - URL PostgreSQL (по умолчанию: postgres://postgres:password@localhost:5432/iam_service?sslmode=disable)
- `REDIS_ADDRESS` - адрес Redis (по умолчанию: localhost:6379)
- `REDIS_PASSWORD` - пароль Redis (по умолчанию: redis_password)
- `SESSION_TTL` - время жизни сессии (по умолчанию: 24h)

### Inventory Service
- `GRPC_PORT` - порт gRPC сервера (по умолчанию: 50051)
- `MONGODB_URL` - URL MongoDB (по умолчанию: mongodb://admin:password@localhost:27017)
- `IAM_GRPC_ADDRESS` - адрес IAM Service (по умолчанию: localhost:50053)

### Payment Service
- `GRPC_PORT` - порт gRPC сервера (по умолчанию: 50052)

### Order Service
- `HTTP_PORT` - порт HTTP сервера (по умолчанию: 8080)
- `DATABASE_URL` - URL PostgreSQL (по умолчанию: postgres://postgres:password@localhost:5432/order_service?sslmode=disable)
- `INVENTORY_GRPC_ADDRESS` - адрес Inventory Service (по умолчанию: localhost:50051)
- `PAYMENT_GRPC_ADDRESS` - адрес Payment Service (по умолчанию: localhost:50052)
- `KAFKA_BOOTSTRAP_SERVERS` - адреса Kafka брокеров (по умолчанию: localhost:9092)

### Notification Service
- `TELEGRAM_BOT_TOKEN` - токен Telegram бота (по умолчанию: установлен)
- `TELEGRAM_CHAT_ID` - ID чата для уведомлений (по умолчанию: 123456789)
- `KAFKA_BOOTSTRAP_SERVERS` - адреса Kafka брокеров (по умолчанию: localhost:9092)

### Assembly Service
- `KAFKA_BOOTSTRAP_SERVERS` - адреса Kafka брокеров (по умолчанию: localhost:9092)

## 🧪 Тестирование

### Unit тесты
```bash
go test ./services/...
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

## 🛠️ Разработка

### Генерация Proto файлов

Используйте Docker для генерации (рекомендуется на Windows):

```bash
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.23-alpine sh -c "apk add --no-cache protoc protobuf-dev && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --proto_path=api/proto api/proto/iam/v1/iam.proto api/proto/inventory/v1/inventory.proto api/proto/payment/v1/payment.proto"
```

### Использование grpcurl

Все gRPC сервисы поддерживают reflection API для работы с `grpcurl`:

```bash
# Список сервисов
grpcurl -plaintext localhost:50053 list

# Список методов
grpcurl -plaintext localhost:50053 list iam.v1.IAMService

# Описание метода
grpcurl -plaintext localhost:50053 describe iam.v1.IAMService.Register
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
- Полная трассировка от HTTP входа до вызовов Payment и Inventory сервисов

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

### Тестирование Observability стека

Запустите скрипт для проверки всех компонентов:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/test-observability.ps1
```

Подробная документация:
- [monitoring/ELASTICSEARCH_SETUP.md](monitoring/ELASTICSEARCH_SETUP.md) - настройка Elasticsearch и Kibana
- [monitoring/QUICK_START_ELASTICSEARCH.md](monitoring/QUICK_START_ELASTICSEARCH.md) - быстрый старт ELK стека
- [monitoring/ALERTING_SETUP.md](monitoring/ALERTING_SETUP.md) - настройка алертов
- [monitoring/GRAFANA_SETUP.md](monitoring/GRAFANA_SETUP.md) - настройка Grafana
- [platform/OBSERVABILITY.md](platform/OBSERVABILITY.md) - использование платформенных компонентов

## 🛑 Остановка

```bash
# Остановить Docker контейнеры
docker-compose down

# Остановить с удалением данных
docker-compose down -v
```

## 📚 Дополнительная документация

- [README_KAFKA.md](README_KAFKA.md) - документация по Kafka
- [README_TESTING.md](README_TESTING.md) - документация по тестированию
- [services/order/README.md](services/order/README.md) - документация Order Service
- [services/notification/README.md](services/notification/README.md) - документация Notification Service

## 🔧 Технологии

### Основные
- **Go 1.25.1** - основной язык программирования
- **gRPC** - межсервисная коммуникация
- **Protocol Buffers** - сериализация данных
- **PostgreSQL** - реляционная БД
- **MongoDB** - документная БД
- **Redis** - кэш и хранение сессий
- **Kafka** - event streaming
- **Docker** - контейнеризация зависимостей

### Observability
- **Prometheus** - сбор метрик
- **Grafana** - визуализация метрик
- **Jaeger** - распределенная трассировка
- **OpenTelemetry** - стандарт для телеметрии
- **Elasticsearch** - хранение логов
- **Kibana** - визуализация логов
- **Filebeat** - сбор логов
- **Alertmanager** - управление алертами

## 📝 Лицензия

MIT
