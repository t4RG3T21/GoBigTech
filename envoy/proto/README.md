# Proto Descriptor для gRPC JSON Transcoding

## Описание

Эта директория содержит скомпилированные proto descriptor файлы для использования с `grpc_json_transcoder` в Envoy.

## Генерация Descriptor

### Требования

```bash
# Установка protoc
# Windows (chocolatey)
choco install protoc

# Linux
apt-get install protobuf-compiler

# macOS
brew install protobuf
```

### Генерация для Inventory и Payment

```bash
# Из корня проекта
protoc \
  --descriptor_set_out=envoy/proto/inventory_payment.pb \
  --include_imports \
  --include_source_info \
  api/proto/inventory/v1/inventory.proto \
  api/proto/payment/v1/payment.proto
```

### Проверка descriptor

```bash
# Просмотр содержимого
protoc --decode_raw < envoy/proto/inventory_payment.pb

# Или используя protoc-gen-grpc-gateway
protoc --decode google.protobuf.FileDescriptorSet envoy/proto/inventory_payment.pb
```

## Использование

Descriptor файл монтируется в Envoy контейнер:
```yaml
volumes:
  - ./envoy/proto:/etc/envoy/proto:ro
```

И используется в конфигурации:
```yaml
proto_descriptor: "/etc/envoy/proto/inventory_payment.pb"
```

## Обновление

При изменении proto файлов необходимо:
1. Перегенерировать descriptor
2. Перезапустить Envoy: `docker-compose restart envoy`

