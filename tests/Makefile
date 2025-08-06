# Makefile for Portal64API Import Feature Tests
# Provides standardized test execution across environments

.PHONY: test test-unit test-integration test-e2e test-all test-benchmark test-coverage
.PHONY: setup-test-env cleanup-test-env help
.PHONY: build-for-test start-test-service stop-test-service

# Variables
GO_VERSION := 1.19
TEST_TIMEOUT := 30m
COVERAGE_FILE := test-coverage/coverage.out
COVERAGE_HTML := test-coverage/coverage.html
TEST_RESULTS_DIR := test-results
MYSQL_HOST := localhost
MYSQL_PORT := 3306
MYSQL_USER := root
MYSQL_PASSWORD := 
API_BASE_URL := http://localhost:8080

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

# Default target
help: ## Show this help message
	@echo "Portal64API Import Feature Test Suite"
	@echo "====================================="
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Setup and cleanup
setup-test-env: ## Set up test environment
	@echo "$(YELLOW)Setting up test environment...$(RESET)"
	@mkdir -p $(TEST_RESULTS_DIR)
	@mkdir -p test-coverage
	@mkdir -p data/import/temp
	@mkdir -p data/import/logs
	@echo "$(GREEN)Test environment ready$(RESET)"

cleanup-test-env: ## Clean up test environment
	@echo "$(YELLOW)Cleaning up test environment...$(RESET)"
	@rm -rf $(TEST_RESULTS_DIR)
	@rm -rf test-coverage
	@rm -rf data/import/temp/*
	@echo "$(GREEN)Test environment cleaned$(RESET)"

# Build targets
build-for-test: ## Build application for testing
	@echo "$(YELLOW)Building Portal64API for testing...$(RESET)"
	@go build -o bin/portal64api.exe ./cmd/server
	@echo "$(GREEN)Build completed$(RESET)"

# Service management for E2E tests
start-test-service: build-for-test ## Start Portal64API service for testing
	@echo "$(YELLOW)Starting Portal64API test service...$(RESET)"
	@if pgrep -f "portal64api.exe" > /dev/null; then \
		echo "$(YELLOW)Service already running$(RESET)"; \
	else \
		nohup ./bin/portal64api.exe > test-service.log 2>&1 & \
		sleep 5; \
		if curl -s $(API_BASE_URL)/health > /dev/null; then \
			echo "$(GREEN)Service started successfully$(RESET)"; \
		else \
			echo "$(RED)Failed to start service$(RESET)"; \
			exit 1; \
		fi \
	fi

stop-test-service: ## Stop Portal64API test service
	@echo "$(YELLOW)Stopping Portal64API test service...$(RESET)"
	@pkill -f "portal64api.exe" || true
	@sleep 2
	@echo "$(GREEN)Service stopped$(RESET)"

# Test targets
test-unit: setup-test-env ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(RESET)"
	@go test -v -timeout=10m ./tests/unit/... | tee $(TEST_RESULTS_DIR)/unit-tests.log
	@if [ $$? -eq 0 ]; then \
		echo "$(GREEN)✅ Unit tests passed$(RESET)"; \
	else \
		echo "$(RED)❌ Unit tests failed$(RESET)"; \
		exit 1; \
	fi

test-integration: setup-test-env ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(RESET)"
	@echo "$(YELLOW)Checking MySQL connection...$(RESET)"
	@mysql -h$(MYSQL_HOST) -P$(MYSQL_PORT) -u$(MYSQL_USER) -e "SELECT 1;" || (echo "$(RED)MySQL connection failed$(RESET)" && exit 1)
	@echo "$(GREEN)MySQL connection OK$(RESET)"
	@go test -v -timeout=15m ./tests/integration/... | tee $(TEST_RESULTS_DIR)/integration-tests.log
	@if [ $$? -eq 0 ]; then \
		echo "$(GREEN)✅ Integration tests passed$(RESET)"; \
	else \
		echo "$(RED)❌ Integration tests failed$(RESET)"; \
		exit 1; \
	fi

test-e2e: start-test-service ## Run end-to-end tests
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@export RUN_E2E_TESTS=true && \
	 export API_BASE_URL=$(API_BASE_URL) && \
	 go test -v -timeout=20m ./tests/e2e/... | tee $(TEST_RESULTS_DIR)/e2e-tests.log
	@if [ $$? -eq 0 ]; then \
		echo "$(GREEN)✅ E2E tests passed$(RESET)"; \
	else \
		echo "$(RED)❌ E2E tests failed$(RESET)"; \
		$(MAKE) stop-test-service; \
		exit 1; \
	fi
	@$(MAKE) stop-test-service

test-benchmark: setup-test-env ## Run performance benchmarks
	@echo "$(BLUE)Running performance benchmarks...$(RESET)"
	@go test -bench=. -benchmem -timeout=$(TEST_TIMEOUT) ./tests/... | tee $(TEST_RESULTS_DIR)/benchmark-results.log
	@echo "$(GREEN)✅ Benchmarks completed$(RESET)"
	@echo ""
	@echo "$(YELLOW)Top benchmark results:$(RESET)"
	@grep "Benchmark" $(TEST_RESULTS_DIR)/benchmark-results.log | head -10

test-coverage: setup-test-env ## Run tests with coverage analysis
	@echo "$(BLUE)Running tests with coverage analysis...$(RESET)"
	@go test -coverprofile=$(COVERAGE_FILE) -covermode=atomic -timeout=$(TEST_TIMEOUT) ./tests/...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@go tool cover -func=$(COVERAGE_FILE) > test-coverage/coverage-summary.txt
	@echo "$(GREEN)✅ Coverage analysis completed$(RESET)"
	@echo ""
	@echo "$(YELLOW)Coverage Summary:$(RESET)"
	@grep "total:" test-coverage/coverage-summary.txt
	@echo ""
	@echo "$(BLUE)Coverage report: $(COVERAGE_HTML)$(RESET)"

test-all: setup-test-env test-unit test-integration test-e2e ## Run all test suites
	@echo "$(GREEN)🎉 All test suites completed successfully!$(RESET)"

test: test-unit ## Default test target (unit tests)

# Stress testing targets
test-stress: start-test-service ## Run stress tests
	@echo "$(BLUE)Running stress tests...$(RESET)"
	@export RUN_E2E_STRESS_TESTS=true && \
	 export API_BASE_URL=$(API_BASE_URL) && \
	 go test -v -timeout=30m ./tests/e2e/... -run "Stress" | tee $(TEST_RESULTS_DIR)/stress-tests.log
	@$(MAKE) stop-test-service
	@echo "$(GREEN)✅ Stress tests completed$(RESET)"

# Configuration testing
test-config: start-test-service ## Run configuration tests
	@echo "$(BLUE)Running configuration tests...$(RESET)"
	@export RUN_E2E_CONFIG_TESTS=true && \
	 export API_BASE_URL=$(API_BASE_URL) && \
	 go test -v -timeout=15m ./tests/e2e/... -run "Configuration" | tee $(TEST_RESULTS_DIR)/config-tests.log
	@$(MAKE) stop-test-service
	@echo "$(GREEN)✅ Configuration tests completed$(RESET)"

# Recovery testing
test-recovery: start-test-service ## Run recovery tests
	@echo "$(BLUE)Running recovery tests...$(RESET)"
	@export RUN_E2E_RECOVERY_TESTS=true && \
	 export API_BASE_URL=$(API_BASE_URL) && \
	 go test -v -timeout=20m ./tests/e2e/... -run "Recovery" | tee $(TEST_RESULTS_DIR)/recovery-tests.log
	@$(MAKE) stop-test-service
	@echo "$(GREEN)✅ Recovery tests completed$(RESET)"

# Documentation testing
test-docs: start-test-service ## Run documentation compliance tests
	@echo "$(BLUE)Running documentation compliance tests...$(RESET)"
	@export RUN_E2E_DOC_TESTS=true && \
	 export API_BASE_URL=$(API_BASE_URL) && \
	 go test -v -timeout=10m ./tests/e2e/... -run "Documentation" | tee $(TEST_RESULTS_DIR)/docs-tests.log
	@$(MAKE) stop-test-service
	@echo "$(GREEN)✅ Documentation tests completed$(RESET)"

# Race condition testing
test-race: setup-test-env ## Run tests with race detection
	@echo "$(BLUE)Running tests with race detection...$(RESET)"
	@go test -race -timeout=20m ./tests/unit/... ./tests/integration/... | tee $(TEST_RESULTS_DIR)/race-tests.log
	@if [ $$? -eq 0 ]; then \
		echo "$(GREEN)✅ Race tests passed$(RESET)"; \
	else \
		echo "$(RED)❌ Race conditions detected$(RESET)"; \
		exit 1; \
	fi

# Memory leak testing
test-memory: setup-test-env ## Run tests with memory profiling
	@echo "$(BLUE)Running tests with memory profiling...$(RESET)"
	@go test -memprofile=test-coverage/mem.prof -timeout=15m ./tests/unit/... ./tests/integration/...
	@go tool pprof -top test-coverage/mem.prof > $(TEST_RESULTS_DIR)/memory-profile.txt
	@echo "$(GREEN)✅ Memory profiling completed$(RESET)"
	@echo "$(BLUE)Memory profile: $(TEST_RESULTS_DIR)/memory-profile.txt$(RESET)"

# CPU profiling
test-cpu: setup-test-env ## Run tests with CPU profiling
	@echo "$(BLUE)Running tests with CPU profiling...$(RESET)"
	@go test -cpuprofile=test-coverage/cpu.prof -timeout=15m ./tests/unit/... ./tests/integration/...
	@go tool pprof -top test-coverage/cpu.prof > $(TEST_RESULTS_DIR)/cpu-profile.txt
	@echo "$(GREEN)✅ CPU profiling completed$(RESET)"
	@echo "$(BLUE)CPU profile: $(TEST_RESULTS_DIR)/cpu-profile.txt$(RESET)"

# Continuous Integration targets
ci-test: ## Run CI test suite (optimized for CI/CD)
	@echo "$(BLUE)Running CI test suite...$(RESET)"
	@$(MAKE) setup-test-env
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@echo "$(GREEN)✅ CI test suite completed$(RESET)"

# Development targets
dev-test: ## Run tests in development mode (with verbose output)
	@echo "$(BLUE)Running development tests...$(RESET)"
	@$(MAKE) setup-test-env
	@go test -v -count=1 ./tests/unit/... ./tests/integration/...

quick-test: ## Run a quick test subset for development
	@echo "$(BLUE)Running quick tests...$(RESET)"
	@go test -short -timeout=5m ./tests/unit/importers/ ./tests/unit/services/

# Utility targets
fmt: ## Format Go code
	@echo "$(YELLOW)Formatting Go code...$(RESET)"
	@go fmt ./tests/...
	@echo "$(GREEN)Code formatted$(RESET)"

lint: ## Run linter on test code
	@echo "$(YELLOW)Running linter on test code...$(RESET)"
	@golangci-lint run ./tests/...
	@echo "$(GREEN)Linting completed$(RESET)"

vet: ## Run go vet on test code
	@echo "$(YELLOW)Running go vet on test code...$(RESET)"
	@go vet ./tests/...
	@echo "$(GREEN)Vet completed$(RESET)"

# Reporting
test-report: ## Generate comprehensive test report
	@echo "$(BLUE)Generating test report...$(RESET)"
	@mkdir -p test-reports
	@echo "# Portal64API Import Feature Test Report" > test-reports/test-report.md
	@echo "Generated: $$(date)" >> test-reports/test-report.md
	@echo "" >> test-reports/test-report.md
	@if [ -f $(TEST_RESULTS_DIR)/unit-tests.log ]; then \
		echo "## Unit Tests" >> test-reports/test-report.md; \
		grep -E "(PASS|FAIL|ok|FAIL)" $(TEST_RESULTS_DIR)/unit-tests.log >> test-reports/test-report.md; \
		echo "" >> test-reports/test-report.md; \
	fi
	@if [ -f $(TEST_RESULTS_DIR)/integration-tests.log ]; then \
		echo "## Integration Tests" >> test-reports/test-report.md; \
		grep -E "(PASS|FAIL|ok|FAIL)" $(TEST_RESULTS_DIR)/integration-tests.log >> test-reports/test-report.md; \
		echo "" >> test-reports/test-report.md; \
	fi
	@if [ -f test-coverage/coverage-summary.txt ]; then \
		echo "## Coverage Summary" >> test-reports/test-report.md; \
		cat test-coverage/coverage-summary.txt >> test-reports/test-report.md; \
	fi
	@echo "$(GREEN)✅ Test report generated: test-reports/test-report.md$(RESET)"

# Docker-based testing (if Docker is available)
docker-test: ## Run tests in Docker container
	@echo "$(BLUE)Running tests in Docker container...$(RESET)"
	@docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	@docker-compose -f docker-compose.test.yml down

# Clean all
clean: cleanup-test-env stop-test-service ## Clean all test artifacts and stop services
	@rm -f test-service.log
	@rm -rf test-reports
	@echo "$(GREEN)All clean!$(RESET)"
