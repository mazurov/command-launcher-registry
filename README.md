# Command Launcher Registry

A complete solution for hosting and managing Command Launcher remote registries, featuring both a REST API server and a powerful CLI client (`cola-regctl`) for seamless registry management.

## Features

### Server
- **Multi-Registry Support**: Isolated namespaces for different teams
- **Package & Version Management**: Full CRUD operations via REST API
- **Gradual Rollout**: Partition-based version distribution (0-9 range)
- **Command Launcher Compatible**: Serves index.json in CDT-compatible format
- **Authentication**: Optional HTTP Basic Auth for write operations
- **Pluggable Storage**: File-based JSON storage or OCI registry storage (ghcr.io, Docker Hub, etc.)
- **Operational Ready**: Health checks, metrics, structured logging

### CLI Client (`cola-regctl`)
- **Full CRUD Operations**: Manage registries, packages, and versions from the command line
- **Secure Credential Storage**: OS-native keychains (macOS Keychain, Windows Credential Manager, Linux protected files)
- **Environment Variable Support**: `COLA_REGISTRY_URL` and `COLA_REGISTRY_SESSION_TOKEN`
- **Multiple Output Formats**: Human-readable tables and machine-parseable JSON
- **Interactive Workflows**: Password prompts and deletion confirmations
- **Cross-Platform**: Works on macOS, Linux, and Windows

## Quick Start

### Prerequisites

- Go 1.24 or later
- Git

### Build

```bash
# Clone and build
git clone <repository-url>
cd command-launcher-registry
make build       # Build server
make build-cli   # Build CLI client
```

### Run Server

```bash
# Start server with default configuration
./bin/cola-registry server

# With explicit storage URI
./bin/cola-registry server --storage-uri file://./data/registry.json

# With path only (auto-prefixed with file://)
./bin/cola-registry server --storage-uri ./data/registry.json

# Full CLI configuration
./bin/cola-registry server \
  --storage-uri file://./data/registry.json \
  --port 8080 \
  --host 0.0.0.0 \
  --log-level info \
  --log-format json \
  --auth-type none

# Or with environment variables
export COLA_REGISTRY_STORAGE_URI=file://./data/registry.json
export COLA_REGISTRY_SERVER_PORT=8080
./bin/cola-registry server
```

Server starts at `http://localhost:8080` by default.

### CLI Client Quick Start

```bash
# Set server URL and credentials via environment variables
export COLA_REGISTRY_URL=http://localhost:8080
export COLA_REGISTRY_SESSION_TOKEN=admin:admin

# Create a registry
./bin/cola-regctl registry create my-tools --description "My tools registry"

# List registries
./bin/cola-regctl registry list

# Create a package
./bin/cola-regctl package create my-tools my-cli \
  --description "My CLI tool" \
  --maintainer "you@example.com"

# Publish a version
./bin/cola-regctl version create my-tools my-cli 1.0.0 \
  --checksum "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
  --url "https://downloads.example.com/my-cli-1.0.0.zip" \
  --start-partition 0 \
  --end-partition 9

# List versions
./bin/cola-regctl version list my-tools my-cli

# Get JSON output (for scripting)
./bin/cola-regctl registry list --json
```

### Populate Test Data

Using curl (direct API calls):
```bash
# Populate sample registry with packages and versions
./scripts/populate-test-data.sh

# Verify data
curl http://localhost:8080/api/v1/registry/company-tools/index.json | jq '.'

# Clean up test data
./scripts/clean-test-data.sh
```

Using CLI client:
```bash
# Populate sample registry with packages and versions
./scripts/populate-test-data-cli.sh

# Verify data using CLI
export COLA_REGISTRY_URL=http://localhost:8080
export COLA_REGISTRY_SESSION_TOKEN=admin:admin
./bin/cola-regctl registry list
./bin/cola-regctl package list company-tools
./bin/cola-regctl version list company-tools deployment-cli

# Clean up test data
./scripts/clean-test-data-cli.sh
```

### Manual Testing

```bash
# Health check
curl http://localhost:8080/api/v1/health | jq '.'

# Metrics
curl http://localhost:8080/api/v1/metrics | jq '.'

# Create a registry (no auth required by default)
curl -X POST http://localhost:8080/api/v1/registry \
  -H "Content-Type: application/json" \
  -d '{"name":"build","description":"Build tools"}' | jq '.'

# Add a package
curl -X POST http://localhost:8080/api/v1/registry/build/package \
  -H "Content-Type: application/json" \
  -d '{"name":"hotfix","description":"Hotfix tool"}' | jq '.'

# Publish a version
curl -X POST http://localhost:8080/api/v1/registry/build/package/hotfix/version \
  -H "Content-Type: application/json" \
  -d '{
    "name":"hotfix",
    "version":"1.0.0",
    "checksum":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "url":"https://example.com/hotfix-1.0.0.zip",
    "startPartition":0,
    "endPartition":9
  }' | jq '.'

# Fetch registry index (Command Launcher format)
curl http://localhost:8080/api/v1/registry/build/index.json | jq '.'
```

## Configuration

Configuration follows [12-factor app](https://12factor.net/) principles:
1. CLI flags (highest priority)
2. Environment variables
3. Defaults (lowest priority)

No configuration file is required.

### CLI Flags

```bash
cola-registry server [flags]

Flags:
  --storage-uri string     Storage URI (file:// for local, oci:// for OCI registry)
                           Default: file://./data/registry.json
  --storage-token string   Storage authentication token (required for OCI storage)
                           Default: (empty)
  --port int               Server port
                           Default: 8080
  --host string            Bind address
                           Default: 0.0.0.0
  --log-level string       Log level (debug|info|warn|error)
                           Default: info
  --log-format string      Log format (json|text)
                           Default: json
  --auth-type string       Authentication type (none|basic)
                           Default: none
```

### Environment Variables

All CLI flags have corresponding environment variables with `COLA_REGISTRY_` prefix:

```bash
export COLA_REGISTRY_STORAGE_URI=file://./data/registry.json
export COLA_REGISTRY_STORAGE_TOKEN=my-token        # Required for OCI storage
export COLA_REGISTRY_SERVER_PORT=8080
export COLA_REGISTRY_SERVER_HOST=0.0.0.0
export COLA_REGISTRY_LOGGING_LEVEL=info
export COLA_REGISTRY_LOGGING_FORMAT=json
export COLA_REGISTRY_AUTH_TYPE=basic
export COLA_REGISTRY_AUTH_USERS_FILE=./users.yaml  # Environment-only (no CLI flag)
```

Priority order: **CLI flags > Environment variables > Defaults**

### Storage URI

The storage backend is configured via URI:

```bash
# File storage
--storage-uri file://./data/registry.json       # Relative path
--storage-uri file:///var/data/registry.json    # Absolute path (Unix)
--storage-uri ./data/registry.json              # Auto-prefixed with file://

# OCI storage (GitHub Container Registry)
--storage-uri oci://ghcr.io/myorg/cola-registry-data
--storage-token ghp_xxxxxxxxxxxxxxxxxxxx

# OCI storage (Docker Hub)
--storage-uri oci://docker.io/myuser/cola-registry-data
--storage-token dckr_pat_xxxxxxxxx

# OCI storage (Azure Container Registry)
--storage-uri oci://myregistry.azurecr.io/cola-registry-data
--storage-token <acr-token>
```

**OCI Storage Notes**:
- OCI storage requires `--storage-token` or `COLA_REGISTRY_STORAGE_TOKEN` environment variable
- The registry data is stored as an OCI artifact with `latest` tag (overwritten on each write)
- Supports any OCI Distribution-compliant registry
- Token format is registry-specific (e.g., GitHub PAT for ghcr.io)

### Docker Usage

```bash
# Build the image
docker build -t cola-registry -f docker/Dockerfile .

# Run with file storage
docker run -p 8080:8080 \
  -e COLA_REGISTRY_STORAGE_URI=file:///data/registry.json \
  -v /host/data:/data \
  cola-registry

# Run with OCI storage
docker run -p 8080:8080 \
  -e COLA_REGISTRY_STORAGE_URI=oci://ghcr.io/myorg/cola-registry-data \
  -e COLA_REGISTRY_STORAGE_TOKEN=ghp_xxxxxxxxxxxx \
  cola-registry
```

### Authentication Setup

To enable basic authentication:

```bash
# Generate password hash
./bin/cola-registry auth hash-password
# Enter password when prompted
# Copy the bcrypt hash

# Create users.yaml
cat > users.yaml <<EOF
users:
  - username: admin
    password_hash: "$2a$10$..." # paste bcrypt hash here
    roles: ["admin"]
EOF

# Update config.yaml or use environment variable
export COLA_REGISTRY_AUTH_TYPE=basic
export COLA_REGISTRY_AUTH_USERS_FILE=./users.yaml

# Start server with authentication enabled
./bin/cola-registry server
```

Now write operations require authentication:
```bash
# Create registry with auth
curl -u admin:yourpassword -X POST http://localhost:8080/api/v1/registry \
  -H "Content-Type: application/json" \
  -d '{"name":"secure","description":"Secure registry"}'
```

## CLI Client (`cola-regctl`)

The `cola-regctl` CLI provides a user-friendly interface for managing registries, packages, and versions.

### Installation

```bash
# Build the CLI
make build-cli

# Optionally install to system PATH
sudo cp bin/cola-regctl /usr/local/bin/
```

### Authentication

The CLI supports multiple authentication methods with the following precedence:

1. `--token` flag (highest priority)
2. `COLA_REGISTRY_SESSION_TOKEN` environment variable
3. Stored credentials from `login` command (lowest priority)

#### Interactive Login

```bash
# Login interactively (stores credentials securely)
./bin/cola-regctl login http://localhost:8080
# Enter username: admin
# Enter password: ****

# Check authentication status
./bin/cola-regctl whoami

# Logout (removes stored credentials)
./bin/cola-regctl logout
```

#### Environment Variables

```bash
# Set once, use everywhere
export COLA_REGISTRY_URL=http://localhost:8080
export COLA_REGISTRY_SESSION_TOKEN=admin:admin

# Now all commands use these credentials automatically
./bin/cola-regctl registry list
./bin/cola-regctl package create my-registry my-package
```

#### Per-Command Authentication

```bash
# Override with flags
./bin/cola-regctl --url http://localhost:8080 --token admin:admin registry list
```

### Command Reference

#### Registry Management

```bash
# Create a registry
cola-regctl registry create <name> \
  --description "Description" \
  --admin "admin@example.com" \
  --custom-value "key=value"

# List all registries
cola-regctl registry list
cola-regctl registry list --json  # JSON output

# Get registry details
cola-regctl registry get <name>

# Update registry
cola-regctl registry update <name> \
  --description "New description" \
  --admin "new-admin@example.com" \
  --clear-admins  # Clear all admins

# Delete registry (with confirmation)
cola-regctl registry delete <name>
cola-regctl registry delete <name> --yes  # Skip confirmation
```

#### Package Management

```bash
# Create a package
cola-regctl package create <registry> <package> \
  --description "Package description" \
  --maintainer "maintainer@example.com" \
  --custom-value "language=go" \
  --custom-value "repo=https://github.com/..."

# List packages
cola-regctl package list <registry>
cola-regctl package list <registry> --json

# Get package details
cola-regctl package get <registry> <package>

# Update package
cola-regctl package update <registry> <package> \
  --description "New description" \
  --maintainer "new@example.com" \
  --clear-maintainers  # Clear all maintainers

# Delete package
cola-regctl package delete <registry> <package>
```

#### Version Management

```bash
# Publish a version
cola-regctl version create <registry> <package> <version> \
  --checksum "sha256:abc123..." \
  --url "https://downloads.example.com/package-1.0.0.zip" \
  --start-partition 0 \
  --end-partition 9

# List versions
cola-regctl version list <registry> <package>
cola-regctl version list <registry> <package> --json

# Get version details
cola-regctl version get <registry> <package> <version>

# Delete version
cola-regctl version delete <registry> <package> <version>
```

### Global Flags

All commands support these global flags:

- `--url <url>` - Server URL (or use `COLA_REGISTRY_URL` env var)
- `--token <user:pass>` - Authentication token (or use `COLA_REGISTRY_SESSION_TOKEN` env var)
- `--json` - Output in JSON format (for scripting)
- `--verbose` - Enable verbose logging
- `--timeout <duration>` - HTTP request timeout (default: 30s)
- `--yes` / `-y` - Skip confirmation prompts

### Examples

#### Complete Workflow

```bash
# Set up environment
export COLA_REGISTRY_URL=http://localhost:8080
export COLA_REGISTRY_SESSION_TOKEN=admin:admin

# Create a complete registry
cola-regctl registry create build-tools \
  --description "Build and deployment tools" \
  --admin "devops@example.com"

# Add a package
cola-regctl package create build-tools deployer \
  --description "Deployment automation tool" \
  --maintainer "devops@example.com" \
  --custom-value "language=go"

# Publish version 1.0.0 (for partitions 0-4 = 50% rollout)
cola-regctl version create build-tools deployer 1.0.0 \
  --checksum "sha256:abc123..." \
  --url "https://cdn.example.com/deployer-1.0.0.tar.gz" \
  --start-partition 0 \
  --end-partition 4

# Publish version 1.1.0 (for partitions 5-9 = 50% rollout)
cola-regctl version create build-tools deployer 1.1.0 \
  --checksum "sha256:def456..." \
  --url "https://cdn.example.com/deployer-1.1.0.tar.gz" \
  --start-partition 5 \
  --end-partition 9

# List all versions
cola-regctl version list build-tools deployer
```

#### Scripting with JSON Output

```bash
#!/bin/bash
# Get all registries and process with jq
registries=$(cola-regctl registry list --json | jq -r '.data[].name')

for registry in $registries; do
  echo "Packages in $registry:"
  cola-regctl package list "$registry" --json | jq -r '.data[].name'
done
```

### Credential Storage

The CLI stores credentials securely using OS-native mechanisms:

- **macOS**: Keychain (token) + file (URL)
- **Windows**: Credential Manager (token) + file (URL)
- **Linux**: Protected file (0600 permissions)

Credentials are stored at: `~/.config/cola-registry/credentials.yaml`

## Documentation

- **[Complete Specification](./docs/spec.md)** - Full specification covering both server and CLI client
- **[OpenAPI Contract](./docs/openapi.yaml)** - REST API specification

### Endpoints

#### Operational
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Server metrics

#### Registries
- `GET /api/v1/registry` - List all registries (auth required)
- `POST /api/v1/registry` - Create registry (auth required)
- `GET /api/v1/registry/:name` - Get registry details
- `PUT /api/v1/registry/:name` - Update registry (auth required)
- `DELETE /api/v1/registry/:name` - Delete registry (auth required, cascade)
- `GET /api/v1/registry/:name/index.json` - Get registry index (CDT format)

#### Packages
- `GET /api/v1/registry/:name/package` - List packages
- `POST /api/v1/registry/:name/package` - Create package (auth required)
- `GET /api/v1/registry/:name/package/:package` - Get package details
- `PUT /api/v1/registry/:name/package/:package` - Update package (auth required)
- `DELETE /api/v1/registry/:name/package/:package` - Delete package (auth required, cascade)

#### Versions
- `GET /api/v1/registry/:name/package/:package/version` - List versions
- `POST /api/v1/registry/:name/package/:package/version` - Create version (auth required)
- `GET /api/v1/registry/:name/package/:package/version/:version` - Get version details
- `DELETE /api/v1/registry/:name/package/:package/version/:version` - Delete version (auth required)

## Development

### Build Commands

```bash
make build        # Build server binary
make build-cli    # Build CLI client binary
make clean        # Remove artifacts
make test         # Run tests
make test-cli     # Run CLI tests
make run          # Build and run server
make fmt          # Format code
make lint         # Run linter
make help         # Show all targets
```

### Project Structure

```
cmd/
├── cola-registry/          # Server binary entry point
└── cola-regctl/            # CLI client binary entry point
internal/
├── server/                 # HTTP server
│   ├── handlers/           # HTTP handlers
│   └── middleware/         # Middleware (auth, logging, etc.)
├── client/                 # CLI client (NEW)
│   ├── commands/           # Cobra commands (registry, package, version, login, etc.)
│   ├── auth/               # Credential storage and authentication
│   ├── config/             # URL and configuration resolution
│   ├── output/             # Output formatters (table, JSON)
│   ├── prompts/            # Interactive prompts
│   ├── validation/         # Client-side validation
│   └── errors/             # Error handling and exit codes
├── storage/                # Storage layer (file, OCI)
├── models/                 # Shared data models
├── auth/                   # Server authentication
├── cli/                    # Server CLI commands
├── config/                 # Server configuration
└── apierrors/              # API error types
scripts/
├── populate-test-data.sh       # curl-based test data
├── clean-test-data.sh          # curl-based cleanup
├── populate-test-data-cli.sh   # CLI-based test data
└── clean-test-data-cli.sh      # CLI-based cleanup
docker/
└── Dockerfile                  # Multi-stage Docker build
docs/
└── spec.md                     # Complete specification
```

## License

[Add license information]

## Contributing

[Add contribution guidelines]
