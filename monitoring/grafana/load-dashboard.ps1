# Скрипт для загрузки дашборда в Grafana через API
# Использование: .\load-dashboard.ps1

$grafanaUrl = "http://localhost:3000"
$username = "admin"
$password = "admin"
$dashboardPath = "monitoring/grafana/provisioning/dashboards/order-service-dashboard.json"

Write-Host "=== Загрузка дашборда в Grafana ===" -ForegroundColor Cyan
Write-Host ""

# Проверка доступности Grafana
try {
    $response = Invoke-WebRequest -Uri "$grafanaUrl/api/health" -Method Get -TimeoutSec 5
    Write-Host "✓ Grafana доступен" -ForegroundColor Green
} catch {
    Write-Host "✗ Grafana недоступен. Убедитесь, что контейнер запущен:" -ForegroundColor Red
    Write-Host "  docker-compose up -d grafana" -ForegroundColor Yellow
    exit 1
}

# Проверка существования файла дашборда
if (-not (Test-Path $dashboardPath)) {
    Write-Host "✗ Файл дашборда не найден: $dashboardPath" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Файл дашборда найден" -ForegroundColor Green

# Чтение JSON файла
$jsonContent = Get-Content -Path $dashboardPath -Raw
$dashboard = $jsonContent | ConvertFrom-Json

# Подготовка тела запроса
$body = @{
    dashboard = $dashboard
    overwrite = $true
} | ConvertTo-Json -Depth 20

# Создание заголовков авторизации
$base64Auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${username}:${password}"))
$headers = @{
    Authorization = "Basic $base64Auth"
    "Content-Type" = "application/json"
}

# Загрузка дашборда
try {
    Write-Host ""
    Write-Host "Загрузка дашборда..." -ForegroundColor Cyan
    $response = Invoke-RestMethod -Uri "$grafanaUrl/api/dashboards/db" -Method Post -Headers $headers -Body $body
    
    Write-Host ""
    Write-Host "✓ Дашборд успешно загружен!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Информация о дашборде:" -ForegroundColor Cyan
    Write-Host "  ID: $($response.id)" -ForegroundColor White
    Write-Host "  UID: $($response.uid)" -ForegroundColor White
    Write-Host "  URL: $grafanaUrl$($response.url)" -ForegroundColor White
    Write-Host ""
    Write-Host "Откройте дашборд в браузере:" -ForegroundColor Cyan
    Write-Host "  $grafanaUrl/d/$($response.uid)" -ForegroundColor Yellow
    Write-Host ""
    
} catch {
    Write-Host ""
    Write-Host "✗ Ошибка при загрузке дашборда:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    if ($_.ErrorDetails) {
        Write-Host $_.ErrorDetails.Message -ForegroundColor Red
    }
    exit 1
}

