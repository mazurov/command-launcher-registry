# Quick Start Guide

## 🚀 Get Started in 5 Minutes

### 1. Install Dependencies
```bash
go mod download
```

### 2. Start the Server
```bash
# Using SQLite (easiest for development)
make run-server

# Or manually (requires building first):
make build-server
./bin/cola-registry-server --db-type sqlite --db-dsn registry.db
```

The server will start on `http://localhost:8080`

### 3. Create Your First Registry
```bash
# In another terminal
./bin/cola-registry-cli registry create \
  --name my-registry \
  --description "My first registry"
```

### 4. Add a Package
```bash
./bin/cola-registry-cli package create \
  --registry my-registry \
  --name my-tool \
  --description "My awesome tool"
```

### 5. Publish a Version
```bash
./bin/cola-registry-cli version publish \
  --registry my-registry \
  --package my-tool \
  --version 1.0.0 \
  --url http://example.com/my-tool-1.0.0.zip \
  --checksum sha256:abc123
```

### 6. Test the CDT-Compatible Endpoint
```bash
curl http://localhost:8080/remote/registries/my-registry/index.json
```

You should see:
```json
[
  {
    "name": "my-tool",
    "version": "1.0.0",
    "url": "http://example.com/my-tool-1.0.0.zip",
    "checksum": "sha256:abc123",
    "startPartition": 0,
    "endPartition": 9
  }
]
```

### 7. (Optional) Populate with Fake Data
```bash
# Populate the registry with sample packages for testing
./scripts/populate-registry.sh

# This creates:
# - 3 registries (development, production, experimental)
# - 10 packages (kubectl, terraform, docker, jq, envoy, prometheus, grafana, hotfix, env, proto-compiler)
# - 28 versions across all packages
```

## 🎯 Next Steps

### Run Tests
```bash
make test
```

### Build Binaries
```bash
make build
./bin/cola-registry-server --help
./bin/cola-registry-cli --help
```

### Use PostgreSQL
```bash
# Start PostgreSQL
docker run --name registry-postgres -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=registry -e POSTGRES_USER=registry -p 5432:5432 -d postgres:16-alpine

# Run server with PostgreSQL
make run-server-pg
```

### View API Documentation
- API Reference: [docs/API.md](docs/API.md)
- Examples: [docs/EXAMPLES.md](docs/EXAMPLES.md)

## 🛠️ Development

### Hot Reload (requires air)
```bash
make install-tools
make dev
```

### Format Code
```bash
make fmt
```

### Run Linters
```bash
make lint
```

## 📖 Full Documentation

- [README.md](README.md) - Project overview
- [docs/API.md](docs/API.md) - Complete API reference
- [docs/EXAMPLES.md](docs/EXAMPLES.md) - Usage examples and deployment guides

## 🐛 Troubleshooting

### Port Already in Use
```bash
# Change the port
./bin/cola-registry-server --port 8081
```

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker ps

# View server logs
./bin/cola-registry-server --log-level debug
```

### CLI Connection Issues
```bash
# Specify API URL
./bin/cola-registry-cli --api-url http://localhost:8080 registry list
```
