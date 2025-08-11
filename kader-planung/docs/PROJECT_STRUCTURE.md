# Kader-Planung Project Structure

## Directory Layout

```
kader-planung/
├── cmd/
│   └── kader-planung/
│       └── main.go                 # Main application entry point
├── internal/
│   ├── api/
│   │   ├── client.go              # Portal64 API client
│   │   └── client_test.go         # API client tests
│   ├── models/
│   │   └── models.go              # Data structures and models
│   ├── processor/
│   │   ├── processor.go           # Core processing logic
│   │   └── processor_test.go      # Processor tests
│   ├── export/
│   │   ├── export.go              # Export functionality (CSV/JSON/Excel)
│   │   └── export_test.go         # Export tests
│   └── resume/
│       ├── manager.go             # Checkpoint and resume management
│       └── manager_test.go        # Resume manager tests
├── docs/
│   ├── kader-planung-design.md    # High-level design document
│   └── api-reference.md           # API usage reference
├── bin/                           # Built binaries (generated)
│   └── kader-planung.exe         # Windows executable
├── build.bat                      # Build script for Windows
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── README.md                     # Project documentation
└── PROJECT_STRUCTURE.md         # This file
```

## Component Overview

### Main Application (`cmd/kader-planung/`)
- **main.go**: Application entry point with CLI interface using Cobra
- Handles command-line arguments, configuration, and orchestrates the entire workflow
- Provides comprehensive help and usage information

### API Client (`internal/api/`)
- **client.go**: HTTP client wrapper for Portal64 REST API
- Handles all API communication including:
  - Club search and profile retrieval  
  - Player search and profile retrieval
  - Rating history fetching
  - Pagination and batch processing
  - Error handling and timeouts

### Data Models (`internal/models/`)
- **models.go**: Core data structures including:
  - `Club`: Chess club information
  - `Player`: Player profile data
  - `RatingHistory`: Historical rating data points
  - `KaderPlanungRecord`: Final export record structure
  - `HistoricalAnalysis`: Calculated statistics
  - Various result and error structures

### Processing Engine (`internal/processor/`)
- **processor.go**: Core business logic including:
  - Concurrent processing using worker pools
  - Historical data analysis algorithms
  - Progress tracking and reporting
  - Error handling and recovery
  - Checkpoint integration

### Export Engine (`internal/export/`)
- **export.go**: Multi-format export functionality:
  - CSV export with proper escaping
  - JSON export with structured format
  - Excel export with formatting
  - Data validation and sanitization

### Resume Manager (`internal/resume/`)
- **manager.go**: Checkpoint and resume functionality:
  - State persistence to JSON files
  - Progress tracking
  - Configuration compatibility validation
  - Atomic checkpoint updates
  - Recovery mechanisms

## Build and Deployment

### Build Artifacts
- `bin/kader-planung.exe`: Windows executable (primary target)
- `bin/kader-planung-linux`: Linux binary (cross-compilation)
- Checkpoint files: `kader-planung-checkpoint-*.json`
- Output files: `kader-planung-{prefix}-{timestamp}.{ext}`

### Build Process
1. **Dependency Management**: `go mod tidy` ensures all dependencies are available
2. **Compilation**: `go build` with optimization flags (`-ldflags="-s -w"`)
3. **Testing**: `go test ./...` runs all unit and integration tests
4. **Cross-compilation**: Support for Windows, Linux, and macOS

## Data Flow

### Phase 1: Discovery
1. Fetch all clubs from API
2. Filter by club prefix if specified
3. Initialize progress tracking

### Phase 2: Processing
1. Create worker pool with configurable concurrency
2. Process clubs concurrently:
   - Fetch club details
   - Retrieve all club players
   - For each player:
     - Get player profile
     - Fetch rating history
     - Perform historical analysis
3. Handle errors gracefully, continue processing

### Phase 3: Analysis
1. Calculate DWZ rating 12 months ago
2. Count games in last 12 months
3. Calculate success rate percentage
4. Handle missing data appropriately

### Phase 4: Export
1. Aggregate all processed records
2. Validate data integrity
3. Export to requested format(s)
4. Generate timestamped output files

## Configuration

### Command-Line Arguments
- Runtime behavior controlled entirely through CLI flags
- No configuration files required (though supported for future enhancement)
- Environment variable override capability

### Default Values
- API Base URL: `http://localhost:8080`
- Concurrency: 1 (conservative default)
- Output Format: CSV
- Timeout: 30 seconds
- Output Directory: Current working directory

## Error Handling Strategy

### Levels of Resilience
1. **Individual Player Failures**: Log and continue
2. **API Timeouts**: Retry with exponential backoff  
3. **Data Validation Errors**: Mark as unavailable, continue
4. **Critical System Failures**: Save checkpoint, exit gracefully

### Logging Strategy
- Structured logging using logrus
- Configurable log levels (INFO, WARN, ERROR, DEBUG)
- Progress indicators for long-running operations
- Comprehensive error context

## Testing Strategy

### Unit Tests
- Each package has corresponding test files
- Mock API responses for isolated testing
- Data validation and transformation testing
- Historical analysis algorithm verification

### Integration Tests  
- End-to-end workflow testing
- Resume functionality validation
- Export format verification
- Performance and memory usage testing

### Test Data
- Synthetic test records for development
- Mock API client for offline testing
- Sample checkpoint files for resume testing

## Performance Considerations

### Memory Management
- Streaming data processing to minimize memory footprint
- Efficient data structures for large datasets
- Garbage collection optimization hints
- Configurable batch sizes

### Network Optimization
- HTTP connection pooling and keep-alive
- Concurrent request processing
- Request/response compression
- Retry mechanisms with backoff

### Disk I/O
- Atomic file operations for data integrity
- Buffered writing for large exports
- Temporary file usage with cleanup
- Checkpoint file optimization

## Security Considerations

### Data Handling
- No persistent storage of sensitive information
- Memory cleanup after processing
- Secure temporary file handling
- Input validation and sanitization

### API Security
- Support for authentication headers (future)
- HTTPS enforcement capabilities
- Rate limiting respect (though not required)
- Error message sanitization

## Future Enhancement Points

### Extensibility
- Pluggable export formats
- Custom analysis modules
- Configurable data sources
- Template-based reporting

### Features
- Web-based interface
- Scheduled execution
- Email report delivery
- Advanced filtering options
- Data quality validation
- Incremental update support

This structure provides a solid foundation for a maintainable, scalable, and robust application that can be extended and modified as requirements evolve.