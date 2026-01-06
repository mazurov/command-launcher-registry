---
marp: true
theme: default
paginate: true
backgroundColor: #0d1117
color: #e6edf3
style: |
  section {
    font-family: 'Segoe UI', 'Helvetica Neue', Arial, sans-serif;
    font-size: 20px;
  }
  h1 {
    color: #58a6ff;
    font-size: 1.5em;
    border-bottom: 2px solid #58a6ff;
    padding-bottom: 8px;
  }
  h2 {
    color: #58a6ff;
    font-size: 1.2em;
  }
  h3 {
    color: #f78166;
    font-size: 1.0em;
  }
  code {
    background-color: #161b22;
    color: #79c0ff;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.8em;
  }
  pre {
    background-color: #fdf6e3 !important;
    border: 1px solid #eee8d5;
    border-radius: 6px;
    padding: 10px !important;
    font-size: 0.6em;
    line-height: 1.35;
  }
  pre code {
    color: #657b83;
  }
  table {
    font-size: 0.65em;
    width: 100%;
  }
  th {
    background-color: #21262d;
    color: #e6edf3;
    border: 1px solid #30363d;
  }
  td {
    background-color: #161b22;
    border: 1px solid #30363d;
  }
  strong {
    color: #58a6ff;
  }
  em {
    color: #f78166;
  }
  ul, ol {
    font-size: 0.85em;
  }
  p {
    margin: 0.4em 0;
  }
---

# COLA Registry Server

## Command Launcher Package Registry

**Multi-Backend Package Distribution Platform**

Registry server for managing and distributing CLI tools with pluggable storage backends

---

# Architecture Overview

```
┌───────────────────────────────────────────────────────────────┐
│                    COLA Registry Server                       │
├───────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  REST API   │  │ Middleware  │  │  Handlers   │            │
│  │   (Chi)     │  │   Stack     │  │             │            │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘            │
│         └────────────────┼────────────────┘                   │
│  ┌───────────────────────┴───────────────────────┐            │
│  │           Storage Interface (Store)           │            │
│  └───────────────────────┬───────────────────────┘            │
│  ┌───────────────┬───────┴───────┬───────────────┐            │
│  │     File      │      OCI      │      S3       │            │
│  │    Storage    │    Storage    │    Storage    │            │
│  └───────────────┴───────────────┴───────────────┘            │
└───────────────────────────────────────────────────────────────┘
```

---

# Data Model

## Hierarchical Structure

```
Registry (Organization/Team Level)
  ├── name: "devops-tools"
  ├── description: "DevOps CLI tools"
  ├── admins: ["admin@company.com"]
  ├── custom_values: {"env": "production"}
  │
  └── Package (Tool Level)
        ├── name: "deployer"
        ├── description: "Deployment automation CLI"
        ├── maintainers: ["dev@company.com"]
        │
        └── Version (Release Level)
              ├── version: "1.2.0"
              ├── checksum: "sha256:abc..."
              ├── url: "https://cdn/deployer-1.2.0.tar.gz"
              ├── startPartition: 0
              └── endPartition: 9
```

---

# Server CLI Commands

## Main Binary: `cola-registry`

### Start Server
```bash
cola-registry server \
  --storage-uri "s3://s3.amazonaws.com/my-bucket/registry.json" \
  --storage-token "ACCESS_KEY:SECRET_KEY" \
  --port 8080 \
  --host 0.0.0.0 \
  --log-level info \
  --log-format json \
  --auth-type basic
```

### Generate Password Hash
```bash
cola-registry auth hash-password
# Interactive prompt for secure password entry
# Outputs: $2a$10$... (bcrypt hash)
```

---

# Configuration Options

## CLI Flags, Environment Variables & Defaults

| Flag | Environment Variable | Default |
|------|---------------------|---------|
| `--storage-uri` | `COLA_REGISTRY_STORAGE_URI` | *required* |
| `--storage-token` | `COLA_REGISTRY_STORAGE_TOKEN` | - |
| `--port` | `COLA_REGISTRY_SERVER_PORT` | `8080` |
| `--host` | `COLA_REGISTRY_SERVER_HOST` | `0.0.0.0` |
| `--log-level` | `COLA_REGISTRY_LOGGING_LEVEL` | `info` |
| `--log-format` | `COLA_REGISTRY_LOGGING_FORMAT` | `json` |
| `--auth-type` | `COLA_REGISTRY_AUTH_TYPE` | `none` |
| - | `COLA_REGISTRY_AUTH_USERS_FILE` | `users.yaml` |

**Priority:** CLI flags > Environment variables > Defaults

---

# Storage Backend #1: File Storage

## Local/Network File System

```bash
# Relative path
cola-registry server --storage-uri "file://./data/registry.json"

# Absolute path
cola-registry server --storage-uri "file:///var/lib/cola/registry.json"
```

### Features

| Feature | Description |
|---------|-------------|
| **Atomic Writes** | Temp file + rename pattern |
| **Auto Directory** | Creates parent directories automatically |
| **Size Monitoring** | Warns at 50MB, max 100MB recommended |
| **Cross-Platform** | Windows, Linux, macOS |

### Configuration

| Parameter | CLI Flag | Environment Variable |
|-----------|----------|---------------------|
| Storage URI | `--storage-uri "file://..."` | `COLA_REGISTRY_STORAGE_URI` |
| Token | Not required | Not required |

---


# Storage Backend #2: OCI Registry

## Container Registry as Storage

```bash
# GitHub Container Registry
cola-registry server --storage-uri "oci://ghcr.io/myorg/cola-registry" --storage-token "ghp_xxxx"

# Docker Hub
cola-registry server --storage-uri "oci://docker.io/myuser/cola-registry" --storage-token "dckr_pat_xxxx"

# Azure Container Registry
cola-registry server --storage-uri "oci://myregistry.azurecr.io/cola-registry" --storage-token "user:pass"
```

### Configuration

| Parameter | CLI Flag | Environment Variable |
|-----------|----------|---------------------|
| Storage URI | `--storage-uri "oci://..."` | `COLA_REGISTRY_STORAGE_URI` |
| Token | `--storage-token "PAT"` | `COLA_REGISTRY_STORAGE_TOKEN` |

Token format: PAT (ghcr.io), `username:password` (ACR, Docker Hub)

---


# Storage Backend #3: S3-Compatible

```bash
# AWS S3
cola-registry server --storage-uri "s3://s3.us-east-1.amazonaws.com/bucket/registry.json" \
  --storage-token "$AWS_ACCESS_KEY_ID:$AWS_SECRET_ACCESS_KEY"

# MinIO (HTTP)
cola-registry server --storage-uri "s3+http://minio.local:9000/bucket/registry.json" \
  --storage-token "minioadmin:minioadmin"

# DigitalOcean Spaces
cola-registry server --storage-uri "s3://nyc3.digitaloceanspaces.com/bucket/registry.json"
```

### Configuration

| Parameter | CLI Flag | Environment Variable |
|-----------|----------|---------------------|
| Storage URI | `--storage-uri "s3://..."` | `COLA_REGISTRY_STORAGE_URI` |
| Credentials | `--storage-token "KEY:SECRET"` | `COLA_REGISTRY_STORAGE_TOKEN` |
| AWS Credentials | - | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| Region | URI query `?region=us-east-1` | `AWS_REGION` |
| IAM Role | Leave token empty | Auto-detected |

Supported: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Wasabi

---


# Storage Interface (Go)

```go
type Store interface {
    // Registry CRUD
    CreateRegistry(ctx context.Context, registry *Registry) error
    GetRegistry(ctx context.Context, name string) (*Registry, error)
    UpdateRegistry(ctx context.Context, registry *Registry) error
    DeleteRegistry(ctx context.Context, name string) error
    ListRegistries(ctx context.Context) ([]*Registry, error)

    // Package CRUD
    CreatePackage(ctx context.Context, registryName string, pkg *Package) error
    GetPackage(ctx context.Context, registryName, packageName string) (*Package, error)
    UpdatePackage(ctx context.Context, registryName string, pkg *Package) error
    DeletePackage(ctx context.Context, registryName, packageName string) error
    ListPackages(ctx context.Context, registryName string) ([]*Package, error)

    // Version CRUD + Index
    CreateVersion(ctx, registryName, packageName string, version *Version) error
    GetVersion(ctx, registryName, packageName, version string) (*Version, error)
    DeleteVersion(ctx, registryName, packageName, version string) error
    ListVersions(ctx, registryName, packageName string) ([]*Version, error)
    GetRegistryIndex(ctx, registryName string) ([]IndexEntry, error)

    Close() error
}
```

---


# BaseStorage Pattern

```go
type BaseStorage struct {
    mu         sync.RWMutex
    data       *StorageData
    persistFn  func([]byte) error  // Backend-specific persistence
}

// Atomic operation pattern
func (s *BaseStorage) CreateRegistry(ctx context.Context, r *Registry) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 1. Validate
    if _, exists := s.data.Registries[r.Name]; exists {
        return ErrAlreadyExists
    }

    // 2. Modify in-memory
    s.data.Registries[r.Name] = r

    // 3. Persist with rollback on failure
    if err := s.persist(); err != nil {
        delete(s.data.Registries, r.Name)  // Rollback
        return err
    }
    return nil
}
```

All backends share this pattern - only `persistFn` differs

---


# REST API Overview

```
/api/v1/
├── health                              GET      Health Check
├── metrics                             GET      Metrics
├── whoami                              GET      Auth Test
├── registry                            GET      List All
├── registry                            POST 🔒  Create
├── registry/:name                      GET      Get Registry
├── registry/:name                      PUT 🔒   Update
├── registry/:name                      DELETE 🔒
├── registry/:name/index.json           GET      Command Launcher Index
├── registry/:name/package              GET POST 🔒
├── registry/:name/package/:pkg         GET PUT 🔒 DELETE 🔒
├── registry/:name/package/:pkg/version GET POST 🔒
└── registry/:name/package/:pkg/version/:v GET DELETE 🔒
```

🔒 = Requires Authentication (write operations only)

---


# API: Registry Operations

## Create Registry
```bash
curl -X POST http://localhost:8080/api/v1/registry -u admin:password \
  -H "Content-Type: application/json" \
  -d '{"name": "devops-tools", "description": "DevOps CLI tools",
       "admins": ["admin@company.com"], "custom_values": {"env": "prod"}}'
```

## Get Registry Index (Command Launcher Format)
```bash
curl http://localhost:8080/api/v1/registry/devops-tools/index.json
```
```json
[{"name": "deployer", "version": "1.0.0", "checksum": "sha256:...",
  "url": "https://cdn/deployer-1.0.0.tar.gz", "startPartition": 0, "endPartition": 9}]
```

---


# API: Package & Version Operations

## Create Package
```bash
curl -X POST http://localhost:8080/api/v1/registry/devops-tools/package -u admin:password \
  -H "Content-Type: application/json" \
  -d '{"name": "deployer", "description": "K8s deployment automation",
       "maintainers": ["dev@company.com"]}'
```

## Publish Version
```bash
curl -X POST http://localhost:8080/api/v1/registry/devops-tools/package/deployer/version \
  -u admin:password -H "Content-Type: application/json" \
  -d '{"name": "deployer", "version": "1.2.0",
       "checksum": "sha256:e3b0c44298fc1c149afbf4c8996fb924...",
       "url": "https://cdn.company.com/deployer-1.2.0-linux-amd64.tar.gz",
       "startPartition": 0, "endPartition": 9}'
```

---


# Authentication System

### User Configuration (`users.yaml`)
```yaml
users:
  - username: admin
    password: "$2a$10$N9qo8uLOickgx2ZMRZoMy..."  # bcrypt hash
    roles: ["admin"]  # Reserved for future RBAC
  - username: deployer
    password: "$2a$10$Qz3RG3yLOickgx2ZMRZoMy..."
    roles: ["publisher"]
```

### Generate Password Hash
```bash
$ cola-registry auth hash-password
Enter password: ********
Confirm password: ********
Password hash (add to users.yaml):
$2a$10$N9qo8uLOickgx2ZMRZoMyeJ4xVXQFqLmLiUXLz3A2DfK/FqPZwFOq
```

### Security Features
- bcrypt cost factor: 10 | Token masking in logs | Failed attempt logging with IP

---


# Authentication Interface

```go
type Authenticator interface {
    Authenticate(ctx context.Context, username, password string) (string, error)
}

// NoAuth - default, allows all
type NoAuthenticator struct{}
func (a *NoAuthenticator) Authenticate(ctx context.Context, _, _ string) (string, error) {
    return "anonymous", nil
}

// BasicAuth - bcrypt-based authentication
type BasicAuthenticator struct {
    users map[string]User
}
func (a *BasicAuthenticator) Authenticate(ctx, user, pass string) (string, error) {
    u, exists := a.users[user]
    if !exists {
        return "", ErrInvalidCredentials
    }
    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass)); err != nil {
        return "", ErrInvalidCredentials
    }
    return user, nil
}
```

---

# Middleware Stack

## Request Processing Pipeline

```
Request → Logging → Rate Limiting → CORS → Auth → Handler → Response
           │            │            │       │
           │            │            │       └── Per-route (write ops)
           │            │            └────────── index.json only
           │            └─────────────────────── 100 req/min per IP
           └──────────────────────────────────── UUID, duration, IP
```

### Logging Middleware
- UUID request ID generation
- Request/response logging
- Duration tracking (ms)
- Remote IP extraction

### Rate Limiting
- Token bucket per IP
- 100 requests/minute default
- Auto cleanup of stale clients
- `429 Too Many Requests` + `Retry-After`

---


# Health & Metrics

### Health Check `GET /api/v1/health`
```json
{"status": "healthy", "checks": {"storage": {"status": "healthy"}}}
```

### Metrics `GET /api/v1/metrics`
```json
{
  "total_requests": 12345,
  "by_type": {"index_requests": 5000, "registry_creates": 10, "registry_reads": 200,
              "package_creates": 50, "version_creates": 150},
  "by_status": {"auth_failures": 5, "rate_limit_exceeded": 2, "validation_errors": 15}
}
```

### Response Codes
- `200 OK` - Server and storage healthy
- `503 Service Unavailable` - Storage unhealthy

---

# Input Validation

## Comprehensive Request Validation

### Registry/Package Names
```go
Pattern: ^[a-z0-9][a-z0-9_-]*$
Length:  1-64 characters
```

### Version Format
```go
Pattern: Semantic Versioning
Valid:   1.0.0, 2.1.3-alpha, 3.0.0-beta.1+build.123
```

### Checksum Format
```go
Pattern: sha256:[64 hex characters]
Example: sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

### Custom Values
```go
Max pairs:    20
Key pattern:  ^[a-zA-Z_][a-zA-Z0-9_-]{0,63}$
Value length: max 1024 characters
```

---


# Error Handling

```go
var (
    ErrNotFound             = errors.New("resource not found")      // → 404
    ErrAlreadyExists        = errors.New("resource already exists") // → 409
    ErrStorageUnavailable   = errors.New("storage unavailable")     // → 500
    ErrImmutabilityViolation = errors.New("cannot modify version")  // → 409
    ErrPartitionOverlap     = errors.New("partition overlap")       // → 409
)
```

### API Error Response Format
```json
{"error": {"code": "RESOURCE_NOT_FOUND", "message": "Registry 'my-registry' not found",
           "details": {"resource_type": "registry", "resource_name": "my-registry"}}}
```

---


# Graceful Shutdown

```go
func (s *Server) Start(ctx context.Context) error {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        slog.Info("shutdown signal received")
        shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()

        s.httpServer.Shutdown(shutdownCtx)  // Stop accepting new requests
        s.store.Close()                      // Close storage connections
        slog.Info("server shutdown complete")
    }()

    return s.httpServer.ListenAndServe()
}
```

- Handles SIGINT, SIGTERM
- 30s timeout for in-flight requests
- Proper storage connection cleanup

---


# Structured Logging

```go
// JSON format (default - for log aggregation)
{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"request completed",
 "request_id":"550e8400-e29b-41d4-a716-446655440000",
 "method":"POST","path":"/api/v1/registry","status":201,"duration_ms":45}

// Text format (for development)
2024-01-15T10:30:00Z INFO request completed request_id=550e8400... status=201 duration_ms=45
```

| Level | Use Case |
|-------|----------|
| `debug` | Detailed debugging information |
| `info` | Normal operations, startup, requests |
| `warn` | Warning conditions (file size, deprecation) |
| `error` | Error conditions, failures |

---

# HTTP Server Configuration

## Production Timeouts

```go
server := &http.Server{
    Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
    Handler:      router,
    ReadTimeout:  30 * time.Second,   // Request read timeout
    WriteTimeout: 120 * time.Second,  // Response write (long for OCI push)
    IdleTimeout:  120 * time.Second,  // Keep-alive timeout
}
```

### Why These Values?
- **ReadTimeout 30s**: Sufficient for large JSON payloads
- **WriteTimeout 120s**: OCI push operations can be slow
- **IdleTimeout 120s**: Efficient connection reuse

---

# Technology Stack

## Core Dependencies

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.24+ | High performance, static binary |
| **HTTP Router** | chi/v5 | Lightweight, composable |
| **Config** | viper | Multi-source configuration |
| **CLI** | cobra | Professional CLI framework |
| **Logging** | slog | Structured logging (stdlib) |
| **UUID** | google/uuid | Request ID generation |

## Storage SDKs

| Backend | Library | Purpose |
|---------|---------|---------|
| **OCI** | oras-go/v2 | OCI registry client |
| **S3** | minio-go/v7 | S3-compatible storage |
| **Auth** | x/crypto/bcrypt | Password hashing |

---


# Deployment Options

```bash
# Build
go build -o cola-registry ./cmd/cola-registry

# Docker
docker run -p 8080:8080 -e COLA_REGISTRY_STORAGE_URI="file:///data/registry.json" \
  -v /data:/data cola-registry:latest
```

```yaml
# Kubernetes
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: cola-registry
        image: cola-registry:latest
        env:
        - name: COLA_REGISTRY_STORAGE_URI
          value: "s3://bucket.s3.amazonaws.com/registry.json"
        - name: COLA_REGISTRY_AUTH_TYPE
          value: "basic"
        livenessProbe:
          httpGet: {path: /api/v1/health}
```

---

# Use Cases

### CLI Tool Distribution
- Centralized tool management across teams
- Version control and rollback capability
- Audit trail via structured logging

### Multi-Region Deployments
- S3 replication for global distribution
- OCI registry mirroring
- Edge caching with CDN

### Air-Gapped Environments
- File storage for offline deployments
- Single binary, no external dependencies
- Self-contained authentication

---

# Roadmap

## Future Enhancements

### Security
- 🔜 OIDC/SSO Authentication
- 🔜 Role-Based Access Control (RBAC)
- 🔜 API Key authentication
- 🔜 Audit logging

### Storage
- 🔜 Azure Blob Storage backend
- 🔜 Google Cloud Storage backend
- 🔜 Caching layer (Redis)

### Features
- 🔜 Webhooks for CI/CD integration
- 🔜 Package signing & verification
- 🔜 Automatic version cleanup
- 🔜 Multi-registry federation

---

# Summary

## Key Features

- **Production Ready** - Health checks, metrics, rate limiting, graceful shutdown
- **Pluggable Storage** - File, S3, OCI backends
- **Security** - Authentication, input validation, structured logging
- **Single Binary** - Minimal footprint, no external dependencies
- **RESTful API** - Clear error messages, comprehensive CLI

---

# Q&A

## Questions?

### Resources

- **Documentation**: `docs/quickstart.md`
- **API Reference**: See REST API slides
- **Source Code**: Browse the codebase

### Contact

Ready to discuss implementation details, integration requirements, or customization needs.

---

# Thank You!

## Let's Build Better CLI Distribution Together =)

**COLA Registry Server**
*Multi-Backend Package Distribution Platform*
