# Настройка Grafana для Order Service

## Обзор

Создан полноценный дашборд Grafana для визуализации метрик Order Service с интеграцией Prometheus.

## Структура файлов

```
monitoring/grafana/provisioning/
├── datasources/
│   └── prometheus.yml          # Конфигурация источника данных Prometheus
└── dashboards/
    ├── dashboard.yml            # Конфигурация автоматической загрузки дашбордов
    └── order-service-dashboard.json  # Дашборд Order Service
```

## Компоненты дашборда

### 1. Ключевые показатели (Stat Panels)
- **Всего заказов** - общее количество созданных заказов
- **Общая выручка** - суммарная выручка в USD
- **Неудачных заказов** - количество заказов с ошибками
- **Процент неудачных заказов** - процентное соотношение неудачных к общему количеству
- **Заказов в минуту** - скорость создания заказов
- **Выручка в минуту** - скорость генерации выручки

### 2. Скорость создания заказов (Time Series)
- График скорости создания заказов (заказов/секунду)
- Использует метрику: `rate(order_orders_created_total[1m])`

### 3. Скорость выручки (Time Series)
- График скорости генерации выручки (USD/секунду)
- Использует метрику: `rate(order_revenue_total[1m])`

### 4. Время обработки заказов (Time Series)
- Перцентили времени обработки: p50, p95, p99, max
- Использует гистограмму: `order_processing_duration_seconds`
- Запросы:
  - `histogram_quantile(0.50, ...)` - медиана
  - `histogram_quantile(0.95, ...)` - 95-й перцентиль
  - `histogram_quantile(0.99, ...)` - 99-й перцентиль
  - `histogram_quantile(1.0, ...)` - максимум

### 5. Неудачные заказы по причинам (Time Series)
- График неудачных заказов, сгруппированных по причинам
- Использует метрику: `sum by (reason) (rate(order_orders_failed_total[5m]))`
- Отображает различные причины ошибок (inventory_check_failed, payment_processing_failed, database_save_failed и т.д.)

### 6. HTTP запросы по методам и эндпоинтам (Time Series)
- График всех HTTP запросов, сгруппированных по методу и эндпоинту
- Использует метрику: `sum by (method, endpoint) (rate(order_http_requests_total[1m]))`
- Показывает нагрузку на различные эндпоинты

### 7. HTTP запросы по статусам (Bar Gauge)
- Визуализация распределения HTTP запросов по статус-кодам
- Использует метрику: `sum by (status) (rate(order_http_requests_total[1m]))`
- Цветовая индикация: зеленый (2xx), желтый (4xx), красный (5xx)

## Используемые метрики

Все метрики собираются из Prometheus и соответствуют метрикам, определенным в `services/order/internal/service/prometheus_metrics.go`:

1. **order_orders_created_total** - счетчик созданных заказов
2. **order_revenue_total** - счетчик выручки
3. **order_processing_duration_seconds** - гистограмма времени обработки
4. **order_orders_failed_total{reason}** - счетчик неудачных заказов по причинам
5. **order_http_requests_total{method, endpoint, status}** - счетчик HTTP запросов

## Запуск и настройка

### 1. Запуск Grafana

```bash
docker-compose up -d grafana
```

### 2. Доступ к Grafana

- URL: http://localhost:3000
- Логин: `admin`
- Пароль: `admin` (изменяется при первом входе)

### 3. Проверка источника данных

1. Откройте Grafana: http://localhost:3000
2. Перейдите в **Configuration** → **Data Sources**
3. Убедитесь, что **Prometheus** источник данных настроен и доступен
4. Нажмите **Save & Test** для проверки подключения

### 4. Просмотр дашборда

1. Перейдите в **Dashboards** → **Browse**
2. Найдите дашборд **"Order Service - Метрики и Мониторинг"**
3. Откройте дашборд

Или напрямую по URL: http://localhost:3000/d/order-service-dashboard

## Автоматическая загрузка

Дашборд автоматически загружается при запуске Grafana благодаря provisioning:

- Конфигурация datasource: `monitoring/grafana/provisioning/datasources/prometheus.yml`
- Конфигурация dashboards: `monitoring/grafana/provisioning/dashboards/dashboard.yml`
- Дашборд: `monitoring/grafana/provisioning/dashboards/order-service-dashboard.json`

## Обновление дашборда

1. Отредактируйте `order-service-dashboard.json`
2. Перезапустите Grafana:
   ```bash
   docker-compose restart grafana
   ```
3. Или обновите дашборд через UI Grafana (изменения сохранятся в JSON)

## Настройка времени обновления

По умолчанию дашборд обновляется каждые 10 секунд. Для изменения:

1. Откройте дашборд
2. Нажмите на иконку часов в правом верхнем углу
3. Выберите интервал обновления (например, 5s, 30s, 1m)

## Troubleshooting

### Дашборд не отображается

1. Проверьте логи Grafana:
   ```bash
   docker logs gobigtech_grafana
   ```

2. Убедитесь, что файлы находятся в правильных директориях:
   ```bash
   ls -la monitoring/grafana/provisioning/datasources/
   ls -la monitoring/grafana/provisioning/dashboards/
   ```

3. Проверьте права доступа к файлам

### Нет данных в панелях

1. Убедитесь, что Prometheus собирает метрики:
   ```bash
   curl http://localhost:9090/api/v1/targets
   ```

2. Проверьте, что Order Service экспортирует метрики:
   ```bash
   curl http://localhost:8080/metrics | grep order_
   ```

3. Проверьте конфигурацию Prometheus:
   ```bash
   docker exec gobigtech_prometheus cat /etc/prometheus/prometheus.yml
   ```

### Источник данных недоступен

1. Проверьте, что Prometheus запущен:
   ```bash
   docker ps | grep prometheus
   ```

2. Проверьте сеть Docker:
   ```bash
   docker network inspect gobigtech_network
   ```

3. Убедитесь, что в `prometheus.yml` указан правильный URL:
   ```yaml
   url: http://prometheus:9090
   ```

## Дополнительные возможности

### Экспорт дашборда

1. Откройте дашборд
2. Нажмите на иконку шестеренки (Settings)
3. Выберите **JSON Model**
4. Скопируйте JSON для сохранения или передачи

### Создание алертов

Алерты можно создать прямо из панелей:

1. Нажмите на панель
2. Выберите **Edit**
3. Перейдите на вкладку **Alert**
4. Настройте условия алерта
5. Сохраните

Альтернативно, используйте Prometheus Alerting Rules (см. `monitoring/alerts.yml`).

## Интеграция с Alertmanager

Дашборд интегрирован с системой алертов:

- Алерты определены в `monitoring/alerts.yml`
- Alertmanager настроен в `monitoring/alertmanager/alertmanager.yml`
- Уведомления отправляются в Telegram через Notification Service

## Следующие шаги

1. **Добавить больше панелей** для других метрик (если появятся)
2. **Настроить алерты** прямо в Grafana
3. **Создать дополнительные дашборды** для других сервисов
4. **Настроить переменные** для фильтрации данных
5. **Добавить аннотации** для отметки важных событий

