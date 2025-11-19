.PHONY: help build test run-server run-cli clean docker

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	go mod download
	go mod tidy

build: build-server build-cli ## Build server and CLI binaries

build-server: ## Build only the server
	go build -o bin/cola-registry-server ./cmd/server

build-cli: ## Build only the CLI
	go build -o bin/cola-registry-cli ./cmd/cli

test: ## Run tests
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

run-server: build-server ## Run the server (SQLite)
	./bin/cola-registry-server --db-type sqlite --db-dsn registry.db --log-level debug

run-server-pg: build-server ## Run the server (PostgreSQL)
	./bin/cola-registry-server \
		--db-type postgres \
		--db-dsn "host=localhost user=registry password=secret dbname=registry sslmode=disable" \
		--log-level debug

run-cli: build-cli ## Run the CLI with help
	./bin/cola-registry-cli --help

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f *.db
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t cola-registry:latest -f deployments/docker/Dockerfile .

docker-run: ## Run Docker container with SQLite
	docker run -p 8080:8080 \
		-e REGISTRY_DB_TYPE=sqlite \
		-e REGISTRY_DB_DSN=/data/registry.db \
		-v $(PWD)/data:/data \
		cola-registry:latest

docker-compose-up: ## Start services with docker-compose
	cd deployments/docker && docker-compose up -d

docker-compose-down: ## Stop docker-compose services
	cd deployments/docker && docker-compose down

docker-compose-logs: ## View docker-compose logs
	cd deployments/docker && docker-compose logs -f

docker-compose-build: ## Build and start services with docker-compose
	cd deployments/docker && docker-compose up -d --build

lint: ## Run linters
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found in PATH. Run 'make install-tools' and add \$$(go env GOPATH)/bin to your PATH"; exit 1; }
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

dev: ## Run in development mode with hot reload (requires air)
	@command -v air >/dev/null 2>&1 || { echo "air not found in PATH. Run 'make install-tools' and add \$$(go env GOPATH)/bin to your PATH"; exit 1; }
	air

install-tools: ## Install development tools
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo ""
	@echo "✅ Tools installed successfully!"
	@echo "Add Go binaries to your PATH by running:"
	@echo "  export PATH=\"\$$(go env GOPATH)/bin:\$$PATH\""
	@echo "Or add it to your ~/.zshrc or ~/.bashrc"

.DEFAULT_GOAL := help
