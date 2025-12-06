# Быстрый старт Elasticsearch и Kibana

## Запуск

```bash
# Запуск всех компонентов ELK стека
docker-compose up -d elasticsearch kibana filebeat

# Проверка статуса
docker-compose ps elasticsearch kibana filebeat
```

## Доступ

- **Kibana UI**: http://localhost:5601
- **Elasticsearch API**: http://localhost:9200

## Первоначальная настройка Kibana

1. Откройте Kibana: http://localhost:5601
2. Перейдите в **Stack Management** → **Data Views**
3. Нажмите **Create data view**
4. Введите:
   - **Name**: `logs-*`
   - **Time field**: `@timestamp`
5. Нажмите **Create data view**

## Просмотр логов

1. Откройте **Discover** в Kibana
2. Выберите data view: `logs-*`
3. Логи отображаются в реальном времени

## Поиск логов

### Поиск по сервису
```
json.service: "order-service"
```

### Поиск ошибок
```
json.level: "error"
```

### Комбинированный поиск
```
json.service: "order-service" AND json.level: "error"
```

## Проверка работы

```bash
# Проверка Elasticsearch
curl http://localhost:9200/_cluster/health

# Проверка индексов
curl http://localhost:9200/_cat/indices?v

# Проверка логов Filebeat
docker logs gobigtech_filebeat
```

## Troubleshooting

### Логи не появляются

1. Убедитесь, что сервисы используют `LOG_LEVEL=info` (для JSON формата)
2. Проверьте логи Filebeat: `docker logs gobigtech_filebeat`
3. Проверьте, что Elasticsearch здоров: 
   ```powershell
   Invoke-WebRequest -Uri http://localhost:9200/_cluster/health -UseBasicParsing
   ```
4. Проверьте индексы:
   ```powershell
   Invoke-WebRequest -Uri http://localhost:9200/_cat/indices?v -UseBasicParsing
   ```
5. Проверьте логи контейнеров: `docker logs gobigtech_order-service`

### Filebeat ошибка с правами доступа

Если Filebeat выдает ошибку о правах доступа к конфигурации, это уже решено добавлением флага `-strict.perms=false` в docker-compose.yml.

### Проверка количества логов

```powershell
# Проверить количество документов в индексе
Invoke-WebRequest -Uri "http://localhost:9200/_cat/indices?v" -UseBasicParsing
```

Вы должны увидеть индекс `logs-YYYY.MM.DD` с количеством документов > 0.

