#!/usr/bin/env bash
#
# populate-test-data.sh - Populate the registry with test data
#
# This script creates a sample registry with packages and versions for testing.
# It assumes the server is running at http://localhost:8080 and uses basic auth
# with admin:admin credentials.
#
# Usage:
#   ./scripts/populate-test-data.sh [base_url] [username:password]
#
# Examples:
#   ./scripts/populate-test-data.sh
#   ./scripts/populate-test-data.sh http://localhost:8080 admin:admin
#

set -e

# Configuration
BASE_URL="${1:-http://localhost:8080}"
AUTH="${2:-admin:admin}"
API_BASE="${BASE_URL}/api/v1"

echo "Populating test data to ${BASE_URL}..."
echo

# Function to make authenticated API calls
api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"

    if [ -n "$data" ]; then
        curl -s -X "${method}" \
            -H "Content-Type: application/json" \
            -u "${AUTH}" \
            -d "${data}" \
            "${API_BASE}${endpoint}"
    else
        curl -s -X "${method}" \
            -u "${AUTH}" \
            "${API_BASE}${endpoint}"
    fi
}

# Create registry
echo "Creating registry 'company-tools'..."
api_call POST "/registry" '{
  "name": "company-tools",
  "description": "Company internal tools registry",
  "admins": ["admin@example.com"],
  "custom_values": {
    "team": "platform",
    "owner": "devops"
  }
}' | jq '.'
echo

# Create package 1: deployment-cli
echo "Creating package 'deployment-cli'..."
api_call POST "/registry/company-tools/package" '{
  "name": "deployment-cli",
  "description": "CLI tool for deploying applications",
  "maintainers": ["deploy-team@example.com"],
  "custom_values": {
    "language": "go",
    "repo": "https://github.com/company/deployment-cli"
  }
}' | jq '.'
echo

# Create versions for deployment-cli
echo "Creating version 1.0.0 for deployment-cli..."
api_call POST "/registry/company-tools/package/deployment-cli/version" '{
  "name": "deployment-cli",
  "version": "1.0.0",
  "checksum": "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "url": "https://downloads.example.com/deployment-cli/1.0.0/deployment-cli",
  "startPartition": 0,
  "endPartition": 4
}' | jq '.'
echo

echo "Creating version 1.1.0 for deployment-cli..."
api_call POST "/registry/company-tools/package/deployment-cli/version" '{
  "name": "deployment-cli",
  "version": "1.1.0",
  "checksum": "sha256:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
  "url": "https://downloads.example.com/deployment-cli/1.1.0/deployment-cli",
  "startPartition": 5,
  "endPartition": 9
}' | jq '.'
echo

# Create package 2: data-sync
echo "Creating package 'data-sync'..."
api_call POST "/registry/company-tools/package" '{
  "name": "data-sync",
  "description": "Data synchronization utility",
  "maintainers": ["data-team@example.com"],
  "custom_values": {
    "language": "python",
    "repo": "https://github.com/company/data-sync"
  }
}' | jq '.'
echo

# Create version for data-sync
echo "Creating version 2.0.0 for data-sync..."
api_call POST "/registry/company-tools/package/data-sync/version" '{
  "name": "data-sync",
  "version": "2.0.0",
  "checksum": "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
  "url": "https://downloads.example.com/data-sync/2.0.0/data-sync",
  "startPartition": 0,
  "endPartition": 9
}' | jq '.'
echo

# Create package 3: monitoring-agent
echo "Creating package 'monitoring-agent'..."
api_call POST "/registry/company-tools/package" '{
  "name": "monitoring-agent",
  "description": "System monitoring agent",
  "maintainers": ["sre-team@example.com"]
}' | jq '.'
echo

# Create version for monitoring-agent
echo "Creating version 3.5.2 for monitoring-agent..."
api_call POST "/registry/company-tools/package/monitoring-agent/version" '{
  "name": "monitoring-agent",
  "version": "3.5.2",
  "checksum": "sha256:9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba",
  "url": "https://downloads.example.com/monitoring-agent/3.5.2/monitoring-agent",
  "startPartition": 0,
  "endPartition": 9
}' | jq '.'
echo

echo "Test data populated successfully!"
echo
echo "To verify:"
echo "  curl ${BASE_URL}/api/v1/registry/company-tools/index.json | jq '.'"
