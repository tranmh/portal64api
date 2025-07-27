# Windows Development Guide

This guide provides Windows-specific instructions for developing the Portal64 API.

## Prerequisites for Windows

- **PowerShell 5.1 or higher** (usually pre-installed on Windows 10/11)
- **Go 1.21+**: Download from [https://golang.org/dl/](https://golang.org/dl/)
- **Git**: Download from [https://git-scm.com/download/win](https://git-scm.com/download/win)
- **MySQL 8.0+**: Download from [https://dev.mysql.com/downloads/mysql/](https://dev.mysql.com/downloads/mysql/)

## Quick Windows Setup

### Option 1: Automated Setup (Recommended)

1. **Clone the repository**:
   ```powershell
   git clone <repository-url>
   cd portal64api
   ```

2. **Run the automated setup** (as Administrator for best results):
   ```powershell
   .\setup-windows.ps1 -All
   ```

3. **Configure your databases** by editing `.env` file

4. **Start the API**:
   ```powershell
   .\build.ps1 run
   ```

### Option 2: Manual Setup

1. **Install development tools manually**:
   - Install Go from [golang.org](https://golang.org/dl/)
   - Install Git from [git-scm.com](https://git-scm.com/download/win)
   - Optionally install [Chocolatey](https://chocolatey.org/install) for easier tool management

2. **Setup the project**:
   ```powershell
   git clone <repository-url>
   cd portal64api
   .\setup-windows.ps1 -SetupProject
   ```

## Windows Build Commands

All commands use the PowerShell script `build.ps1`:

### Basic Commands
```powershell
# Show help
.\build.ps1 help

# Build the application
.\build.ps1 build

# Run in development mode
.\build.ps1 run

# Run in production mode
.\build.ps1 run-prod
```

### Development Commands
```powershell
# Install/update dependencies
.\build.ps1 deps

# Format code
.\build.ps1 format

# Run linter
.\build.ps1 lint

# Generate Swagger docs
.\build.ps1 swagger

# Clean build artifacts
.\build.ps1 clean
```

### Testing Commands
```powershell
# Run all tests
.\build.ps1 test

# Run only unit tests
.\build.ps1 test-unit

# Run only integration tests
.\build.ps1 test-integration

# Run tests with coverage report
.\build.ps1 test-coverage

# Run benchmarks
.\build.ps1 bench
```

### Docker Commands
```powershell
# Build Docker image
.\build.ps1 docker-build

# Run in Docker container
.\build.ps1 docker-run

# Start with docker-compose (includes MySQL)
.\build.ps1 docker-compose-up

# Stop docker-compose services
.\build.ps1 docker-compose-down
```

### Advanced Commands
```powershell
# Setup complete development environment
.\build.ps1 setup

# Install development tools
.\build.ps1 install-tools

# Run quality checks (format + lint + test)
.\build.ps1 check

# Build release binaries for all platforms
.\build.ps1 release

# Start development server with hot reload
.\build.ps1 dev
```

## Alternative Ways to Run Commands

### Using Batch File Wrapper
For Command Prompt users:
```cmd
build.bat help
build.bat run
build.bat test
```

### Using Cross-Platform Script
If you have WSL or Git Bash:
```bash
./build.sh help
./build.sh run
./build.sh test
```

### Direct Go Commands
You can also use Go directly:
```powershell
# Run the application
go run cmd/server/main.go

# Run tests
go test ./...

# Build binary
go build -o bin/portal64api.exe cmd/server/main.go
```

## Windows-Specific Configuration

### Environment Variables
Create a `.env` file in the project root:
```env
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
ENVIRONMENT=development

# HTTPS (optional)
ENABLE_HTTPS=false
CERT_FILE=C:\path\to\certificate.crt
KEY_FILE=C:\path\to\private.key

# Database Configuration
MVDSB_HOST=localhost
MVDSB_PORT=3306
MVDSB_USERNAME=your_username
MVDSB_PASSWORD=your_password
MVDSB_DATABASE=mvdsb

# ... (repeat for other databases)
```

### Windows Firewall
If you're running the API locally, you may need to allow it through Windows Firewall:
1. Go to Windows Defender Firewall settings
2. Click "Allow an app or feature through Windows Defender Firewall"
3. Add the Go executable or allow on port 8080

## Development Workflow

### Daily Development
```powershell
# Start development server with hot reload
.\build.ps1 dev

# In another terminal, run tests while developing
.\build.ps1 test-unit

# Format code before committing
.\build.ps1 format

# Run all quality checks before pushing
.\build.ps1 check
```

### Working with Docker on Windows

#### Docker Desktop Setup
1. Install [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)
2. Enable WSL 2 backend (recommended)
3. Start Docker Desktop

#### Development with Docker Compose
```powershell
# Start all services (API + MySQL + Redis)
.\build.ps1 docker-compose-up

# View logs
docker-compose logs -f portal64api

# Stop all services
.\build.ps1 docker-compose-down
```

## Troubleshooting Windows Issues

### PowerShell Execution Policy
If you get execution policy errors:
```powershell
# Check current policy
Get-ExecutionPolicy

# Allow scripts for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or bypass for single command
powershell -ExecutionPolicy Bypass -File build.ps1 run
```

### Go Command Not Found
If `go` command is not recognized:
1. Ensure Go is installed from [golang.org](https://golang.org/dl/)
2. Add Go to your PATH:
   - Go to System Properties → Advanced → Environment Variables
   - Add `C:\Go\bin` to your PATH
   - Restart PowerShell

### Long Path Issues
Windows has a 260-character path limit. To enable long paths:
1. Run as Administrator:
   ```powershell
   New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force
   ```
2. Or enable via Group Policy: Computer Configuration → Administrative Templates → System → Filesystem → Enable Win32 long paths

### Antivirus Software
Some antivirus software may interfere with Go compilation:
- Add the project directory to antivirus exclusions
- Add Go installation directory to exclusions
- Temporarily disable real-time protection during development

## IDE Setup for Windows

### Visual Studio Code
1. Install [VS Code](https://code.visualstudio.com/)
2. Install the Go extension
3. Configure workspace settings in `.vscode/settings.json`:
   ```json
   {
     "go.buildTags": "",
     "go.lintTool": "golangci-lint",
     "go.formatTool": "goimports"
   }
   ```

### GoLand
1. Install [GoLand](https://www.jetbrains.com/go/)
2. Open project directory
3. Configure Go SDK path
4. Enable Go modules support

## Performance Tips for Windows

1. **Use SSD**: Store the project on an SSD for faster compilation
2. **Exclude from Windows Defender**: Add project directory to exclusions
3. **Use PowerShell ISE or Windows Terminal**: Better than Command Prompt
4. **Enable Developer Mode**: Allows creating symlinks without admin privileges

## Getting Help

If you encounter Windows-specific issues:
1. Check this guide first
2. Review the main README.md
3. Check project issues on GitHub
4. Ensure all prerequisites are properly installed

## File Structure Summary

Windows-specific files in the project:
```
portal64api/
├── build.ps1              # Main PowerShell build script
├── build.bat              # Batch file wrapper
├── setup-windows.ps1      # Windows environment setup
├── .air.toml              # Hot reload configuration
└── README-WINDOWS.md      # This file
```

Use `.\build.ps1 help` to see all available commands!
