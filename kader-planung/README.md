# Kader-Planung Application

A standalone Golang application that generates comprehensive player roster reports by consuming the Portal64 REST API. The application produces detailed CSV/JSON/Excel reports with historical player data for club management purposes.

## Features

- **Multi-format Export**: Generate reports in CSV, JSON, or Excel format
- **Club Filtering**: Filter clubs by ID prefix (e.g., `C0*` for all clubs starting with C0)
- **Concurrent Processing**: Configurable worker pool for efficient data collection
- **Resume Functionality**: Automatically resume interrupted runs using checkpoint files
- **Historical Analysis**: Calculate DWZ ratings from 12 months ago, recent game statistics, and success rates
- **Progress Tracking**: Real-time progress bars and detailed logging
- **Error Resilience**: Continue processing when individual items fail

## Installation

### Prerequisites

- Go 1.21 or higher
- Access to Portal64 API (default: `http://localhost:8080`)

### Build from Source

1. Clone the repository:
```bash
git clone <repository-url>
cd kader-planung
```

2. Install dependencies:
```bash
.\build.bat deps
```

3. Build the application:
```bash
.\build.bat build
```

The compiled binary will be available at `.\bin\kader-planung.exe`.

## Usage

### Basic Usage

```bash
# Generate report for all clubs starting with C0
.\bin\kader-planung.exe --club-prefix="C0"

# Generate Excel report for all clubs
.\bin\kader-planung.exe --output-format=excel

# Use higher concurrency for faster processing
.\bin\kader-planung.exe --club-prefix="C03" --concurrency=5
```

### Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--club-prefix` | Filter clubs by ID prefix | All clubs |
| `--output-format` | Output format (csv\|json\|excel) | csv |
| `--output-dir` | Output directory | Current directory |
| `--concurrency` | Number of concurrent API requests | 1 |
| `--resume` | Resume from previous run | false |
| `--checkpoint-file` | Custom checkpoint file path | Auto-generated |
| `--api-base-url` | Portal64 API base URL | http://localhost:8080 |
| `--timeout` | API request timeout (seconds) | 30 |
| `--verbose` | Enable detailed logging | false |

### Resume Functionality

If a run is interrupted, you can resume it:

```bash
# Resume with same parameters
.\bin\kader-planung.exe --resume

# Resume with different output format
.\bin\kader-planung.exe --resume --output-format=excel
```

### Examples

```bash
# Generate CSV report for all clubs starting with C0
.\bin\kader-planung.exe --club-prefix="C0" --verbose

# Generate Excel report with higher concurrency
.\bin\kader-planung.exe --output-format=excel --concurrency=10 --output-dir="./reports"

# Resume interrupted run and change to JSON format
.\bin\kader-planung.exe --resume --output-format=json

# Process specific club prefix with custom API URL
.\bin\kader-planung.exe --club-prefix="C0327" --api-base-url="https://api.example.com" --timeout=60
```

## Output Format

The application generates reports with the following columns:

| Column | Description | Source |
|--------|-------------|--------|
| club_name | Full club name | Club profile |
| club_id | Club ID (e.g., C0327) | Club profile |
| player_id | Player ID (e.g., C0327-297) | Player profile |
| lastname | Player's last name | Player profile |
| firstname | Player's first name | Player profile |
| birthyear | Player's birth year | Player profile |
| current_dwz | Current DWZ rating | Player profile |
| dwz_12_months_ago | DWZ rating ~12 months ago | Rating history analysis |
| games_last_12_months | Number of games played | Rating history analysis |
| success_rate_last_12_months | Win percentage (0-100) | Rating history analysis |

### Output Files

Files are named using the pattern:
`kader-planung-{prefix}-{timestamp}.{extension}`

Examples:
- `kader-planung-C0-20240811-143022.csv`
- `kader-planung-all-20240811-143022.xlsx`
- `kader-planung-C03-20240811-143022.json`

## Data Analysis

### Historical DWZ Calculation

The application finds the DWZ rating closest to (but before) 12 months ago from the current date. If no historical data is available, the field shows `DATA_NOT_AVAILABLE`.

### Games and Success Rate

- **Games Count**: Total number of rated games played in the last 12 months
- **Success Rate**: Calculated as `(Wins + 0.5 × Draws) / Total Games × 100`
- Missing data is marked as `DATA_NOT_AVAILABLE`

## Logging and Monitoring

The application provides comprehensive logging:

- **INFO**: Progress updates, milestones
- **WARN**: Recoverable errors, missing data
- **ERROR**: API failures, processing errors  
- **DEBUG**: Detailed processing information (with `--verbose`)

Example log output:
```
[INFO] Starting Kader-Planung data collection...
[INFO] Found 150 clubs matching prefix 'C0'
[INFO] Processing with concurrency: 5
[INFO] Progress: Club 75/150 (50%) - Players: 2,450/estimated 5,000
[WARN] Player C0327-297: Historical data incomplete, using DATA_NOT_AVAILABLE
[INFO] Export complete: kader-planung-C0-20240811-143022.csv (5,127 players)
```

## Error Handling

The application is designed to be resilient:

- **Individual Player Failures**: Logged but processing continues
- **API Timeouts**: Retry with exponential backoff
- **Missing Data**: Marked as `DATA_NOT_AVAILABLE` in output
- **Critical Failures**: Checkpoint saved for resume capability

## Performance

### Optimization Tips

1. **Concurrency**: Increase `--concurrency` for faster processing (recommended: 5-10)
2. **Network**: Ensure stable connection to the API
3. **Disk Space**: Large datasets may require significant temporary storage
4. **Memory**: Application uses streaming processing to minimize memory usage

### Typical Performance

- **Small datasets** (< 50 clubs): 1-5 minutes
- **Medium datasets** (50-200 clubs): 10-30 minutes  
- **Large datasets** (> 200 clubs): 30+ minutes

## Development

### Project Structure

```
kader-planung/
├── cmd/kader-planung/     # Main application entry point
├── internal/
│   ├── api/              # API client
│   ├── models/           # Data models
│   ├── processor/        # Core processing logic
│   ├── export/           # Export functionality
│   └── resume/           # Resume/checkpoint management
├── docs/                 # Documentation
├── bin/                  # Built binaries
├── build.bat             # Build script
├── go.mod               # Go module definition
└── README.md            # This file
```

### Build Commands

```bash
# Install dependencies
.\build.bat deps

# Build application
.\build.bat build

# Run tests
.\build.bat test

# Clean build artifacts
.\build.bat clean

# Run application directly
.\build.bat run --club-prefix="C0"
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./internal/processor

# Run with coverage
go test -cover ./...
```

## Troubleshooting

### Common Issues

1. **API Connection Failed**
   - Verify API URL with `--api-base-url`
   - Check network connectivity
   - Ensure Portal64 API is running

2. **Out of Memory**
   - Reduce `--concurrency` value
   - Process smaller club prefixes
   - Ensure sufficient system memory

3. **Slow Performance**
   - Increase `--concurrency` (try 5-10)
   - Check API response times
   - Verify database performance

4. **Resume Not Working**
   - Ensure checkpoint file exists
   - Verify configuration compatibility
   - Check file permissions

### Debug Mode

Enable verbose logging for troubleshooting:
```bash
.\bin\kader-planung.exe --verbose --club-prefix="C0327"
```

## License

[Add license information here]

## Contributing

[Add contribution guidelines here]

## Support

For issues and questions:
- Check the logs with `--verbose` flag
- Review the troubleshooting section
- [Add contact/support information]