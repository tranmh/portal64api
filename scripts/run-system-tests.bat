@echo off
REM Simple Windows batch script to run system tests

echo Portal64 API System Tests
echo =========================
echo.

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed or not in PATH
    exit /b 1
)

REM Check server health
echo Checking server health at http://test.svw.info:8080...
curl -f -s "http://test.svw.info:8080/health" >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: Cannot reach server at http://test.svw.info:8080
    echo.
    echo Please ensure:
    echo   1. The server is running at http://test.svw.info:8080
    echo   2. Your network connection is working
    echo   3. No firewall is blocking the connection
    echo.
    echo You can also run the tests manually with:
    echo   go test ./tests/integration/ -run TestSystemSuite -v -timeout=60s
    exit /b 1
)

echo Server is healthy!
echo.

REM Set environment variable for tests
set PORTAL64_TEST_BASE_URL=http://test.svw.info:8080

REM Run the tests
echo Running system tests...
echo Command: go test ./tests/integration/ -run TestSystemSuite -v -timeout=60s
echo.

go test ./tests/integration/ -run TestSystemSuite -v -timeout=60s
set EXIT_CODE=%ERRORLEVEL%

echo.
if %EXIT_CODE% equ 0 (
    echo SUCCESS: All system tests passed!
) else (
    echo FAILED: Some system tests failed (exit code: %EXIT_CODE%)
    echo.
    echo Troubleshooting tips:
    echo   1. Check server logs for errors
    echo   2. Verify test data exists in the database  
    echo   3. Check network connectivity to the test server
    echo   4. Review the test output above for specific failures
)

echo.
echo Test run completed.
exit /b %EXIT_CODE%
