# Скрипт запуска Notification Service с настройками Telegram
# Использование: .\run.ps1

# Устанавливаем переменные окружения
$env:KAFKA_BOOTSTRAP_SERVERS = "localhost:9092"
$env:KAFKA_PAYMENT_TOPIC = "orders.payment"
$env:KAFKA_ASSEMBLY_TOPIC = "orders.assembly"
$env:KAFKA_CONSUMER_GROUP_ID = "notification-service"

# Telegram настройки
$env:TELEGRAM_BOT_TOKEN = "8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo"
$env:TELEGRAM_CHAT_ID = "620586579"

# Логирование
$env:LOG_LEVEL = "debug"

Write-Host "Запуск Notification Service..." -ForegroundColor Green
Write-Host "Telegram Bot Token: $($env:TELEGRAM_BOT_TOKEN.Substring(0, 10))..." -ForegroundColor Yellow
Write-Host "Telegram Chat ID: $env:TELEGRAM_CHAT_ID" -ForegroundColor Yellow
Write-Host ""

# Запускаем сервис
go run ./cmd/notification/main.go

