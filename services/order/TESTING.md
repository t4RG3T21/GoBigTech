# Инструкция по запуску тестов для Order Service

## Требования

1. **Go** версии 1.25.1 или выше
2. **Git Bash** или **PowerShell** (Windows)
3. Установленные зависимости проекта (запустите `go mod download`)

## Установка зависимостей

Перед запуском тестов убедитесь, что все зависимости установлены:

### PowerShell:
```powershell
cd services/order
go mod download
```

### Git Bash:
```bash
cd services/order
go mod download
```

## Запуск всех тестов

### PowerShell:
```powershell
# Переход в директорию сервиса
cd services/order

# Запуск всех тестов с подробным выводом
go test ./... -v

# Запуск всех тестов с покрытием кода
go test ./... -v -cover

# Запуск всех тестов с HTML отчетом о покрытии
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Git Bash:
```bash
# Переход в директорию сервиса
cd services/order

# Запуск всех тестов с подробным выводом
go test ./... -v

# Запуск всех тестов с покрытием кода
go test ./... -v -cover

# Запуск всех тестов с HTML отчетом о покрытии
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Запуск конкретных тестов

### PowerShell:
```powershell
cd services/order

# Запуск тестов только для service пакета
go test ./internal/service -v

# Запуск тестов только для api пакета
go test ./internal/api -v

# Запуск конкретного теста
go test ./internal/service -v -run TestOrderService_CreateOrder_Success

# Запуск тестов, соответствующих паттерну
go test ./internal/service -v -run TestOrderService_CreateOrder
```

### Git Bash:
```bash
cd services/order

# Запуск тестов только для service пакета
go test ./internal/service -v

# Запуск тестов только для api пакета
go test ./internal/api -v

# Запуск конкретного теста
go test ./internal/service -v -run TestOrderService_CreateOrder_Success

# Запуск тестов, соответствующих паттерну
go test ./internal/service -v -run TestOrderService_CreateOrder
```

## Структура тестов

### Unit-тесты сервиса (`internal/service/order_service_test.go`)
- `TestOrderService_CreateOrder_Success` - успешное создание заказа
- `TestOrderService_CreateOrder_InventoryUnavailable` - товар недоступен
- `TestOrderService_CreateOrder_PaymentFailed` - ошибка платежа
- `TestOrderService_CreateOrder_EmptyItems` - пустой список товаров
- `TestOrderService_GetOrderByID_Success` - успешное получение заказа
- `TestOrderService_GetOrderByID_NotFound` - заказ не найден
- `TestOrderService_GetOrderByID_EmptyID` - пустой ID заказа

### Unit-тесты API handler (`internal/api/order_handler_test.go`)
- `TestOrderHandler_PostOrders_Success` - успешное создание заказа через API
- `TestOrderHandler_PostOrders_InvalidJSON` - невалидный JSON
- `TestOrderHandler_PostOrders_ServiceError` - ошибка сервиса
- `TestOrderHandler_PostOrders_EmptyUserID` - пустой user ID
- `TestOrderHandler_PostOrders_EmptyItems` - пустой список товаров
- `TestOrderHandler_GetOrdersId_Success` - успешное получение заказа
- `TestOrderHandler_GetOrdersId_NotFound` - заказ не найден

## Моки

Проект использует библиотеку `testify/mock` для создания моков зависимостей:

- `internal/service/mocks/inventory_client.go` - моки для InventoryClient
- `internal/service/mocks/inventory_client.go` - моки для PaymentClient (в том же файле)
- `internal/repository/mocks/order_repository.go` - моки для OrderRepository

## Полезные команды

### Просмотр покрытия кода:
```powershell
# PowerShell
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

```bash
# Git Bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Запуск тестов в режиме watch (требует установки `modd` или `entr`):
```powershell
# Установка modd (опционально)
go install github.com/cortesi/modd/cmd/modd@latest

# Запуск в режиме watch
modd
```

### Запуск тестов с таймаутом:
```powershell
# PowerShell
$env:GO_TEST_TIMEOUT="30s"; go test ./... -v
```

```bash
# Git Bash
GO_TEST_TIMEOUT=30s go test ./... -v
```

## Устранение проблем

### Ошибка: "package not found"
Убедитесь, что вы находитесь в правильной директории и зависимости установлены:
```powershell
cd services/order
go mod download
go mod tidy
```

### Ошибка: "cannot find package"
Проверьте, что GOPATH и GOROOT настроены правильно:
```powershell
go env GOPATH
go env GOROOT
```

### Тесты не компилируются
Очистите кэш и пересоберите:
```powershell
go clean -cache
go test ./... -v
```

## Быстрый старт

1. Откройте PowerShell или Git Bash
2. Перейдите в директорию проекта:
   ```powershell
   cd C:\Users\t4rg3t\GolandProjects\GoBigTech\services\order
   ```
3. Запустите тесты:
   ```powershell
   go test ./... -v
   ```

Вы должны увидеть вывод всех тестов с результатами PASS/FAIL.

