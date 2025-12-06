# Troubleshooting Envoy gRPC Routing

## Проблема: Не удается подключиться к Envoy

### 1. Проверка Docker Desktop

Убедитесь, что Docker Desktop запущен:

```powershell
# Проверка статуса Docker
docker ps

# Если ошибка "cannot connect to Docker daemon", запустите Docker Desktop
```

### 2. Проверка запущенных контейнеров

```bash
# Проверка всех контейнеров
docker ps -a

# Проверка конкретно Envoy
docker ps --filter "name=envoy"

# Проверка логов Envoy
docker logs gobigtech_envoy --tail 100
```

### 3. Запуск сервисов

Если контейнеры не запущены:

```bash
# Запуск всех сервисов
docker-compose up -d

# Или только Envoy
docker-compose up -d envoy

# Проверка статуса
docker-compose ps
```

### 4. Проверка портов

```bash
# Windows
netstat -ano | findstr ":8084 :9091 :80"

# Linux/Mac
netstat -tuln | grep -E ":(8084|9091|80)"
```

### 5. Проверка конфигурации Envoy

```bash
# Проверка валидности конфигурации
docker exec gobigtech_envoy envoy --config-path /etc/envoy/envoy.yaml --mode validate

# Просмотр конфигурации
docker exec gobigtech_envoy cat /etc/envoy/envoy.yaml
```

### 6. Проверка proto descriptor

```bash
# Проверка наличия файла
ls -la envoy/proto/inventory_payment.pb

# Проверка внутри контейнера
docker exec gobigtech_envoy ls -la /etc/envoy/proto/

# Если файл отсутствует, сгенерируйте его
.\scripts\generate-proto-descriptor.ps1
```

### 7. Частые ошибки

#### Ошибка: "proto descriptor not found"
**Решение**: Сгенерируйте proto descriptor:
```powershell
.\scripts\generate-proto-descriptor.ps1
docker-compose restart envoy
```

#### Ошибка: "address already in use"
**Решение**: Проверьте, какой процесс использует порт:
```bash
# Windows
netstat -ano | findstr ":8084"
taskkill /PID <PID> /F

# Linux/Mac
lsof -i :8084
kill -9 <PID>
```

#### Ошибка: "failed to connect"
**Решение**: 
1. Убедитесь, что Docker Desktop запущен
2. Проверьте, что контейнер запущен: `docker ps`
3. Проверьте логи: `docker logs gobigtech_envoy`

#### Ошибка: "invalid configuration"
**Решение**: 
1. Проверьте синтаксис YAML
2. Проверьте валидность конфигурации:
   ```bash
   docker exec gobigtech_envoy envoy --config-path /etc/envoy/envoy.yaml --mode validate
   ```

### 8. Тестирование подключения

После запуска всех сервисов:

```bash
# Проверка HTTP API Gateway
curl http://localhost:80/health

# Проверка gRPC (требует grpcurl)
grpcurl -plaintext localhost:9091 list

# Проверка REST→gRPC Gateway
curl -X POST http://localhost:8084/inventory.v1.InventoryService/GetStock \
  -H "Content-Type: application/json" \
  -d '{"product_id": "test"}'
```

### 9. Полный перезапуск

Если ничего не помогает:

```bash
# Остановка всех сервисов
docker-compose down

# Очистка (опционально)
docker-compose down -v

# Запуск заново
docker-compose up -d

# Проверка логов
docker-compose logs -f envoy
```

### 10. Проверка сетевых настроек

```bash
# Проверка сети Docker
docker network ls
docker network inspect gobigtech_network

# Проверка подключения контейнеров
docker exec gobigtech_envoy ping inventory-service
docker exec gobigtech_envoy ping payment-service
```

