# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cola-registry ./cmd/cola-registry

# Runtime stage
FROM alpine:3.19

# Add ca-certificates for HTTPS and OCI registry communication
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -u 1000 cola
USER cola

# Copy binary from builder
COPY --from=builder /cola-registry /usr/local/bin/cola-registry

# Default data directory
WORKDIR /data

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["cola-registry"]
CMD ["server"]
