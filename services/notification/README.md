# Notification Service

Сервис для отправки уведомлений в Telegram о событиях оплаты и сборки заказов.

## Конфигурация

Сервис использует переменные окружения для настройки. Текущие значения:

- **TELEGRAM_BOT_TOKEN**: `8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo`
- **TELEGRAM_CHAT_ID**: `620586579`

## Запуск

### Способ 1: Через скрипт (рекомендуется)

```powershell
cd services/notification
.\run.ps1
```

### Способ 2: Через переменные окружения

```powershell
$env:TELEGRAM_BOT_TOKEN = "8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo"
$env:TELEGRAM_CHAT_ID = "620586579"
$env:KAFKA_BOOTSTRAP_SERVERS = "localhost:9092"
go run ./cmd/notification/main.go
```

### Способ 3: Через .env файл (если используете библиотеку для загрузки .env)

Создайте файл `.env` в папке `services/notification/`:

```
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_PAYMENT_TOPIC=orders.payment
KAFKA_ASSEMBLY_TOPIC=orders.assembly
KAFKA_CONSUMER_GROUP_ID=notification-service
TELEGRAM_BOT_TOKEN=8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo
TELEGRAM_CHAT_ID=620586579
LOG_LEVEL=debug
```

## Функциональность

Сервис подписывается на два топика Kafka:
- `orders.payment` - события оплаты заказов
- `orders.assembly` - события сборки заказов

При получении событий отправляет уведомления в Telegram.

## Проверка работы

1. Убедитесь, что Kafka запущена: `docker-compose ps`
2. Убедитесь, что топики созданы: `.\check-topics.ps1`
3. Запустите сервис: `.\run.ps1`
4. Создайте заказ через Order Service - должны прийти уведомления в Telegram

