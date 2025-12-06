# Troubleshooting: Telegram уведомления не приходят

## Возможные причины и решения

### 1. Notification Service не запущен

**Проверка:**
```powershell
# Проверить, запущен ли сервис
Get-Process | Where-Object {$_.ProcessName -like "*notification*"}
```

**Решение:**
```powershell
cd services\notification\cmd\notification
go run main.go
```

### 2. Неправильный Telegram Chat ID

**Проблема:** По умолчанию используется `123456789`, который не является вашим реальным Chat ID.

**Как узнать свой Chat ID:**
1. Найдите бота `@userinfobot` в Telegram
2. Отправьте ему `/start`
3. Он покажет ваш Chat ID (например, `620586579`)

**Решение:**
```powershell
# Установить правильный Chat ID
$env:TELEGRAM_CHAT_ID = "620586579"
cd services\notification\cmd\notification
go run main.go
```

Или создайте файл `.env` в `services/notification/`:
```
TELEGRAM_BOT_TOKEN=8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo
TELEGRAM_CHAT_ID=620586579
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
```

### 3. Kafka не работает или топики не созданы

**Проверка:**
```powershell
# Проверить, запущен ли Kafka
docker ps | Select-String "kafka"

# Проверить топики
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --list
```

**Решение:**
```powershell
# Запустить Kafka через docker-compose
cd C:\Users\t4rg3t\GolandProjects\GoBigTech
docker-compose up -d kafka

# Подождать 30 секунд, затем проверить топики
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --list
```

Должны быть видны:
- `orders.payment`
- `orders.assembly`

### 4. Order Service не отправляет события в Kafka

**Проверка:**
Проверьте логи Order Service при создании заказа. Должно быть:
```
Payment event sent to Kafka
  order_id: ...
  topic: orders.payment
```

**Решение:**
- Убедитесь, что Kafka запущен
- Проверьте, что `KAFKA_BOOTSTRAP_SERVERS=localhost:9092` в Order Service
- Проверьте логи на ошибки отправки в Kafka

### 5. Notification Service не получает сообщения из Kafka

**Проверка:**
Проверьте логи Notification Service. Должно быть:
```
Starting payment messages consumer
  topic: orders.payment
```

И при получении события:
```
Received payment event
  topic: orders.payment
  value: {...}
```

**Решение:**
- Убедитесь, что Notification Service запущен
- Проверьте, что `KAFKA_BOOTSTRAP_SERVERS=localhost:9092` в Notification Service
- Проверьте, что топики `orders.payment` и `orders.assembly` существуют

### 6. Проблемы с Telegram Bot API

**Проверка:**
При запуске Notification Service должно быть:
```
Telegram bot authorized
  bot_username: ...
  chat_id: 620586579
```

Если видите ошибку:
```
Failed to create Telegram bot: ...
```

**Решение:**
- Проверьте, что токен правильный
- Убедитесь, что бот активен
- Проверьте, что вы отправили `/start` боту перед использованием

### 7. Неправильный формат Chat ID

**Проблема:** Chat ID должен быть числом (int64), а не строкой.

**Проверка:**
В логах Notification Service должно быть:
```
telegram_chat_id: 620586579
```

Если видите `123456789` (значение по умолчанию), значит Chat ID не установлен.

**Решение:**
Установите правильный Chat ID через переменную окружения.

## Пошаговая проверка

### Шаг 1: Проверьте все сервисы запущены

```powershell
# IAM Service (порт 50053, 8082)
# Order Service (порт 8080)
# Inventory Service (порт 50051)
# Payment Service (порт 50052)
# Notification Service (без порта, только Kafka consumer)
# Assembly Service (без порта, только Kafka consumer)
```

### Шаг 2: Проверьте Kafka

```powershell
# Проверить статус
docker ps | Select-String "kafka"

# Проверить топики
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### Шаг 3: Проверьте Chat ID

```powershell
# Узнать свой Chat ID через бота @userinfobot
# Затем установить:
$env:TELEGRAM_CHAT_ID = "ВАШ_CHAT_ID"
```

### Шаг 4: Перезапустите Notification Service

```powershell
cd services\notification\cmd\notification
go run main.go
```

### Шаг 5: Создайте тестовый заказ

```powershell
# В Postman:
POST http://localhost:8080/orders
Authorization: Bearer <ваш_session_id>
Content-Type: application/json

{
  "user_id": "sasha",
  "items": [
    {
      "product_id": "p1",
      "quantity": 2
    }
  ]
}
```

### Шаг 6: Проверьте логи

**Order Service:**
```
Payment event sent to Kafka
  order_id: ...
  topic: orders.payment
```

**Notification Service:**
```
Received payment event
  topic: orders.payment
  ...
Payment notification sent
  order_id: ...
```

## Быстрая проверка

Запустите эту команду для проверки всех компонентов:

```powershell
Write-Host "=== Notification Service Check ===" -ForegroundColor Green
Write-Host ""
Write-Host "1. Kafka:" -ForegroundColor Cyan
docker ps --filter "name=kafka" --format "table {{.Names}}\t{{.Status}}"
Write-Host ""
Write-Host "2. Kafka Topics:" -ForegroundColor Cyan
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --list 2>&1
Write-Host ""
Write-Host "3. Notification Service:" -ForegroundColor Cyan
Write-Host "   - Запущен ли? (проверьте вручную)" -ForegroundColor Yellow
Write-Host "   - Chat ID установлен? (проверьте TELEGRAM_CHAT_ID)" -ForegroundColor Yellow
Write-Host ""
Write-Host "4. Order Service:" -ForegroundColor Cyan
Write-Host "   - Запущен ли? (проверьте вручную)" -ForegroundColor Yellow
Write-Host "   - Kafka подключен? (проверьте логи)" -ForegroundColor Yellow
```

## Частые ошибки

### Ошибка: "Failed to read payment message from Kafka"
**Причина:** Kafka не доступен или топик не существует
**Решение:** Запустите `docker-compose up -d kafka` и проверьте топики

### Ошибка: "Failed to send payment notification"
**Причина:** Проблема с Telegram Bot API или неправильный Chat ID
**Решение:** Проверьте токен и Chat ID

### Ошибка: "Missing metadata in request"
**Причина:** Не относится к Notification Service, это ошибка Inventory Service
**Решение:** Проверьте передачу session_id (уже исправлено)

## Тестирование вручную

Если хотите протестировать отправку уведомления напрямую:

```powershell
# Создайте тестовый скрипт test_notification.ps1
$env:TELEGRAM_BOT_TOKEN = "8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo"
$env:TELEGRAM_CHAT_ID = "620586579"

# Запустите Notification Service
cd services\notification\cmd\notification
go run main.go
```

Затем создайте заказ через Postman - уведомление должно прийти.

