# Session Verification Lua Script

## Описание

Скрипт `verify_session.lua` проверяет валидность сессии пользователя через IAM Service перед пропуском запроса к защищенным эндпоинтам.

## Функциональность

1. **Извлечение токена**: Извлекает токен из заголовка `Authorization` (формат `Bearer <token>` или просто токен)
2. **Проверка публичных эндпоинтов**: Пропускает запросы к публичным эндпоинтам без проверки:
   - `/api/iam/login`
   - `/api/iam/register`
   - `/v1/iam/login`
   - `/v1/iam/register`
   - `/health`
3. **Валидация сессии**: Выполняет HTTP запрос к IAM Service для проверки валидности сессии
4. **Добавление заголовков**: При успешной валидации добавляет заголовки:
   - `X-User-Id` - ID пользователя
   - `X-Session-Token` - токен сессии
5. **Обработка ошибок**: Возвращает 401 Unauthorized при невалидной сессии

## Интеграция в Envoy

Скрипт интегрирован в `envoy.yaml` как HTTP filter:

```yaml
http_filters:
  - name: envoy.filters.http.lua
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
      default_source_code:
        filename: /etc/envoy/lua/verify_session.lua
  - name: envoy.filters.http.router
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
```

## Использование

### Запрос с валидным токеном

```bash
curl -X GET http://localhost/api/orders \
  -H "Authorization: Bearer <session_token>"
```

Если токен валиден, запрос будет пропущен с заголовками:
- `X-User-Id: <user_id>`
- `X-Session-Token: <session_token>`

### Запрос без токена или с невалидным токеном

```bash
curl -X GET http://localhost/api/orders
# Ответ: 401 Unauthorized
```

### Публичные эндпоинты (не требуют токен)

```bash
curl -X POST http://localhost/api/iam/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'
```

## Конфигурация IAM Service

Скрипт делает запрос к:
- **Cluster**: `iam-service` (определен в envoy.yaml)
- **Endpoint**: `/v1/iam/validate-session`
- **Method**: POST
- **Body**: `{"session_id": "<token>"}`
- **Timeout**: 5 секунд

## Формат ответа IAM Service

IAM Service должен возвращать JSON в формате:

```json
{
  "valid": true,
  "user_id": "user-123",
  "email": "user@example.com"
}
```

Или при невалидной сессии:

```json
{
  "valid": false
}
```

## Логирование

Скрипт логирует:
- Публичные эндпоинты (info)
- Отсутствие токена (warn)
- Ошибки валидации (warn)
- Успешную валидацию (info)

Логи доступны через:
```bash
docker logs gobigtech_envoy
```

## Troubleshooting

### Сессия не валидируется

1. Проверьте, что IAM Service доступен:
   ```bash
   docker exec gobigtech_envoy wget -O- http://iam-service:8082/v1/iam/validate-session
   ```

2. Проверьте формат токена в заголовке Authorization

3. Проверьте логи Envoy:
   ```bash
   docker logs gobigtech_envoy | grep -i session
   ```

### Ошибка "cluster not found"

Убедитесь, что в `envoy.yaml` определен cluster `iam-service`:

```yaml
clusters:
  - name: iam-service
    # ... конфигурация
```

### Timeout при валидации

Увеличьте timeout в скрипте (по умолчанию 5 секунд):

```lua
local response = request_handle:httpCall(
  "iam-service",
  headers,
  request_body,
  10000  -- 10 секунд
)
```

