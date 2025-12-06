# Скрипт для генерации proto descriptor файла для Envoy gRPC JSON Transcoding
# Использование: .\scripts\generate-proto-descriptor.ps1

$ErrorActionPreference = "Stop"

Write-Host "Generating proto descriptor for Envoy gRPC JSON Transcoding..." -ForegroundColor Cyan

# Проверка наличия protoc
$protocPath = Get-Command protoc -ErrorAction SilentlyContinue
if (-not $protocPath) {
    Write-Host "Error: protoc not found. Please install Protocol Buffers compiler." -ForegroundColor Red
    Write-Host "Windows: choco install protoc" -ForegroundColor Yellow
    Write-Host "Linux: apt-get install protobuf-compiler" -ForegroundColor Yellow
    Write-Host "macOS: brew install protobuf" -ForegroundColor Yellow
    exit 1
}

# Создание директории для proto descriptor
$protoDir = "envoy/proto"
if (-not (Test-Path $protoDir)) {
    New-Item -ItemType Directory -Path $protoDir -Force | Out-Null
    Write-Host "Created directory: $protoDir" -ForegroundColor Green
}

# Генерация descriptor для inventory и payment
$descriptorFile = "$protoDir/inventory_payment.pb"
$inventoryProto = "api/proto/inventory/v1/inventory.proto"
$paymentProto = "api/proto/payment/v1/payment.proto"

Write-Host "Generating descriptor from:" -ForegroundColor Yellow
Write-Host "  - $inventoryProto" -ForegroundColor Gray
Write-Host "  - $paymentProto" -ForegroundColor Gray

protoc `
    --descriptor_set_out=$descriptorFile `
    --include_imports `
    --include_source_info `
    $inventoryProto `
    $paymentProto

if ($LASTEXITCODE -eq 0) {
    Write-Host "Proto descriptor generated successfully: $descriptorFile" -ForegroundColor Green
    
    # Проверка размера файла
    $fileSize = (Get-Item $descriptorFile).Length
    Write-Host "File size: $([math]::Round($fileSize / 1KB, 2)) KB" -ForegroundColor Gray
} else {
    Write-Host "Error: Failed to generate proto descriptor" -ForegroundColor Red
    exit 1
}

Write-Host "`nProto descriptor is ready for Envoy!" -ForegroundColor Green
Write-Host "Restart Envoy to use the new descriptor:" -ForegroundColor Yellow
Write-Host "  docker-compose restart envoy" -ForegroundColor Gray

