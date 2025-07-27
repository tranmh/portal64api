# Portal64 API Makefile

.PHONY: help build run test test-unit test-integration clean deps swagger docker docker-build docker-run

# Variables
BINARY_NAME=portal64api
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=portal64api
DOCKER_TAG=latest

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

## help: Show this help message
help:
	@echo "Portal64 API - Available commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin
	@go build -ldflags="-w -s" -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)Binary built: $(BINARY_PATH)$(NC)"

## run: Run the application in development mode
run:
	@echo "$(GREEN)Starting development server...$(NC)"
	@go run $(MAIN_PATH)

## run-prod: Run the application with production config
run-prod:
	@echo "$(GREEN)Starting production server...$(NC)"
	@ENVIRONMENT=production go run $(MAIN_PATH)

## deps: Download and tidy dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

## test: Run unit and integration tests (excluding system tests)
test: test-unit test-integration

## test-all: Run all tests including system tests
test-all: test-unit test-integration test-system

## test-unit: Run unit tests
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v ./tests/unit/...

## test-integration: Run integration tests
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	@go test -v ./tests/integration/ -run "^((?!TestSystemSuite).)*$$"

## test-system: Run system tests against live deployment
test-system:
	@echo "$(GREEN)Running system tests against http://test.svw.info:8080...$(NC)"
	@echo "$(YELLOW)Checking server health...$(NC)"
	@curl -f http://test.svw.info:8080/health > /dev/null 2>&1 || (echo "$(RED)Error: Test server not reachable at http://test.svw.info:8080$(NC)" && exit 1)
	@echo "$(GREEN)Server is healthy, running system tests...$(NC)"
	@go test -v ./tests/integration/ -run TestSystemSuite -timeout=300s

## test-coverage: Run tests with coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

## swagger: Generate Swagger documentation
swagger:
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@swag init -g cmd/server/main.go -o docs/generated
	@echo "$(GREEN)Swagger docs generated in docs/generated/$(NC)"

## lint: Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run

## format: Format code
format:
	@echo "$(GREEN)Formatting code...$(NC)"
	@gofmt -s -w .
	@go mod tidy

## clean: Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf docs/generated/
	@echo "$(GREEN)Clean complete$(NC)"

## install-tools: Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Development tools installed$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

## docker-run: Run application in Docker
docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

## docker-compose-up: Start services with docker-compose
docker-compose-up:
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	@docker-compose up -d

## docker-compose-down: Stop services with docker-compose
docker-compose-down:
	@echo "$(GREEN)Stopping services with docker-compose...$(NC)"
	@docker-compose down

## setup: Setup development environment
setup: deps install-tools
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@cp .env.example .env
	@echo "$(YELLOW)Please update .env with your database credentials$(NC)"
	@echo "$(GREEN)Setup complete!$(NC)"

## release: Build release version
release: clean test
	@echo "$(GREEN)Building release version...$(NC)"
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.version=$(shell git describe --tags --always)" -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s -X main.version=$(shell git describe --tags --always)" -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s -X main.version=$(shell git describe --tags --always)" -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "$(GREEN)Release binaries built in bin/$(NC)"

## check: Run quality checks
check: format lint test
	@echo "$(GREEN)All quality checks passed!$(NC)"

## dev: Start development server with hot reload
dev:
	@echo "$(GREEN)Starting development server with hot reload...$(NC)"
	@which air >/dev/null || (echo "$(RED)Air not found. Install with: go install github.com/cosmtrek/air@latest$(NC)" && exit 1)
	@air

## install-air: Install air for hot reload
install-air:
	@echo "$(GREEN)Installing air for hot reload...$(NC)"
	@go install github.com/cosmtrek/air@latest
	@echo "$(GREEN)Air installed$(NC)"
