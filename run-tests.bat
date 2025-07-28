@echo off
echo Portal64 API Frontend Testing
echo ================================

REM Check if Node.js is installed
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    pause
    exit /b 1
)

REM Check if package.json exists
if not exist package.json (
    echo ERROR: package.json not found
    echo Make sure you're in the Portal64API root directory
    pause
    exit /b 1
)

REM Check if node_modules exists, install if not
if not exist node_modules (
    echo Installing dependencies...
    npm install
    if %errorlevel% neq 0 (
        echo ERROR: Failed to install dependencies
        pause
        exit /b 1
    )
)

REM Run the tests
echo.
echo Running frontend tests...
echo.

REM Menu for different test options
echo Select test option:
echo 1. Run all tests once
echo 2. Run tests with coverage report
echo 3. Run tests in watch mode
echo 4. Run specific test file
echo 5. Exit

set /p choice="Enter your choice (1-5): "

if "%choice%"=="1" (
    npm test
) else if "%choice%"=="2" (
    npm run test:coverage
    echo.
    echo Coverage report generated in: coverage\lcov-report\index.html
) else if "%choice%"=="3" (
    echo Press Ctrl+C to stop watching...
    npm run test:watch
) else if "%choice%"=="4" (
    echo Available test files:
    echo - tests\frontend\unit\api\api.test.js
    echo - tests\frontend\unit\pages\index.test.js
    echo.
    set /p testfile="Enter test file path: "
    npm test -- "%testfile%"
) else if "%choice%"=="5" (
    exit /b 0
) else (
    echo Invalid choice. Please run the script again.
)

echo.
echo Test execution completed.
pause
