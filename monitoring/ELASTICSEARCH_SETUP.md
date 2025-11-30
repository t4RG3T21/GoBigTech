# Настройка Elasticsearch и Kibana для сбора логов

## Архитектура

```
Docker Containers (сервисы)
    ↓ (stdout/stderr → JSON logs)
Docker Logging Driver (json-file)
    ↓
Filebeat (сбор логов)
    ↓
Elasticsearch (хранение)
    ↓
Kibana (визуализация и поиск)
```

## Компоненты

### 1. Elasticsearch
- Хранит логи в индексах
- Индексы создаются автоматически по паттерну `logs-YYYY.MM.DD`
- Порт: `9200` (HTTP API), `9300` (Transport)

### 2. Kibana
- Web UI для просмотра и поиска логов
- Порт: `5601`
- Автоматически подключается к Elasticsearch

### 3. Filebeat
- Собирает логи из Docker контейнеров
- Парсит JSON логи
- Отправляет в Elasticsearch

## Запуск

### 1. Запуск всех компонентов

```bash
docker-compose up -d elasticsearch kibana filebeat
```

**Примечание:** Если возникает ошибка 403 при загрузке образов, убедитесь, что используются публичные образы из Docker Hub (уже настроено в docker-compose.yml).

### 2. Проверка статуса

**Elasticsearch:**
```bash
curl http://localhost:9200/_cluster/health
```

**Kibana:**
```bash
curl http://localhost:5601/api/status
```

### 3. Доступ к Kibana UI

Откройте в браузере: `http://localhost:5601`

## Настройка индекс-паттернов в Kibana

### Автоматическое создание

Filebeat автоматически создаст индекс-паттерн при первом запуске.

### Ручное создание

1. Откройте Kibana: `http://localhost:5601`
2. Перейдите в **Stack Management** → **Index Patterns**
3. Нажмите **Create index pattern**
4. Введите паттерн: `logs-*`
5. Выберите поле времени: `@timestamp`
6. Нажмите **Create index pattern**

## Просмотр логов

### 1. Discover

1. Откройте Kibana: `http://localhost:5601`
2. Перейдите в **Discover**
3. Выберите индекс-паттерн: `logs-*`
4. Логи отображаются в реальном времени

### 2. Поиск по сервису

Используйте фильтры:
- `container.name: "gobigtech_order-service"`
- `json.service: "order-service"`
- `json.level: "error"`

### 3. Поиск по тексту

Введите в поисковую строку:
- `"error"` - найти все ошибки
- `"order"` - найти все логи с упоминанием "order"
- `json.service: "order-service" AND json.level: "error"` - ошибки в Order Service

## Структура логов

Логи в Elasticsearch имеют следующую структуру:

```json
{
  "@timestamp": "2024-01-15T10:30:00.000Z",
  "container": {
    "name": "gobigtech_order-service",
    "id": "...",
    "labels": {...}
  },
  "json": {
    "level": "info",
    "ts": 1705315800.123,
    "caller": "main.go:123",
    "msg": "Service started",
    "service": "order-service",
    "http_port": "8080"
  },
  "message": "{\"level\":\"info\",\"ts\":1705315800.123,\"msg\":\"Service started\"}",
  "log": {
    "file": {
      "path": "/var/lib/docker/containers/.../...-json.log"
    }
  }
}
```

## Полезные запросы KQL (Kibana Query Language)

### Все логи Order Service
```
json.service: "order-service"
```

### Только ошибки
```
json.level: "error"
```

### Ошибки в Order Service за последний час
```
json.service: "order-service" AND json.level: "error"
```

### Логи с определенным сообщением
```
message: "failed"
```

### Комбинация условий
```
json.service: "order-service" AND (json.level: "error" OR json.level: "warn")
```

## Настройка логгера

Логгер уже настроен для вывода JSON в production режиме:

```go
logger, err := platformlogger.New("order-service", "info") // JSON формат
logger, err := platformlogger.New("order-service", "debug") // Консольный формат
```

Для Elasticsearch рекомендуется использовать уровень `info` или выше, чтобы логи были в JSON формате.

## Мониторинг

### Проверка индексов в Elasticsearch

```bash
curl http://localhost:9200/_cat/indices?v
```

### Проверка статуса Filebeat

```bash
docker logs gobigtech_filebeat
```

### Проверка количества логов

В Kibana:
1. Откройте **Discover**
2. Выберите индекс-паттерн `logs-*`
3. Внизу страницы отображается количество документов

## Troubleshooting

### Логи не появляются в Kibana

1. Проверьте, что Filebeat запущен:
   ```bash
   docker ps | grep filebeat
   ```

2. Проверьте логи Filebeat:
   ```bash
   docker logs gobigtech_filebeat
   ```

3. Проверьте, что Elasticsearch доступен:
   ```bash
   curl http://localhost:9200/_cluster/health
   ```

4. Проверьте, что логи пишутся в контейнеры:
   ```bash
   docker logs gobigtech_order-service | head -20
   ```

### Filebeat не может подключиться к Elasticsearch

1. Убедитесь, что Elasticsearch запущен и здоров
2. Проверьте сеть Docker:
   ```bash
   docker network inspect gobigtech_network
   ```

### Индекс-паттерн не создается автоматически

1. Создайте вручную через Kibana UI
2. Или используйте API:
   ```bash
   curl -X POST "http://localhost:5601/api/index_patterns/index_pattern" \
     -H "kbn-xsrf: true" \
     -H "Content-Type: application/json" \
     -d '{
       "index_pattern": {
         "title": "logs-*",
         "timeFieldName": "@timestamp"
       }
     }'
   ```

## Оптимизация

### Уменьшение размера индексов

В `monitoring/filebeat/filebeat.yml`:
- `index.number_of_shards: 1` - один шард
- `index.number_of_replicas: 0` - без реплик (для разработки)
- `index.codec: best_compression` - лучшее сжатие

### Очистка старых логов

Для production рекомендуется настроить ILM (Index Lifecycle Management):
- Автоматическое удаление логов старше N дней
- Оптимизация индексов

## Переменные окружения

Для сервисов рекомендуется установить:
- `LOG_LEVEL=info` - для JSON формата логов
- `ENVIRONMENT=production` - для production режима

