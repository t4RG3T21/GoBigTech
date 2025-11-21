# GoBigTech - Microservices Project

Микросервисная архитектура на Go с использованием gRPC и HTTP.

## Архитектура

- **Order Service** (HTTP) - управление заказами, порт 8080
- **Inventory Service** (gRPC) - управление складом, порт 50051
- **Payment Service** (gRPC) - обработка платежей, порт 50052

## Быстрый старт

### 1. Запуск зависимостей (PostgreSQL и MongoDB)

```bash
docker-compose up -d
```

Это запустит:
- PostgreSQL на порту 5432
- MongoDB на порту 27017

### 2. Запуск сервисов

В отдельных терминалах:

**Inventory Service:**
```bash
cd services/inventory/cmd/inventory
go run main.go
```

**Payment Service:**
```bash
cd services/payment/cmd/payment
go run main.go
```

**Order Service:**
```bash
cd services/order/cmd/order
go run main.go
```

### 3. Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Создание заказа
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

## Структура проекта

```
.
├── docker-compose.yml          # Docker Compose для зависимостей
├── services/
│   ├── order/                  # Order Service (HTTP)
│   ├── inventory/              # Inventory Service (gRPC)
│   └── payment/                # Payment Service (gRPC)
└── platform/
    └── logger/                 # Платформенная библиотека логирования
```

## Конфигурация

Все сервисы используют переменные окружения с разумными значениями по умолчанию. Подробнее см. `services/START_SERVICES.md`.

## Тестирование

```bash
# Unit тесты
go test ./services/...

# Интеграционные тесты (требуют запущенные БД)
go test -tags=integration ./services/order/internal/repository

# E2E тесты (требуют запущенные БД и сервисы)
go test -tags=e2e ./services/inventory/internal/e2e
```

## Остановка

```bash
# Остановить Docker контейнеры
docker-compose down

# Остановить с удалением данных
docker-compose down -v
```

