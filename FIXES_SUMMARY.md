# Исправления: Notification Service и Alertmanager

## ✅ Решенные проблемы

### 1. Конфликт портов: Kafka и Alertmanager (9093)
**Проблема**: Kafka использовал порт 9093 для CONTROLLER listener и был проброшен на хост, конфликтовал с Alertmanager.

**Решение**: 
- Убран проброс порта `9093:9093` для Kafka в `docker-compose.yml`
- Kafka теперь использует порт 9093 только внутри Docker сети
- Alertmanager успешно запущен на порту 9093

### 2. Ошибка конфигурации Alertmanager
**Проблема**: `yaml: unmarshal errors: line 56: field timeout not found in type config.plain`

**Решение**: 
- Удалено неверное поле `timeout` из `http_config` в `monitoring/alertmanager/alertmanager.yml`
- Alertmanager теперь запускается без ошибок

### 3. Notification Service: порт 8083 занят
**Проблема**: Docker контейнер `gobigtech_notification_service` занимал порт 8083.

**Решение**: 
- Контейнер остановлен и удален
- Порт 8083 теперь свободен для локального запуска

### 4. Зависимость Alertmanager от Notification Service
**Проблема**: Alertmanager не запускался из-за зависимости от notification-service.

**Решение**: 
- Убрана зависимость `depends_on: notification-service` из Alertmanager
- Alertmanager может работать независимо, webhook будет вызываться только когда notification-service доступен

## 🚀 Текущий статус

✅ **Alertmanager**: Запущен и работает на `http://localhost:9093`
✅ **Kafka**: Запущен, порт 9093 используется только внутри Docker
✅ **Порт 8083**: Свободен для локального запуска Notification Service

## 📝 Инструкции

### Запуск Notification Service локально

1. **Убедитесь, что Docker контейнер остановлен**:
   ```bash
   docker stop gobigtech_notification_service
   docker rm gobigtech_notification_service
   ```

2. **Запустите Notification Service локально**:
   ```bash
   cd services/notification
   go run cmd/notification/main.go
   ```

3. **Проверьте работу**:
   - Health endpoint: `http://localhost:8083/health`
   - Alert endpoint: `http://localhost:8083/alert` (для Alertmanager)

### Запуск Notification Service в Docker

Если хотите запустить в Docker вместо локального запуска:

```bash
docker-compose up -d notification-service
```

**Важно**: Нельзя запускать одновременно локально и в Docker на одном порту!

### Проверка Alertmanager

1. **Откройте веб-интерфейс**: `http://localhost:9093`
2. **Проверьте статус**: Должен быть доступен UI
3. **Проверьте конфигурацию**: В разделе "Status" → "Configuration" должна быть видна конфигурация с webhook на `notification-service:8083/alert`

## 🔧 Измененные файлы

1. `docker-compose.yml`:
   - Убран проброс порта 9093 для Kafka
   - Убрана зависимость Alertmanager от notification-service

2. `monitoring/alertmanager/alertmanager.yml`:
   - Удалено неверное поле `timeout` из `http_config`

3. `QUICK_START.md`:
   - Добавлена информация о решенных проблемах

## ⚠️ Важные замечания

1. **Порт 8083**: Может использоваться либо локально, либо в Docker, но не одновременно
2. **Alertmanager webhook**: Использует `http://notification-service:8083/alert` (имя сервиса в Docker сети)
   - Если Notification Service запущен локально, Alertmanager не сможет до него достучаться
   - Для локального запуска нужно либо изменить конфигурацию Alertmanager, либо использовать `host.docker.internal:8083`
3. **Kafka порт 9093**: Теперь используется только внутри Docker, не пробрасывается на хост

## 🧪 Тестирование

1. **Проверьте Alertmanager**:
   ```bash
   curl http://localhost:9093/-/healthy
   # Должен вернуть: OK
   ```

2. **Проверьте Notification Service** (если запущен локально):
   ```bash
   curl http://localhost:8083/health
   # Должен вернуть: {"status":"ok"}
   ```

3. **Проверьте интеграцию**:
   - Создайте тестовый алерт в Prometheus
   - Убедитесь, что Alertmanager получает алерт
   - Проверьте, что Notification Service получает webhook от Alertmanager
   - Проверьте Telegram на наличие уведомления

