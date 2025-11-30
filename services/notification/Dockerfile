# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем platform модуль (необходим для replace директивы в go.mod)
COPY platform ./platform

# Копируем go.mod и go.sum notification service
COPY services/notification/go.mod services/notification/go.sum ./services/notification/

# Переходим в директорию notification service
WORKDIR /app/services/notification

# Загружаем зависимости (platform доступен через replace в go.mod)
# Не используем go.work, так как replace директива уже указывает на platform
RUN go mod download

# Копируем исходный код notification service
COPY services/notification/ ./

# Синхронизируем зависимости (загрузит недостающие)
RUN go mod tidy

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/notification ./cmd/notification

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Копируем бинарник из builder stage
COPY --from=builder /app/notification .

# Expose port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8083/health || exit 1

# Run the application
CMD ["./notification"]

