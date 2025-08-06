@echo off
REM Test runner script for Portal64API SCP Import Feature tests
REM Usage: run-import-tests.bat [unit|integration|e2e|all]

setlocal EnableDelayedExpansion

echo ====================================
echo Portal64API Import Feature Test Runner
echo ====================================

REM Default to unit tests if no argument provided
set TEST_TYPE=%1
if "%TEST_TYPE%"=="" set TEST_TYPE=unit

REM Set environment variables
set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=1

REM Test configuration
set MYSQL_HOST=localhost
set MYSQL_PORT=3306
set MYSQL_USER=root
set MYSQL_PASSWORD=
set REDIS_HOST=localhost
set REDIS_PORT=6379

REM Create test output directories
if not exist "test-results" mkdir test-results
if not exist "test-coverage" mkdir test-coverage

echo Test Type: %TEST_TYPE%
echo.

if "%TEST_TYPE%"=="unit" goto run_unit
if "%TEST_TYPE%"=="integration" goto run_integration  
if "%TEST_TYPE%"=="e2e" goto run_e2e
if "%TEST_TYPE%"=="all" goto run_all
if "%TEST_TYPE%"=="benchmark" goto run_benchmark
if "%TEST_TYPE%"=="coverage" goto run_coverage

echo Invalid test type: %TEST_TYPE%
echo Valid options: unit, integration, e2e, all, benchmark, coverage
exit /b 1

:run_unit
echo Running Unit Tests...
echo.
go test -v -timeout=10m ./tests/unit/... > test-results\unit-tests.log 2>&1
if %ERRORLEVEL% neq 0 (
    echo ❌ Unit tests failed! Check test-results\unit-tests.log for details.
    type test-results\unit-tests.log | findstr "FAIL\|Error\|panic"
    exit /b 1
) else (
    echo ✅ All unit tests passed!
)
goto end

:run_integration
echo Running Integration Tests...
echo.
REM Check if MySQL is available
echo Testing MySQL connection...
c:\xampp\mysql\bin\mysql -h%MYSQL_HOST% -P%MYSQL_PORT% -u%MYSQL_USER% -e "SELECT 1;" 2>nul
if %ERRORLEVEL% neq 0 (
    echo ❌ MySQL connection failed! Ensure MySQL is running on %MYSQL_HOST%:%MYSQL_PORT%
    exit /b 1
)
echo ✅ MySQL connection successful

go test -v -timeout=15m ./tests/integration/... > test-results\integration-tests.log 2>&1
if %ERRORLEVEL% neq 0 (
    echo ❌ Integration tests failed! Check test-results\integration-tests.log for details.
    type test-results\integration-tests.log | findstr "FAIL\|Error\|panic"
    exit /b 1
) else (
    echo ✅ All integration tests passed!
)
goto end

:run_e2e
echo Running End-to-End Tests...
echo.

REM Check if service is running
echo Checking if Portal64API service is running...
curl -s http://localhost:8080/health >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Portal64API service not detected. Starting service...
    start /B "" ".\bin\portal64api.exe"
    echo Waiting for service to start...
    timeout /t 10 /nobreak >nul
    
    REM Check again
    curl -s http://localhost:8080/health >nul 2>&1
    if !ERRORLEVEL! neq 0 (
        echo ❌ Failed to start Portal64API service
        exit /b 1
    )
)
echo ✅ Portal64API service is running

set RUN_E2E_TESTS=true
set API_BASE_URL=http://localhost:8080
go test -v -timeout=20m ./tests/e2e/... > test-results\e2e-tests.log 2>&1
if %ERRORLEVEL% neq 0 (
    echo ❌ E2E tests failed! Check test-results\e2e-tests.log for details.
    type test-results\e2e-tests.log | findstr "FAIL\|Error\|panic"
    exit /b 1
) else (
    echo ✅ All E2E tests passed!
)
goto end

:run_all
echo Running All Tests...
echo.

call :run_unit
if %ERRORLEVEL% neq 0 exit /b 1

call :run_integration  
if %ERRORLEVEL% neq 0 exit /b 1

call :run_e2e
if %ERRORLEVEL% neq 0 exit /b 1

echo.
echo ✅ All test suites passed successfully!
goto end

:run_benchmark
echo Running Performance Benchmarks...
echo.
go test -bench=. -benchmem -timeout=30m ./tests/... > test-results\benchmark-results.log 2>&1
if %ERRORLEVEL% neq 0 (
    echo ❌ Benchmarks failed! Check test-results\benchmark-results.log for details.
    exit /b 1
) else (
    echo ✅ Benchmarks completed successfully!
    echo.
    echo Top benchmark results:
    type test-results\benchmark-results.log | findstr "Benchmark"
)
goto end

:run_coverage
echo Running Tests with Coverage Analysis...
echo.

REM Run all tests with coverage
go test -coverprofile=test-coverage\coverage.out -covermode=atomic -timeout=25m ./tests/...
if %ERRORLEVEL% neq 0 (
    echo ❌ Coverage test run failed!
    exit /b 1
)

REM Generate coverage report
go tool cover -html=test-coverage\coverage.out -o test-coverage\coverage.html
if %ERRORLEVEL% neq 0 (
    echo ❌ Failed to generate coverage report!
    exit /b 1
)

REM Show coverage summary
go tool cover -func=test-coverage\coverage.out > test-coverage\coverage-summary.txt
echo Coverage Summary:
type test-coverage\coverage-summary.txt | findstr "total:"

echo.
echo ✅ Coverage analysis completed!
echo Report generated: test-coverage\coverage.html
goto end

:end
echo.
echo ====================================
echo Test run completed: %DATE% %TIME%
echo ====================================
