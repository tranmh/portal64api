# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Portal64 API is a comprehensive REST API for the DWZ (Deutsche Wertungszahl) chess rating system. It provides access to player ratings, club information, and tournament data from the SVW (Schachverband Württemberg) chess federation databases.

The application is built with Go using the Gin web framework and GORM ORM, connecting to two MySQL databases (`mvdsb` and `portal64_bdw`).

**Critical Context**: This is a Go rewrite of legacy PHP code. When making changes to query logic, always verify behavior matches the original PHP implementation to avoid bugs.

## Development Commands

The project supports cross-platform development with identical commands across Linux/Mac and Windows:

### Linux/Mac (using Makefile):
```bash
make build              # Build the application
make run                # Run in development mode
make dev                # Run with hot reload (requires air)
make test               # Run all tests (unit + integration)
make test-unit          # Run unit tests only
make test-integration   # Run integration tests only
make test-coverage      # Run tests with coverage report
make swagger            # Generate Swagger documentation
make lint               # Run linter
make format             # Format code
make clean              # Clean build artifacts
make setup              # Setup development environment (first time)
```

### Windows (using PowerShell):
```powershell
.\build.ps1 build              # Build the application
.\build.ps1 run                # Run in development mode
.\build.ps1 dev                # Run with hot reload
.\build.ps1 test               # Run all tests
.\build.ps1 test-unit          # Run unit tests only
.\build.ps1 test-integration   # Run integration tests only
.\build.ps1 swagger            # Generate Swagger documentation
.\build.ps1 lint               # Run linter
.\build.ps1 format             # Format code
.\build.ps1 clean              # Clean build artifacts
.\setup-windows.ps1 -All       # Setup development environment (first time)
```

### Cross-Platform:
```bash
./build.sh <command>   # Auto-detects OS and uses appropriate tool
```

## Architecture

### Application Entry Point
- **cmd/server/main.go**: Main entry point that initializes logging, databases, cache, services (import, Kader-Planung), sets up routes, and starts the HTTP/HTTPS server with graceful shutdown

### Core Layers (Standard Go project structure)

1. **Handlers** (internal/api/handlers/): HTTP request handlers for all endpoints
   - Player, club, tournament, address handlers
   - Admin, import, Kader-Planung, somatogramm handlers
   - Each handler receives service dependencies via constructor injection

2. **Services** (internal/services/): Business logic layer
   - PlayerService, ClubService, TournamentService, AddressService
   - ImportService: Automated data import from external sources
   - KaderPlanungService: Generates roster planning reports with statistical analysis
   - SomatogrammService (deprecated): Growth chart generation (now integrated into Kader-Planung)
   - Services depend on repositories and cache

3. **Repositories** (internal/repositories/): Data access layer
   - Direct database access using GORM
   - Each repository handles queries for a specific domain (players, clubs, tournaments, addresses)

4. **Models** (internal/models/): Domain models and data structures
   - Player, Club, Tournament, Address models
   - Request/response DTOs

5. **Middleware** (internal/api/middleware/): HTTP middleware
   - CORS, logging, error handling, response format negotiation (JSON/CSV)

6. **Database** (internal/database/): Database connection management
   - Manages connections to two databases: MVDSB and Portal64_BDW
   - Connection pooling configuration

7. **Cache** (internal/cache/): Redis-based caching layer
   - CacheService interface with Redis implementation
   - Key generation, metrics collection
   - Used throughout services to cache frequently accessed data

8. **Configuration** (internal/config/): Application configuration
   - Loads from environment variables and .env file
   - Server, database, cache, import, Kader-Planung settings

### Routes Setup (internal/api/routes.go)
All API routes are defined in `SetupRoutes()` which:
- Creates repositories
- Injects repositories into services
- Injects services into handlers
- Configures API routes under `/api/v1`
- Sets up Swagger documentation at `/swagger/`
- Sets up demo interface at `/demo/`

### Two Database Architecture
The application connects to two separate MySQL databases:
- **MVDSB**: Main database containing player (`person`), club (`organisation`), membership (`mitgliedschaft`) data
- **Portal64_BDW**: Database with DWZ evaluations (`evaluation`, 2.8M+ records), tournaments (`tournament`, `tournamentmaster`), games (`game`, 7.8M+ records), and addresses

Database connections are established in `database.Connect()` and stored in a `Databases` struct passed to repositories.

**Critical Database Query Gotchas**:
1. **ALWAYS use `tournamentmaster` table for evaluation JOINs**, not `tournament` table. These have overlapping IDs but completely different tournament data. Incorrect JOINs cause wrong tournament data to appear.
   - Correct: `FROM evaluation e INNER JOIN tournamentmaster tm ON e.idMaster = tm.id`
   - Wrong: `FROM evaluation e INNER JOIN tournament t ON e.idMaster = t.id` (causes data corruption appearance)
2. **Include future-ending memberships**: Use `bis IS NULL OR bis > CURDATE()` in membership queries to match PHP behavior
3. **Use EXISTS subqueries for active players** to avoid MySQL's 65K placeholder limit: `EXISTS (SELECT 1 FROM mitgliedschaft WHERE mitgliedschaft.person = person.id AND (mitgliedschaft.bis IS NULL OR mitgliedschaft.bis > CURDATE()))`
4. **Range-based prefix matching** for better index usage: `name >= 'query' AND name < 'queryzz'` instead of `LIKE 'query%'`

### ID Formats (Critical for API Usage)
- **Player IDs**: `{VKZ}-{PersonID}` format (e.g., `C0101-1014`)
  - VKZ is the club identifier
  - PersonID is the unique person identifier
- **Club IDs**: `{C}{NNNN}` format (e.g., `C0101` for "Post-SV Ulm")
- **Tournament IDs**: `{Code}-{SubCode}-{Type}` format (e.g., `B718-A08-BEL`, `C529-K00-HT1`)
  - Code: Letter (A-Z) + year digit + week number

### Response Formats
All endpoints support both JSON and CSV output:
- Default: JSON
- Specify `format=csv` query parameter or `Accept: text/csv` header for CSV
- Response format handling is done in middleware: `internal/api/middleware/response_format.go`
- CSV responses include UTF-8 BOM and proper Content-Type headers for German Excel compatibility

## Specialized Sub-Projects

### Kader-Planung (kader-planung/)
Standalone Go application for generating comprehensive player roster reports with statistical analysis and somatogram percentiles. Key features:
- Fetches player data from Portal64 API
- Calculates statistical metrics (12-month ratings, game statistics)
- Generates somatogram percentiles (Germany-wide rankings by age/gender)
- Outputs CSV reports optimized for German Excel (with UTF-8 BOM)
- **Always uses hybrid mode** (detailed + statistical analysis)
- **Always outputs CSV format**
- **Always processes complete German dataset** (~50,000 players) regardless of club filter
- Integrated into main API via KaderPlanungService
- Binary path: `kader-planung/bin/kader-planung.exe` (Windows) or `kader-planung/bin/kader-planung` (Linux)

**Important Implementation Details**:
- The Kader-Planung service runs the standalone binary as a subprocess
- Uses completion callbacks to auto-execute after ImportService completes
- Environment config obsolete parameters: `KADER_PLANUNG_MODE`, `KADER_PLANUNG_INCLUDE_STATISTICS`, `KADER_PLANUNG_OUTPUT_FORMAT` (see docs/MIGRATION_GUIDE_KADER_PLANUNG.md)
- Default `MIN_SAMPLE_SIZE` changed from 100 to 10 for better data coverage

### Somatogramm (somatogramm/) - DEPRECATED
Legacy standalone application for growth chart generation. **Now integrated into Kader-Planung as "statistical mode"**.
- Backward compatibility maintained via SomatogrammCompatibilityHandler
- All new development should use Kader-Planung service
- API routes at `/api/v1/somatogramm/*` redirect to Kader-Planung statistical mode

## Configuration

### Environment Variables (.env file)
Key configuration from `.env.example`:
- **Server**: `SERVER_PORT`, `SERVER_HOST`, `ENVIRONMENT`, `ENABLE_HTTPS`, `CERT_FILE`, `KEY_FILE`
- **MVDSB Database**: `MVDSB_HOST`, `MVDSB_PORT`, `MVDSB_USERNAME`, `MVDSB_PASSWORD`, `MVDSB_DATABASE`, `MVDSB_CHARSET` (use `utf8mb4`)
- **Portal64_BDW Database**: `PORTAL64_BDW_HOST`, `PORTAL64_BDW_PORT`, `PORTAL64_BDW_USERNAME`, `PORTAL64_BDW_PASSWORD`, `PORTAL64_BDW_DATABASE`, `PORTAL64_BDW_CHARSET` (use `utf8mb4`)
- **Redis Cache**: `CACHE_ENABLED`, `CACHE_ADDRESS`, `CACHE_PASSWORD`, etc.
- **Import Service**: `IMPORT_ENABLED`, `IMPORT_SCHEDULE` (cron format), `IMPORT_SCP_*` (remote server access), `IMPORT_FRESHNESS_*` (skip if not newer)
- **Kader-Planung**: `KADER_PLANUNG_ENABLED`, `KADER_PLANUNG_BINARY_PATH`, `KADER_PLANUNG_OUTPUT_DIR`, `KADER_PLANUNG_API_BASE_URL`, `KADER_PLANUNG_CONCURRENCY`, `KADER_PLANUNG_MIN_SAMPLE_SIZE`

**Critical**: Always use `utf8mb4` charset for database connections to avoid encoding issues with German umlauts (ü, ö, ä, ß). Database creation uses `utf8mb4_general_ci` collation.

### First-Time Setup
1. Copy `.env.example` to `.env`
2. Update database credentials in `.env`
3. **Ensure both databases exist**: MVDSB and Portal64_BDW must be available
4. Run setup command: `make setup` (Linux/Mac) or `.\setup-windows.ps1 -All` (Windows)
5. Start development server: `make dev` or `.\build.ps1 dev`

## Testing

### Test Structure
- **tests/unit/**: Unit tests for utility functions
- **tests/integration/**: Integration tests for API endpoints (requires running server)
- **E2E Tests**: Playwright-based frontend tests (see package.json)

### Running Tests
```bash
make test               # All tests (unit + integration)
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-coverage      # With coverage report
npm run test:e2e        # Frontend E2E tests (Playwright)
```

### Test Files for Kader-Planung
Located in `kader-planung/internal/*/` with `_test.go` suffix. Includes unit tests, integration tests, benchmark tests, and regression tests.

## API Documentation

Swagger documentation is auto-generated and available at:
- Development: `http://localhost:8080/swagger/`
- Swagger annotations are in handler methods using `@` comments

To regenerate Swagger docs after handler changes:
```bash
make swagger           # Linux/Mac
.\build.ps1 swagger    # Windows
```

## Common Development Patterns

### Adding a New Endpoint
1. Define the model in `internal/models/`
2. Add repository methods in `internal/repositories/`
3. Add service methods in `internal/services/` (inject repository + cache)
4. Add handler in `internal/api/handlers/` with Swagger annotations
5. Register route in `internal/api/routes.go` in `SetupRoutes()`
6. Regenerate Swagger docs: `make swagger` or `.\build.ps1 swagger`

### Database Queries
- Use GORM for all database queries
- Always specify which database connection to use: `dbs.MVDSB` or `dbs.Portal64BDW`
- **Verify JOIN logic against PHP code** when working with cross-table queries
- Use proper error handling and logging
- Consider caching for frequently accessed data
- Avoid N+1 queries: prefer single optimized queries with JOINs over multiple sequential queries

### Caching Pattern
```go
// Check cache first
cacheKey := cache.GenerateKey("prefix", params...)
if cachedData, err := service.cache.Get(ctx, cacheKey); err == nil {
    return cachedData, nil
}

// Fetch from database
data, err := repository.FetchData(params...)
if err != nil {
    return nil, err
}

// Store in cache with appropriate TTL
service.cache.Set(ctx, cacheKey, data, 15*time.Minute)
return data, nil
```

### Service Integration Pattern
Services that execute external binaries (ImportService, KaderPlanungService):
- Implement `Start()/Stop()` lifecycle methods
- Use completion callbacks for chaining service executions
- Implement `ImportCompleteCallback` interface: `OnImportComplete()` method
- Register callbacks via `AddCompletionCallback(callback)`
- Log execution details for debugging
- Handle subprocess output and errors properly

### Import Service Workflow (6 Phases)
When working with data import functionality:
1. **Freshness Check**: Check if remote files are newer than last import
2. **Download**: SCP download from remote server to temp directory
3. **Extraction**: Extract ZIP files to find database dumps
4. **Database Import**: Import SQL dumps using mysql command-line tool
5. **Cache Cleanup**: Flush all Redis cache after successful import
6. **Cleanup**: Save metadata, remove temp files if configured

Import completion triggers registered callbacks (e.g., Kader-Planung auto-execution).

## Windows Development

This project has comprehensive Windows support:
- **setup-windows.ps1**: Automated environment setup (installs Go, Git, tools)
- **build.ps1**: Full feature parity with Makefile
- **quickstart-windows.bat**: Interactive setup guide
- See README-WINDOWS.md for detailed Windows instructions

## Production Deployment

- Use `ENVIRONMENT=production` in .env
- Enable HTTPS with `ENABLE_HTTPS=true` and provide cert/key files
- Configure Redis for caching (recommended)
- Use reverse proxy (Nginx) for SSL termination
- Ensure database user has SELECT-only permissions
- Health check endpoint: `/health`

## Critical Bug History (from TASKS.md)

Understanding past bugs helps avoid repeating them:

1. **Wrong Database JOIN Bug**: Used `tournament` instead of `tournamentmaster` table, showing completely wrong tournament data for all players. Always use `tournamentmaster` for evaluation JOINs.

2. **N+1 Query Performance Issue**: Service layer made separate API calls for each evaluation to get tournament details. Fixed by including tournament data in single optimized query with proper SELECT.

3. **Character Encoding Issues**: Mixing utf8mb3 vs utf8mb4 caused missing players. Always use utf8mb4 and include UTF-8 BOM in CSV exports.

4. **Browser Caching Problems**: Frontend cached API responses causing wrong data display. Fixed with cache-control headers and timestamp-based cache busting.

5. **Import Service Missing Backup File**: Portal64_BDW database import failed when backup file missing on remote server. Always verify both database backup files exist before import.

## Debugging Tips

### Common Issues
- **"Table doesn't exist" errors**: Check if Portal64_BDW database was imported successfully. Verify both `mvdsb-*.zip` and `portal64_bdw_*.zip` exist on remote server.
- **Wrong tournament data**: Verify using `tournamentmaster` table, not `tournament`
- **Missing players**: Check charset is utf8mb4, check membership date logic includes future-ending memberships
- **Cache issues**: Import service should flush cache after database updates
- **Encoding problems in CSV**: Ensure UTF-8 BOM is written at file start

### Useful SQL Queries for Debugging
```sql
-- Check table structure
SHOW TABLES FROM mvdsb;
SHOW TABLES FROM portal64_bdw;

-- Verify tournament table differences
SELECT COUNT(*) FROM portal64_bdw.tournament;        -- 67,802 specific instances
SELECT COUNT(*) FROM portal64_bdw.tournamentmaster;  -- 62,601 master records

-- Check player membership
SELECT * FROM mvdsb.mitgliedschaft
WHERE organisation = (SELECT id FROM mvdsb.organisation WHERE vkz = 'C0101')
  AND spielernummer = 1014
  AND (bis IS NULL OR bis > CURDATE());

-- Verify evaluation data
SELECT COUNT(*) FROM portal64_bdw.evaluation;  -- Should be ~2.8M records
SELECT COUNT(*) FROM portal64_bdw.game;        -- Should be ~7.8M records
```

## Git Workflow

- Default branch: `master`
- Recent commits show work on Kader-Planung/Somatogramm integration
- Use descriptive commit messages following the pattern: "action: description"

## Performance Optimization Notes

- **Query Optimization**: Use single queries with JOINs instead of N+1 query patterns
- **Cache Strategy**: Cache frequently accessed data (player profiles, club info) with 15-minute TTL
- **Concurrency**: Kader-Planung uses concurrent processing (default: 16 workers, configurable)
- **Database Connections**: Connection pool configured with 10 idle / 100 max open connections per database
