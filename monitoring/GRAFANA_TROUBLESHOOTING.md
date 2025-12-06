# Решение проблемы "Dashboard not found" в Grafana

## Проблема

При попытке открыть дашборд Order Service в Grafana появляется ошибка:
```
Dashboard not found
We're looking but can't seem to find this dashboard.
```

## Причины

1. **Дашборд не был автоматически загружен через provisioning**
   - Grafana provisioning может не загружать дашборды при первом запуске
   - Файлы были добавлены после запуска контейнера

2. **Проблемы с конфигурацией datasource**
   - Отсутствие UID в конфигурации datasource
   - Несоответствие UID в дашборде и конфигурации

3. **Проблемы с путями и монтированием**
   - Неправильное монтирование volumes в docker-compose
   - Проблемы с правами доступа к файлам

## Решения

### Решение 1: Загрузка через API (Быстрое решение)

Используйте скрипт для загрузки дашборда:

```powershell
.\monitoring\grafana\load-dashboard.ps1
```

Или вручную через PowerShell:

```powershell
$json = Get-Content -Path "monitoring/grafana/provisioning/dashboards/order-service-dashboard.json" -Raw
$body = @{dashboard = ($json | ConvertFrom-Json); overwrite = $true} | ConvertTo-Json -Depth 20
Invoke-RestMethod -Uri "http://localhost:3000/api/dashboards/db" -Method Post `
    -Headers @{Authorization="Basic " + [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("admin:admin"))} `
    -ContentType "application/json" -Body $body
```

### Решение 2: Импорт через UI Grafana

1. Откройте Grafana: http://localhost:3000
2. Войдите (admin/admin)
3. Перейдите в **Dashboards** → **Import**
4. Нажмите **Upload JSON file**
5. Выберите файл: `monitoring/grafana/provisioning/dashboards/order-service-dashboard.json`
6. Нажмите **Load**
7. Выберите источник данных **Prometheus**
8. Нажмите **Import**

### Решение 3: Перезапуск Grafana с правильной конфигурацией

1. Убедитесь, что все файлы на месте:
   ```bash
   ls -la monitoring/grafana/provisioning/datasources/
   ls -la monitoring/grafana/provisioning/dashboards/
   ```

2. Проверьте конфигурацию datasource (должен быть UID):
   ```yaml
   uid: prometheus
   ```

3. Перезапустите Grafana:
   ```bash
   docker-compose restart grafana
   ```

4. Проверьте логи:
   ```bash
   docker logs gobigtech_grafana --tail 50 | grep -i dashboard
   ```

### Решение 4: Проверка и исправление конфигурации

1. **Проверьте UID datasource**:
   ```bash
   docker exec gobigtech_grafana cat /etc/grafana/provisioning/datasources/prometheus.yml
   ```
   
   Должно быть:
   ```yaml
   uid: prometheus
   ```

2. **Проверьте структуру файлов в контейнере**:
   ```bash
   docker exec gobigtech_grafana ls -la /etc/grafana/provisioning/dashboards/
   ```

3. **Проверьте JSON файл дашборда**:
   ```bash
   docker exec gobigtech_grafana cat /etc/grafana/provisioning/dashboards/order-service-dashboard.json | head -50
   ```

## Проверка работоспособности

После загрузки дашборда проверьте:

1. **Доступность через URL**:
   - http://localhost:3000/d/order-service-dashboard

2. **Через меню Grafana**:
   - Dashboards → Browse → Order Service - Метрики и Мониторинг

3. **Проверка источника данных**:
   - Configuration → Data Sources → Prometheus
   - Нажмите **Save & Test** - должно быть "Data source is working"

4. **Проверка метрик**:
   - Откройте дашборд
   - Убедитесь, что панели отображают данные (или "No data" если метрики еще не собраны)

## Автоматическая загрузка при следующем запуске

После успешной загрузки дашборда через API, он будет сохранен в базе данных Grafana и автоматически появится при следующих запусках.

Для обеспечения автоматической загрузки через provisioning:

1. Убедитесь, что конфигурация правильная
2. Перезапустите Grafana:
   ```bash
   docker-compose restart grafana
   ```
3. Проверьте логи на наличие ошибок:
   ```bash
   docker logs gobigtech_grafana 2>&1 | grep -i "dashboard\|provisioning"
   ```

## Предотвращение проблемы в будущем

1. **Всегда добавляйте UID в datasource конфигурацию**:
   ```yaml
   uid: prometheus
   ```

2. **Используйте скрипт загрузки** после изменений:
   ```powershell
   .\monitoring\grafana\load-dashboard.ps1
   ```

3. **Проверяйте логи при запуске**:
   ```bash
   docker logs gobigtech_grafana --tail 50
   ```

## Дополнительная информация

- Документация Grafana Provisioning: https://grafana.com/docs/grafana/latest/administration/provisioning/
- API документация: https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/

