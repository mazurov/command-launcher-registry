# Scripts

Utility scripts for the Command Launcher Registry.

## populate-registry.sh

Populates a running registry with fake/test data for development and testing purposes.

### Usage

```bash
# Start the server first
make run-server

# In another terminal, run the populate script
./scripts/populate-registry.sh

# Or specify a custom API URL
./scripts/populate-registry.sh http://custom-host:8080
```

### What it Creates

The script creates **3 registries** with **10 packages** and **28 versions** total:

#### Development Registry
- **kubectl** - Kubernetes CLI (3 versions: 1.28.0, 1.29.0, 1.30.0)
- **terraform** - Infrastructure as Code tool (3 versions: 1.6.0, 1.6.1, 1.7.0)
- **docker** - Container platform CLI (3 versions: 24.0.0, 24.0.5, 25.0.0)
- **jq** - JSON processor (2 versions: 1.6, 1.7)

#### Production Registry
- **envoy** - Cloud-native proxy (3 versions with partition ranges)
- **prometheus** - Monitoring system (3 versions: 2.45.0, 2.46.0, 2.47.0)
- **grafana** - Observability platform (3 versions: 10.0.0, 10.1.0, 10.2.0)

#### Experimental Registry
- **hotfix** - Emergency hotfix tool (3 versions with build numbers)
- **env** - Environment configuration tool (3 versions including alpha/beta)
- **proto-compiler** - Protocol Buffer compiler (2 versions including RC)

### Testing the Results

After running the script, test the populated data:

```bash
# Get all versions from development registry
curl http://localhost:8080/remote/registries/development/index.json

# Get all versions from production registry
curl http://localhost:8080/remote/registries/production/index.json

# Get all versions from experimental registry
curl http://localhost:8080/remote/registries/experimental/index.json

# List all registries
./bin/cola-registry-cli registry list

# List packages in a specific registry
./bin/cola-registry-cli package list development
```

### Requirements

- Server must be running (on localhost:8080 by default)
- Go must be installed (script uses `go run` to execute CLI commands)

### Notes

- The script filters out `go: downloading` messages for cleaner output
- All URLs and checksums are fake/example data
- Some versions include partition ranges for testing canary deployments
- The script uses `set -e` to stop on first error
