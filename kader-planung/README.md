# Kader-Planung Application

A standalone Golang application that generates comprehensive player roster reports with somatogram percentiles by consuming the Portal64 REST API. The application produces detailed CSV reports with historical player data and Germany-wide percentile rankings for chess club management.

## Features

- **Somatogram Integration**: Includes Germany-wide percentile rankings by age and gender
- **Historical Analysis**: DWZ ratings from 12 months ago, recent game statistics, and success rates
- **Club Filtering**: Filter clubs by ID prefix (e.g., `C0*` for all clubs starting with C0)
- **Concurrent Processing**: Configurable worker pool for efficient data collection
- **Resume Functionality**: Automatically resume interrupted runs using checkpoint files
- **Progress Tracking**: Real-time progress bars and detailed logging
- **Error Resilience**: Continue processing when individual items fail
- **German Excel Compatible**: CSV output with semicolon separators for direct Excel import

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

## Usage

### Basic Usage

```bash
# Generate report for all clubs starting with C0
.\bin\kader-planung.exe --club-prefix="C0"

# Generate report for all German chess clubs (with somatogram percentiles)
.\bin\kader-planung.exe

# Specify custom output directory
.\bin\kader-planung.exe --output-dir="./reports" --club-prefix="C03"

# Resume from previous interrupted run
.\bin\kader-planung.exe --resume --club-prefix="C0"
```

### Available Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--club-prefix` | `""` | Filter clubs by ID prefix (e.g., `C0` for C0327, C0401, etc.) |
| `--output-dir` | `"."` | Output directory for generated files |
| `--concurrency` | `CPU cores` | Number of concurrent API requests |
| `--resume` | `false` | Resume from previous run using checkpoint file |
| `--checkpoint-file` | `""` | Custom checkpoint file path |
| `--api-base-url` | `"http://localhost:8080"` | Base URL for Portal64 API |
| `--timeout` | `30` | API request timeout in seconds |
| `--min-sample-size` | `10` | Minimum sample size for somatogram percentile accuracy |
| `--verbose` | `false` | Enable detailed logging |
### Resume Functionality

If a run is interrupted, you can resume it:

```bash
# Resume with same parameters
.\bin\kader-planung.exe --resume

# Resume with custom checkpoint file
.\bin\kader-planung.exe --resume --checkpoint-file="custom-checkpoint.json"
```

### Examples

```bash
# Generate CSV report for all clubs starting with C0
.\bin\kader-planung.exe --club-prefix="C0" --verbose

# Generate report with higher concurrency and custom output directory
.\bin\kader-planung.exe --concurrency=10 --output-dir="./reports"

# Resume interrupted run
.\bin\kader-planung.exe --resume

# Process specific club prefix with custom API URL and higher sample size threshold
.\bin\kader-planung.exe --club-prefix="C0327" --api-base-url="https://api.example.com" --timeout=60 --min-sample-size=50
```

## Output Format

The application generates CSV reports with the following columns:

| Column | Description | Source |
|--------|-------------|--------|
| club_id_prefix1 | First character of club ID (e.g., "C") | Club filtering |
| club_id_prefix2 | First 2 characters of club ID (e.g., "C0") | Club filtering |
| club_id_prefix3 | First 3 characters of club ID (e.g., "C03") | Club filtering |
| club_name | Full club name | Club profile |
| club_id | Club ID (e.g., C0327) | Club profile |
| player_id | Player ID (e.g., C0327-297) | Player profile |
| pkz | Player code number | Player profile |
| lastname | Player's last name | Player profile |
| firstname | Player's first name | Player profile |
| birthyear | Player's birth year | Player profile |
| gender | Player's gender (m/f) | Player profile |
| current_dwz | Current DWZ rating | Player profile |
| list_ranking | Current ranking in club | Player profile |
| dwz_12_months_ago | DWZ rating ~12 months ago | Rating history analysis |
| games_last_12_months | Number of games played | Rating history analysis |
| success_rate_last_12_months | Success rate percentage (0-100) | Rating history analysis |
| somatogram_percentile | Germany-wide percentile by age/gender | Somatogram analysis |
| dwz_age_relation | DWZ relative to age group | Somatogram analysis |

### Output Files

Files are named using the pattern:
`kader-planung-{prefix}-{timestamp}.csv`

Examples:
- `kader-planung-C0-20240811-143022.csv`
- `kader-planung-all-20240811-143022.csv`
- `kader-planung-C03-20240811-143022.csv`

## Data Analysis

### Historical DWZ Calculation

The application finds the DWZ rating closest to (but before) 12 months ago from the current date. If no historical data is available, the field shows `DATA_NOT_AVAILABLE`.

### Games and Success Rate

- **Games Count**: Total number of rated games played in the last 12 months
- **Success Rate**: Calculated as `(Wins + 0.5 × Draws) / Total Games × 100`
- Missing data is marked as `DATA_NOT_AVAILABLE`

### Somatogram Percentile Calculation

The `somatogram_percentile` column provides Germany-wide percentile rankings:

#### **Data Collection Scope**
- **Always processes ALL German players** (~50,000+ players from ~4,300 clubs)
- **Germany-wide percentiles** regardless of `--club-prefix` filter
- **Complete dataset** ensures meaningful and accurate percentile comparisons

#### **Percentile Calculation Method**
- **Age-Gender Grouping**: Players grouped by birth year and gender (m/f)
- **DWZ-based Ranking**: Percentiles calculated from DWZ ratings within each group
- **Statistical Interpolation**: Uses precise interpolation for accurate percentile values
- **Sample Size Threshold**: Groups with fewer than `--min-sample-size` (default: 10) players show `DATA_NOT_AVAILABLE`

#### **Output Format**
- **Floating-point values**: Percentiles displayed with one decimal place (e.g., `96.4`, `67.2`, `89.0`)
- **Range**: Valid percentiles are 0.0-100.0
- **Missing data**: Shows `DATA_NOT_AVAILABLE` for insufficient sample sizes or missing birth year

#### **Example Interpretation**
- `somatogram_percentile: 96.4` → Player's DWZ is higher than 96.4% of German players in their age/gender group
- `somatogram_percentile: 23.1` → Player's DWZ is higher than 23.1% of German players in their age/gender group
- `somatogram_percentile: DATA_NOT_AVAILABLE` → Fewer than 10 players in this age/gender group for reliable calculation

#### **Sample Size Considerations**
- **Default threshold (10)**: Balances accuracy with data availability for most age groups
- **Lower thresholds (5-9)**: Include more age groups but may have less statistical reliability
- **Higher thresholds (20-50)**: More statistically reliable but exclude smaller age groups
- **Very high thresholds (100+)**: Only include age groups with substantial player populations

#### **Performance Impact**
- **Additional processing time**: ~30-60 seconds for complete German dataset collection
- **Memory usage**: ~200-300MB for full player dataset
- **API calls**: ~500-800 API requests to collect all German club data

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