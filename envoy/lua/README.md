# Envoy Lua Scripts

Эта папка содержит Lua скрипты для кастомной логики в Envoy.

## Использование

Lua скрипты можно использовать для:
- Кастомной маршрутизации
- Трансформации запросов/ответов
- Логирования
- Аутентификации
- Rate limiting
- Добавления заголовков

## Примеры

### request_logger.lua
Логирует детальную информацию о запросах.

### auth_header.lua
Добавляет заголовки аутентификации к запросам.

### rate_limiter.lua
Реализует rate limiting на основе IP адреса.

## Подключение в envoy.yaml

```yaml
http_filters:
  - name: envoy.filters.http.lua
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
      inline_code: |
        -- Lua код здесь
      default_source_code:
        inline_string: |
          -- Или загрузить из файла
```

Или использовать file_system:
```yaml
http_filters:
  - name: envoy.filters.http.lua
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
      default_source_code:
        filename: /etc/envoy/lua/request_logger.lua
```

