# Envoy API Gateway Configuration

## Обзор

Envoy API Gateway настроен как единая точка входа для всех HTTP сервисов проекта GoBigTech.

## Конфигурация

### Порты
- **HTTP внешний порт**: 80 (на хосте) → 8080 (в контейнере)
- **gRPC внешний порт**: 9091 (на хосте) → 9090 (в контейнере)
- **Admin порт**: 9901 (внутренний, для мониторинга)

### Маршрутизация

Envoy маршрутизирует запросы к следующим сервисам:

1. **Order Service** (`order-service:8080`)
   - `/orders/*` → Order Service
   - `/health` → Order Service (health check)
   - `/` → Order Service (default route)

2. **IAM Service** (`iam-service:8082`)
   - `/v1/iam/*` → IAM Service HTTP Gateway

3. **Notification Service** (`notification-service:8083`)
   - `/notifications/*` → Notification Service

### Health Checks

Envoy проверяет здоровье всех upstream сервисов:
- Order Service: `/health`
- IAM Service: `/v1/iam/health`
- Notification Service: `/health`

## Использование

### Запуск Envoy

```bash
docker-compose up -d envoy
```

### Проверка статуса

```bash
# Проверка health check
docker-compose ps envoy

# Просмотр логов
docker logs gobigtech_envoy

# Admin интерфейс (внутри контейнера)
docker exec gobigtech_envoy curl http://localhost:9901/server_info
```

### Доступ к сервисам через Envoy

Все запросы теперь идут через Envoy на порту 80:

```bash
# Order Service
curl http://localhost/orders
curl http://localhost/health

# IAM Service
curl http://localhost/v1/iam/register
curl http://localhost/v1/iam/login

# Notification Service
curl http://localhost/notifications/health
```

## Структура файлов

```
envoy/
├── envoy.yaml          # Основная конфигурация Envoy
└── lua/                # Lua скрипты для кастомной логики
    └── .gitkeep
```

## Важные замечания

1. **DNS имена**: Сервисы должны быть доступны по DNS именам в Docker сети:
   - `order-service` (порт 8080)
   - `iam-service` (порт 8082)
   - `notification-service` (порт 8083)

2. **Сеть**: Все сервисы должны быть в сети `gobigtech_network`

3. **Health checks**: Сервисы должны иметь рабочие health check endpoints

## Расширение конфигурации

### Добавление нового сервиса

1. Добавьте cluster в `envoy.yaml`:
```yaml
- name: new-service
  connect_timeout: 5s
  type: LOGICAL_DNS
  dns_lookup_family: V4_ONLY
  lb_policy: ROUND_ROBIN
  load_assignment:
    cluster_name: new-service
    endpoints:
      - lb_endpoints:
          - endpoint:
              address:
                socket_address:
                  address: new-service
                  port_value: 8080
```

2. Добавьте route в `virtual_hosts`:
```yaml
- match:
    prefix: "/new-service"
  route:
    cluster: new-service
```

3. Перезапустите Envoy:
```bash
docker-compose restart envoy
```

### Lua скрипты

Lua скрипты можно разместить в `envoy/lua/` и использовать для:
- Кастомной логики маршрутизации
- Трансформации запросов/ответов
- Логирования
- Аутентификации

## Troubleshooting

### Envoy не запускается

1. Проверьте конфигурацию:
```bash
docker run --rm -v $(pwd)/envoy/envoy.yaml:/etc/envoy/envoy.yaml envoyproxy/envoy:v1.30-latest envoy --config-path /etc/envoy/envoy.yaml --mode validate
```

2. Проверьте логи:
```bash
docker logs gobigtech_envoy
```

### Сервисы недоступны

1. Убедитесь, что сервисы запущены:
```bash
docker-compose ps
```

2. Проверьте DNS резолюцию:
```bash
docker exec gobigtech_envoy nslookup order-service
```

3. Проверьте сеть:
```bash
docker network inspect gobigtech_network
```

### Health checks не проходят

1. Проверьте, что health endpoints доступны:
```bash
curl http://order-service:8080/health
```

2. Проверьте конфигурацию health checks в `envoy.yaml`

