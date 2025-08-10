# Log Rotation Implementation

## Overview

Portal64API now includes comprehensive log rotation functionality using the `lumberjack` library for automatic file rotation and management.

## Features Implemented

✅ **Dual Log Files**:
- Main application logs: `./logs/portal64api.log`
- Import service logs: `./logs/portal64api-import.log` (with `[IMPORT]` prefix)

✅ **Automatic Log Rotation**:
- Max file size: 100MB (configurable)
- Max backups: 5 files (configurable) 
- Max age: 30 days (configurable)
- Compression: enabled (configurable)

✅ **Environment-Based Output**:
- Development: Logs to both file and stdout
- Production: Logs to file only

✅ **Log Levels**:
- Configurable log level (DEBUG, INFO, WARN, ERROR, FATAL)
- Default: INFO level

## Configuration

All logging settings are configured via environment variables in `.env`:

```env
# Logging Configuration
LOG_ENABLED=true
LOG_LEVEL=info
LOG_FILE_PATH=./logs/portal64api.log
LOG_IMPORT_FILE_PATH=./logs/portal64api-import.log
LOG_MAX_SIZE_MB=100
LOG_MAX_BACKUPS=5
LOG_MAX_AGE_DAYS=30
LOG_COMPRESS=true
```

## File Structure

```
./logs/
├── portal64api.log                    # Main application logs
├── portal64api-import.log            # Import service logs
├── portal64api-2025-08-09.log.gz     # Rotated/compressed backups
├── portal64api-2025-08-08.log.gz
└── ...
```

## Implementation Details

### Components Added:

1. **Configuration** (`internal/config/config.go`):
   - Added `LoggingConfig` struct with all log rotation settings

2. **Logging Package** (`internal/logging/setup.go`):
   - `SetupLogging()` - Initializes main application logging with rotation
   - `CreateImportLogger()` - Creates separate logger for import service
   - Automatic directory creation and log rotation setup

3. **Main Application** (`cmd/server/main.go`):
   - Initializes logging system on startup
   - Creates separate logger for import service

4. **Dependencies**:
   - Added `gopkg.in/natefinch/lumberjack.v2` for log rotation

### Log Format

```
# Main Application Logs
2025/08/10 16:25:25.671779 Logging system initialized
2025/08/10 16:25:25.699814 Successfully connected to MVDSB database

# Import Service Logs  
[IMPORT] 2025/08/10 16:25:25.702488 Starting import service...
[IMPORT] 2025/08/10 16:25:25.703681 Import service started with schedule: 0 20 * * *
```

## Benefits

- **Automatic Management**: No manual log file cleanup needed
- **Space Efficient**: Automatic compression of old log files
- **Separation**: Import service logs are isolated for easier troubleshooting
- **Configurable**: All rotation settings can be tuned via environment variables
- **Production Ready**: Minimal performance impact, battle-tested rotation library

## Rotation Behavior

- When log file reaches 100MB, it's rotated to `portal64api-YYYY-MM-DD.log`
- Old rotated files are compressed to `.gz` format
- Only the 5 most recent backup files are kept
- Files older than 30 days are automatically deleted

## Usage

No code changes required for existing log statements. All existing `log.Printf()`, `log.Println()`, etc. calls work unchanged and are automatically rotated.

The import service automatically gets its own rotating log file separate from the main application.
