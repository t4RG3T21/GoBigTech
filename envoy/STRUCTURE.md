# Структура Envoy Configuration

## Обзор

Эта директория содержит конфигурацию Envoy API Gateway для проекта GoBigTech.

## Структура файлов

```
envoy/
├── envoy.yaml              # Основная конфигурация Envoy
├── lua/                    # Lua скрипты для кастомной логики
│   ├── README.md          # Документация по Lua скриптам
│   ├── request_logger.lua # Логирование запросов
│   ├── auth_header.lua    # Обработка заголовков аутентификации
│   └── rate_limiter.lua   # Rate limiting
├── README.md              # Основная документация
└── STRUCTURE.md           # Этот файл
```

## Конфигурация

### HTTP Listener (порт 8080)

Маршрутизация HTTP запросов:
- `/api/orders/*` → `order-service:8080`
- `/api/iam/*` → `iam-service:8082`
- `/api/notifications/*` → `notification-service:8083`
- `/health` → `order-service:8080` (health check)

### gRPC Listener (порт 9090)

Маршрутизация gRPC запросов:
- `/iam.v1.IAMService/*` → `iam-service:50053`
- `/inventory.v1.InventoryService/*` → `inventory-service:50051`
- `/payment.v1.PaymentService/*` → `payment-service:50052`

### Особенности конфигурации

1. **Retry Policies**
   - Автоматические повторы при ошибках 5xx, gateway-error, connect-failure
   - Максимум 3 попытки
   - Exponential backoff: 0.25s → 3s

2. **Circuit Breakers**
   - Максимум 1000 соединений на кластер
   - Максимум 1000 pending запросов
   - Максимум 3 retry

3. **Outlier Detection**
   - Автоматическое исключение нездоровых хостов
   - 5 последовательных ошибок 5xx → исключение на 30 секунд
   - Максимум 50% хостов могут быть исключены

4. **Health Checks**
   - HTTP health checks для всех HTTP сервисов
   - Интервал: 10 секунд
   - Timeout: 1 секунда
   - 3 неудачных проверки → unhealthy

5. **Access Logging**
   - JSON формат логов
   - Включает: timestamp, method, path, status code, duration, upstream info
   - Выводится в stdout (можно перенаправить в файл)

6. **Timeouts**
   - HTTP запросы: 30 секунд
   - gRPC запросы: 30 секунд
   - Health checks: 5 секунд

## Использование Lua скриптов

Lua скрипты можно использовать для:
- Кастомной логики маршрутизации
- Трансформации запросов/ответов
- Логирования
- Аутентификации
- Rate limiting

Пример подключения в `envoy.yaml`:
```yaml
http_filters:
  - name: envoy.filters.http.lua
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
      default_source_code:
        filename: /etc/envoy/lua/request_logger.lua
```

## Порты

| Порт (внешний) | Порт (внутренний) | Протокол | Назначение |
|----------------|-------------------|----------|------------|
| 80             | 8080              | HTTP     | HTTP API Gateway |
| 9091           | 9090              | gRPC     | gRPC API Gateway |
| -              | 9901              | HTTP     | Admin интерфейс (внутренний) |

**Примечание**: Внешний порт 9091 используется вместо 9090, так как 9090 занят Prometheus.

## Проверка конфигурации

```bash
# Валидация конфигурации
docker run --rm -v $(pwd)/envoy/envoy.yaml:/etc/envoy/envoy.yaml \
  envoyproxy/envoy:v1.30-latest \
  envoy --config-path /etc/envoy/envoy.yaml --mode validate
```

## Обновление конфигурации

После изменения `envoy.yaml`:
```bash
# Перезапуск Envoy
docker-compose restart envoy

# Или hot reload (если настроен)
docker exec gobigtech_envoy curl -X POST http://localhost:9901/hot_restart_epoch
```

