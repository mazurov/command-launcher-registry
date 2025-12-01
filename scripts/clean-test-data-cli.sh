#!/usr/bin/env bash
#
# clean-test-data-cli.sh - Clean test data from the registry using cola-regctl CLI
#
# This script deletes the test registry and all its packages and versions.
# It demonstrates CLI usage and provides a convenient way to clean up test data.
#
# Usage:
#   ./scripts/clean-test-data-cli.sh [server_url] [username:password]
#
# Examples:
#   ./scripts/clean-test-data-cli.sh
#   ./scripts/clean-test-data-cli.sh http://localhost:8080 admin:admin
#
# Prerequisites:
#   - cola-regctl binary in PATH or bin/cola-regctl
#   - Server running at specified URL (default: http://localhost:8080)
#

set -e

# Configuration
SERVER_URL="${1:-http://localhost:8080}"
CREDENTIALS="${2:-admin:admin}"

# Set environment variables for CLI authentication
export COLA_REGISTRY_URL="${SERVER_URL}"
export COLA_REGISTRY_SESSION_TOKEN="${CREDENTIALS}"

# Find cola-regctl binary
if command -v cola-regctl &> /dev/null; then
    CLI="cola-regctl"
elif [ -f "./bin/cola-regctl" ]; then
    CLI="./bin/cola-regctl"
else
    echo "Error: cola-regctl not found in PATH or ./bin/cola-regctl"
    echo "Please build the CLI first: make build-cli"
    exit 1
fi

echo "Using CLI: ${CLI}"
echo "Server URL: ${COLA_REGISTRY_URL}"
echo "Cleaning test data using cola-regctl..."
echo

# Check authentication
echo "Authenticating to ${COLA_REGISTRY_URL}..."
${CLI} registry list > /dev/null 2>&1 || {
    echo "Error: Failed to authenticate to server"
    echo "Please ensure server is running and credentials are correct"
    exit 1
}
echo "✓ Authentication successful"
echo

# Delete the entire registry (cascade deletes all packages and versions)
echo "Deleting registry 'company-tools' (this will also delete all packages and versions)..."
${CLI} registry delete company-tools --yes
echo

echo "✓ Test data cleaned successfully!"
echo
echo "To verify registry was deleted, run:"
echo "  ${CLI} registry list"
echo
echo "Or check the registry index (should return 404):"
echo "  curl -i ${COLA_REGISTRY_URL}/api/v1/registry/company-tools/index.json"
echo
