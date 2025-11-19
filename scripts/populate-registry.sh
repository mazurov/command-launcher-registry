#!/bin/bash

# populate-registry.sh
# Populates the registry with fake data for testing and development
# Usage: ./scripts/populate-registry.sh [api-url]

set -e

API_URL="${1:-http://localhost:8080}"
CLI="./bin/cola-registry-cli --api-url $API_URL"

echo "🚀 Populating registry at $API_URL with fake data..."
echo ""
echo "Building CLI..."
make build-cli > /dev/null 2>&1
echo ""

# Function to create a registry
create_registry() {
    local name=$1
    local description=$2
    echo "📦 Creating registry: $name"
    $CLI registry create --name "$name" --description "$description" 2>&1 | grep -v "^go: downloading" || true
}

# Function to create a package
create_package() {
    local registry=$1
    local package=$2
    local description=$3
    echo "  📄 Creating package: $package"
    $CLI package create --registry "$registry" --name "$package" --description "$description" 2>&1 | grep -v "^go: downloading" || true
}

# Function to publish a version
publish_version() {
    local registry=$1
    local package=$2
    local version=$3
    local url=$4
    local checksum=$5
    local start_partition=${6:-0}
    local end_partition=${7:-9}
    echo "    🔖 Publishing version: $version"
    $CLI version publish \
        --registry "$registry" \
        --package "$package" \
        --version "$version" \
        --url "$url" \
        --checksum "$checksum" \
        --start-partition "$start_partition" \
        --end-partition "$end_partition" 2>&1 | grep -v "^go: downloading" || true
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Creating Registries"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Create development registry
create_registry "development" "Development tools and utilities"

# Create production registry
create_registry "production" "Production-ready packages"

# Create experimental registry
create_registry "experimental" "Experimental and beta packages"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Populating Development Registry"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Package: kubectl
create_package "development" "kubectl" "Kubernetes command-line tool"
publish_version "development" "kubectl" "1.28.0" "https://dl.k8s.io/release/v1.28.0/bin/linux/amd64/kubectl.zip" "sha256:aa939ca3e423f0dc9e76c9e8fdb8e1b6f1e3f1c4e9f8d9e8c7f6e5d4c3b2a190"
publish_version "development" "kubectl" "1.29.0" "https://dl.k8s.io/release/v1.29.0/bin/linux/amd64/kubectl.zip" "sha256:bb939ca3e423f0dc9e76c9e8fdb8e1b6f1e3f1c4e9f8d9e8c7f6e5d4c3b2a191"
publish_version "development" "kubectl" "1.30.0" "https://dl.k8s.io/release/v1.30.0/bin/linux/amd64/kubectl.zip" "sha256:cc939ca3e423f0dc9e76c9e8fdb8e1b6f1e3f1c4e9f8d9e8c7f6e5d4c3b2a192"

# Package: terraform
create_package "development" "terraform" "Infrastructure as Code tool"
publish_version "development" "terraform" "1.6.0" "https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip" "sha256:d117883fd98b960c5d0f012b0d4b21801e1c9f7e50e7f5f5c1f0e9d8c7b6a593"
publish_version "development" "terraform" "1.6.1" "https://releases.hashicorp.com/terraform/1.6.1/terraform_1.6.1_linux_amd64.zip" "sha256:e117883fd98b960c5d0f012b0d4b21801e1c9f7e50e7f5f5c1f0e9d8c7b6a594"
publish_version "development" "terraform" "1.7.0" "https://releases.hashicorp.com/terraform/1.7.0/terraform_1.7.0_linux_amd64.zip" "sha256:f117883fd98b960c5d0f012b0d4b21801e1c9f7e50e7f5f5c1f0e9d8c7b6a595"

# Package: docker
create_package "development" "docker" "Container platform CLI"
publish_version "development" "docker" "24.0.0" "https://download.docker.com/linux/static/stable/x86_64/docker-24.0.0.zip" "sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
publish_version "development" "docker" "24.0.5" "https://download.docker.com/linux/static/stable/x86_64/docker-24.0.5.zip" "sha256:2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
publish_version "development" "docker" "25.0.0" "https://download.docker.com/linux/static/stable/x86_64/docker-25.0.0.zip" "sha256:3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

# Package: jq
create_package "development" "jq" "JSON processor"
publish_version "development" "jq" "1.6" "https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64.zip" "sha256:af986793a515d500ab2d35f8d2aecd656e764504b789b66d7e1a0b727a124c44"
publish_version "development" "jq" "1.7" "https://github.com/stedolan/jq/releases/download/jq-1.7/jq-linux64.zip" "sha256:bf986793a515d500ab2d35f8d2aecd656e764504b789b66d7e1a0b727a124c45"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Populating Production Registry"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Package: envoy
create_package "production" "envoy" "Cloud-native proxy"
publish_version "production" "envoy" "1.28.0" "https://github.com/envoyproxy/envoy/releases/download/v1.28.0/envoy-1.28.0-linux-x86_64.zip" "sha256:a1a2a3a4a5a6a7a8a9a0b1b2b3b4b5b6b7b8b9b0c1c2c3c4c5c6c7c8c9c0d1d2" 0 3
publish_version "production" "envoy" "1.29.0" "https://github.com/envoyproxy/envoy/releases/download/v1.29.0/envoy-1.29.0-linux-x86_64.zip" "sha256:b1b2b3b4b5b6b7b8b9b0c1c2c3c4c5c6c7c8c9c0d1d2d3d4d5d6d7d8d9d0e1e2" 4 6
publish_version "production" "envoy" "1.30.0" "https://github.com/envoyproxy/envoy/releases/download/v1.30.0/envoy-1.30.0-linux-x86_64.zip" "sha256:c1c2c3c4c5c6c7c8c9c0d1d2d3d4d5d6d7d8d9d0e1e2e3e4e5e6e7e8e9e0f1f2" 7 9

# Package: prometheus
create_package "production" "prometheus" "Monitoring system"
publish_version "production" "prometheus" "2.45.0" "https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.zip" "sha256:d1d2d3d4d5d6d7d8d9d0e1e2e3e4e5e6e7e8e9e0f1f2f3f4f5f6f7f8f9f0a1a2"
publish_version "production" "prometheus" "2.46.0" "https://github.com/prometheus/prometheus/releases/download/v2.46.0/prometheus-2.46.0.linux-amd64.zip" "sha256:e1e2e3e4e5e6e7e8e9e0f1f2f3f4f5f6f7f8f9f0a1a2a3a4a5a6a7a8a9a0b1b2"
publish_version "production" "prometheus" "2.47.0" "https://github.com/prometheus/prometheus/releases/download/v2.47.0/prometheus-2.47.0.linux-amd64.zip" "sha256:f1f2f3f4f5f6f7f8f9f0a1a2a3a4a5a6a7a8a9a0b1b2b3b4b5b6b7b8b9b0c1c2"

# Package: grafana
create_package "production" "grafana" "Observability platform"
publish_version "production" "grafana" "10.0.0" "https://dl.grafana.com/oss/release/grafana-10.0.0.linux-amd64.zip" "sha256:1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2"
publish_version "production" "grafana" "10.1.0" "https://dl.grafana.com/oss/release/grafana-10.1.0.linux-amd64.zip" "sha256:2a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2"
publish_version "production" "grafana" "10.2.0" "https://dl.grafana.com/oss/release/grafana-10.2.0.linux-amd64.zip" "sha256:3a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Populating Experimental Registry"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Package: hotfix
create_package "experimental" "hotfix" "Emergency hotfix tool"
publish_version "experimental" "hotfix" "1.0.0-44733" "https://example.com/hotfix/hotfix-44733.zip" "5f5f47e4966b984a4c7d33003dd2bbe8fff5d31bf2bee0c6db3add099e4542b3" 0 9
publish_version "experimental" "hotfix" "1.0.0-45149" "https://example.com/hotfix/hotfix-45149.zip" "773a919429e50346a7a002eb3ecbf2b48d058bae014df112119a67fc7d9a3598" 6 8
publish_version "experimental" "hotfix" "1.0.0-45200" "https://example.com/hotfix/hotfix-45200.zip" "8a3a919429e50346a7a002eb3ecbf2b48d058bae014df112119a67fc7d9a3599" 0 5

# Package: env
create_package "experimental" "env" "Environment configuration tool"
publish_version "experimental" "env" "0.0.1" "https://example.com/env/package.zip" "c87a417cce3d26777bcc6b8b0dea2ec43a0d78486438b1bf3f3fbd2cafc2c7cc" 0 9
publish_version "experimental" "env" "0.0.2-alpha" "https://example.com/env/package-0.0.2.zip" "d87a417cce3d26777bcc6b8b0dea2ec43a0d78486438b1bf3f3fbd2cafc2c7cd"
publish_version "experimental" "env" "0.1.0-beta" "https://example.com/env/package-0.1.0.zip" "e87a417cce3d26777bcc6b8b0dea2ec43a0d78486438b1bf3f3fbd2cafc2c7ce"

# Package: proto-compiler
create_package "experimental" "proto-compiler" "Protocol Buffer compiler"
publish_version "experimental" "proto-compiler" "0.1.0" "https://example.com/proto/proto-compiler-0.1.0.pkg" "1234abcd5678efab9012cdef3456fabe7890dcba1234567890abcdef12345678"
publish_version "experimental" "proto-compiler" "0.2.0-rc1" "https://example.com/proto/proto-compiler-0.2.0-rc1.pkg" "2234abcd5678efab9012cdef3456fabe7890dcba1234567890abcdef12345678"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Registry Population Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Summary:"
echo "  • 3 registries created (development, production, experimental)"
echo "  • 10 packages created"
echo "  • 28 versions published"
echo ""
echo "Test the API:"
echo "  curl $API_URL/remote/registries/development/index.json"
echo "  curl $API_URL/remote/registries/production/index.json"
echo "  curl $API_URL/remote/registries/experimental/index.json"
echo ""
