# Скрипт для валидации proto файлов

Write-Host "=== Валидация proto файлов ===" -ForegroundColor Cyan

$errors = 0

# Проверка наличия protoc
try {
    $protocVersion = protoc --version 2>&1
    Write-Host "✓ protoc найден: $protocVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ protoc не найден. Установите protoc или используйте Docker" -ForegroundColor Red
    $errors++
    exit 1
}

# Проверка IAM proto
Write-Host "`nПроверка api/proto/iam/v1/iam.proto..." -ForegroundColor Yellow
try {
    # Скачиваем googleapis если нужно
    if (-not (Test-Path "/tmp/googleapis")) {
        Write-Host "Скачивание googleapis..." -ForegroundColor Gray
        git clone --depth 1 https://github.com/googleapis/googleapis.git /tmp/googleapis 2>&1 | Out-Null
    }
    
    protoc --proto_path=api/proto --proto_path=/tmp/googleapis --descriptor_set_out=/dev/null api/proto/iam/v1/iam.proto 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ iam.proto валиден" -ForegroundColor Green
    } else {
        Write-Host "✗ Ошибка в iam.proto" -ForegroundColor Red
        $errors++
    }
} catch {
    Write-Host "✗ Ошибка проверки iam.proto: $_" -ForegroundColor Red
    $errors++
}

# Проверка Inventory proto
Write-Host "`nПроверка api/proto/inventory/v1/inventory.proto..." -ForegroundColor Yellow
try {
    protoc --proto_path=api/proto --descriptor_set_out=/dev/null api/proto/inventory/v1/inventory.proto 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ inventory.proto валиден" -ForegroundColor Green
    } else {
        Write-Host "✗ Ошибка в inventory.proto" -ForegroundColor Red
        $errors++
    }
} catch {
    Write-Host "✗ Ошибка проверки inventory.proto: $_" -ForegroundColor Red
    $errors++
}

# Проверка Payment proto
Write-Host "`nПроверка api/proto/payment/v1/payment.proto..." -ForegroundColor Yellow
try {
    protoc --proto_path=api/proto --descriptor_set_out=/dev/null api/proto/payment/v1/payment.proto 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ payment.proto валиден" -ForegroundColor Green
    } else {
        Write-Host "✗ Ошибка в payment.proto" -ForegroundColor Red
        $errors++
    }
} catch {
    Write-Host "✗ Ошибка проверки payment.proto: $_" -ForegroundColor Red
    $errors++
}

Write-Host "`n=== Результат ===" -ForegroundColor Cyan
if ($errors -eq 0) {
    Write-Host "✓ Все proto файлы валидны" -ForegroundColor Green
    exit 0
} else {
    Write-Host "✗ Найдено ошибок: $errors" -ForegroundColor Red
    exit 1
}

