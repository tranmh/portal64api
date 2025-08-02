# System Tests for Portal64 API

This directory contains comprehensive automated system tests for the Portal64 API based on the swagger.json definitions.

## Test Coverage

The system tests cover all endpoints defined in the swagger specification:

### Clubs Endpoints
- `/api/v1/clubs` - Search clubs with various parameters
- `/api/v1/clubs/all` - Get all clubs
- `/api/v1/clubs/{id}` - Get specific club by ID
- `/api/v1/clubs/{id}/players` - Get players by club ID

### Players Endpoints  
- `/api/v1/players` - Search players
- `/api/v1/players/{id}` - Get specific player by ID
- `/api/v1/players/{id}/rating-history` - Get player rating history

### Tournaments Endpoints
- `/api/v1/tournaments` - Search tournaments
- `/api/v1/tournaments/date-range` - Get tournaments by date range
- `/api/v1/tournaments/recent` - Get recent tournaments
- `/api/v1/tournaments/{id}` - Get specific tournament by ID

## Test Types

### Functional Tests
- ✅ All API endpoints with valid parameters
- ✅ JSON and CSV response formats
- ✅ Pagination and sorting
- ✅ Filtering and search queries
- ✅ Parameter validation

### Error Handling Tests
- ✅ Invalid parameter values
- ✅ Malformed IDs
- ✅ Non-existent endpoints
- ✅ Unsupported HTTP methods
- ✅ Boundary value testing

### Non-Functional Tests
- ✅ Response time limits (< 5 seconds)
- ✅ Concurrent request handling
- ✅ CORS headers validation
- ✅ Swagger documentation accessibility

## Running the Tests

### Prerequisites
1. Ensure the test server is running at `http://test.svw.info:8080`
2. The server should be accessible and respond to health checks

### Run System Tests
```bash
# Run only system tests
go test ./tests/integration -run TestSystemSuite -v

# Run with verbose output and race detection
go test ./tests/integration -run TestSystemSuite -v -race

# Run with timeout
go test ./tests/integration -run TestSystemSuite -v -timeout=300s

# Run specific test groups
go test ./tests/integration -run TestSystemSuite/TestClubsSearch -v
go test ./tests/integration -run TestSystemSuite/TestPlayersSearch -v
go test ./tests/integration -run TestSystemSuite/TestTournamentsSearch -v
```

### Using Make
Add this target to your Makefile:
```makefile
.PHONY: test-system
test-system:
	@echo "Running system tests against http://test.svw.info:8080"
	go test ./tests/integration -run TestSystemSuite -v -timeout=300s
```

Then run:
```bash
make test-system
```

## Test Configuration

The tests are configured to:
- **Base URL**: `http://test.svw.info:8080`
- **Timeout**: 30 seconds per request
- **Max Response Time**: 5 seconds (performance test)
- **Concurrent Requests**: 10 simultaneous requests (load test)

## Test Results Interpretation

### Success Scenarios
- **200 OK**: Endpoint returned data successfully
- **404 Not Found**: Valid request format but resource doesn't exist (acceptable)
- **CSV Format**: Proper Content-Type and Content-Disposition headers

### Failure Scenarios  
- **400 Bad Request**: Invalid parameter format or values
- **405 Method Not Allowed**: Unsupported HTTP method
- **500 Internal Server Error**: Server-side error (indicates bug)
- **Timeout**: Response took longer than 30 seconds

### Test Skipping
Tests will be skipped if:
- Server is not reachable at the configured URL
- Health check endpoint returns non-200 status

## Adding New Tests

To add tests for new endpoints:

1. **Add endpoint definition** to the appropriate test section
2. **Create test cases** covering positive and negative scenarios  
3. **Include format tests** for both JSON and CSV if supported
4. **Add boundary tests** for parameters with limits
5. **Update documentation** with the new endpoint coverage

Example test case structure:
```go
func (suite *SystemTestSuite) TestNewEndpoint() {
    testCases := []struct {
        name   string
        params map[string]string
        status int
    }{
        // Add test cases here
    }
    
    for _, tc := range testCases {
        suite.Run(tc.name, func() {
            // Test implementation
        })
    }
}
```

## Debugging Failed Tests

1. **Check server status**: Verify the test server is running and accessible
2. **Review logs**: Check application logs for error details
3. **Validate data**: Ensure test data exists in the database
4. **Network issues**: Check for connectivity problems
5. **Rate limiting**: Verify the server isn't rate limiting requests

## CI/CD Integration

For continuous integration, add this to your pipeline:

```yaml
- name: Run System Tests
  run: |
    # Wait for test server to be ready
    timeout 300 bash -c 'until curl -f http://test.svw.info:8080/health; do sleep 5; done'
    
    # Run system tests
    go test ./tests/integration -run TestSystemSuite -v -timeout=300s
```

The tests are designed to be robust and will automatically skip if the test environment is not available, making them safe for CI/CD pipelines.
