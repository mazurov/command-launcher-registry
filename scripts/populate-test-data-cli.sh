#!/usr/bin/env bash
#
# populate-test-data-cli.sh - Populate the registry with test data using cola-regctl CLI
#
# This script creates a sample registry with packages and versions for testing.
# It demonstrates CLI usage and provides a convenient way to populate test data.
#
# Usage:
#   ./scripts/populate-test-data-cli.sh [server_url] [username:password]
#
# Examples:
#   ./scripts/populate-test-data-cli.sh
#   ./scripts/populate-test-data-cli.sh http://localhost:8080 admin:admin
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
echo "Populating test data using cola-regctl..."
echo

# Login to the server
echo "Logging in to ${COLA_REGISTRY_URL}..."

# Test authentication
${CLI} registry list > /dev/null 2>&1 || {
    echo "Error: Failed to authenticate to server"
    echo "Please ensure server is running and credentials are correct"
    exit 1
}
echo "✓ Authentication successful"
echo

# Create registry 'company-tools'
echo "Creating registry 'company-tools'..."
${CLI} registry create company-tools \
    --description "Company internal tools registry" \
    --admin "admin@example.com" \
    --custom-value "team=platform" \
    --custom-value "owner=devops"
echo

# Create package 'deployment-cli'
echo "Creating package 'deployment-cli'..."
${CLI} package create company-tools deployment-cli \
    --description "CLI tool for deploying applications" \
    --maintainer "deploy-team@example.com" \
    --custom-value "language=go" \
    --custom-value "repo=https://github.com/company/deployment-cli"
echo

# Create version 1.0.0 for deployment-cli
echo "Creating version 1.0.0 for deployment-cli..."
${CLI} version create company-tools deployment-cli 1.0.0 \
    --checksum "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" \
    --url "https://downloads.example.com/deployment-cli/1.0.0/deployment-cli.zip" \
    --start-partition 0 \
    --end-partition 4
echo

# Create version 1.1.0 for deployment-cli
echo "Creating version 1.1.0 for deployment-cli..."
${CLI} version create company-tools deployment-cli 1.1.0 \
    --checksum "sha256:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210" \
    --url "https://downloads.example.com/deployment-cli/1.1.0/deployment-cli.zip" \
    --start-partition 5 \
    --end-partition 9
echo

# Create package 'data-sync'
echo "Creating package 'data-sync'..."
${CLI} package create company-tools data-sync \
    --description "Data synchronization utility" \
    --maintainer "data-team@example.com" \
    --custom-value "language=python" \
    --custom-value "repo=https://github.com/company/data-sync"
echo

# Create version 2.0.0 for data-sync
echo "Creating version 2.0.0 for data-sync..."
${CLI} version create company-tools data-sync 2.0.0 \
    --checksum "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" \
    --url "https://downloads.example.com/data-sync/2.0.0/data-sync.pkg" \
    --start-partition 0 \
    --end-partition 9
echo

# Create package 'monitoring-agent'
echo "Creating package 'monitoring-agent'..."
${CLI} package create company-tools monitoring-agent \
    --description "System monitoring agent" \
    --maintainer "sre-team@example.com"
echo

# Create version 3.5.2 for monitoring-agent
echo "Creating version 3.5.2 for monitoring-agent..."
${CLI} version create company-tools monitoring-agent 3.5.2 \
    --checksum "sha256:9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba" \
    --url "https://downloads.example.com/monitoring-agent/3.5.2/monitoring-agent.pkg" \
    --start-partition 0 \
    --end-partition 9
echo

echo "✓ Test data populated successfully!"
echo
echo "To verify, run:"
echo "  ${CLI} registry list"
echo "  ${CLI} package list company-tools"
echo "  ${CLI} version list company-tools deployment-cli"
echo
echo "Or check the registry index:"
echo "  curl ${COLA_REGISTRY_URL}/api/v1/registry/company-tools/index.json | jq '.'"
echo
