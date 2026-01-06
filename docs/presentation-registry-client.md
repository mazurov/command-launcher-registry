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

# COLA Registry Client

## cola-regctl Command Line Tool

**Manage packages and registries from the command line**

Full-featured CLI client for interacting with COLA Registry servers

---

# Client Overview

```
┌───────────────────────────────────────────────────────────────┐
│                      cola-regctl CLI                          │
├───────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ Commands    │  │ Auth Layer  │  │ Output      │            │
│  │             │  │             │  │ Formatters  │            │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘            │
│         └────────────────┼────────────────┘                   │
│  ┌───────────────────────┴───────────────────────┐            │
│  │           HTTP Client + Credential Store       │            │
│  └───────────────────────┬───────────────────────┘            │
│  ┌───────────────┬───────┴───────┬───────────────┐            │
│  │   macOS       │   Windows     │    Linux      │            │
│  │   Keychain    │   Credential  │   File-based  │            │
│  │               │   Manager     │   (~/.cola/)  │            │
│  └───────────────┴───────────────┴───────────────┘            │
└───────────────────────────────────────────────────────────────┘
```

---

# Command Structure

| Command | Subcommands | Description |
|---------|-------------|-------------|
| `login` | - | Authenticate with registry server |
| `logout` | - | Remove stored credentials |
| `whoami` | - | Display current authenticated user |
| `registry` | `list`, `add`, `remove`, `set-default` | Manage registry configurations |
| `package` | `list`, `get`, `push`, `delete` | Manage packages |
| `version` | `list`, `get`, `push`, `delete` | Manage package versions |

**8 main commands** with **17 subcommands**

---

# Authentication Commands

### Login

```bash
# Interactive login (prompts for password)
cola-regctl login --url https://registry.example.com --user admin

# Non-interactive login (for scripts/CI)
cola-regctl login --url https://registry.example.com \
    --user admin --password secret

# Using environment variables
export COLA_REGISTRY_URL=https://registry.example.com
export COLA_REGISTRY_SESSION_TOKEN=admin:password
cola-regctl login
```

---

# Authentication Commands (continued)

### Session Management

```bash
# Check current authenticated user
cola-regctl whoami
# Output: Logged in as: admin

# Logout from current registry
cola-regctl logout

# Logout from specific registry
cola-regctl logout --url https://registry.example.com
```

---

# Credential Storage

Platform-specific secure credential storage:

| Platform | Storage Location | Security |
|----------|------------------|----------|
| macOS | System Keychain | Hardware-backed encryption |
| Windows | Credential Manager | DPAPI encryption |
| Linux | `~/.cola/credentials` | File permissions (0600) |

### Credential Flow

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│    Login     │───▶│   Validate   │───▶│    Store     │
│   Command    │    │   with API   │    │  Credential  │
└──────────────┘    └──────────────┘    └──────────────┘
                                               │
                    ┌──────────────┐           │
                    │ Future Cmds  │◀──────────┘
                    │  Auto-Auth   │
                    └──────────────┘
```

---

# Registry Management

### Adding Registries

```bash
# Add a new registry
cola-regctl registry add --name prod \
    --url https://registry.example.com

# Add and set as default
cola-regctl registry add --name dev \
    --url https://dev-registry.example.com --default
```

### Managing Registries

```bash
# List all configured registries
cola-regctl registry list

# Set default registry
cola-regctl registry set-default prod

# Remove a registry
cola-regctl registry remove dev
```

---

# Registry List Output

```bash
$ cola-regctl registry list
┌──────────┬─────────────────────────────────┬─────────┐
│ NAME     │ URL                             │ DEFAULT │
├──────────┼─────────────────────────────────┼─────────┤
│ prod     │ https://registry.example.com    │ ✓       │
│ dev      │ https://dev-registry.local:8080 │         │
│ staging  │ https://staging.example.com     │         │
└──────────┴─────────────────────────────────┴─────────┘
```

---

# Package Management

### Listing Packages

```bash
# List all packages
cola-regctl package list

# List with JSON output
cola-regctl package list --output json

# List packages from specific registry
cola-regctl package list --url https://registry.example.com
```

### Getting Package Details

```bash
# Get package metadata
cola-regctl package get mypackage

# Get as JSON for scripting
cola-regctl package get mypackage --output json
```

---

# Package Push

### Pushing New Packages

```bash
# Push a package (creates if not exists)
cola-regctl package push mypackage \
    --description "My CLI tool"

# Push with metadata
cola-regctl package push mypackage \
    --description "My CLI tool" \
    --maintainer "team@example.com" \
    --repository "https://github.com/org/repo"
```

### Package Push Options

| Option | Description |
|--------|-------------|
| `--description` | Package description |
| `--maintainer` | Maintainer email/name |
| `--repository` | Source repository URL |
| `--homepage` | Project homepage |

---

# Version Management

### Listing Versions

```bash
# List all versions of a package
cola-regctl version list mypackage

# Filter by platform
cola-regctl version list mypackage --platform linux

# Filter by architecture
cola-regctl version list mypackage --arch amd64
```

### Version Output

```bash
$ cola-regctl version list mypackage
┌─────────┬──────────┬─────────┬────────────────────────────┐
│ VERSION │ PLATFORM │ ARCH    │ ARTIFACT                   │
├─────────┼──────────┼─────────┼────────────────────────────┤
│ 1.2.0   │ linux    │ amd64   │ mypackage-linux-amd64      │
│ 1.2.0   │ linux    │ arm64   │ mypackage-linux-arm64      │
│ 1.2.0   │ darwin   │ amd64   │ mypackage-darwin-amd64     │
│ 1.2.0   │ darwin   │ arm64   │ mypackage-darwin-arm64     │
│ 1.1.0   │ linux    │ amd64   │ mypackage-linux-amd64      │
└─────────┴──────────┴─────────┴────────────────────────────┘
```

---

# Version Push

### Pushing Package Versions

```bash
# Push a new version
cola-regctl version push mypackage 1.2.0 \
    --platform linux --arch amd64 \
    --artifact ./build/mypackage-linux-amd64

# Push with checksum
cola-regctl version push mypackage 1.2.0 \
    --platform darwin --arch arm64 \
    --artifact ./build/mypackage-darwin-arm64 \
    --checksum sha256:abc123...
```

---

# Version Push Options

| Option | Required | Description |
|--------|----------|-------------|
| `--platform` | Yes | Target OS (linux, darwin, windows) |
| `--arch` | Yes | Target architecture (amd64, arm64, 386) |
| `--artifact` | Yes | Path to binary artifact |
| `--checksum` | No | SHA256 checksum for verification |
| `--url` | No | Download URL (if external hosting) |

### Supported Platforms & Architectures

| Platforms | Architectures |
|-----------|---------------|
| `linux`, `darwin`, `windows` | `amd64`, `arm64`, `386`, `arm` |

---

# Delete Operations

### Deleting Packages

```bash
# Delete a package (requires confirmation)
cola-regctl package delete mypackage

# Force delete (no confirmation)
cola-regctl package delete mypackage --force
```

### Deleting Versions

```bash
# Delete specific version
cola-regctl version delete mypackage 1.0.0 \
    --platform linux --arch amd64

# Force delete
cola-regctl version delete mypackage 1.0.0 \
    --platform linux --arch amd64 --force
```

---

# Global Flags

Available for all commands:

| Flag | Short | Environment Variable | Description |
|------|-------|---------------------|-------------|
| `--url` | `-u` | `COLA_REGISTRY_URL` | Registry server URL |
| `--user` | | `COLA_REGISTRY_USER` | Username |
| `--password` | | `COLA_REGISTRY_PASSWORD` | Password |
| `--output` | `-o` | | Output format (table, json) |
| `--verbose` | `-v` | | Enable verbose output |
| `--help` | `-h` | | Show help |

---

# Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `COLA_REGISTRY_URL` | Default registry URL | `https://registry.example.com` |
| `COLA_REGISTRY_USER` | Username for auth | `admin` |
| `COLA_REGISTRY_PASSWORD` | Password for auth | `secret` |
| `COLA_REGISTRY_SESSION_TOKEN` | Session token (user:pass) | `admin:secret` |
| `COLA_CONFIG_DIR` | Config directory | `~/.cola` |

### Priority Order

1. Command-line flags (highest)
2. Environment variables
3. Stored credentials
4. Config file defaults (lowest)

---

# Output Formats

### Table Output (Default)

```bash
$ cola-regctl package list
┌─────────────┬─────────────────────────┬─────────────┐
│ NAME        │ DESCRIPTION             │ VERSIONS    │
├─────────────┼─────────────────────────┼─────────────┤
│ mytool      │ CLI utility             │ 3           │
│ helper      │ Helper scripts          │ 5           │
└─────────────┴─────────────────────────┴─────────────┘
```

### JSON Output

```bash
$ cola-regctl package list --output json
[
  {"name": "mytool", "description": "CLI utility", "versions": 3},
  {"name": "helper", "description": "Helper scripts", "versions": 5}
]
```

---

# Exit Codes

| Code | Meaning | Example |
|------|---------|---------|
| `0` | Success | Command completed successfully |
| `1` | General error | Unspecified error |
| `2` | Authentication failed | Invalid credentials |
| `3` | Not found | Package/version not found |
| `4` | Conflict | Resource already exists |
| `5` | Validation error | Invalid input |
| `6` | Network error | Connection failed |

### Scripting Example

```bash
if cola-regctl package get mypackage > /dev/null 2>&1; then
    echo "Package exists"
else
    echo "Package not found"
fi
```

---

# CI/CD Integration

### GitHub Actions Example

```yaml
- name: Push to Registry
  env:
    COLA_REGISTRY_URL: ${{ secrets.REGISTRY_URL }}
    COLA_REGISTRY_SESSION_TOKEN: ${{ secrets.REGISTRY_TOKEN }}
  run: |
    cola-regctl version push mypackage ${{ github.ref_name }} \
      --platform linux --arch amd64 \
      --artifact ./dist/mypackage-linux-amd64
```

---

# CI/CD Integration (continued)

### GitLab CI Example

```yaml
deploy:
  script:
    - |
      cola-regctl login \
        --url $COLA_REGISTRY_URL \
        --user $COLA_REGISTRY_USER \
        --password $COLA_REGISTRY_PASSWORD
    - |
      cola-regctl version push mypackage $CI_COMMIT_TAG \
        --platform linux --arch amd64 \
        --artifact ./dist/mypackage-linux-amd64
```

---

# Multi-Platform Builds

### Build Matrix Example

```bash
#!/bin/bash
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")
VERSION=$1

for PLATFORM in "${PLATFORMS[@]}"; do
    OS="${PLATFORM%/*}"
    ARCH="${PLATFORM#*/}"
    ARTIFACT="./dist/mypackage-${OS}-${ARCH}"

    cola-regctl version push mypackage "$VERSION" \
        --platform "$OS" --arch "$ARCH" \
        --artifact "$ARTIFACT"
done
```

---

# Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Connection refused | Check `--url` or `COLA_REGISTRY_URL` |
| Authentication failed | Run `cola-regctl login` or check credentials |
| Permission denied | Verify user has required permissions |
| Package not found | Check package name and registry |

### Verbose Mode

```bash
# Enable verbose output for debugging
cola-regctl --verbose package list
```

---

# Quick Reference

```bash
# Authentication
cola-regctl login --url URL --user USER
cola-regctl logout
cola-regctl whoami

# Registry management
cola-regctl registry list
cola-regctl registry add --name NAME --url URL
cola-regctl registry set-default NAME

# Package operations
cola-regctl package list
cola-regctl package get NAME
cola-regctl package push NAME --description DESC

# Version operations
cola-regctl version list PACKAGE
cola-regctl version push PACKAGE VERSION --platform OS --arch ARCH --artifact PATH
```

---

# Summary

### Key Features

- **Platform-specific secure credential storage**
- **Multiple registry management**
- **Full package lifecycle management**
- **Multi-platform version support**
- **JSON output for scripting**
- **Environment variable configuration**
- **CI/CD ready**

### Installation

```bash
# Download for your platform
curl -L https://registry.example.com/cola-regctl-linux-amd64 -o cola-regctl
chmod +x cola-regctl
./cola-regctl --help
```
