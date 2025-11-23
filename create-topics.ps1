# Скрипт для создания топиков в Kafka (PowerShell версия)
# Использование: .\create-topics.ps1

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

# Функция для создания топика
function Create-Topic {
    param(
        [string]$TopicName,
        [string]$Description
    )
    
    Write-Host "Создание топика: $TopicName ($Description)" -ForegroundColor Yellow
    
    # Проверяем, существует ли топик
    $existingTopics = docker exec $KAFKA_CONTAINER kafka-topics --bootstrap-server $BOOTSTRAP_SERVER --list
    if ($existingTopics -match "^${TopicName}$") {
        Write-Host "Топик $TopicName уже существует, пропускаем..." -ForegroundColor Yellow
    } else {
        docker exec $KAFKA_CONTAINER kafka-topics `
            --bootstrap-server $BOOTSTRAP_SERVER `
            --create `
            --topic $TopicName `
            --partitions 1 `
            --replication-factor 1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Топик $TopicName успешно создан" -ForegroundColor Green
        } else {
            Write-Host "✗ Ошибка при создании топика $TopicName" -ForegroundColor Red
            return $false
        }
    }
    Write-Host ""
    return $true
}

# Создаем топики
Create-Topic -TopicName "orders.payment" -Description "события оплаты заказов"
Create-Topic -TopicName "orders.assembly" -Description "события сборки заказов"

Write-Host "Готово! Список всех топиков:" -ForegroundColor Green
docker exec $KAFKA_CONTAINER kafka-topics --bootstrap-server $BOOTSTRAP_SERVER --list

