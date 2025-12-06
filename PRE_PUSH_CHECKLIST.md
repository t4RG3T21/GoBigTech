# Pre-Push Checklist

Перед пушем на GitHub убедитесь, что:

## ✅ Проверки

### 1. Proto файлы
- [x] Все proto файлы синтаксически корректны
- [x] Proto файлы сгенерированы для всех сервисов
- [x] Proto descriptor для Envoy создан (`envoy/proto/inventory_payment.pb`)

### 2. Тесты
- [x] Все unit тесты проходят: `go test ./services/...`
- [x] Покрытие кода >= 40% для каждого сервиса
- [x] Интеграционные тесты проходят (если есть)

### 3. Docker
- [x] Все Dockerfile корректны
- [x] `docker-compose up -d` запускает все сервисы
- [x] Все сервисы становятся healthy

### 4. Envoy
- [x] Envoy конфигурация валидна (`envoy/envoy.yaml`)
- [x] Lua скрипты работают (`envoy/lua/verify_session.lua`)
- [x] Proto descriptor существует (`envoy/proto/inventory_payment.pb`)

### 5. Документация
- [x] README.md обновлен с Envoy архитектурой
- [x] Примеры API запросов актуальны
- [x] Troubleshooting раздел добавлен

### 6. CI/CD
- [x] `.github/workflows/ci.yml` корректна
- [x] Все пути к proto файлам правильные
- [x] Тесты запускаются для всех сервисов

### 7. Очистка
- [x] Удалены временные файлы (*.txt, temp_*)
- [x] Удалены бинарники (*.exe)
- [x] Удалены старые fix документации

## 🚀 Команды для проверки

```bash
# 1. Проверка proto файлов
docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.25-alpine sh -c \
  "apk add --no-cache protoc protobuf-dev git && \
   git clone --depth 1 https://github.com/googleapis/googleapis.git /tmp/googleapis && \
   protoc --proto_path=api/proto --proto_path=/tmp/googleapis \
          api/proto/iam/v1/iam.proto \
          api/proto/inventory/v1/inventory.proto \
          api/proto/payment/v1/payment.proto"

# 2. Запуск тестов
go test ./services/...

# 3. Проверка Docker
docker-compose config
docker-compose up -d
docker ps  # Проверить что все healthy

# 4. Тестирование через Envoy
.\test-envoy.ps1

# 5. Проверка Envoy конфигурации
docker exec gobigtech_envoy /usr/local/bin/envoy --config-path /etc/envoy/envoy.yaml --mode validate
```

## 📝 Что было обновлено

1. **README.md** - добавлена архитектура с Envoy, примеры API, troubleshooting
2. **Удалены временные файлы** - protoc_output*.txt, temp_*.txt, fix документация
3. **Проверены proto файлы** - все синтаксически корректны
4. **Проверены тесты** - 83 теста в 17 файлах
5. **CI конфигурация** - проверена, должна пройти успешно

## ⚠️ Важно

- Все сервисы должны быть доступны через Envoy на порту 80
- Прямой доступ к сервисам должен быть ограничен (только внутри Docker сети)
- Proto descriptor должен быть сгенерирован перед запуском Envoy

