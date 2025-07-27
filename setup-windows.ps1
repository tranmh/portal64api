# Portal64 API Windows Setup Script
# This script helps set up the development environment on Windows

param(
    [switch]$InstallChocolatey,
    [switch]$InstallTools,
    [switch]$SetupProject,
    [switch]$All
)

$GREEN = "Green"
$YELLOW = "Yellow"
$RED = "Red"
$CYAN = "Cyan"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Test-IsAdmin {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Install-Chocolatey {
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-ColorOutput "Chocolatey is already installed" $GREEN
        return
    }
    
    Write-ColorOutput "Installing Chocolatey..." $GREEN
    Set-ExecutionPolicy Bypass -Scope Process -Force
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
    
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-ColorOutput "Chocolatey installed successfully" $GREEN
    } else {
        Write-ColorOutput "Failed to install Chocolatey" $RED
        exit 1
    }
}

function Install-DevelopmentTools {
    Write-ColorOutput "Installing development tools..." $GREEN
    
    # Check if Go is installed
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Installing Go..." $YELLOW
        if (Get-Command choco -ErrorAction SilentlyContinue) {
            choco install golang -y
        } else {
            Write-ColorOutput "Please install Go manually from: https://golang.org/dl/" $RED
            Write-ColorOutput "Or install Chocolatey first with -InstallChocolatey" $YELLOW
        }
    } else {
        Write-ColorOutput "Go is already installed: $(go version)" $GREEN
    }
    
    # Check if Git is installed
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Installing Git..." $YELLOW
        if (Get-Command choco -ErrorAction SilentlyContinue) {
            choco install git -y
        } else {
            Write-ColorOutput "Please install Git manually from: https://git-scm.com/download/win" $RED
        }
    } else {
        Write-ColorOutput "Git is already installed: $(git --version)" $GREEN
    }
    
    # Install Docker Desktop (optional)
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Docker not found. To install:" $YELLOW
        Write-ColorOutput "  choco install docker-desktop -y" $WHITE
        Write-ColorOutput "  or download from: https://www.docker.com/products/docker-desktop" $WHITE
    } else {
        Write-ColorOutput "Docker is already installed: $(docker --version)" $GREEN
    }
    
    # Install Make (optional, for cross-platform compatibility)
    if (-not (Get-Command make -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Installing Make..." $YELLOW
        if (Get-Command choco -ErrorAction SilentlyContinue) {
            choco install make -y
        }
    }
    
    # Install golangci-lint
    if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Installing golangci-lint..." $YELLOW
        if (Get-Command choco -ErrorAction SilentlyContinue) {
            choco install golangci-lint -y
        }
    } else {
        Write-ColorOutput "golangci-lint is already installed" $GREEN
    }
}

function Setup-Project {
    Write-ColorOutput "Setting up Portal64 API project..." $GREEN
    
    # Check if we're in the right directory
    if (-not (Test-Path "go.mod")) {
        Write-ColorOutput "Error: go.mod not found. Please run this script from the project root directory." $RED
        exit 1
    }
    
    # Download dependencies
    Write-ColorOutput "Downloading Go dependencies..." $YELLOW
    go mod download
    go mod tidy
    
    # Create .env file if it doesn't exist
    if (-not (Test-Path ".env") -and (Test-Path ".env.example")) {
        Write-ColorOutput "Creating .env file from .env.example..." $YELLOW
        Copy-Item ".env.example" ".env"
        Write-ColorOutput "Please edit .env file with your database credentials" $YELLOW
    }
    
    # Create necessary directories
    $directories = @("bin", "logs", "tmp")
    foreach ($dir in $directories) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
            Write-ColorOutput "Created directory: $dir" $GREEN
        }
    }
    
    # Install Go tools
    Write-ColorOutput "Installing Go development tools..." $YELLOW
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/cosmtrek/air@latest
    
    Write-ColorOutput "Project setup complete!" $GREEN
    Write-ColorOutput "Next steps:" $CYAN
    Write-ColorOutput "  1. Edit .env file with your database credentials" $WHITE
    Write-ColorOutput "  2. Run: .\build.ps1 run" $WHITE
    Write-ColorOutput "  3. Open: http://localhost:8080/swagger/index.html" $WHITE
}

function Show-Help {
    Write-ColorOutput "Portal64 API Windows Setup Script" $CYAN
    Write-ColorOutput ""
    Write-ColorOutput "Usage:" $GREEN
    Write-ColorOutput "  .\setup-windows.ps1 [options]" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Options:" $GREEN
    Write-ColorOutput "  -InstallChocolatey    Install Chocolatey package manager" $WHITE
    Write-ColorOutput "  -InstallTools         Install development tools (Go, Git, etc.)" $WHITE
    Write-ColorOutput "  -SetupProject         Setup the Portal64 API project" $WHITE
    Write-ColorOutput "  -All                  Run all setup steps" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Examples:" $YELLOW
    Write-ColorOutput "  .\setup-windows.ps1 -All" $WHITE
    Write-ColorOutput "  .\setup-windows.ps1 -InstallTools -SetupProject" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "Manual Steps:" $CYAN
    Write-ColorOutput "1. Install Chocolatey (package manager):" $WHITE
    Write-ColorOutput "   .\setup-windows.ps1 -InstallChocolatey" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "2. Install development tools:" $WHITE
    Write-ColorOutput "   .\setup-windows.ps1 -InstallTools" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "3. Setup project:" $WHITE
    Write-ColorOutput "   .\setup-windows.ps1 -SetupProject" $WHITE
    Write-ColorOutput ""
    Write-ColorOutput "4. Build and run:" $WHITE
    Write-ColorOutput "   .\build.ps1 build" $WHITE
    Write-ColorOutput "   .\build.ps1 run" $WHITE
}

# Main execution
if ($All) {
    $InstallChocolatey = $true
    $InstallTools = $true
    $SetupProject = $true
}

if (-not $InstallChocolatey -and -not $InstallTools -and -not $SetupProject) {
    Show-Help
    exit 0
}

Write-ColorOutput "Portal64 API Windows Setup" $CYAN
Write-ColorOutput "=========================" $CYAN

if ($InstallChocolatey) {
    if (-not (Test-IsAdmin)) {
        Write-ColorOutput "WARNING: Installing Chocolatey requires administrator privileges." $YELLOW
        Write-ColorOutput "Please run PowerShell as Administrator or install Chocolatey manually." $YELLOW
    } else {
        Install-Chocolatey
    }
}

if ($InstallTools) {
    Install-DevelopmentTools
}

if ($SetupProject) {
    Setup-Project
}

Write-ColorOutput ""
Write-ColorOutput "Setup completed!" $GREEN
Write-ColorOutput "Use .\build.ps1 help to see available build commands" $CYAN
