# Command LauncherRemote Registry

[![CI](https://github.com/mazurov/command-launcher-registry/actions/workflows/ci.yml/badge.svg)](https://github.com/mazurov/command-launcher-registry/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mazurov/command-launcher-registry)](https://goreportcard.com/report/github.com/mazurov/command-launcher-registry)
[![codecov](https://codecov.io/gh/mazurov/command-launcher-registry/branch/main/graph/badge.svg)](https://codecov.io/gh/mazurov/command-launcher-registry)

Modern implementation of the Command Launcher remote registry with Gin web framework and GORM ORM.

## Features

- 🚀 **High-Performance API**: Built with Gin web framework
- 💾 **Flexible Storage**: GORM with PostgreSQL/SQLite support
- 🔐 **Secure Authentication**: JWT tokens + API key support
- 🎯 **Backward Compatible**: Maintains compatibility with existing Command Launcher clients (provides index.json)
- 🛠️ **Powerful CLI**: Cobra-based management tool
- 📦 **Clean Architecture**: Layered design with separation of concerns

## Architecture

```
.
├── cmd/
│   ├── server/          # Gin web server
│   └── cli/             # Registry management CLI
├── internal/
│   ├── api/             # HTTP handlers & routing
│   ├── service/         # Business logic
│   ├── repository/      # Data access layer (GORM)
│   ├── models/          # Database entities
│   ├── auth/            # Authentication & authorization
│   └── config/          # Configuration management
├── pkg/
│   └── types/           # Shared types
└── migrations/          # Database migrations
```

## Quick Start

### Start Server

```bash
# PostgreSQL
go run cmd/server/main.go --db-type postgres --db-dsn "host=localhost user=registry password=secret dbname=registry"

# SQLite (development)
go run cmd/server/main.go --db-type sqlite --db-dsn "registry.db"
```

### Use CLI

```bash
# Create registry
./bin/cola-registry-cli registry create --name my-registry --description "My Registry"

# Add package
./bin/cola-registry-cli package create --registry my-registry --name my-package

# Publish version
./bin/cola-registry-cli version publish --registry my-registry --package my-package --version 1.0.0 --url http://example.com/pkg.zip
```

## API Endpoints

### Backward Compatible (CDT v1)
- `GET /remote/registry/:name/index.json` - Get registry index in CDT format

### New API (v2)
- `POST /remote/registries` - Create registry
- `GET /remote/registries` - List registries
- `GET /remote/registries/:name` - Get registry details
- `PUT /remote/registries/:name` - Update registry
- `DELETE /remote/registries/:name` - Delete registry
- `POST /remote/registries/:name/packages` - Create package
- `POST /remote/registries/:name/packages/:package/versions` - Publish version

## Configuration

Environment variables:
- `REGISTRY_DB_TYPE` - Database type (postgres, sqlite)
- `REGISTRY_DB_DSN` - Database connection string
- `REGISTRY_PORT` - Server port (default: 8080)
- `REGISTRY_JWT_SECRET` - JWT signing secret
- `REGISTRY_LOG_LEVEL` - Log level (debug, info, warn, error)

## Docker Deployment

📖 See [deployments/docker/README.md](deployments/docker/README.md) for complete Docker deployment documentation.


### Makefile Commands

```bash
make docker-build              # Build Docker image
make docker-run               # Run with SQLite (single container)
make docker-compose-up        # Start with PostgreSQL
make docker-compose-down      # Stop services
make docker-compose-logs      # View logs
make docker-compose-build     # Rebuild and start
```

## Development
See [QUICKSTART.md](QUICKSTART.md) for detailed setup and usage instructions.

### 🔜 Future Enhancements (Optional)

- [ ] API authentication enforcement (currently optional)
- [ ] Rate limiting middleware
- [ ] Prometheus metrics
- [ ] OpenAPI/Swagger documentation generation
- [ ] Package upload endpoint (receive files, not just URLs)
- [ ] Package search/filtering
- [ ] User management & RBAC
- [ ] Webhook notifications
