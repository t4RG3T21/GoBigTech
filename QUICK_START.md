# Быстрый запуск всех сервисов

## Проблемы и решения

### Проблема 1: Конфликт портов
- **Kafka** использовал порт 9093 (CONTROLLER listener), конфликтовал с **Alertmanager**
- **Решение**: Убран проброс порта 9093 для Kafka (используется только внутри Docker)

### Проблема 2: Ошибка конфигурации Alertmanager
- **Ошибка**: `field timeout not found in type config.plain`
- **Решение**: Удалено неверное поле `timeout` из `http_config` в `alertmanager.yml`

### Проблема 3: Notification Service порт занят
- **Решение**: Остановите Docker контейнер `gobigtech_notification_service` перед запуском локально:
  ```bash
  docker stop gobigtech_notification_service
  docker rm gobigtech_notification_service
  ```

## Решение

### 1. Запустите Docker Desktop
Убедитесь, что Docker Desktop запущен и работает.

### 2. Запустите все необходимые сервисы

```bash
# Запуск всех сервисов инфраструктуры
docker-compose up -d

# Или только необходимые для мониторинга и алертов:
docker-compose up -d prometheus alertmanager otel-collector kafka notification-service
```

### 3. Проверьте статус сервисов

```bash
docker-compose ps
```

Должны быть запущены:
- ✅ `gobigtech_prometheus` (порт 9090)
- ✅ `gobigtech_alertmanager` (порт 9093)
- ✅ `gobigtech_otel_collector` (порты 4317, 4318, 8889)
- ✅ `gobigtech_kafka` (порт 9092)
- ✅ `gobigtech_notification_service` (порт 8083)

### 4. Проверьте доступность

После запуска проверьте:
- Prometheus: http://localhost:9090
- Alertmanager: http://localhost:9093
- Notification Service: http://localhost:8083/health

### 5. Перезапустите Order Service

После того как все Docker сервисы запущены, перезапустите Order Service:

```powershell
# Остановите текущий процесс (Ctrl+C)
# Затем запустите снова:
cd services/order/cmd/order
go run main.go
```

## Порты, которые должны быть доступны

| Сервис | Порт | URL |
|--------|------|-----|
| Prometheus | 9090 | http://localhost:9090 |
| Alertmanager | 9093 | http://localhost:9093 |
| OpenTelemetry Collector | 4317 | gRPC (внутренний) |
| OpenTelemetry Collector | 8889 | http://localhost:8889/metrics |
| Kafka | 9092 | localhost:9092 |
| Notification Service | 8083 | http://localhost:8083/health |

## Troubleshooting

### Если порты заняты

```bash
# Проверьте, что занимает порт
netstat -ano | findstr :9090
netstat -ano | findstr :9093
netstat -ano | findstr :4317
netstat -ano | findstr :9092

# Остановите контейнеры
docker-compose down

# Запустите заново
docker-compose up -d
```

### Если Docker Desktop не запускается

1. Перезапустите Docker Desktop
2. Проверьте, что WSL2 установлен (для Windows)
3. Убедитесь, что виртуализация включена в BIOS

### Если Order Service все еще не подключается

1. Убедитесь, что все Docker контейнеры запущены: `docker ps`
2. Проверьте логи контейнеров: `docker logs gobigtech_kafka`
3. Проверьте сеть: `docker network ls`
4. Убедитесь, что Order Service использует правильные адреса:
   - Kafka: `localhost:9092` (если запущен локально) или `kafka:29092` (если в Docker)
   - OTLP: `localhost:4317`

