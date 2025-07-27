@echo off
REM Quick Start Script for Portal64 API on Windows
REM This script helps new users get started quickly

echo.
echo =====================================
echo  Portal64 API - Windows Quick Start
echo =====================================
echo.

REM Check if we're in the right directory
if not exist "go.mod" (
    echo ERROR: go.mod not found!
    echo Please run this script from the Portal64 API project root directory.
    echo.
    pause
    exit /b 1
)

echo Step 1: Checking PowerShell availability...
powershell -Command "Get-Host" >nul 2>&1
if !errorlevel! neq 0 (
    echo ERROR: PowerShell not available!
    echo Please install PowerShell or add it to your PATH.
    pause
    exit /b 1
)
echo   ✓ PowerShell is available

echo.
echo Step 2: Setting up development environment...
echo This will install Go tools and setup the project.
echo.
echo Would you like to run the automated setup? (Y/N)
set /p SETUP_CHOICE="> "

if /i "%SETUP_CHOICE%"=="Y" (
    echo Running automated setup...
    powershell -ExecutionPolicy Bypass -File "setup-windows.ps1" -SetupProject
    if !errorlevel! neq 0 (
        echo Setup failed. Please check the error messages above.
        pause
        exit /b 1
    )
    echo   ✓ Setup completed successfully
) else (
    echo Skipping automated setup.
    echo You can run it later with: .\setup-windows.ps1 -SetupProject
)

echo.
echo Step 3: Database Configuration
echo.
if exist ".env" (
    echo   ✓ .env file already exists
) else (
    if exist ".env.example" (
        copy ".env.example" ".env" >nul
        echo   ✓ Created .env file from .env.example
    ) else (
        echo   ! .env.example not found, creating minimal .env file
        echo SERVER_PORT=8080 > .env
        echo ENVIRONMENT=development >> .env
    )
)

echo.
echo IMPORTANT: Please edit the .env file with your database credentials:
echo   - MVDSB_HOST, MVDSB_USERNAME, MVDSB_PASSWORD, etc.
echo.
echo Would you like to open the .env file now? (Y/N)
set /p EDIT_CHOICE="> "

if /i "%EDIT_CHOICE%"=="Y" (
    if exist "%EDITOR%" (
        "%EDITOR%" .env
    ) else if exist "%PROGRAMFILES%\Notepad++\notepad++.exe" (
        "%PROGRAMFILES%\Notepad++\notepad++.exe" .env
    ) else (
        notepad .env
    )
)

echo.
echo Step 4: Testing the setup
echo.
echo Testing if Go is available...
go version >nul 2>&1
if !errorlevel! neq 0 (
    echo   ! Go not found. Please install Go from: https://golang.org/dl/
    echo     Then restart this script.
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
echo   ✓ %GO_VERSION%

echo.
echo Testing build system...
powershell -ExecutionPolicy Bypass -File "build.ps1" deps >nul 2>&1
if !errorlevel! neq 0 (
    echo   ! Failed to download dependencies
    echo     Please check your internet connection and Go installation
    pause
    exit /b 1
)
echo   ✓ Dependencies downloaded successfully

echo.
echo =====================================
echo  Setup Complete!
echo =====================================
echo.
echo Next steps:
echo   1. Edit .env file with your database credentials
echo   2. Start the API: .\build.ps1 run
echo   3. Open: http://localhost:8080/swagger/index.html
echo.
echo Available commands:
echo   .\build.ps1 help     - Show all available commands
echo   .\build.ps1 run      - Start the development server
echo   .\build.ps1 test     - Run tests
echo   .\build.ps1 build    - Build the application
echo.
echo For detailed documentation, see:
echo   - README.md (general documentation)
echo   - README-WINDOWS.md (Windows-specific guide)
echo.
echo Would you like to start the API now? (Y/N)
set /p START_CHOICE="> "

if /i "%START_CHOICE%"=="Y" (
    echo.
    echo Starting Portal64 API...
    echo Press Ctrl+C to stop the server
    echo.
    powershell -ExecutionPolicy Bypass -File "build.ps1" run
) else (
    echo.
    echo To start the API later, run: .\build.ps1 run
    echo.
    pause
)
