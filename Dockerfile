# ===========================================
# GoPlan Backend Dockerfile
# Multi-stage build for Go application
# Optimized for layer caching and security
# ===========================================

# ----- Build Stage -----
FROM golang:1.24-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && update-ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum* ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s \
        -X main.Version=${VERSION} \
        -X main.BuildTime=${BUILD_TIME} \
        -X main.GitCommit=${GIT_COMMIT} \
        -extldflags '-static'" \
    -o /app/goplan \
    ./cmd/server

# ----- Development Stage -----
FROM golang:1.24-alpine AS development

# Install development tools
RUN apk add --no-cache git make curl

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

# Install migrate for database migrations
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code (will be overridden by volume mount in docker-compose)
COPY . .

# Expose port
EXPOSE 8080

# Default command for development (hot reload with air)
CMD ["air", "-c", ".air.toml"]

# ----- Production Stage -----
FROM alpine:3.19 AS production

# Install runtime dependencies and security updates
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# Create non-root user with specific UID/GID for security
RUN addgroup -g 1001 -S goplan && \
    adduser -u 1001 -S goplan -G goplan -h /app -s /sbin/nologin

WORKDIR /app

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder
COPY --from=builder /app/goplan /app/goplan

# Copy migrations for runtime migration option
COPY --from=builder /app/migrations /app/migrations

# Set ownership
RUN chown -R goplan:goplan /app && \
    chmod +x /app/goplan

# Create tmp directory for any temporary files
RUN mkdir -p /tmp && chown goplan:goplan /tmp

# Switch to non-root user
USER goplan

# Expose port
EXPOSE 8080

# Set environment defaults
ENV PORT=8080 \
    HOST=0.0.0.0 \
    ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

# Labels for container metadata
LABEL org.opencontainers.image.title="GoPlan Backend" \
      org.opencontainers.image.description="AI-Powered Project Management Platform - Backend API" \
      org.opencontainers.image.vendor="GoPlan" \
      org.opencontainers.image.source="https://github.com/goplan/goplan" \
      org.opencontainers.image.licenses="MIT"

# Run the application
ENTRYPOINT ["/app/goplan"]
