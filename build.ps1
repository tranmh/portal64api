# Portal64 API Build Script for Windows PowerShell
# Usage: .\build.ps1 <command>
# Example: .\build.ps1 help

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Variables
$BINARY_NAME = "portal64api"
$BINARY_PATH = "bin\$BINARY_NAME.exe"
$MAIN_PATH = "cmd\server\main.go"
$DOCKER_IMAGE = "portal64api"
$DOCKER_TAG = "latest"

# Colors for output
$GREEN = "Green"
$YELLOW = "Yellow"
$RED = "Red"
$CYAN = "Cyan"
$WHITE = "White"

# Helper function to write colored output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Helper function to check if command exists
function Test-CommandExists {
    param([string]$Command)
    $null = Get-Command $Command -ErrorAction SilentlyContinue
    return $?
}

# Helper function to run command with error checking
function Invoke-SafeCommand {
    param(
        [string]$Command,
        [string]$Arguments = "",
        [string]$ErrorMessage = "Command failed"
    )
    
    try {
        if ($Arguments) {
            $process = Start-Process -FilePath $Command -ArgumentList $Arguments -Wait -PassThru -NoNewWindow
        } else {
            $process = Start-Process -FilePath $Command -Wait -PassThru -NoNewWindow
        }
        
        if ($process.ExitCode -ne 0) {
            Write-ColorOutput "$ErrorMessage (Exit code: $($process.ExitCode))" $RED
            exit $process.ExitCode
        }
    }
    catch {
        Write-ColorOutput "$ErrorMessage : $_" $RED
        exit 1
    }
}

# Check prerequisites
function Test-Prerequisites {
    $missing = @()
    
    if (-not (Test-CommandExists "go")) {
        $missing += "Go (https://golang.org/dl/)"
    }
    
    if ($missing.Count -gt 0) {
        Write-ColorOutput "Missing prerequisites:" $RED
        foreach ($item in $missing) {
            Write-ColorOutput "  - $item" $RED
        }
        exit 1
    }
}

# Command functions
function Show-Help {
    Write-ColorOutput "Portal64 API - Available commands:" $CYAN
    Write-ColorOutput ""
    Write-ColorOutput "Build Commands:" $GREEN
    Write-ColorOutput "  help          Show this help message" $WHITE
    Write-ColorOutput "  build         Build the application binary" $WHITE
    Write-ColorOutput "  run           Run the application in development mode" $WHITE
    Write-ColorOutput "  run-prod      Run the application with production config" $WHITE
    Write-ColorOutput "  release       Build release versions for multiple platforms" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Development Commands:" $GREEN
    Write-ColorOutput "  deps          Download and tidy dependencies" $WHITE
    Write-ColorOutput "  format        Format code and tidy modules" $WHITE
    Write-ColorOutput "  lint          Run linter (requires golangci-lint)" $WHITE
    Write-ColorOutput "  swagger       Generate Swagger documentation" $WHITE
    Write-ColorOutput "  clean         Clean build artifacts" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Testing Commands:" $GREEN
    Write-ColorOutput "  test          Run all tests" $WHITE
    Write-ColorOutput "  test-unit     Run unit tests only" $WHITE
    Write-ColorOutput "  test-integration Run integration tests only" $WHITE
    Write-ColorOutput "  test-coverage Run tests with coverage report" $WHITE
    Write-ColorOutput "  bench         Run benchmarks" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Docker Commands:" $GREEN
    Write-ColorOutput "  docker-build  Build Docker image" $WHITE
    Write-ColorOutput "  docker-run    Run application in Docker" $WHITE
    Write-ColorOutput "  docker-compose-up   Start services with docker-compose" $WHITE
    Write-ColorOutput "  docker-compose-down Stop services with docker-compose" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Setup Commands:" $GREEN
    Write-ColorOutput "  setup         Setup development environment" $WHITE
    Write-ColorOutput "  install-tools Install development tools" $WHITE
    Write-ColorOutput "  install-air   Install air for hot reload" $WHITE
    Write-ColorOutput "  check         Run quality checks (format, lint, test)" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Development Server:" $GREEN
    Write-ColorOutput "  dev           Start development server with hot reload" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Examples:" $YELLOW
    Write-ColorOutput "  .\build.ps1 build" $WHITE
    Write-ColorOutput "  .\build.ps1 run" $WHITE
    Write-ColorOutput "  .\build.ps1 test" $WHITE
    Write-ColorOutput "  .\build.ps1 docker-build" $WHITE
}

function Build-Application {
    Test-Prerequisites
    Write-ColorOutput "Building $BINARY_NAME..." $GREEN
    
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" -Force | Out-Null
    }
    
    $env:CGO_ENABLED = "0"
    go build -ldflags="-w -s" -o $BINARY_PATH $MAIN_PATH
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Binary built: $BINARY_PATH" $GREEN
    } else {
        Write-ColorOutput "Build failed" $RED
        exit 1
    }
}

function Run-Application {
    Test-Prerequisites
    Write-ColorOutput "Starting development server..." $GREEN
    go run $MAIN_PATH
}

function Run-Production {
    Test-Prerequisites
    Write-ColorOutput "Starting production server..." $GREEN
    $env:ENVIRONMENT = "production"
    go run $MAIN_PATH
}

function Get-Dependencies {
    Test-Prerequisites
    Write-ColorOutput "Downloading dependencies..." $GREEN
    go mod download
    if ($LASTEXITCODE -ne 0) { exit 1 }
    
    go mod tidy
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Dependencies updated" $GREEN
    } else {
        Write-ColorOutput "Failed to update dependencies" $RED
        exit 1
    }
}

function Test-All {
    Test-Prerequisites
    Write-ColorOutput "Running all tests..." $GREEN
    Test-Unit
    Test-Integration
}

function Test-Unit {
    Test-Prerequisites
    Write-ColorOutput "Running unit tests..." $GREEN
    go test -v .\tests\unit\...
}

function Test-Integration {
    Test-Prerequisites
    Write-ColorOutput "Running integration tests..." $GREEN
    go test -v .\tests\integration\...
}

function Test-Coverage {
    Test-Prerequisites
    Write-ColorOutput "Running tests with coverage..." $GREEN
    go test -v -coverprofile=coverage.out .\...
    if ($LASTEXITCODE -eq 0) {
        go tool cover -html=coverage.out -o coverage.html
        Write-ColorOutput "Coverage report generated: coverage.html" $GREEN
    }
}

function Run-Benchmarks {
    Test-Prerequisites
    Write-ColorOutput "Running benchmarks..." $GREEN
    go test -bench=. -benchmem .\...
}

function Generate-Swagger {
    if (-not (Test-CommandExists "swag")) {
        Write-ColorOutput "swag command not found. Installing..." $YELLOW
        go install github.com/swaggo/swag/cmd/swag@latest
    }
    
    Write-ColorOutput "Generating Swagger documentation..." $GREEN
    swag init -g cmd/server/main.go -o docs/generated
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Swagger docs generated in docs/generated/" $GREEN
    }
}

function Run-Linter {
    if (-not (Test-CommandExists "golangci-lint")) {
        Write-ColorOutput "golangci-lint not found. Please install it first:" $YELLOW
        Write-ColorOutput "https://golangci-lint.run/usage/install/" $WHITE
        return
    }
    
    Write-ColorOutput "Running linter..." $GREEN
    golangci-lint run
}

function Format-Code {
    Test-Prerequisites
    Write-ColorOutput "Formatting code..." $GREEN
    go fmt .\...
    go mod tidy
    Write-ColorOutput "Code formatted" $GREEN
}

function Clean-Build {
    Write-ColorOutput "Cleaning build artifacts..." $GREEN
    
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    
    if (Test-Path "coverage.out") {
        Remove-Item "coverage.out"
    }
    
    if (Test-Path "coverage.html") {
        Remove-Item "coverage.html"
    }
    
    if (Test-Path "docs\generated") {
        Remove-Item -Recurse -Force "docs\generated"
    }
    
    Write-ColorOutput "Clean complete" $GREEN
}

function Install-Tools {
    Test-Prerequisites
    Write-ColorOutput "Installing development tools..." $GREEN
    
    Write-ColorOutput "Installing swag..." $WHITE
    go install github.com/swaggo/swag/cmd/swag@latest
    
    Write-ColorOutput "Installing golangci-lint..." $WHITE
    if ($IsWindows -or $env:OS -eq "Windows_NT") {
        # Install golangci-lint for Windows
        $temp = New-TemporaryFile
        Invoke-WebRequest -Uri "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" -OutFile $temp
        # Note: You may need to install via Chocolatey or Scoop instead
        Write-ColorOutput "Please install golangci-lint manually:" $YELLOW
        Write-ColorOutput "  choco install golangci-lint" $WHITE
        Write-ColorOutput "  or" $WHITE
        Write-ColorOutput "  scoop install golangci-lint" $WHITE
    }
    
    Write-ColorOutput "Development tools installation attempted" $GREEN
}

function Build-Docker {
    if (-not (Test-CommandExists "docker")) {
        Write-ColorOutput "Docker not found. Please install Docker first." $RED
        return
    }
    
    Write-ColorOutput "Building Docker image..." $GREEN
    docker build -t "${DOCKER_IMAGE}:${DOCKER_TAG}" .
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Docker image built: ${DOCKER_IMAGE}:${DOCKER_TAG}" $GREEN
    }
}

function Run-Docker {
    if (-not (Test-CommandExists "docker")) {
        Write-ColorOutput "Docker not found. Please install Docker first." $RED
        return
    }
    
    Write-ColorOutput "Running Docker container..." $GREEN
    docker run -p 8080:8080 --env-file .env "${DOCKER_IMAGE}:${DOCKER_TAG}"
}

function Start-DockerCompose {
    if (-not (Test-CommandExists "docker-compose")) {
        Write-ColorOutput "docker-compose not found. Please install Docker Compose first." $RED
        return
    }
    
    Write-ColorOutput "Starting services with docker-compose..." $GREEN
    docker-compose up -d
}

function Stop-DockerCompose {
    if (-not (Test-CommandExists "docker-compose")) {
        Write-ColorOutput "docker-compose not found. Please install Docker Compose first." $RED
        return
    }
    
    Write-ColorOutput "Stopping services with docker-compose..." $GREEN
    docker-compose down
}

function Setup-Environment {
    Write-ColorOutput "Setting up development environment..." $GREEN
    Get-Dependencies
    Install-Tools
    
    if (-not (Test-Path ".env")) {
        if (Test-Path ".env.example") {
            Copy-Item ".env.example" ".env"
            Write-ColorOutput "Created .env file from .env.example" $GREEN
            Write-ColorOutput "Please update .env with your database credentials" $YELLOW
        }
    }
    
    Write-ColorOutput "Setup complete!" $GREEN
}

function Build-Release {
    Test-Prerequisites
    Clean-Build
    Test-All
    
    Write-ColorOutput "Building release versions..." $GREEN
    
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" -Force | Out-Null
    }
    
    # Get version from git or use default
    $version = "1.0.0"
    try {
        $version = git describe --tags --always 2>$null
        if (-not $version) { $version = "1.0.0" }
    } catch {
        $version = "1.0.0"
    }
    
    $ldflags = "-w -s -X main.version=$version"
    
    # Build for different platforms
    Write-ColorOutput "Building for Linux (amd64)..." $WHITE
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -ldflags=$ldflags -o "bin\$BINARY_NAME-linux-amd64" $MAIN_PATH
    
    Write-ColorOutput "Building for Windows (amd64)..." $WHITE
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -ldflags=$ldflags -o "bin\$BINARY_NAME-windows-amd64.exe" $MAIN_PATH
    
    Write-ColorOutput "Building for macOS (amd64)..." $WHITE
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    go build -ldflags=$ldflags -o "bin\$BINARY_NAME-darwin-amd64" $MAIN_PATH
    
    Write-ColorOutput "Building for macOS (arm64)..." $WHITE
    $env:GOOS = "darwin"
    $env:GOARCH = "arm64"
    go build -ldflags=$ldflags -o "bin\$BINARY_NAME-darwin-arm64" $MAIN_PATH
    
    # Reset environment variables
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    
    Write-ColorOutput "Release binaries built in bin/" $GREEN
}

function Run-QualityChecks {
    Write-ColorOutput "Running quality checks..." $GREEN
    Format-Code
    Run-Linter
    Test-All
    Write-ColorOutput "All quality checks completed!" $GREEN
}

function Start-Development {
    if (-not (Test-CommandExists "air")) {
        Write-ColorOutput "Air not found. Installing..." $YELLOW
        go install github.com/cosmtrek/air@latest
        if ($LASTEXITCODE -ne 0) {
            Write-ColorOutput "Failed to install air. Please install manually:" $RED
            Write-ColorOutput "go install github.com/cosmtrek/air@latest" $WHITE
            return
        }
    }
    
    Write-ColorOutput "Starting development server with hot reload..." $GREEN
    air
}

function Install-Air {
    Test-Prerequisites
    Write-ColorOutput "Installing air for hot reload..." $GREEN
    go install github.com/cosmtrek/air@latest
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "Air installed" $GREEN
    }
}

# Main command dispatcher
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "build" { Build-Application }
    "run" { Run-Application }
    "run-prod" { Run-Production }
    "deps" { Get-Dependencies }
    "test" { Test-All }
    "test-unit" { Test-Unit }
    "test-integration" { Test-Integration }
    "test-coverage" { Test-Coverage }
    "bench" { Run-Benchmarks }
    "swagger" { Generate-Swagger }
    "lint" { Run-Linter }
    "format" { Format-Code }
    "clean" { Clean-Build }
    "install-tools" { Install-Tools }
    "docker-build" { Build-Docker }
    "docker-run" { Run-Docker }
    "docker-compose-up" { Start-DockerCompose }
    "docker-compose-down" { Stop-DockerCompose }
    "setup" { Setup-Environment }
    "release" { Build-Release }
    "check" { Run-QualityChecks }
    "dev" { Start-Development }
    "install-air" { Install-Air }
    default {
        Write-ColorOutput "Unknown command: $Command" $RED
        Write-ColorOutput "Use '.\build.ps1 help' to see available commands" $YELLOW
        exit 1
    }
}
