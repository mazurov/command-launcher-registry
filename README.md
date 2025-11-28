# Command Launcher Remote Registry

[![CI](https://github.com/mazurov/command-launcher-registry/actions/workflows/ci.yml/badge.svg)](https://github.com/mazurov/command-launcher-registry/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mazurov/command-launcher-registry)](https://goreportcard.com/report/github.com/mazurov/command-launcher-registry)

<!-- [![codecov](https://codecov.io/gh/mazurov/command-launcher-registry/branch/main/graph/badge.svg)](https://codecov.io/gh/mazurov/command-launcher-registry) -->

Modern implementation of the Command Launcher remote registry with Gin web framework and GORM ORM.

## Features

- 🚀 **API**: Built with Gin web framework
- 💾 **Flexible Storage**: GORM with PostgreSQL/SQLite support
- 🔐 **GitHub OAuth Authentication**: Device Flow for CLI, PAT Exchange for CI/CD
- 🔒 **Team-Based Authorization**: GitHub Teams for access control
- 🎯 **Backward Compatible**: Maintains compatibility with existing Command Launcher clients (provides index.json)
- 🛠️ **CLI**: Cobra-based management tool
- 📦 **Architecture**: Layered design with separation of concerns

## Quick Start

**📖 See [QUICKSTART.md](QUICKSTART.md) for detailed setup and usage instructions.**

### Start Server

```bash
# Build binaries
make build

# PostgreSQL
./bin/cola-registry-server --db-type postgres --db-dsn "host=localhost user=registry password=secret dbname=registry"

# SQLite (development)
./bin/cola-registry-server --db-type sqlite --db-dsn "registry.db"
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

## API Endpoints

### Backward Compatibility with Command Launcher remote registry index
- `GET /remote/registry/:name/index.json` - Get registry index in CDT format

### API
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

## Authentication & Authorization

The registry uses GitHub OAuth with team-based authorization:

- **Read Access**: Any authenticated user
- **Write Access**: Only users in configured GitHub Teams

### Authentication Methods

| Method | Use Case | How it Works |
|--------|----------|--------------|
| **Device Flow** | Interactive CLI | Enter code in browser, CLI polls for token automatically |
| **PAT Exchange** | CI/CD pipelines | Exchange GitHub PAT for JWT token |

### Server Setup

```bash
# Set environment variables
export REGISTRY_AUTH_STRATEGY=github
export REGISTRY_GITHUB_CLIENT_ID="your_client_id"
export REGISTRY_GITHUB_CLIENT_SECRET="your_client_secret"
export REGISTRY_GITHUB_WRITE_TEAMS="backend-team,infrastructure"
export REGISTRY_JWT_SECRET="your_jwt_secret"  # Optional in dev mode

# Start server
./bin/cola-registry-server --auth-strategy github
```

### CLI Authentication (Device Flow)

```bash
# Login - opens browser, enter the displayed code
./bin/cola-registry-cli auth login

# Check current user
./bin/cola-registry-cli auth whoami

# Logout
./bin/cola-registry-cli auth logout

# Use CLI normally (token is saved automatically)
./bin/cola-registry-cli registry list
```

**How Device Flow works:**
1. CLI requests a device code from server
2. CLI displays a code (e.g., `ABCD-1234`) and opens browser
3. You enter the code in browser and authorize with GitHub
4. CLI automatically receives the JWT token (no manual pasting needed!)
5. Token is saved to `~/.config/cola-registry/config.yaml`

### CI/CD Authentication (PAT Exchange)

For non-interactive environments, exchange a GitHub Personal Access Token for a JWT:

```bash
# 1. Create GitHub PAT at https://github.com/settings/tokens
#    Required scope: read:org (for team membership)

# 2. Exchange PAT for JWT token
curl -X POST http://localhost:8080/auth/github/token \
  -H "Content-Type: application/json" \
  -d '{"token": "ghp_your_github_pat"}'

# Response: {"access_token": "eyJhbGc...", "expires_in": 86400}

# 3. Use the JWT token
export COLA_REGISTRY_TOKEN="eyJhbGc..."
./bin/cola-registry-cli registry list
```

### Token Lifecycle

| Token Type | Lifetime | Refresh |
|------------|----------|---------|
| Device Code | 15 minutes | One-time use, request new code |
| JWT Token | 24 hours (configurable) | Re-authenticate with `auth login` or PAT exchange |

**📖 Documentation:**
- [Token Lifecycle](docs/TOKEN_LIFECYCLE.md) - Token expiration and refresh details
- [PAT Exchange](docs/AUTH_PAT_EXCHANGE.md) - CI/CD integration guide

## Development
See [QUICKSTART.md](QUICKSTART.md) for detailed setup and usage instructions.

### 🔜 Future Enhancements (Optional)

- [ ] Rate limiting middleware
- [ ] Prometheus metrics
- [ ] OpenAPI/Swagger documentation generation
- [ ] Package upload endpoint (receive files, not just URLs)
- [ ] Package search/filtering
- [ ] Webhook notifications
