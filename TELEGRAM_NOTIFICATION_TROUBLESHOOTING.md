# Troubleshooting: Telegram Notifications

Если уведомления не приходят в Telegram бота, проверьте следующие моменты:

## Шаг 1: Проверка запуска Notification Service

Убедитесь, что Notification Service запущен:

```powershell
cd services\notification\cmd\notification
go run main.go
```

В логах должно быть:
```
Starting Notification Service
Telegram bot authorized
Starting Kafka consumers
```

## Шаг 2: Проверка Telegram Chat ID

**ВАЖНО:** По умолчанию используется `TELEGRAM_CHAT_ID=123456789`, это **не реальный** chat ID!

### Как получить реальный Chat ID:

1. **Способ 1: Через бота @userinfobot**
   - Добавьте бота @userinfobot в ваш чат или напишите ему в личку
   - Бот покажет ваш Chat ID

2. **Способ 2: Через API Telegram**
   - Напишите вашему боту любое сообщение
   - Вызовите API: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Найдите `"chat":{"id":<YOUR_CHAT_ID>}` в ответе

3. **Способ 3: Через бота @RawDataBot**
   - Добавьте бота @RawDataBot в ваш чат
   - Бот покажет всю информацию о чате, включая Chat ID

### Установка правильного Chat ID:

```powershell
$env:TELEGRAM_CHAT_ID = "YOUR_REAL_CHAT_ID"
cd services\notification\cmd\notification
go run main.go
```

Или создайте файл `.env` в `services/notification/`:
```
TELEGRAM_BOT_TOKEN=8301618810:AAGrWANT0ZgatbGxgpc59TysKbn-6jydkZo
TELEGRAM_CHAT_ID=YOUR_REAL_CHAT_ID
```

## Шаг 3: Проверка Kafka

Убедитесь, что Kafka запущена и топики созданы:

```powershell
# Проверить, что Kafka работает
docker ps | findstr kafka

# Проверить топики (если есть доступ к Kafka CLI)
docker exec -it gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --list
```

Должны быть топики:
- `orders.payment`
- `orders.assembly`

Если топиков нет, они создадутся автоматически при первой отправке сообщения (если `KAFKA_AUTO_CREATE_TOPICS_ENABLE=true`).

## Шаг 4: Проверка отправки событий из Order Service

Убедитесь, что Order Service отправляет события в Kafka. В логах Order Service должно быть:

```
Payment event sent to Kafka
  order_id: ...
  topic: orders.payment
```

Если этого нет, значит Order Service не отправляет события (возможно, заказ не дошел до этапа оплаты).

## Шаг 5: Проверка получения сообщений в Notification Service

В логах Notification Service должно быть:

```
Received payment event
  topic: orders.payment
  value: {"order_id":"...","user_id":"...","amount":...}

Payment notification sent
  order_id: ...
```

Если видите "Received payment event", но нет "Payment notification sent", значит проблема с отправкой в Telegram.

## Шаг 6: Проверка Telegram бота

1. **Убедитесь, что бот запущен:**
   - В логах должно быть: `Telegram bot authorized, bot_username: @YourBotName`

2. **Убедитесь, что бот добавлен в чат:**
   - Добавьте бота в ваш чат или группу
   - Отправьте боту команду `/start` в личку

3. **Проверьте токен бота:**
   - Токен должен быть правильным (получен от @BotFather)
   - В логах не должно быть ошибок типа "Unauthorized"

## Шаг 7: Тестирование вручную

Можно протестировать отправку уведомления напрямую:

```powershell
# Отправьте тестовое сообщение в Kafka топик
# (требуется установленный kafka-console-producer или Docker)

docker exec -it gobigtech_kafka kafka-console-producer --bootstrap-server localhost:9092 --topic orders.payment

# Затем введите JSON:
{"order_id":"test-123","user_id":"test-user","amount":100.50,"transaction_id":"tx-test","timestamp":"2025-01-01T12:00:00Z"}
```

Если после этого пришло уведомление в Telegram, значит проблема была в Order Service или в создании заказа.

## Частые ошибки

### Ошибка: "chat not found"
**Причина:** Неправильный Chat ID или бот не добавлен в чат  
**Решение:** Получите правильный Chat ID и добавьте бота в чат

### Ошибка: "Unauthorized"
**Причина:** Неправильный токен бота  
**Решение:** Проверьте `TELEGRAM_BOT_TOKEN` в конфигурации

### Ошибка: "Failed to read payment message from Kafka"
**Причина:** Kafka не запущена или неправильный адрес  
**Решение:** Проверьте `KAFKA_BOOTSTRAP_SERVERS` и убедитесь, что Kafka работает

### Нет ошибок, но уведомления не приходят
**Причина:** События не отправляются в Kafka или Chat ID неправильный  
**Решение:** 
1. Проверьте логи Order Service - отправляются ли события
2. Проверьте логи Notification Service - получаются ли сообщения
3. Проверьте Chat ID - он должен быть реальным числом

## Диагностика через логи

Включите debug логирование:

```powershell
$env:LOG_LEVEL = "debug"
cd services\notification\cmd\notification
go run main.go
```

Это покажет подробную информацию о всех операциях.

