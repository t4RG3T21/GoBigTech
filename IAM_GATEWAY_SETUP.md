# Инструкция по настройке HTTP Gateway для IAM Service

Эта инструкция поможет вам настроить HTTP Gateway для IAM Service, чтобы тестировать API через Postman.

## ⚠️ ВАЖНО: Порядок выполнения

**Обязательно выполните шаги в указанном порядке!** Gateway код должен быть сгенерирован ПЕРЕД компиляцией сервиса.

## Шаг 1: Обновление зависимостей

Сначала нужно добавить зависимости для gRPC Gateway в `go.mod`:

```powershell
cd services/iam
go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
go get github.com/grpc-ecosystem/grpc-gateway/v2/runtime
go get golang.org/x/net/http2
go get golang.org/x/net/http2/h2c
go mod tidy
```

## Шаг 2: Генерация Gateway кода через Docker

Запустите скрипт для генерации gateway кода:

```powershell
cd C:\Users\t4rg3t\GolandProjects\GoBigTech
.\tools\generate-iam-gateway.ps1
```

Этот скрипт:
- Установит необходимые инструменты в Docker контейнере
- Сгенерирует Go код для proto файлов
- Сгенерирует gateway код (`iam.pb.gw.go`)
- Сгенерирует OpenAPI спецификацию (`iam.swagger.json`)

**Ожидаемый результат:**
- `services/iam/v1/iam.pb.go` (обновлен)
- `services/iam/v1/iam_grpc.pb.go` (обновлен)
- `services/iam/v1/iam.pb.gw.go` (новый файл - gateway)
- `services/iam/v1/iam.swagger.json` (новый файл - OpenAPI)

## Шаг 3: Проверка сгенерированных файлов

Убедитесь, что файлы созданы:

```powershell
ls services/iam/v1/
```

Должны быть видны:
- `iam.pb.go`
- `iam_grpc.pb.go`
- `iam.pb.gw.go` ← новый файл gateway
- `iam.swagger.json` ← новый файл OpenAPI

## Шаг 4: Запуск IAM Service

Запустите IAM Service:

```powershell
cd services/iam/cmd/iam
go run main.go
```

Вы должны увидеть в логах:
```
Starting IAM Service
  grpc_port: 50053
  http_port: 8082  ← HTTP Gateway порт
```

## Шаг 5: Тестирование через Postman

### 5.1. Регистрация пользователя

**Метод:** `POST`  
**URL:** `http://localhost:8082/v1/iam/register`  
**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "email": "test@example.com",
  "password": "password123",
  "username": "testuser"
}
```

**Ожидаемый ответ:**
```json
{
  "success": true,
  "user_id": "uuid-here",
  "message": ""
}
```

### 5.2. Вход пользователя (по email)

**Метод:** `POST`  
**URL:** `http://localhost:8082/v1/iam/login`  
**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

**Ожидаемый ответ:**
```json
{
  "success": true,
  "session_id": "session-uuid-here",
  "user_id": "user-uuid-here",
  "message": ""
}
```

### 5.3. Вход пользователя (по username)

**Метод:** `POST`  
**URL:** `http://localhost:8082/v1/iam/login`  
**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**Ожидаемый ответ:**
```json
{
  "success": true,
  "session_id": "session-uuid-here",
  "user_id": "user-uuid-here",
  "message": ""
}
```

### 5.4. Валидация сессии

**Метод:** `POST`  
**URL:** `http://localhost:8082/v1/iam/validate-session`  
**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "session_id": "your-session-id-from-login"
}
```

**Ожидаемый ответ:**
```json
{
  "valid": true,
  "user_id": "user-uuid-here",
  "email": "test@example.com"
}
```

### 5.5. Получение информации о пользователе

**Метод:** `GET`  
**URL:** `http://localhost:8082/v1/iam/user/{user_id}`  

**Пример:**
```
GET http://localhost:8082/v1/iam/user/123e4567-e89b-12d3-a456-426614174000
```

**Ожидаемый ответ:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "test@example.com",
  "name": "testuser",
  "created_at": 1234567890
}
```

## Шаг 6: Просмотр OpenAPI спецификации

После генерации gateway кода, вы можете открыть файл `services/iam/v1/iam.swagger.json` в любом редакторе или использовать Swagger UI для просмотра API документации.

## Troubleshooting

### Ошибка: "failed to register gateway handlers"

**Причина:** gRPC сервер еще не запущен или порт занят.

**Решение:**
1. Убедитесь, что gRPC сервер запущен на порту 50053
2. Проверьте, что порт 8082 свободен для HTTP Gateway

### Ошибка: "no such file or directory" при генерации

**Причина:** Docker не может найти файлы.

**Решение:**
1. Убедитесь, что вы запускаете скрипт из корня проекта
2. Проверьте, что Docker запущен: `docker ps`

### Ошибка компиляции: "package not found"

**Причина:** Зависимости не установлены.

**Решение:**
```powershell
cd services/iam
go mod tidy
go mod download
```

## Полезные команды

### Проверка портов
```powershell
# Проверить, занят ли порт 8082
netstat -ano | findstr :8082

# Проверить, занят ли порт 50053
netstat -ano | findstr :50053
```

### Остановка процесса на порту
```powershell
# Найти PID процесса
netstat -ano | findstr :8082

# Остановить процесс (замените <PID> на реальный PID)
taskkill /PID <PID> /F
```

## Дополнительная информация

- **gRPC порт:** 50053 (для прямых gRPC вызовов)
- **HTTP Gateway порт:** 8082 (для HTTP/REST вызовов через Postman)
- **OpenAPI спецификация:** `services/iam/v1/iam.swagger.json`

Все HTTP endpoints автоматически проксируются на gRPC сервер, поэтому вы можете использовать как gRPC клиенты, так и обычные HTTP клиенты (Postman, curl, и т.д.).

