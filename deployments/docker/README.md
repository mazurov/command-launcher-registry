# Docker Deployment

This directory contains Docker configurations for deploying the CoLA Registry.

## Quick Start

### Using Docker Compose (Recommended)

1. **Copy environment file:**
```bash
cp .env.example .env
# Edit .env and set your passwords and configuration
```

2. **Start services:**
```bash
# From project root
make docker-compose-up

# Or from this directory
docker-compose up -d
```

3. **View logs:**
```bash
make docker-compose-logs
```

4. **Stop services:**
```bash
make docker-compose-down
```

### Using Docker Only

Build and run with SQLite (single container):
```bash
# Build image
make docker-build

# Run container
make docker-run
```

## Configuration

### Environment Variables

All configuration is done via environment variables. See [.env.example](.env.example) for available options.

**PostgreSQL:**
- `POSTGRES_PASSWORD` - Database password (default: registry_secret)
- `POSTGRES_PORT` - PostgreSQL port (default: 5432)

**Registry Server:**
- `REGISTRY_PORT` - Server port (default: 8080)
- `REGISTRY_MODE` - Server mode: `debug` or `release` (default: release)
- `REGISTRY_LOG_LEVEL` - Log level: `debug`, `info`, `warn`, `error` (default: info)
- `REGISTRY_LOG_FORMAT` - Log format: `text` or `json` (default: json)
- `REGISTRY_JWT_SECRET` - JWT secret for authentication (REQUIRED in production)

### Volumes

**PostgreSQL Data:**
- Volume: `postgres_data`
- Persists database data across container restarts

**SQLite Data (docker-run only):**
- Mounted from `./data` to `/data` in container
- Database file: `./data/registry.db`

## Architecture

The docker-compose setup includes:

1. **PostgreSQL Database** (`postgres`)
   - PostgreSQL 16 Alpine
   - Health checks enabled
   - Persistent data volume

2. **Registry Server** (`registry`)
   - Built from multi-stage Dockerfile
   - Non-root user for security
   - Health check endpoint: `/health`
   - Automatic restart policy

## Dockerfile Details

The Dockerfile uses a multi-stage build:

**Stage 1: Builder**
- Go 1.24 Alpine base with GCC and musl-dev for CGO
- Builds both `cola-registry-server` and `cola-registry-cli`
- CGO enabled to support SQLite (go-sqlite3 requires CGO)

**Stage 2: Runtime**
- Minimal Alpine base with SQLite runtime libraries
- Non-root user (`registry:registry`)
- Includes compiled binaries, CA certificates, and SQLite libs
- Health check on `/health` endpoint

## Health Checks

**PostgreSQL:**
```bash
docker exec cola-registry-postgres pg_isready -U registry
```

**Registry Server:**
```bash
curl http://localhost:8080/health
```

## Production Deployment

1. **Set secure passwords:**
```bash
# Generate strong passwords
openssl rand -base64 32  # For POSTGRES_PASSWORD
openssl rand -base64 64  # For REGISTRY_JWT_SECRET
```

2. **Update .env file:**
```env
POSTGRES_PASSWORD=<secure-password>
REGISTRY_JWT_SECRET=<secure-jwt-secret>
REGISTRY_MODE=release
REGISTRY_LOG_LEVEL=info
REGISTRY_LOG_FORMAT=json
```

3. **Start in production mode:**
```bash
docker-compose up -d
```

4. **Monitor logs:**
```bash
docker-compose logs -f registry
```

## Accessing the CLI in Container

To run CLI commands against the containerized server:

```bash
# Execute CLI from registry container
docker exec -it cola-registry-server /app/cola-registry-cli registry list

# Or from host if you built locally
./bin/cola-registry-cli --api-url http://localhost:8080 registry list
```

## Network

Services communicate on the `cola-network` bridge network:
- PostgreSQL: accessible at `postgres:5432` (internal)
- Registry: accessible at `localhost:8080` (external)

## Troubleshooting

**Container won't start:**
```bash
# Check logs
docker-compose logs registry

# Check PostgreSQL health
docker-compose ps postgres
```

**Database connection issues:**
```bash
# Verify PostgreSQL is healthy
docker exec cola-registry-postgres pg_isready -U registry

# Check connection from registry container
docker exec cola-registry-server env | grep REGISTRY_DB_DSN
```

**Reset everything:**
```bash
# Stop and remove containers, networks, and volumes
docker-compose down -v

# Rebuild and start fresh
make docker-compose-build
```

## Makefile Targets

From project root:

- `make docker-build` - Build Docker image
- `make docker-run` - Run single container with SQLite
- `make docker-compose-up` - Start docker-compose services
- `make docker-compose-down` - Stop docker-compose services
- `make docker-compose-logs` - View logs
- `make docker-compose-build` - Rebuild and start services
