# Multi-stage Dockerfile for Payment Service
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy platform module with go.mod and go.sum (required for replace directive in go.mod)
COPY platform/go.mod platform/go.sum ./platform/
COPY platform ./platform

# Copy service go.mod and go.sum
COPY services/payment/go.mod services/payment/go.sum ./services/payment/

# Set working directory to service
WORKDIR /build/services/payment

# Copy service source code
COPY services/payment/ ./

# Download dependencies for platform module first
WORKDIR /build/platform
RUN go mod download

# Download dependencies for payment service (platform available via replace directive in go.mod)
WORKDIR /build/services/payment
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/service \
    ./cmd/payment

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/service /app/service

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose gRPC port
EXPOSE 50052

# Health check (gRPC services don't have HTTP health check)
# Using a simple process check instead
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ps aux | grep "[s]ervice" > /dev/null || exit 1

# Run the application
CMD ["/app/service"]

