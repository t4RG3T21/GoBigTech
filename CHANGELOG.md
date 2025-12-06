# Changelog

## [Unreleased] - 2025-12-05

### Added
- **Envoy API Gateway** - единая точка входа для всех HTTP и gRPC запросов
  - HTTP API на порту 80
  - gRPC на порту 9091
  - REST→gRPC Gateway на порту 8084
- **Lua аутентификация** - проверка сессий через IAM Service
- **Автоматическое тестирование** - скрипт `test-envoy.ps1` для проверки всей системы
- **Docker контейнеризация** - все сервисы запускаются через Docker Compose
- **Multi-stage Dockerfiles** - оптимизированные образы для всех сервисов

### Changed
- **Архитектура** - все запросы теперь идут через Envoy API Gateway
- **Маршрутизация** - HTTP запросы маршрутизируются через Envoy
- **Аутентификация** - централизована через Envoy Lua фильтры
- **Документация** - полностью обновлена с новой архитектурой

### Fixed
- Исправлены проблемы с зависимостями в Docker builds
- Исправлены конфликты метрик Prometheus
- Исправлены проблемы с migrations в Docker образах
- Исправлены проблемы с proto descriptor для Envoy

### Removed
- Удалены временные файлы (protoc_output*.txt, temp_*.txt)
- Удалены старые fix документации
- Удалены бинарники (*.exe)

### Documentation
- Обновлен README.md с архитектурой Envoy
- Добавлены примеры API запросов через Envoy
- Добавлен раздел Troubleshooting
- Создан PRE_PUSH_CHECKLIST.md для проверки перед пушем

