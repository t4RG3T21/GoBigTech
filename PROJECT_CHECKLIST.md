# 📋 Полная инструкция по проверке проекта GoBigTech

## Этап 1: Проверка зависимостей и генерация Proto файлов

### 1.1. Проверка наличия необходимых инструментов

```powershell
# Проверка Go версии
go version
# Должно быть: go version go1.25.1 или выше

# Проверка Docker
docker --version
docker-compose --version
```

### 1.2. Генерация Proto файлов (если нужно)

**Важно:** Proto файлы должны быть сгенерированы перед сборкой Docker образов.

```powershell
# Генерация proto файлов для всех сервисов через Docker
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.25-alpine sh -c "apk add --no-cache protoc protobuf-dev git && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --proto_path=api/proto api/proto/payment/v1/payment.proto && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --proto_path=api/proto api/proto/inventory/v1/inventory.proto && echo 'Proto files generated!'"

# Для IAM сервиса (требует googleapis)
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.25-alpine sh -c "apk add --no-cache protoc protobuf-dev git && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && if [ ! -d /tmp/googleapis ]; then git clone --depth 1 https://github.com/googleapis/googleapis.git /tmp/googleapis; fi && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --grpc-gateway_out=services --grpc-gateway_opt=paths=source_relative --openapiv2_out=services/iam/v1 --openapiv2_opt=logtostderr=true,allow_merge=true --proto_path=api/proto --proto_path=/tmp/googleapis api/proto/iam/v1/iam.proto && echo 'IAM proto files generated!'"
```

**Проверка сгенерированных файлов:**
```powershell
# Проверка наличия proto файлов
Test-Path services/payment/v1/payment_grpc.pb.go
Test-Path services/inventory/v1/inventory_grpc.pb.go
Test-Path services/iam/v1/iam_grpc.pb.go
Test-Path services/iam/v1/iam.pb.gw.go
```

### 1.3. Проверка go.mod файлов

Проверьте, что все сервисы имеют правильные `replace` директивы:

```powershell
# Проверка replace директив
Select-String -Path "services/*/go.mod" -Pattern "^replace"
```

**Ожидаемый результат:**
- Все сервисы должны иметь: `replace github.com/t4RG3T21/GoBigTech/platform => ../../platform`
- Inventory service должен иметь: `replace github.com/t4RG3T21/GoBigTech/services/iam => ../iam`

**Проверка require секций:**
```powershell
# Проверка наличия platform в require
Select-String -Path "services/*/go.mod" -Pattern "github.com/t4RG3T21/GoBigTech/platform"
```

---

## Этап 2: Локальная проверка Go модулей

### 2.1. Проверка зависимостей для каждого сервиса

```powershell
# Проверка payment service
cd services/payment
go mod download
go mod verify
cd ../..

# Проверка order service
cd services/order
go mod download
go mod verify
cd ../..

# Проверка inventory service
cd services/inventory
go mod download
go mod verify
cd ../..

# Проверка iam service
cd services/iam
go mod download
go mod verify
cd ../..

# Проверка assembly service
cd services/assembly
go mod download
go mod verify
cd ../..

# Проверка notification service
cd services/notification
go mod download
go mod verify
cd ../..
```

### 2.2. Проверка компиляции сервисов (опционально)

```powershell
# Компиляция payment service
cd services/payment
go build ./cmd/payment
cd ../..

# Компиляция order service
cd services/order
go build ./cmd/order
cd ../..

# И так далее для остальных сервисов
```

---

## Этап 3: Проверка Docker конфигурации

### 3.1. Проверка docker-compose.yml

```powershell
# Валидация docker-compose.yml
docker-compose config
```

### 3.2. Проверка Dockerfile

Убедитесь, что все Dockerfile:
- Копируют platform модуль
- Имеют правильные `replace` директивы в go.mod
- Не используют `go mod tidy` (который может удалить replace)

**Проверка:**
```powershell
# Проверка отсутствия go mod tidy в Dockerfile
Select-String -Path "services/*/Dockerfile" -Pattern "go mod tidy"
# Должно быть пусто (или только в комментариях)
```

---

## Этап 4: Сборка Docker образов

### 4.1. Остановка и очистка предыдущих контейнеров

```powershell
# Остановка всех контейнеров
docker-compose down

# Удаление старых образов (опционально, если нужно пересобрать)
docker-compose build --no-cache
```

### 4.2. Сборка образов

```powershell
# Сборка всех образов
docker-compose build

# Или сборка конкретного сервиса
docker-compose build payment-service
```

**Если есть ошибки сборки:**
1. Проверьте логи: `docker-compose build payment-service 2>&1 | Select-String -Pattern "error" -Context 5`
2. Проверьте, что proto файлы сгенерированы
3. Проверьте, что `replace` директивы присутствуют в go.mod

### 4.3. Запуск всех сервисов

```powershell
# Запуск в фоновом режиме
docker-compose up -d

# Проверка статуса
docker-compose ps
```

**Проверка логов:**
```powershell
# Логи всех сервисов
docker-compose logs

# Логи конкретного сервиса
docker-compose logs payment-service
docker-compose logs order-service
docker-compose logs inventory-service
```

---

## Этап 5: Проверка работоспособности сервисов

### 5.1. Проверка health checks

```powershell
# Проверка статуса health checks
docker-compose ps

# Все сервисы должны иметь статус "healthy" или "starting"
```

### 5.2. Проверка портов

```powershell
# Проверка открытых портов
netstat -ano | Select-String -Pattern "8080|8082|8083|50051|50052|50053|80|8084|9091"
```

**Ожидаемые порты:**
- `80` - Envoy HTTP Gateway
- `8084` - Envoy gRPC REST Gateway
- `9091` - Envoy gRPC (native)
- `8080` - Order Service
- `8082` - IAM Service HTTP Gateway
- `8083` - Notification Service
- `50051` - Inventory Service gRPC
- `50052` - Payment Service gRPC
- `50053` - IAM Service gRPC

### 5.3. Тестирование через Envoy

**HTTP запросы:**
```powershell
# Health check через Envoy
curl http://localhost/health

# Order Service через Envoy
curl http://localhost/api/orders/health

# IAM Service через Envoy
curl http://localhost/api/iam/health
```

**gRPC через Envoy (требует grpcurl):**
```powershell
# Установка grpcurl (если нет)
# Windows: choco install grpcurl
# Или через Go: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Список сервисов
grpcurl -plaintext localhost:9091 list

# Вызов метода
grpcurl -plaintext -d '{"product_id": "prod-123"}' localhost:9091 inventory.v1.InventoryService/GetStock
```

**REST-to-gRPC через Envoy:**
```powershell
# Inventory Service через REST Gateway
curl -X POST http://localhost:8084/inventory.v1.InventoryService/GetStock `
  -H "Content-Type: application/json" `
  -d '{"product_id": "prod-123"}'

# Payment Service через REST Gateway
curl -X POST http://localhost:8084/payment.v1.PaymentService/ProcessPayment `
  -H "Content-Type: application/json" `
  -d '{"order_id": "order-123", "user_id": "user-123", "amount": 100.0, "method": "card"}'
```

---

## Этап 6: Проверка интеграции между сервисами

### 6.1. Проверка работы IAM Service

```powershell
# Регистрация пользователя
curl -X POST http://localhost/api/iam/v1/register `
  -H "Content-Type: application/json" `
  -d '{"email": "test@example.com", "password": "password123"}'

# Логин
$response = curl -X POST http://localhost/api/iam/v1/login `
  -H "Content-Type: application/json" `
  -d '{"email": "test@example.com", "password": "password123"}'
$token = ($response | ConvertFrom-Json).token
```

### 6.2. Проверка работы Order Service

```powershell
# Создание заказа (требует токен)
curl -X POST http://localhost/api/orders `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $token" `
  -d '{"items": [{"product_id": "prod-123", "quantity": 2}]}'
```

### 6.3. Проверка работы Inventory Service

```powershell
# Проверка наличия товара
curl -X POST http://localhost:8084/inventory.v1.InventoryService/GetStock `
  -H "Content-Type: application/json" `
  -d '{"product_id": "prod-123"}'
```

---

## Этап 7: Проверка Observability стека

### 7.1. Prometheus

```powershell
# Проверка метрик
curl http://localhost:9090/metrics

# Проверка UI (откройте в браузере)
# http://localhost:9090
```

### 7.2. Grafana

```powershell
# Проверка UI (откройте в браузере)
# http://localhost:3000
# Логин: admin / admin
```

### 7.3. Jaeger

```powershell
# Проверка UI (откройте в браузере)
# http://localhost:16686
```

### 7.4. Kibana

```powershell
# Проверка UI (откройте в браузере)
# http://localhost:5601
```

---

## Этап 8: Финальная проверка

### 8.1. Проверка всех контейнеров

```powershell
# Статус всех контейнеров
docker-compose ps

# Должны быть запущены:
# - order-service
# - iam-service
# - inventory-service
# - payment-service
# - assembly-service
# - notification-service
# - envoy
# - postgres
# - mongodb
# - redis
# - kafka
# - zookeeper
# - prometheus
# - grafana
# - jaeger
# - elasticsearch
# - kibana
# - otel-collector
```

### 8.2. Проверка логов на ошибки

```powershell
# Поиск ошибок в логах
docker-compose logs | Select-String -Pattern "error|Error|ERROR|fatal|Fatal|FATAL" -Context 3
```

### 8.3. Проверка использования ресурсов

```powershell
# Использование ресурсов контейнерами
docker stats --no-stream
```

---

## 🔧 Решение типичных проблем

### Проблема: "module provides package but is replaced but not required"

**Решение:** Добавьте модуль в `require` секцию `go.mod`:
```powershell
# В go.mod файле сервиса добавьте:
require (
    github.com/t4RG3T21/GoBigTech/platform v0.0.0
    # другие зависимости...
)
```

### Проблема: "no required module provides package"

**Решение:** Добавьте недостающую зависимость в `require` секцию `go.mod`.

### Проблема: "port is already allocated"

**Решение:** Измените порт в `docker-compose.yml` или остановите процесс, использующий порт:
```powershell
# Найти процесс на порту
netstat -ano | Select-String -Pattern ":8080"
```

### Проблема: Proto файлы не найдены

**Решение:** Сгенерируйте proto файлы (см. Этап 1.2).

### Проблема: "replace directive removed by go mod tidy"

**Решение:** Не используйте `go mod tidy` в Dockerfile. Используйте только `go mod download`.

---

## ✅ Чеклист готовности проекта

- [ ] Все proto файлы сгенерированы
- [ ] Все `go.mod` файлы содержат правильные `replace` директивы
- [ ] Все `go.mod` файлы содержат необходимые зависимости в `require`
- [ ] Все Dockerfile не используют `go mod tidy`
- [ ] Docker образы успешно собираются
- [ ] Все контейнеры запускаются без ошибок
- [ ] Health checks проходят успешно
- [ ] Сервисы доступны через Envoy
- [ ] Observability стек работает
- [ ] Нет критических ошибок в логах

---

## 📝 Полезные команды

```powershell
# Полная пересборка проекта
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d

# Просмотр логов в реальном времени
docker-compose logs -f

# Остановка всех сервисов
docker-compose down

# Очистка всех данных (ОСТОРОЖНО!)
docker-compose down -v
docker system prune -a
```

