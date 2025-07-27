# Portal64 API

A comprehensive REST API for the DWZ (Deutsche Wertungszahl) chess rating system, providing access to player ratings, club information, and tournament data from the SVW (Schachverband Württemberg) chess federation databases.

## Features

- **Player Management**: Search players, get detailed player information, rating history
- **Club Management**: Search clubs, get club details, member listings with statistics  
- **Tournament Management**: Search tournaments, get upcoming/recent tournaments, tournament results
- **Multiple Response Formats**: JSON and CSV output support
- **Cross-Origin Support**: Full CORS implementation for web applications
- **Comprehensive Documentation**: OpenAPI/Swagger documentation
- **Production Ready**: HTTPS support, proper error handling, logging, health checks

## API Overview

### Player ID Format
Player IDs follow the format: `{VKZ}-{PersonID}` (e.g., `C0101-1014`)
- `VKZ`: Club identifier (e.g., `C0101` for "Post-SV Ulm")  
- `PersonID`: Unique person identifier

### Club ID Format
Club IDs use the VKZ format: `{C}{NNNN}` (e.g., `C0101`)

### Tournament ID Format  
Tournament IDs follow the format: `{Code}-{SubCode}-{Type}` (e.g., `C529-K00-HT1`)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- Access to the three databases: `mvdsb`, `portal64_bdw`, `portal64_svw`

### Installation

#### Linux/Mac

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd portal64api
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Setup environment**
   ```bash
   make setup
   ```

4. **Configure database connections**
   Edit the `.env` file with your database credentials

5. **Run the application**
   ```bash
   make run
   ```

#### Windows

1. **Clone the repository**
   ```powershell
   git clone <repository-url>
   cd portal64api
   ```

2. **Setup development environment (one-time setup)**
   ```powershell
   # Install tools and setup project (run as Administrator for best results)
   .\setup-windows.ps1 -All
   ```

   Or do it step by step:
   ```powershell
   # Install development tools
   .\setup-windows.ps1 -InstallTools
   
   # Setup the project
   .\setup-windows.ps1 -SetupProject
   ```

3. **Configure database connections**
   Edit the `.env` file with your database credentials:
   ```env
   # MVDSB Database
   MVDSB_HOST=localhost
   MVDSB_PORT=3306
   MVDSB_USERNAME=your_username
   MVDSB_PASSWORD=your_password
   MVDSB_DATABASE=mvdsb

   # Portal64_BDW Database  
   PORTAL64_BDW_HOST=localhost
   PORTAL64_BDW_PORT=3306
   PORTAL64_BDW_USERNAME=your_username
   PORTAL64_BDW_PASSWORD=your_password
   PORTAL64_BDW_DATABASE=portal64_bdw

   # Portal64_SVW Database
   PORTAL64_SVW_HOST=localhost
   PORTAL64_SVW_PORT=3306
   PORTAL64_SVW_USERNAME=your_username
   PORTAL64_SVW_PASSWORD=your_password
   PORTAL64_SVW_DATABASE=portal64_svw
   ```

4. **Run the application**
   ```powershell
   .\build.ps1 run
   ```

   Or using the batch file wrapper:
   ```cmd
   build.bat run
   ```

#### Cross-Platform (Auto-Detection)

For systems with both make and PowerShell available:
```bash
# Automatically detects your OS and uses the appropriate build tool
./build.sh run
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Core Endpoints

#### Players
- `GET /api/v1/players` - Search players
- `GET /api/v1/players/{id}` - Get player by ID (e.g., `C0101-1014`)
- `GET /api/v1/players/{id}/rating-history` - Get player's rating history

#### Clubs  
- `GET /api/v1/clubs` - Search clubs
- `GET /api/v1/clubs/{id}` - Get club by ID (e.g., `C0101`)
- `GET /api/v1/clubs/{club_id}/players` - Get players in a club
- `GET /api/v1/clubs/all` - Get all clubs

#### Tournaments
- `GET /api/v1/tournaments` - Search tournaments
- `GET /api/v1/tournaments/{id}` - Get tournament by ID
- `GET /api/v1/tournaments/upcoming` - Get upcoming tournaments
- `GET /api/v1/tournaments/recent` - Get recent tournaments
- `GET /api/v1/tournaments/date-range` - Get tournaments by date range

#### System
- `GET /health` - Health check
- `GET /swagger/*` - API documentation

### Response Formats

All endpoints support both JSON and CSV formats:

**JSON (default):**
```bash
curl "http://localhost:8080/api/v1/players/C0101-1014"
```

**CSV:**
```bash
curl "http://localhost:8080/api/v1/players/C0101-1014?format=csv"
# or
curl -H "Accept: text/csv" "http://localhost:8080/api/v1/players/C0101-1014"
```

### Query Parameters

Most endpoints support these common parameters:
- `query` - Search term
- `limit` - Maximum results (default: 20, max: 100)
- `offset` - Results to skip (default: 0)
- `sort_by` - Field to sort by
- `sort_order` - Sort direction (`asc`/`desc`)
- `format` - Response format (`json`/`csv`)

## Examples

### Get a specific player
```bash
curl "http://localhost:8080/api/v1/players/C0101-1014"
```

### Search for players named "Müller"
```bash
curl "http://localhost:8080/api/v1/players?query=Müller&limit=10"
```

### Get all players in a club
```bash
curl "http://localhost:8080/api/v1/clubs/C0101/players"
```

### Search clubs in CSV format
```bash
curl "http://localhost:8080/api/v1/clubs?query=Ulm&format=csv"
```

### Get upcoming tournaments
```bash
curl "http://localhost:8080/api/v1/tournaments/upcoming?limit=5"
```

## Development

### Build Tools Overview

The project provides multiple build tools for different environments:

| File | Platform | Description |
|------|----------|-------------|
| `Makefile` | Linux/Mac | Traditional Unix build automation |
| `build.ps1` | Windows | PowerShell build script with full feature parity |
| `build.bat` | Windows | Batch file wrapper for PowerShell script |
| `build.sh` | Cross-platform | Auto-detects OS and uses appropriate tool |
| `setup-windows.ps1` | Windows | One-time Windows development environment setup |

### Windows Development Setup

For Windows developers, we provide comprehensive PowerShell-based tools:

1. **One-time setup** (installs tools and configures project):
   ```powershell
   # Run as Administrator for best results
   .\setup-windows.ps1 -All
   ```

2. **Development workflow**:
   ```powershell
   # Build the application
   .\build.ps1 build
   
   # Run with hot reload
   .\build.ps1 dev
   
   # Run tests
   .\build.ps1 test
   
   # Generate documentation
   .\build.ps1 swagger
   ```

3. **Alternative using batch file**:
   ```cmd
   REM Works in Command Prompt
   build.bat help
   build.bat run
   build.bat test
   ```

### Available Commands

**Linux/Mac (using Makefile):**
```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run in development mode
make test          # Run all tests
make test-unit     # Run unit tests only
make test-integration # Run integration tests only
make swagger       # Generate Swagger docs
make lint          # Run linter
make format        # Format code
make clean         # Clean build artifacts
```

**Windows (using PowerShell):**
```powershell
.\build.ps1 help          # Show all available commands
.\build.ps1 build         # Build the application
.\build.ps1 run           # Run in development mode
.\build.ps1 test          # Run all tests
.\build.ps1 test-unit     # Run unit tests only
.\build.ps1 test-integration # Run integration tests only
.\build.ps1 swagger       # Generate Swagger docs
.\build.ps1 lint          # Run linter
.\build.ps1 format        # Format code
.\build.ps1 clean         # Clean build artifacts
```

**Cross-platform (auto-detection):**
```bash
./build.sh help    # Automatically uses make or PowerShell based on OS
```

### Project Structure

```
portal64api/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/     # HTTP request handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── routes.go     # Route definitions
│   ├── services/         # Business logic
│   ├── repositories/     # Data access layer
│   ├── models/           # Domain models
│   ├── database/         # Database connections
│   └── config/           # Configuration
├── pkg/
│   ├── utils/            # Utility functions
│   └── errors/           # Error handling
├── tests/
│   ├── unit/             # Unit tests
│   └── integration/      # Integration tests
├── docs/                 # API documentation
├── configs/              # Configuration files
└── scripts/              # Build and utility scripts
```

### Running Tests

**Linux/Mac:**
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v ./tests/unit/utils_test.go
```

**Windows:**
```powershell
# Run all tests
.\build.ps1 test

# Run with coverage
.\build.ps1 test-coverage

# Run specific test
go test -v .\tests\unit\utils_test.go
```

### Hot Reload Development

**Linux/Mac:**
```bash
make dev
```

**Windows:**
```powershell
.\build.ps1 dev
```

This starts the development server with automatic reload on file changes using [Air](https://github.com/cosmtrek/air).

### Docker Support

**Build and run with Docker (Linux/Mac):**
```bash
make docker-build
make docker-run
```

**Build and run with Docker (Windows):**
```powershell
.\build.ps1 docker-build
.\build.ps1 docker-run
```

**Use Docker Compose (includes MySQL) - Cross-platform:**
```bash
# Linux/Mac
make docker-compose-up

# Windows
.\build.ps1 docker-compose-up

# Or directly with docker-compose
docker-compose up -d
```

## Configuration

### Environment Variables

The application can be configured using environment variables or a `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8080` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `ENABLE_HTTPS` | Enable HTTPS | `false` |
| `CERT_FILE` | SSL certificate file path | `` |
| `KEY_FILE` | SSL private key file path | `` |
| `MVDSB_HOST` | MVDSB database host | `localhost` |
| `MVDSB_PORT` | MVDSB database port | `3306` |
| ... | (and similar for other databases) | |

### YAML Configuration

Alternatively, use YAML configuration files in the `configs/` directory:
- `config.yaml` - Development configuration
- `config.prod.yaml` - Production configuration

## Production Deployment

### HTTPS Setup

1. **Enable HTTPS in configuration:**
   ```env
   ENABLE_HTTPS=true
   CERT_FILE=/path/to/certificate.crt
   KEY_FILE=/path/to/private.key
   ```

2. **Use environment-specific config:**
   ```bash
   ENVIRONMENT=production ./portal64api
   ```

### Reverse Proxy (Recommended)

For production, use a reverse proxy like Nginx:

```nginx
server {
    listen 80;
    server_name api.svw.info;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Database Permissions

Ensure the API user has appropriate permissions:

```sql
-- Minimum required permissions
GRANT SELECT ON mvdsb.* TO 'portal64api'@'%';
GRANT SELECT ON portal64_bdw.* TO 'portal64api'@'%';  
GRANT SELECT ON portal64_svw.* TO 'portal64api'@'%';
```

## API Documentation

### Swagger/OpenAPI

Interactive API documentation is available at:
- **Development**: http://localhost:8080/swagger/index.html
- **Production**: https://api.svw.info/swagger/index.html

### Response Format

All API responses follow a consistent format:

**Success Response:**
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "count": 20
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Resource not found",
  "message": "The requested player could not be found"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Quality

- Follow Go conventions and idioms
- Write tests for new functionality
- Run `make check` before submitting
- Ensure all tests pass
- Update documentation as needed

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Email: support@svw.info

## Changelog

### v1.0.0
- Initial release
- Complete REST API for DWZ system
- Support for players, clubs, and tournaments
- JSON and CSV response formats
- Comprehensive documentation
- Docker support
- Production-ready configuration
