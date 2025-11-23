# Работа с Kafka топиками

## Автоматическое создание топиков

Топики создаются автоматически при запуске `docker-compose up` через сервис `kafka-init`. Этот сервис:
- Ждет, пока Kafka станет доступной
- Создает необходимые топики (`orders.payment`, `orders.assembly`)
- Завершает работу после создания

## Проверка существования топиков

### Способ 1: Через скрипт (рекомендуется)

```powershell
.\check-topics.ps1
```

Скрипт покажет:
- Список всех топиков
- Какие из необходимых топиков существуют
- Какие отсутствуют

### Способ 2: Через команду Docker

```powershell
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### Способ 3: Через Kafka UI

1. Откройте браузер: http://localhost:8081
2. Перейдите в раздел "Topics"
3. Увидите список всех топиков

## Ручное создание топиков

Если автоматическое создание не сработало, можно запустить скрипт вручную:

```powershell
.\create-topics.ps1
```

## Необходимые топики

- `orders.payment` - события оплаты заказов (1 партиция, replication factor 1)
- `orders.assembly` - события сборки заказов (1 партиция, replication factor 1)

## Проверка детальной информации о топике

```powershell
# Информация о топике orders.payment
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic orders.payment

# Информация о топике orders.assembly
docker exec gobigtech_kafka kafka-topics --bootstrap-server localhost:9092 --describe --topic orders.assembly
```

## Автоматическое создание топиков (KAFKA_AUTO_CREATE_TOPICS_ENABLE)

В `docker-compose.yml` установлено `KAFKA_AUTO_CREATE_TOPICS_ENABLE: 'true'`, что означает:
- Kafka автоматически создаст топик при первой записи
- Но это не рекомендуется для production, так как нельзя контролировать параметры (партиции, replication factor)

**Рекомендация**: Используйте сервис `kafka-init` для явного создания топиков с нужными параметрами.

