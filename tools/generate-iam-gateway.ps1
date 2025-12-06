# Скрипт для генерации gRPC Gateway кода для IAM Service через Docker
# Использует Docker для избежания проблем с кодировкой на Windows

Write-Host "=== Generating IAM Service gRPC Gateway Code ===" -ForegroundColor Green
Write-Host ""

# Проверяем, что Docker запущен
$null = docker ps 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Docker is not running. Please start Docker first." -ForegroundColor Red
    exit 1
}

Write-Host "Step 1: Installing protoc-gen-grpc-gateway and protoc-gen-openapiv2..." -ForegroundColor Cyan
Write-Host "Using golang:1.25-alpine (required for grpc-gateway v2)" -ForegroundColor Yellow
Write-Host ""

# Генерируем gateway код используя Docker
# Используем golang:1.25-alpine, так как grpc-gateway требует Go 1.24+
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.25-alpine sh -c "apk add --no-cache protoc protobuf-dev git && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0 && if [ ! -d /tmp/googleapis ]; then git clone --depth 1 https://github.com/googleapis/googleapis.git /tmp/googleapis; fi && protoc --go_out=services --go_opt=paths=source_relative --go-grpc_out=services --go-grpc_opt=paths=source_relative --grpc-gateway_out=services --grpc-gateway_opt=paths=source_relative --grpc-gateway_opt=logtostderr=true --openapiv2_out=services/iam/v1 --openapiv2_opt=logtostderr=true,allow_merge=true --proto_path=api/proto --proto_path=/tmp/googleapis api/proto/iam/v1/iam.proto && echo 'Gateway code generation completed!'"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Gateway code generated successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Generated files:" -ForegroundColor Cyan
    Write-Host "  - services/iam/v1/iam.pb.go" -ForegroundColor White
    Write-Host "  - services/iam/v1/iam_grpc.pb.go" -ForegroundColor White
    Write-Host "  - services/iam/v1/iam.pb.gw.go (gateway)" -ForegroundColor White
    Write-Host "  - services/iam/v1/iam.swagger.json (OpenAPI)" -ForegroundColor White
    Write-Host ""
    Write-Host "You can now compile the IAM service!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Failed to generate gateway code" -ForegroundColor Red
    Write-Host "Check the error messages above" -ForegroundColor Yellow
    exit 1
}
