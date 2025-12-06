# Скрипт для тестирования всего Observability стека

Write-Host "=== Тестирование Observability стека ===" -ForegroundColor Cyan

# 1. Проверка метрик Prometheus
Write-Host "`n1. Проверка метрик Prometheus..." -ForegroundColor Yellow
try {
    $promTargets = Invoke-WebRequest -Uri "http://localhost:9090/api/v1/targets" -UseBasicParsing | ConvertFrom-Json
    $upTargets = ($promTargets.data.activeTargets | Where-Object { $_.health -eq "up" }).Count
    Write-Host "   ✓ Prometheus targets: $upTargets активных" -ForegroundColor Green
    
    # Проверка метрик Order Service
    try {
        $orderMetrics = Invoke-WebRequest -Uri "http://localhost:8080/metrics" -UseBasicParsing
        if ($orderMetrics.Content -match "order_http_requests_total") {
            Write-Host "   ✓ Order Service метрики доступны" -ForegroundColor Green
        } else {
            Write-Host "   ✗ Order Service метрики не найдены" -ForegroundColor Red
        }
    } catch {
        Write-Host "   ✗ Order Service недоступен: $_" -ForegroundColor Red
    }
} catch {
    Write-Host "   ✗ Ошибка проверки Prometheus: $_" -ForegroundColor Red
}

# 2. Проверка алертов
Write-Host "`n2. Проверка алертов..." -ForegroundColor Yellow
try {
    $alerts = Invoke-WebRequest -Uri "http://localhost:9090/api/v1/alerts" -UseBasicParsing | ConvertFrom-Json
    $activeAlerts = ($alerts.data.alerts | Where-Object { $_.state -eq "firing" }).Count
    $pendingAlerts = ($alerts.data.alerts | Where-Object { $_.state -eq "pending" }).Count
    Write-Host "   ✓ Активных алертов: $activeAlerts" -ForegroundColor Green
    Write-Host "   ✓ Pending алертов: $pendingAlerts" -ForegroundColor Yellow
    
    # Проверка Alertmanager
    try {
        $alertmanager = Invoke-WebRequest -Uri "http://localhost:9093/api/v2/status" -UseBasicParsing
        Write-Host "   ✓ Alertmanager доступен" -ForegroundColor Green
    } catch {
        Write-Host "   ✗ Alertmanager недоступен" -ForegroundColor Red
    }
} catch {
    Write-Host "   ✗ Ошибка проверки алертов: $_" -ForegroundColor Red
}

# 3. Проверка трассировки Jaeger
Write-Host "`n3. Проверка трассировки Jaeger..." -ForegroundColor Yellow
try {
    $services = Invoke-WebRequest -Uri "http://localhost:16686/api/services" -UseBasicParsing | ConvertFrom-Json
    Write-Host "   ✓ Найдено сервисов в Jaeger: $($services.data.Count)" -ForegroundColor Green
    foreach ($service in $services.data) {
        Write-Host "     - $service" -ForegroundColor Gray
    }
} catch {
    Write-Host "   ✗ Ошибка проверки Jaeger: $_" -ForegroundColor Red
}

# 4. Проверка логов Elasticsearch/Kibana
Write-Host "`n4. Проверка логов Elasticsearch/Kibana..." -ForegroundColor Yellow
try {
    $indices = Invoke-WebRequest -Uri "http://localhost:9200/_cat/indices?v" -UseBasicParsing
    if ($indices.Content -match "logs-") {
        $logCount = ($indices.Content -split "`n" | Select-String "logs-" | Measure-Object).Count
        Write-Host "   ✓ Найдено индексов логов: $logCount" -ForegroundColor Green
        
        # Подсчет документов
        $indexInfo = Invoke-WebRequest -Uri "http://localhost:9200/_cat/indices/logs-*?v&h=index,docs.count" -UseBasicParsing
        Write-Host "   $($indexInfo.Content)" -ForegroundColor Gray
    } else {
        Write-Host "   ✗ Индексы логов не найдены" -ForegroundColor Red
    }
    
    # Проверка Kibana
    try {
        $kibana = Invoke-WebRequest -Uri "http://localhost:5601/api/status" -UseBasicParsing
        Write-Host "   ✓ Kibana доступен" -ForegroundColor Green
    } catch {
        Write-Host "   ✗ Kibana недоступен" -ForegroundColor Red
    }
} catch {
    Write-Host "   ✗ Ошибка проверки логов: $_" -ForegroundColor Red
}

# 5. Проверка Grafana
Write-Host "`n5. Проверка Grafana..." -ForegroundColor Yellow
try {
    $grafana = Invoke-WebRequest -Uri "http://localhost:3000/api/health" -UseBasicParsing
    Write-Host "   ✓ Grafana доступен" -ForegroundColor Green
} catch {
    Write-Host "   ✗ Grafana недоступен" -ForegroundColor Red
}

# 6. Проверка OpenTelemetry Collector
Write-Host "`n6. Проверка OpenTelemetry Collector..." -ForegroundColor Yellow
try {
    $otelMetrics = Invoke-WebRequest -Uri "http://localhost:8889/metrics" -UseBasicParsing
    if ($otelMetrics.Content -match "order_otel") {
        Write-Host "   ✓ OpenTelemetry Collector метрики доступны" -ForegroundColor Green
    } else {
        Write-Host "   ⚠ OpenTelemetry Collector метрики пусты" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   ✗ Ошибка проверки OpenTelemetry Collector: $_" -ForegroundColor Red
}

Write-Host "`n=== Тестирование завершено ===" -ForegroundColor Cyan
