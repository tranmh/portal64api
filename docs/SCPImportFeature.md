# SCP Database Import Feature - High Level Design

## Overview

This document outlines the design for an SCP-based database import feature that downloads, extracts, and imports database content from portal.svw.info on a daily basis. The system is designed for simplicity and reliability in a read-only environment.

## Architecture Overview

```
┌────────────────────────────────────────────────────────────┐
│                    Portal64API                             │
│                                                            │
│  ┌─────────────────┐                                       │
│  │   Main App      │ ◄─── Loosely Coupled Integration      │
│  │   (existing)    │                                       │
│  └─────────────────┘                                       │
│           │                                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Import Service                         │   │
│  │                                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │    SCP      │  │    ZIP      │  │  Database   │  │   │
│  │  │ Downloader  │─▶│ Extractor   │─▶│  Importer  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │                                          │          │   │
│  │  ┌─────────────┐  ┌─────────────┐        ▼          │   │
│  │  │   Status    │  │   Cache     │  ┌─────────────┐  │   │
│  │  │  Tracker    │  │   Cleaner   │  │   Logger    │  │   │
│  │  │ (In-Memory) │  │             │  │             │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────┐   │
│                                                        │   │
│  HTTP API Routes:                                      │   │
│  ├─ GET /api/v1/import/status                          │   │
│  ├─ POST /api/v1/import/start                          │   │
│  └─ GET /api/v1/import/logs                            │   │
└────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Import Service (`internal/services/import_service.go`)

**Responsibilities:**
- Orchestrates the complete import process
- Manages cron-style scheduling
- Maintains in-memory status tracking
- Handles API load detection and delay logic
- Provides status information for API endpoints

**Key Methods:**
```go
type ImportService interface {
    Start() error                              // Start the scheduled service
    Stop() error                               // Stop the scheduled service
    TriggerManualImport() error                // Manual import trigger
    GetStatus() *ImportStatus                  // Get current status
    GetLogs() []ImportLogEntry                 // Get recent log entries
    IsAPIUnderHeavyLoad() bool                 // Check API load
}
```

### 2. SCP Downloader (`internal/importers/scp_downloader.go`)

**Responsibilities:**
- Establishes SCP connection using username/password authentication
- Lists and checks file metadata (timestamps, sizes) for freshness comparison
- Downloads files matching configurable wildcard patterns
- Reports download progress
- Handles network timeouts and retries

**Configuration Support:**
```go
type SCPConfig struct {
    Host         string   `yaml:"host"`         // portal.svw.info
    Username     string   `yaml:"username"`     // username
    Password     string   `yaml:"password"`     // password
    Port         int      `yaml:"port"`         // 22
    RemotePath   string   `yaml:"remote_path"`  // /data/exports/
    FilePatterns []string `yaml:"file_patterns"` // ["mvdsb_*.zip", "portal64_bdw_*.zip"]
    Timeout      string   `yaml:"timeout"`      // 300s
}

type FileMetadata struct {
    Filename    string    `json:"filename"`
    Size        int64     `json:"size"`
    ModTime     time.Time `json:"mod_time"`
    Checksum    string    `json:"checksum,omitempty"`
}
```

### 3. File Freshness Checker (`internal/importers/freshness_checker.go`)

**Responsibilities:**
- Compares remote file metadata with last successful import
- Stores and retrieves import metadata from local file
- Determines if import should proceed based on file freshness
- Provides detailed comparison results for logging

**Features:**
- Persistent metadata storage in JSON file
- Multiple comparison criteria (timestamp, size, checksum)
- Graceful handling of missing metadata (first import)
- Protection against broken or corrupted remote files

### 4. ZIP Extractor (`internal/importers/zip_extractor.go`)

**Responsibilities:**
- Extracts password-protected zip files
- Validates extracted content structure
- Reports extraction progress

**Features:**
- Single password for both zip files
- Automatic file type detection
- Progress reporting for large archives

### 5. Database Importer (`internal/importers/database_importer.go`)

**Responsibilities:**
- Performs complete database drop and recreation
- Imports SQL dumps or data files
- Validates import completion
- Reports import progress

**Strategy:**
```sql
-- Simple drop and recreate approach
DROP DATABASE IF EXISTS mvdsb;
CREATE DATABASE mvdsb;
-- Import data
SOURCE /path/to/extracted/mvdsb_dump.sql;
```

### 6. Status Tracker (`internal/importers/status_tracker.go`)

**Responsibilities:**
- Maintains in-memory import status
- Thread-safe status updates
- Provides status snapshots for API endpoints

**No Database Persistence:**
- Status stored only in memory
- Logs written to structured log files
- Service restart clears status (acceptable for this use case)

### 7. Cache Cleaner Integration

**Integration with Existing Cache:**
- Uses existing `internal/cache/redis_service.go`
- Calls `FlushAll()` or selective key pattern deletion
- Triggered only after successful database import

## API Endpoints

### GET /api/v1/import/status
**Purpose:** Asynchronous status reporting
```json
{
  "status": "idle|running|success|failed|skipped",
  "progress": 75,
  "current_step": "importing_database_portal64_bdw",
  "started_at": "2025-08-06T02:00:00Z",
  "completed_at": null,
  "last_success": "2025-08-05T02:00:00Z", 
  "next_scheduled": "2025-08-07T02:00:00Z",
  "retry_count": 1,
  "max_retries": 3,
  "error": null,
  "skip_reason": null,
  "files_info": {
    "remote_files": [
      {
        "filename": "mvdsb_20250806.zip",
        "size": 15728640,
        "mod_time": "2025-08-06T01:30:00Z",
        "is_newer": true
      }
    ],
    "last_imported": [
      {
        "filename": "mvdsb_20250805.zip", 
        "size": 15624832,
        "mod_time": "2025-08-05T01:30:00Z"
      }
    ],
    "downloaded": ["mvdsb_20250806.zip", "portal64_bdw_20250806.zip"],
    "extracted": ["mvdsb_dump.sql", "portal64_bdw_dump.sql"],
    "imported": ["mvdsb"]
  }
}
```

### POST /api/v1/import/start
**Purpose:** Manual import trigger
```json
// Request: Empty body or {}
// Response:
{
  "message": "Manual import started",
  "started_at": "2025-08-06T14:30:00Z"
}
```

### GET /api/v1/import/logs
**Purpose:** Recent log entries
```json
{
  "logs": [
    {
      "timestamp": "2025-08-06T02:00:00Z",
      "level": "INFO",
      "message": "Import process started",
      "step": "initialization"
    },
    {
      "timestamp": "2025-08-06T02:01:30Z", 
      "level": "INFO",
      "message": "Downloaded mvdsb_20250806.zip (15.2MB)",
      "step": "download"
    }
  ]
}
```

## Configuration Design

### Main Configuration (`configs/config.yaml`)

```yaml
import:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM (crontab format)
  
  # Load-based delay configuration
  load_check:
    enabled: true
    delay_duration: "1h"        # Delay 1 hour if under heavy load
    max_delays: 3               # Maximum 3 delays (total 3 hours)
    load_threshold: 100         # Concurrent requests threshold
  
  # SCP Configuration
  scp:
    host: "portal.svw.info"
    port: 22
    username: "portal64user"
    password: "${IMPORT_SCP_PASSWORD}"  # From environment
    remote_path: "/data/exports/"
    file_patterns:
      - "mvdsb_*.zip"             # Matches mvdsb_20250806.zip
      - "portal64_bdw_*.zip"      # Matches portal64_bdw_20250806.zip
    timeout: "300s"
  
  # ZIP Configuration
  zip:
    password: "${IMPORT_ZIP_PASSWORD}"    # Single password for both files
    extract_timeout: "60s"
  
  # Local Storage
  storage:
    temp_dir: "./data/import/temp"        # Configurable download location
    metadata_file: "./data/import/last_import.json"  # Last successful import metadata
    cleanup_on_success: true              # Clean old files after success
    keep_failed_files: true               # Keep files for debugging on failure
  
  # File Freshness Check
  freshness:
    enabled: true                         # Enable file freshness checking
    compare_timestamp: true               # Compare file modification time
    compare_size: true                    # Compare file size
    compare_checksum: false               # Optional: Compare file checksum (slower)
    skip_if_not_newer: true               # Skip import if files are not newer
  
  # Database Import
  database:
    import_timeout: "600s"
    target_databases:
      - name: "mvdsb"
        file_pattern: "mvdsb_*"
      - name: "portal64_bdw" 
        file_pattern: "portal64_bdw_*"
  
  # Error Handling
  retry:
    enabled: true
    max_attempts: 2               # Original + 1 retry = 2 total
    retry_delay: "5m"             # Wait 5 minutes before retry
    fail_fast: true               # Don't continue on critical errors
```

### Environment Variables

```bash
# Import Configuration
IMPORT_ENABLED=true
IMPORT_SCP_PASSWORD=secret_password
IMPORT_ZIP_PASSWORD=zip_password
IMPORT_SCHEDULE="0 2 * * *"

# Import Storage
IMPORT_TEMP_DIR=./data/import/temp
IMPORT_CLEANUP_ON_SUCCESS=true
```

## File Freshness Checking

### Metadata Storage Format

The last successful import metadata is stored in `./data/import/last_import.json`:

```json
{
  "last_import": {
    "timestamp": "2025-08-05T02:15:30Z",
    "success": true,
    "files": [
      {
        "filename": "mvdsb_20250805.zip",
        "pattern": "mvdsb_*.zip",
        "size": 15624832,
        "mod_time": "2025-08-05T01:30:00Z",
        "checksum": "sha256:abc123...",
        "database": "mvdsb",
        "imported": true
      },
      {
        "filename": "portal64_bdw_20250805.zip", 
        "pattern": "portal64_bdw_*.zip",
        "size": 8453120,
        "mod_time": "2025-08-05T01:30:00Z",
        "checksum": "sha256:def456...",
        "database": "portal64_bdw",
        "imported": true
      }
    ]
  }
}
```

### Freshness Comparison Logic

```go
type FreshnessChecker struct {
    config       *ImportConfig
    metadataFile string
    logger       *logrus.Logger
}

func (fc *FreshnessChecker) CheckFreshness(remoteFiles []FileMetadata) (*FreshnessResult, error) {
    lastImport, err := fc.loadLastImportMetadata()
    if err != nil {
        // First import - no metadata exists
        return &FreshnessResult{ShouldImport: true, Reason: "first_import"}, nil
    }
    
    result := &FreshnessResult{
        ShouldImport: false,
        RemoteFiles:  remoteFiles,
        LastImported: lastImport.Files,
        Comparisons:  make([]FileComparison, 0),
    }
    
    for _, remoteFile := range remoteFiles {
        lastFile := fc.findMatchingFile(lastImport.Files, remoteFile)
        comparison := fc.compareFiles(remoteFile, lastFile)
        result.Comparisons = append(result.Comparisons, comparison)
        
        if comparison.IsNewer {
            result.ShouldImport = true
            result.Reason = "newer_files_available"
        }
    }
    
    if !result.ShouldImport {
        result.Reason = "no_newer_files"
    }
    
    return result, nil
}

func (fc *FreshnessChecker) compareFiles(remote FileMetadata, last *FileMetadata) FileComparison {
    comparison := FileComparison{
        RemoteFile: remote,
        LastFile:   last,
        IsNewer:    false,
        Reasons:    make([]string, 0),
    }
    
    if last == nil {
        comparison.IsNewer = true
        comparison.Reasons = append(comparison.Reasons, "file_not_found_in_last_import")
        return comparison
    }
    
    // Compare modification time
    if fc.config.Freshness.CompareTimestamp && remote.ModTime.After(last.ModTime) {
        comparison.IsNewer = true
        comparison.Reasons = append(comparison.Reasons, "newer_timestamp")
    }
    
    // Compare file size
    if fc.config.Freshness.CompareSize && remote.Size != last.Size {
        comparison.IsNewer = true
        comparison.Reasons = append(comparison.Reasons, "different_size")
    }
    
    // Compare checksum (optional, slower)
    if fc.config.Freshness.CompareChecksum && remote.Checksum != "" && 
       last.Checksum != "" && remote.Checksum != last.Checksum {
        comparison.IsNewer = true
        comparison.Reasons = append(comparison.Reasons, "different_checksum")
    }
    
    return comparison
}
```

### Skip Logic Integration

```go
func (s *ImportService) executeImport() error {
    // ... existing code ...
    
    // File freshness check phase
    if s.config.Freshness.Enabled {
        s.updateStatus("checking_file_freshness", 10)
        
        remoteFiles, err := s.downloader.ListFiles()
        if err != nil {
            return fmt.Errorf("failed to list remote files: %w", err)
        }
        
        freshnessResult, err := s.freshnessChecker.CheckFreshness(remoteFiles)
        if err != nil {
            s.logger.Warnf("Freshness check failed, proceeding with import: %v", err)
        } else if !freshnessResult.ShouldImport && s.config.Freshness.SkipIfNotNewer {
            s.updateStatusSkipped(freshnessResult.Reason)
            s.logger.Infof("Import skipped: %s", freshnessResult.Reason)
            return nil // Successfully skipped
        }
        
        s.logger.Infof("Freshness check result: %s", freshnessResult.Reason)
    }
    
    // Continue with download phase...
}
```

## Process Flow

### Daily Scheduled Import

```
1. Cron Trigger (2 AM daily)
   ↓
2. Check if import already running → Exit if yes
   ↓  
3. Check API load → Delay 1h if heavy (max 3 delays)
   ↓
4. Initialize import process
   ↓
5. File Freshness Check Phase
   ├─ Connect to portal.svw.info
   ├─ List files matching patterns
   ├─ Get file timestamps/sizes
   ├─ Compare with last successful import
   └─ Skip if files are not newer → Exit with "no update needed"
   ↓
6. SCP Download Phase
   ├─ Download each matching file
   └─ Verify file integrity
   ↓
7. ZIP Extraction Phase
   ├─ Extract each zip file with password
   ├─ Validate extracted content
   └─ Identify database dump files
   ↓
8. Database Import Phase
   ├─ For each target database:
   │  ├─ DROP DATABASE IF EXISTS
   │  ├─ CREATE DATABASE
   │  ├─ Import SQL dump
   │  └─ Verify import success
   └─ Log import statistics
   ↓
9. Cache Cleanup Phase
   ├─ Clear all Redis cache keys
   └─ Reset cache statistics
   ↓
10. Cleanup & Update Phase
   ├─ Remove temporary files (if success)
   ├─ Update last successful import metadata
   └─ Log completion
```

### Error Handling Strategy

**Fail-Fast Principle:**
- Critical errors immediately abort the process
- Non-critical errors are logged but process continues

**Retry Logic:**
- Single retry attempt after 5-minute delay
- Retry only on recoverable errors (network, temporary file issues)
- No retry on authentication, permission, or data corruption errors

**Error Categories:**
```go
const (
    ErrorCritical     = "critical"      // No retry, abort immediately
    ErrorRecoverable  = "recoverable"   // Retry once
    ErrorWarning      = "warning"       // Log but continue
)
```

**Example Error Handling:**
```go
func (s *ImportService) executeImport() error {
    // Freshness check phase
    if s.config.Freshness.Enabled {
        if freshnessResult, err := s.checkFreshness(); err != nil {
            s.logger.Warnf("Freshness check failed, proceeding anyway: %v", err)
        } else if !freshnessResult.ShouldImport {
            s.updateStatusSkipped(freshnessResult.Reason)
            return nil // Successfully skipped, no error
        }
    }
    
    // Download phase
    if err := s.downloadFiles(); err != nil {
        if IsRecoverableError(err) && s.retryCount == 0 {
            s.scheduleRetry()
            return nil
        }
        return fmt.Errorf("download failed: %w", err)
    }
    
    // Continue with other phases...
}
```

## Load Detection & Delay Logic

### API Load Detection

```go
type LoadDetector interface {
    GetCurrentLoad() int                    // Current concurrent requests
    IsUnderHeavyLoad() bool                // Above threshold?
    GetActiveConnections() int             // Database connections
    GetMemoryUsage() float64               // Memory usage percentage
}

type LoadMetrics struct {
    ConcurrentRequests   int     `json:"concurrent_requests"`
    ActiveDBConnections  int     `json:"active_db_connections"`
    MemoryUsagePercent   float64 `json:"memory_usage_percent"`
    CPUUsagePercent      float64 `json:"cpu_usage_percent"`
}
```

### Delay Strategy

```go
func (s *ImportService) checkLoadAndDelay() error {
    if !s.config.LoadCheck.Enabled {
        return nil
    }
    
    for attempt := 0; attempt < s.config.LoadCheck.MaxDelays; attempt++ {
        if !s.loadDetector.IsUnderHeavyLoad() {
            return nil // Proceed with import
        }
        
        s.logger.Warnf("API under heavy load, delaying import for %v (attempt %d/%d)", 
            s.config.LoadCheck.DelayDuration, attempt+1, s.config.LoadCheck.MaxDelays)
        
        time.Sleep(s.config.LoadCheck.DelayDuration)
    }
    
    return errors.New("max delays exceeded, API still under heavy load")
}
```

## Integration Points

### Loose Coupling with Main Application

The import service integrates loosely with the main application:

1. **Shared Database Configuration:** Uses same MySQL connection settings
2. **Shared Cache Service:** Uses existing Redis cache infrastructure  
3. **Independent API Routes:** Separate route group `/api/v1/import/*`
4. **Optional Dependency:** Main app works normally if import service disabled

### Integration in Main Application

```go
// In cmd/server/main.go
func main() {
    // ... existing initialization ...
    
    // Initialize import service if enabled
    var importService services.ImportService
    if cfg.Import.Enabled {
        importService = services.NewImportService(cfg.Import, db, cache, logger)
        go importService.Start() // Start in background
    }
    
    // Add import routes
    api.SetupImportRoutes(router, importService)
    
    // ... rest of main app ...
}
```

## Logging Strategy

### Structured Logging

```go
type ImportLogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`      // INFO, WARN, ERROR
    Step      string    `json:"step"`       // download, extract, import, cleanup
    Message   string    `json:"message"`
    Error     string    `json:"error,omitempty"`
    Duration  string    `json:"duration,omitempty"`
    FileSize  int64     `json:"file_size,omitempty"`
}
```

### Log Categories

- **Import Lifecycle:** Start, complete, retry, failure
- **Download Progress:** File sizes, transfer rates, completion
- **Extraction Progress:** Archive validation, content discovery
- **Database Operations:** Drop, create, import statistics
- **Performance Metrics:** Operation durations, resource usage

### Log Retention

```yaml
logging:
  import:
    file: "./logs/import.log"
    level: "info" 
    max_size: 100      # MB
    max_backups: 5     # Keep 5 old log files
    max_age: 30        # Keep logs for 30 days
    compress: true     # Compress old logs
```

## Dependencies

### New Go Modules Required

```go
// Add to go.mod
require (
    github.com/pkg/sftp v1.13.6           // SCP/SFTP client
    golang.org/x/crypto v0.11.0           // SSH authentication
    github.com/yeka/zip v0.0.0-20180914125537-d046722c6feb // Password-protected zip
    github.com/robfig/cron/v3 v3.0.1      // Cron scheduling
)
```

### System Dependencies

- **SCP Access:** Network connectivity to portal.svw.info
- **Disk Space:** Sufficient space for zip files and extracted content
- **MySQL Access:** Same credentials as main application
- **Redis Access:** Same instance as main application cache

## File Structure

```
internal/
├── services/
│   └── import_service.go              # Main orchestration service
├── importers/
│   ├── scp_downloader.go              # SCP file downloading
│   ├── freshness_checker.go           # File freshness comparison
│   ├── zip_extractor.go               # Password-protected zip extraction
│   ├── database_importer.go           # Database drop/create/import
│   ├── status_tracker.go              # In-memory status tracking
│   └── load_detector.go               # API load detection
├── api/
│   └── import_handler.go              # HTTP handlers for import API
└── models/
    └── import_models.go               # Import-related data structures

configs/
└── import.yaml                        # Import-specific configuration

data/
└── import/
    ├── temp/                          # Temporary download/extraction
    ├── logs/                          # Import-specific logs
    └── last_import.json               # Last successful import metadata
```

## Security Considerations

### Credential Management

- SCP credentials stored in environment variables
- ZIP passwords stored in environment variables  
- No credentials stored in configuration files or code
- Secure credential rotation capability

### File System Security

- Temporary files created with restricted permissions (0600)
- Automatic cleanup of sensitive temporary files
- Separate directory for import operations
- No shared directories with main application

### Network Security

- SCP connection with username/password authentication
- Connection timeouts to prevent hanging connections
- No permanent connections maintained

## Performance Considerations

### Resource Usage

- **CPU:** ZIP extraction and database imports are CPU-intensive
- **Memory:** Large files loaded into memory during processing
- **Disk I/O:** High during download, extraction, and import phases
- **Network:** Sustained download bandwidth required

### Performance Optimizations

- **Streaming Processing:** Process files as they download when possible
- **Parallel Operations:** Extract and import different databases concurrently
- **Progress Reporting:** Regular status updates for monitoring
- **Resource Monitoring:** Track CPU, memory, and disk usage

### Impact Minimization

- **Scheduled During Low Activity:** Default 2 AM execution
- **Load-Based Delays:** Postpone if API under heavy load
- **Quick Cache Invalidation:** Efficient Redis cache clearing
- **Fail-Fast:** Quick abort on unrecoverable errors

## Testing Strategy

### Unit Tests

- Mock SCP connections for download testing
- Mock file system for extraction testing  
- In-memory databases for import testing
- Status tracker behavior validation

### Integration Tests

- End-to-end import process with test files
- Error scenario testing (network failures, corrupted files)
- Load detection and delay logic validation
- Cache invalidation verification

### Manual Testing Checklist

- [ ] Manual import trigger works
- [ ] Status API returns correct information
- [ ] Log API shows structured entries  
- [ ] File freshness checking works correctly
- [ ] Import skipped when files are not newer
- [ ] Import proceeds when files are newer
- [ ] First import works when no metadata exists
- [ ] Files downloaded successfully
- [ ] ZIP extraction works with password
- [ ] Databases dropped and recreated
- [ ] Cache cleared after import
- [ ] Temporary files cleaned up
- [ ] Last import metadata saved correctly
- [ ] Error handling works as expected
- [ ] Load-based delays function correctly

## Deployment Checklist

### Configuration

- [ ] Import schedule configured
- [ ] SCP credentials provided in environment
- [ ] ZIP password provided in environment
- [ ] Temporary directory exists and writable
- [ ] Database permissions verified
- [ ] Redis cache accessible

### Verification

- [ ] Import service starts without errors
- [ ] Status endpoint returns valid response
- [ ] Manual import trigger works
- [ ] Log files created and writable
- [ ] SCP connectivity tested
- [ ] Database import tested with sample data

This design provides a simple, reliable, and maintainable solution for automated database imports while maintaining loose coupling with the main application and providing comprehensive monitoring capabilities.
