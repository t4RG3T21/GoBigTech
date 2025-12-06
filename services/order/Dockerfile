# Multi-stage Dockerfile for Order Service
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy platform module with go.mod and go.sum (required for replace directive in go.mod)
COPY platform/go.mod platform/go.sum ./platform/
COPY platform ./platform

# Copy iam service (needed by inventory) - v1 directory and go.mod/go.sum
COPY services/iam/v1 ./services/iam/v1
COPY services/iam/go.mod services/iam/go.sum ./services/iam/

# Copy inventory service - v1 directory and go.mod/go.sum (depends on iam)
COPY services/inventory/v1 ./services/inventory/v1
COPY services/inventory/go.mod services/inventory/go.sum ./services/inventory/

# Copy payment service - v1 directory and go.mod/go.sum
COPY services/payment/v1 ./services/payment/v1
COPY services/payment/go.mod services/payment/go.sum ./services/payment/

# Copy service go.mod and go.sum
COPY services/order/go.mod services/order/go.sum ./services/order/

# Set working directory to service
WORKDIR /build/services/order

# Copy service source code
COPY services/order/ ./

# Download dependencies for platform module first
WORKDIR /build/platform
RUN go mod download

# Download dependencies for iam service (needed by inventory)
WORKDIR /build/services/iam
RUN go mod download

# Download dependencies for inventory service (needed for order replace directive, depends on iam)
WORKDIR /build/services/inventory
RUN go mod download

# Download dependencies for payment service (needed for order replace directive)
WORKDIR /build/services/payment
RUN go mod download

# Download dependencies for order service (platform, inventory, payment available via replace directives in go.mod)
WORKDIR /build/services/order
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/service \
    ./cmd/order

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/service /app/service

# Copy migrations directory (needed for database migrations)
COPY --from=builder /build/services/order/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/service"]

