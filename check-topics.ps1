# Скрипт для проверки существования топиков в Kafka
# Использование: .\check-topics.ps1

$KAFKA_CONTAINER = "gobigtech_kafka"
$BOOTSTRAP_SERVER = "localhost:9092"

Write-Host "Проверка доступности Kafka контейнера..." -ForegroundColor Yellow

# Проверяем, запущен ли контейнер
$containerRunning = docker ps --format "{{.Names}}" | Select-String -Pattern $KAFKA_CONTAINER

if (-not $containerRunning) {
    Write-Host "Ошибка: Контейнер $KAFKA_CONTAINER не запущен." -ForegroundColor Red
    Write-Host "Запустите Kafka с помощью: docker-compose up -d kafka" -ForegroundColor Yellow
    exit 1
}

Write-Host "Контейнер $KAFKA_CONTAINER найден." -ForegroundColor Green
Write-Host ""

# Получаем список всех топиков
Write-Host "Список всех топиков в Kafka:" -ForegroundColor Cyan
$topics = docker exec $KAFKA_CONTAINER kafka-topics --bootstrap-server $BOOTSTRAP_SERVER --list

if ($topics) {
    Write-Host $topics
} else {
    Write-Host "Топики не найдены" -ForegroundColor Yellow
}

Write-Host ""

# Проверяем наличие нужных топиков
$requiredTopics = @("orders.payment", "orders.assembly")
$allTopicsExist = $true

foreach ($topic in $requiredTopics) {
    if ($topics -match "^${topic}$") {
        Write-Host "✓ Топик '$topic' существует" -ForegroundColor Green
    } else {
        Write-Host "✗ Топик '$topic' НЕ найден" -ForegroundColor Red
        $allTopicsExist = $false
    }
}

Write-Host ""

if (-not $allTopicsExist) {
    Write-Host "Некоторые топики отсутствуют. Запустите скрипт создания:" -ForegroundColor Yellow
    Write-Host "  .\create-topics.ps1" -ForegroundColor Cyan
    exit 1
} else {
    Write-Host "Все необходимые топики существуют!" -ForegroundColor Green
}

