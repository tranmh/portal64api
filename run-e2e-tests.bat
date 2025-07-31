@echo off
echo ============================================
echo  Portal64 API - E2E Tests with Playwright
echo ============================================
echo.

REM Check if Node.js is installed
node --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    echo.
    pause
    exit /b 1
)

echo Node.js version:
node --version
echo.

REM Check if dependencies are installed
if not exist "node_modules\@playwright\test" (
    echo Installing dependencies...
    npm install
    if errorlevel 1 (
        echo ERROR: Failed to install dependencies
        pause
        exit /b 1
    )
)

REM Install Playwright browsers if needed
if not exist "node_modules\playwright" (
    echo Installing Playwright browsers...
    npx playwright install
    if errorlevel 1 (
        echo ERROR: Failed to install Playwright browsers
        pause
        exit /b 1
    )
)

echo.
echo Starting E2E Tests...
echo.

REM Run Playwright tests
npx playwright test

if errorlevel 1 (
    echo.
    echo FAILED: Some E2E tests failed
    echo Run "npm run test:e2e:report" to view detailed results
) else (
    echo.
    echo SUCCESS: All E2E tests passed!
)

echo.
echo To view the test report, run: npm run test:e2e:report
echo To run tests with UI mode, run: npm run test:e2e:ui
echo To debug tests, run: npm run test:e2e:debug
echo.
pause
