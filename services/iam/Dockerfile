# Multi-stage Dockerfile for IAM Service
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy platform module with go.mod and go.sum (required for replace directive in go.mod)
COPY platform/go.mod platform/go.sum ./platform/
COPY platform ./platform

# Copy service go.mod and go.sum
COPY services/iam/go.mod services/iam/go.sum ./services/iam/

# Set working directory to service
WORKDIR /build/services/iam

# Copy service source code
COPY services/iam/ ./

# Download dependencies for platform module first
WORKDIR /build/platform
RUN go mod download

# Download dependencies for iam service (platform available via replace directive in go.mod)
WORKDIR /build/services/iam
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/service \
    ./cmd/iam

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
COPY --from=builder /build/services/iam/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports (gRPC and HTTP Gateway)
EXPOSE 50053 8082

# Health check (process check since HTTP health endpoint may not be available)
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD pgrep -f /app/service > /dev/null || exit 1

# Run the application
CMD ["/app/service"]

