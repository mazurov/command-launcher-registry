# Multi-stage build for cola-registry-server
FROM golang:1.24-alpine AS builder

# Install build dependencies including gcc for CGO (required for SQLite)
RUN apk add --no-cache git ca-certificates gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both server and CLI with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o cola-registry-server ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o cola-registry-cli ./cmd/cli

# Final stage
FROM alpine:latest

# Install runtime dependencies including SQLite libraries for CGO
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

# Create non-root user
RUN addgroup -g 1000 registry && \
    adduser -D -u 1000 -G registry registry

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/cola-registry-server .
COPY --from=builder /build/cola-registry-cli .

# Change ownership
RUN chown -R registry:registry /app

# Switch to non-root user
USER registry

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run server
ENTRYPOINT ["/app/cola-registry-server"]
CMD ["--port", "8080"]
