@echo off
setlocal enabledelayedexpansion

set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=build

if "%COMMAND%"=="build" (
    echo Building Kader-Planung application...
    go mod tidy
    if !errorlevel! neq 0 (
        echo Error: Failed to tidy modules
        exit /b 1
    )
    
    go build -ldflags="-s -w" -o .\bin\kader-planung.exe .\cmd\kader-planung
    if !errorlevel! neq 0 (
        echo Error: Build failed
        exit /b 1
    )

    set GOARCH=amd64
    set GOOS=linux
    go build -ldflags="-s -w" -o .\bin\kader-planung .\cmd\kader-planung
    if !errorlevel! neq 0 (
        echo Error: Build failed
        exit /b 1
    )
    
    echo Build successful: .\bin\kader-planung.exe
    echo Build successful: .\bin\kader-planung
    goto :eof
)

if "%COMMAND%"=="clean" (
    echo Cleaning build artifacts...
    if exist .\bin rmdir /s /q .\bin    if exist *.exe del *.exe
    if exist *.log del *.log
    if exist *checkpoint*.json del *checkpoint*.json
    echo Clean complete
    goto :eof
)

if "%COMMAND%"=="test" (
    echo Running tests...
    go test -v ./...
    goto :eof
)

if "%COMMAND%"=="run" (
    echo Running application...
    shift
    .\bin\kader-planung.exe %*
    goto :eof
)

if "%COMMAND%"=="deps" (
    echo Installing dependencies...
    go mod download
    go mod verify
    echo Dependencies installed
    goto :eof
)

echo Usage: build.bat [command]
echo Commands:
echo   build   - Build the application (default)
echo   clean   - Clean build artifacts
echo   test    - Run tests
echo   run     - Run the built application
echo   deps    - Install dependencies
exit /b 1