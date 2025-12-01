#!/usr/bin/env bash
#
# clean-test-data.sh - Clean test data from the registry
#
# This script deletes the test registry and all its packages and versions.
# It assumes the server is running at http://localhost:8080 and uses basic auth
# with admin:admin credentials.
#
# Usage:
#   ./scripts/clean-test-data.sh [base_url] [username:password]
#
# Examples:
#   ./scripts/clean-test-data.sh
#   ./scripts/clean-test-data.sh http://localhost:8080 admin:admin
#

set -e

# Configuration
BASE_URL="${1:-http://localhost:8080}"
AUTH="${2:-admin:admin}"
API_BASE="${BASE_URL}/api/v1"

echo "Cleaning test data from ${BASE_URL}..."
echo

# Function to make authenticated API calls
api_delete() {
    local endpoint="$1"

    curl -s -X DELETE \
        -u "${AUTH}" \
        "${API_BASE}${endpoint}" \
        -w "\nHTTP Status: %{http_code}\n"
}

# Delete the entire registry (cascade deletes all packages and versions)
echo "Deleting registry 'company-tools'..."
api_delete "/registry/company-tools"
echo

echo "Test data cleaned successfully!"
echo
echo "To verify:"
echo "  curl ${BASE_URL}/api/v1/registry/company-tools/index.json"
echo "  (Should return 404 Not Found)"
