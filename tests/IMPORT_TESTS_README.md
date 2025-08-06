# SCP Import Feature - Test Documentation

This document describes the comprehensive test suite for the SCP Import Feature, covering unit tests, integration tests, and end-to-end tests.

## Overview

The SCP Import Feature test suite provides complete coverage of the import functionality, including:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions and workflows
- **End-to-End Tests**: Test complete functionality against running services
- **Handler Tests**: Test HTTP API endpoints
- **Performance Tests**: Test under load and stress conditions

## Test Structure

```
tests/
├── unit/
│   ├── importers/           # Tests for import components
│   │   ├── scp_downloader_test.go
│   │   ├── zip_extractor_test.go
│   │   ├── database_importer_test.go
│   │   ├── freshness_checker_test.go
│   │   └── status_tracker_test.go
│   ├── services/            # Tests for import service
│   │   └── import_service_test.go
│   └── handlers/            # Tests for API handlers
│       └── import_handler_test.go
├── integration/             # Integration tests
│   └── import_integration_test.go
├── e2e/                    # End-to-end tests
│   └── import_e2e_test.go
├── testconfig/             # Test configurations
│   └── import_test_config.go
└── testutils/              # Test utilities
    └── import_test_utils.go
```

## Test Categories

### Unit Tests

#### SCP Downloader Tests (`tests/unit/importers/scp_downloader_test.go`)

Tests the SCP file download functionality:

- **Configuration Validation**: Tests valid and invalid SCP configurations
- **File Pattern Matching**: Tests wildcard pattern matching for file selection
- **Connection Handling**: Tests SCP connection establishment and error handling
- **File Metadata**: Tests file metadata extraction and validation
- **Progress Reporting**: Tests download progress reporting

**Key Test Cases:**
- `TestSCPDownloader_ListFiles`: File listing with pattern matching
- `TestSCPDownloader_Configuration`: Configuration validation
- `TestSCPDownloader_FilePatternMatching`: Pattern matching logic

#### ZIP Extractor Tests (`tests/unit/importers/zip_extractor_test.go`)

Tests the ZIP file extraction functionality:

- **Password Protection**: Tests extraction with correct and incorrect passwords
- **Content Validation**: Tests validation of extracted SQL files
- **Progress Reporting**: Tests extraction progress tracking
- **File Type Detection**: Tests identification of database files

**Key Test Cases:**
- `TestZIPExtractor_ExtractPasswordProtectedZip`: Password-protected extraction
- `TestZIPExtractor_ValidateExtractedContent`: Content validation
- `TestZIPExtractor_Configuration`: Configuration validation

#### Database Importer Tests (`tests/unit/importers/database_importer_test.go`)

Tests the database import functionality:

- **Database Operations**: Tests drop, create, and import operations
- **File Mapping**: Tests mapping of SQL files to target databases
- **Error Handling**: Tests handling of database connection and import errors
- **Progress Tracking**: Tests import progress reporting

**Key Test Cases:**
- `TestDatabaseImporter_Configuration`: Configuration validation
- `TestDatabaseImporter_FileToDatabase Mapping`: File-to-database mapping
- `TestDatabaseImporter_ImportValidation`: SQL content validation

#### Freshness Checker Tests (`tests/unit/importers/freshness_checker_test.go`)

Tests the file freshness comparison functionality:

- **First Import**: Tests behavior when no previous import exists
- **File Comparison**: Tests timestamp, size, and checksum comparisons
- **Metadata Handling**: Tests metadata loading and saving
- **Skip Logic**: Tests import skipping when no newer files exist

**Key Test Cases:**
- `TestFreshnessChecker_CheckFreshness_FirstImport`: First import detection
- `TestFreshnessChecker_CheckFreshness_NewerFilesAvailable`: File freshness comparison
- `TestFreshnessChecker_ComparisonMethods`: Different comparison methods

#### Status Tracker Tests (`tests/unit/importers/status_tracker_test.go`)

Tests the import status tracking functionality:

- **Status Updates**: Tests status transitions and progress updates
- **Log Management**: Tests log entry creation and rotation
- **Thread Safety**: Tests concurrent access to status and logs
- **Memory Management**: Tests log entry limits and cleanup

**Key Test Cases:**
- `TestStatusTracker_UpdateStatus`: Status update functionality
- `TestStatusTracker_ThreadSafety`: Concurrent access handling
- `TestStatusTracker_LogRotation`: Log entry rotation

#### Import Service Tests (`tests/unit/services/import_service_test.go`)

Tests the main import service orchestration:

- **Service Lifecycle**: Tests service start, stop, and configuration
- **Manual Triggers**: Tests manual import triggering and concurrency
- **Scheduling**: Tests cron-based scheduling integration
- **Error Recovery**: Tests error handling and retry logic

**Key Test Cases:**
- `TestImportService_Start_Stop`: Service lifecycle management
- `TestImportService_TriggerManualImport`: Manual import triggering
- `TestImportService_ThreadSafety`: Concurrent access handling

#### Import Handler Tests (`tests/unit/handlers/import_handler_test.go`)

Tests the HTTP API endpoints:

- **Status Endpoint**: Tests `/api/v1/import/status` responses
- **Start Endpoint**: Tests `/api/v1/import/start` functionality
- **Logs Endpoint**: Tests `/api/v1/import/logs` responses
- **Error Handling**: Tests error responses and HTTP status codes

**Key Test Cases:**
- `TestImportHandler_GetStatus`: Status endpoint functionality
- `TestImportHandler_StartImport`: Manual import trigger endpoint
- `TestImportHandler_GetLogs`: Log retrieval endpoint

### Integration Tests

#### Import Integration Tests (`tests/integration/import_integration_test.go`)

Tests complete import workflows:

- **End-to-End Workflow**: Tests complete import process from start to finish
- **API Integration**: Tests HTTP API with running import service
- **Concurrency Testing**: Tests concurrent API access and import operations
- **Configuration Testing**: Tests various configuration scenarios

**Key Test Cases:**
- `TestImportWorkflow_Integration`: Complete import workflow testing
- `TestImportAPI_Integration`: HTTP API integration testing
- `TestImportConcurrency_Integration`: Concurrency and thread safety

### End-to-End Tests

#### E2E Tests (`tests/e2e/import_e2e_test.go`)

Tests against running services:

- **Manual Import Trigger**: Tests manual import against live service
- **Status Monitoring**: Tests status monitoring during import
- **Log Retrieval**: Tests log access during and after import
- **Stress Testing**: Tests system under load
- **Configuration Verification**: Tests configuration compliance

**Key Test Cases:**
- `TestImportEndToEnd`: Complete end-to-end import scenarios
- `TestImportStressTest`: Performance and stress testing
- `TestImportConfiguration`: Configuration verification
- `TestImportRecovery`: Error recovery testing

## Running Tests

### Prerequisites

1. **Go Environment**: Go 1.19 or later
2. **Database**: MySQL running on localhost:3306
3. **Dependencies**: All Go modules installed (`go mod download`)

### Unit Tests

Run all unit tests:
```bash
go test ./tests/unit/... -v
```

Run specific unit test categories:
```bash
# Import component tests
go test ./tests/unit/importers/... -v

# Service tests
go test ./tests/unit/services/... -v

# Handler tests
go test ./tests/unit/handlers/... -v
```

Run individual test files:
```bash
go test ./tests/unit/importers/scp_downloader_test.go -v
go test ./tests/unit/services/import_service_test.go -v
```

### Integration Tests

Run integration tests (requires database):
```bash
go test ./tests/integration/... -v
```

### End-to-End Tests

E2E tests require a running Portal64API service:

1. Start the service:
```bash
./bin/portal64api.exe
```

2. Run E2E tests:
```bash
# Enable E2E tests
export RUN_E2E_TESTS=true
export API_BASE_URL=http://localhost:8080

go test ./tests/e2e/... -v
```

3. Run stress tests (optional):
```bash
export RUN_E2E_STRESS_TESTS=true
go test ./tests/e2e/... -v -run "Stress"
```

### All Tests

Run all tests (requires running service):
```bash
export RUN_E2E_TESTS=true
export API_BASE_URL=http://localhost:8080
go test ./tests/... -v
```

## Test Configuration

### Environment Variables

Tests support various environment variables for configuration:

- `RUN_E2E_TESTS=true`: Enable end-to-end tests
- `RUN_E2E_STRESS_TESTS=true`: Enable stress tests  
- `RUN_E2E_CONFIG_TESTS=true`: Enable configuration tests
- `API_BASE_URL=http://localhost:8080`: Set API base URL for E2E tests
- `CI=true`: Indicates running in CI environment (some tests may be skipped)

### Test Configurations

The `tests/testconfig/` directory contains test-specific configurations:

- `GetTestImportConfig()`: Full test configuration with mock SCP settings
- `GetMinimalImportConfig()`: Minimal configuration for basic tests
- `GetTestDatabaseConfig()`: Database configuration for testing

### Test Utilities

The `tests/testutils/` directory provides utilities for test setup:

- `CreateTestZipFile()`: Create test ZIP files with SQL content
- `CreateTestMetadataFile()`: Create import metadata files
- `WaitForCondition()`: Wait for conditions with timeout
- `AssertImportStatus()`: Assert import status meets expectations

## Test Coverage

The test suite covers the following scenarios:

### Success Scenarios

- **First Import**: Successful import when no previous import exists
- **Regular Import**: Import with newer files available
- **Multiple Databases**: Import targeting multiple databases
- **Large Files**: Import with large ZIP files and SQL dumps

### Skip Scenarios

- **No Newer Files**: Import skipped when files haven't changed
- **Freshness Check Disabled**: Import proceeds when freshness check is disabled
- **Manual Skip**: Import skipped due to manual intervention

### Error Scenarios

- **Connection Failures**: SCP connection timeouts and authentication failures
- **Invalid Passwords**: ZIP extraction with wrong passwords
- **Database Errors**: Database connection and import failures
- **File Corruption**: Handling of corrupted ZIP files and SQL content
- **Disk Space**: Handling of insufficient disk space
- **Permission Errors**: File system permission issues

### Edge Cases

- **Empty Files**: Handling empty ZIP files and SQL files
- **Special Characters**: Files with Unicode and special characters
- **Concurrent Access**: Multiple simultaneous import attempts
- **Service Restart**: Import state recovery after service restart
- **Clock Changes**: Handling system clock changes during import

## Continuous Integration

### GitHub Actions Integration

The test suite integrates with CI/CD pipelines:

```yaml
- name: Run Unit Tests
  run: go test ./tests/unit/... -v

- name: Run Integration Tests  
  run: go test ./tests/integration/... -v
  env:
    MYSQL_HOST: localhost
    MYSQL_PORT: 3306

- name: Run E2E Tests
  run: |
    ./bin/portal64api.exe &
    sleep 10
    go test ./tests/e2e/... -v
  env:
    RUN_E2E_TESTS: true
    API_BASE_URL: http://localhost:8080
```

### Test Reports

Generate test coverage reports:

```bash
# Generate coverage report
go test ./tests/... -coverprofile=coverage.out -covermode=atomic

# View coverage in browser
go tool cover -html=coverage.out
```

## Mock Services

For testing without external dependencies:

### Mock SCP Server

Tests include mock SCP server functionality for isolated testing:

- Simulated file listings
- Configurable download scenarios
- Error injection capabilities

### Mock Database

Database tests use in-memory or containerized databases:

- Isolated test databases
- Transaction rollback for cleanup
- Schema validation

## Performance Benchmarks

Performance tests measure:

- **Download Speed**: SCP download performance
- **Extraction Speed**: ZIP extraction performance  
- **Import Speed**: Database import performance
- **Memory Usage**: Peak memory consumption
- **Concurrent Load**: Performance under concurrent requests

Run benchmarks:
```bash
go test -bench=. ./tests/... -benchmem
```

## Debugging Tests

### Verbose Output

Enable verbose logging in tests:
```bash
go test ./tests/... -v -args -test.v=true
```

### Test-Specific Logs

Tests create detailed logs in temporary directories:
- Import logs: Captured during test execution
- Status changes: Tracked throughout test lifecycle  
- Error details: Full error context preserved

### IDE Integration

Tests are compatible with Go IDE test runners:
- VS Code: Go extension with test discovery
- GoLand: Built-in test runner with debugging
- Vim: vim-go test integration

## Troubleshooting

### Common Issues

1. **Database Connection Failures**:
   - Ensure MySQL is running on localhost:3306
   - Check credentials in test configuration
   - Verify test database permissions

2. **SCP Connection Issues**:
   - Tests use localhost SCP for integration
   - Verify SSH service is available
   - Check test user credentials

3. **Port Conflicts**:
   - Ensure test ports are available
   - Stop other Portal64API instances
   - Check for port binding conflicts

4. **File Permission Issues**:
   - Ensure test directories are writable
   - Check temporary directory permissions
   - Verify test file creation rights

### Debug Commands

```bash
# Run single test with full output
go test -v -run TestSpecificTest ./tests/unit/...

# Run with race detection  
go test -race ./tests/...

# Run with timeout
go test -timeout 30m ./tests/e2e/...

# Run with CPU profiling
go test -cpuprofile cpu.prof ./tests/...
```

## Contributing

When adding new import functionality:

1. **Add Unit Tests**: Test new components in isolation
2. **Update Integration Tests**: Test component interactions  
3. **Add E2E Scenarios**: Test against running service
4. **Update Documentation**: Document new test cases
5. **Verify Coverage**: Ensure adequate test coverage

### Test Naming Conventions

- Test functions: `TestComponentName_Functionality`
- Test cases: Descriptive scenario names
- Mock objects: `Mock` prefix (e.g., `MockImportService`)
- Test helpers: `test` prefix (e.g., `testSetupDirectories`)

### Best Practices

- Use `testify/require` for assertions that should stop test execution
- Use `testify/assert` for assertions that should continue test execution
- Create isolated test environments with temporary directories
- Clean up resources in test cleanup functions
- Use table-driven tests for multiple scenarios
- Mock external dependencies in unit tests
- Use real services sparingly in integration tests

This comprehensive test suite ensures the SCP Import Feature is robust, reliable, and maintainable across different environments and usage scenarios.
