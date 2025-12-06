# Карта зависимостей между сервисами

## Структура зависимостей

```
platform (базовый модуль)
├── order-service
│   ├── inventory-service (через replace)
│   │   └── iam-service (через replace)
│   └── payment-service (через replace)
├── inventory-service
│   └── iam-service (через replace)
├── payment-service
├── iam-service
├── assembly-service
└── notification-service
```

## Детальные зависимости

### Order Service
- **Прямые зависимости:**
  - `platform` (replace: `../../platform`)
  - `inventory/v1` (replace: `../inventory`)
  - `payment/v1` (replace: `../payment`)
- **Транзитивные зависимости:**
  - `iam/v1` (через inventory)

### Inventory Service
- **Прямые зависимости:**
  - `platform` (replace: `../../platform`)
  - `iam/v1` (replace: `../iam`)

### Payment Service
- **Прямые зависимости:**
  - `platform` (replace: `../../platform`)

### IAM Service
- **Прямые зависимости:**
  - `platform` (replace: `../../platform`)

### Assembly Service
- **Прямые зависимости:**
  - `platform` (replace: `../../platform`)

### Notification Service
- **Прямые зависимости:**
  - `platform` (replace: `../../platform`)

## Dockerfile требования

### Order Service Dockerfile
Должен копировать:
1. `platform/go.mod`, `platform/go.sum`, `platform/`
2. `services/iam/v1/`, `services/iam/go.mod`, `services/iam/go.sum` (для inventory)
3. `services/inventory/v1/`, `services/inventory/go.mod`, `services/inventory/go.sum`
4. `services/payment/v1/`, `services/payment/go.mod`, `services/payment/go.sum`
5. `services/order/go.mod`, `services/order/go.sum`, `services/order/`

Порядок загрузки зависимостей:
1. `platform` → `go mod download`
2. `iam` → `go mod download`
3. `inventory` → `go mod download`
4. `payment` → `go mod download`
5. `order` → `go mod download`

### Inventory Service Dockerfile
Должен копировать:
1. `platform/go.mod`, `platform/go.sum`, `platform/`
2. `services/iam/v1/`, `services/iam/go.mod`, `services/iam/go.sum`
3. `services/inventory/go.mod`, `services/inventory/go.sum`, `services/inventory/`

Порядок загрузки зависимостей:
1. `platform` → `go mod download`
2. `iam` → `go mod download`
3. `inventory` → `go mod download`

### Payment Service Dockerfile
Должен копировать:
1. `platform/go.mod`, `platform/go.sum`, `platform/`
2. `services/payment/go.mod`, `services/payment/go.sum`, `services/payment/`

Порядок загрузки зависимостей:
1. `platform` → `go mod download`
2. `payment` → `go mod download`

### IAM Service Dockerfile
Должен копировать:
1. `platform/go.mod`, `platform/go.sum`, `platform/`
2. `services/iam/go.mod`, `services/iam/go.sum`, `services/iam/`

Порядок загрузки зависимостей:
1. `platform` → `go mod download`
2. `iam` → `go mod download`

### Assembly Service Dockerfile
Должен копировать:
1. `platform/go.mod`, `platform/go.sum`, `platform/`
2. `services/assembly/go.mod`, `services/assembly/go.sum`, `services/assembly/`

Порядок загрузки зависимостей:
1. `platform` → `go mod download`
2. `assembly` → `go mod download`

### Notification Service Dockerfile
Должен копировать:
1. `platform/go.mod`, `platform/go.sum`, `platform/`
2. `services/notification/go.mod`, `services/notification/go.sum`, `services/notification/`

Порядок загрузки зависимостей:
1. `platform` → `go mod download`
2. `notification` → `go mod download`

## Проверка зависимостей

### Команды для проверки

```powershell
# Проверка replace директив
Get-ChildItem -Path services -Recurse -Filter "go.mod" | ForEach-Object {
    Write-Host "`n=== $($_.Name) ==="
    Get-Content $_.FullName | Select-String -Pattern "^replace"
}

# Проверка require директив для platform
Get-ChildItem -Path services -Recurse -Filter "go.mod" | ForEach-Object {
    Write-Host "`n=== $($_.Name) ==="
    Get-Content $_.FullName | Select-String -Pattern "platform"
}
```

### Обновление всех go.sum

```powershell
# Обновить все go.sum файлы
@('platform', 'services/order', 'services/inventory', 'services/payment', 'services/iam', 'services/assembly', 'services/notification') | ForEach-Object {
    Write-Host "Updating $_..."
    Set-Location $_
    go mod download
    go mod tidy
    Set-Location ..
}
```

## Важные замечания

1. **Всегда копируйте `go.mod` и `go.sum` перед копированием исходного кода** - это позволяет кэшировать слои Docker
2. **Загружайте зависимости в правильном порядке** - сначала базовые модули (platform), потом зависимые сервисы
3. **Не используйте `go mod tidy` в Dockerfile** - он может удалить `replace` директивы
4. **Копируйте все транзитивные зависимости** - если сервис A зависит от B, а B зависит от C, то для сборки A нужны и B, и C

