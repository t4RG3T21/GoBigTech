# Настройка системы алертов

## Архитектура

```
Order Service (метрики) 
    ↓
Prometheus (сбор метрик + правила алертов)
    ↓
Alertmanager (маршрутизация алертов)
    ↓
Notification Service (/alert endpoint)
    ↓
Telegram Bot
```

## Компоненты

1. **Prometheus** (порт 9090)
   - Собирает метрики с Order Service
   - Вычисляет правила алертов из `alerts.yml`
   - Отправляет алерты в Alertmanager

2. **Alertmanager** (порт 9093)
   - Получает алерты от Prometheus
   - Маршрутизирует по уровню серьезности
   - Отправляет webhook в Notification Service

3. **Notification Service** (порт 8083)
   - Принимает алерты через `/alert` endpoint
   - Форматирует и отправляет в Telegram

## Правила алертов

### Бизнес-метрики

- **HighOrderRate**: >10 заказов/минуту (warning, 2 мин)
- **PaymentErrors**: Ошибки обработки платежей (critical, 1 мин)
- **HighOrderFailureRate**: >10% ошибок заказов (warning, 3 мин)
- **NoOrdersReceived**: Нет заказов 10+ минут (warning, 10 мин)

### Производительность

- **SlowOrderProcessing**: 95-й перцентиль >5 сек (warning, 5 мин)
- **VerySlowOrderProcessing**: 95-й перцентиль >10 сек (critical, 3 мин)

### Инфраструктура

- **InventoryCheckFailures**: Ошибки проверки инвентаря (warning, 2 мин)
- **DatabaseSaveFailures**: Ошибки сохранения в БД (critical, 1 мин)
- **HighHTTPRequestRate**: >10 запросов/сек (info, 2 мин)

## Запуск

### 1. Запуск всех компонентов

```bash
docker-compose up -d prometheus alertmanager notification-service
```

### 2. Проверка статуса

```bash
# Prometheus UI
open http://localhost:9090

# Alertmanager UI
open http://localhost:9093

# Проверка алертов в Prometheus
open http://localhost:9090/alerts

# Проверка алертов в Alertmanager
open http://localhost:9093/#/alerts
```

### 3. Тестирование алертов

#### Тест 1: Высокая частота заказов

Создайте более 10 заказов за минуту через Order Service API.

#### Тест 2: Ошибки платежей

Имитируйте ошибки платежей в Payment Service или создайте заказы с невалидными данными.

#### Тест 3: Ручной тест через Alertmanager

```bash
curl -X POST http://localhost:9093/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "service": "order-service"
    },
    "annotations": {
      "summary": "Тестовый алерт",
      "description": "Это тестовое сообщение для проверки системы алертов"
    }
  }]'
```

## Конфигурация

### Prometheus (`monitoring/prometheus.yml`)

- Скрапит метрики с Order Service (порт 8080)
- Загружает правила из `alerts.yml`
- Отправляет алерты в Alertmanager (порт 9093)

### Alertmanager (`monitoring/alertmanager/alertmanager.yml`)

- Группирует алерты по имени и сервису
- Маршрутизирует по уровню серьезности:
  - **critical**: отправка каждые 30 минут
  - **warning**: отправка каждые 2 часа
  - **info**: отправка каждые 6 часов
- Отправляет webhook в `http://notification-service:8083/alert`

### Notification Service

- Принимает POST запросы на `/alert`
- Форматирует сообщения с эмодзи
- Отправляет в Telegram через бота

## Переменные окружения

Для Notification Service:

```bash
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
HTTP_PORT=8083
```

## Мониторинг

### Проверка метрик

```bash
# Метрики Order Service
curl http://localhost:8080/metrics | grep order_

# Метрики через OTLP Collector
curl http://localhost:8889/metrics | grep order_otel_
```

### Проверка алертов

```bash
# Список активных алертов в Prometheus
curl http://localhost:9090/api/v1/alerts

# Список алертов в Alertmanager
curl http://localhost:9093/api/v2/alerts
```

## Troubleshooting

### Алерты не приходят в Telegram

1. Проверьте логи Notification Service:
   ```bash
   docker logs gobigtech_notification_service
   ```

2. Проверьте логи Alertmanager:
   ```bash
   docker logs gobigtech_alertmanager
   ```

3. Проверьте, что Prometheus видит метрики:
   ```bash
   curl http://localhost:9090/api/v1/targets
   ```

4. Проверьте правила алертов:
   ```bash
   curl http://localhost:9090/api/v1/rules
   ```

### Алерты не срабатывают

1. Проверьте, что метрики собираются:
   ```bash
   curl http://localhost:9090/api/v1/query?query=order_orders_created_total
   ```

2. Проверьте правила в Prometheus UI:
   ```
   http://localhost:9090/alerts
   ```

3. Убедитесь, что условия алертов выполнены (например, создайте достаточно заказов)

## Примеры сообщений в Telegram

### Критический алерт
```
🚨 PaymentErrors

🔴 Обнаружены ошибки при обработке платежей. Частота ошибок: 0.5/сек за последние 5 минут

🚨 Статус: АКТИВЕН
📦 Сервис: order-service
```

### Предупреждение
```
🟡 HighOrderRate

🟡 Создается более 10 заказов в минуту. Текущая частота: 0.2 заказов/сек

🚨 Статус: АКТИВЕН
📦 Сервис: order-service
```

### Разрешенный алерт
```
✅ PaymentErrors

🔴 Обнаружены ошибки при обработке платежей. Частота ошибок: 0/сек за последние 5 минут

✅ Статус: РАЗРЕШЕН
📦 Сервис: order-service
```

