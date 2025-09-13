# Somatogramm (Growth Chart) Feature - Implementation Plan

## Overview
A standalone Golang application that generates DWZ percentile charts by age and gender, creating comprehensive growth analysis reports for chess players. The feature analyzes DWZ distribution patterns across age groups to identify talent development trends.

## Application Architecture

### Core Components

1. **CLI Interface** (`somatogramm/cmd/somatogramm/`)
   - Command-line argument parsing similar to kader-planung
   - Configuration validation
   - Application lifecycle management

2. **API Client** (`somatogramm/internal/api/`)
   - REST API wrapper for Portal64 endpoints
   - HTTP client with timeout handling  
   - Response parsing and error handling

3. **Data Models** (`somatogramm/internal/models/`)
   - Player data structures with age/gender/DWZ
   - Percentile calculation structures
   - Export data structures

4. **Processing Engine** (`somatogramm/internal/processor/`)
   - Age-based data aggregation
   - Percentile calculations (0-100%)
   - Statistical analysis with minimum sample filtering

5. **Export Engine** (`somatogramm/internal/export/`)
   - CSV, JSON format writers
   - Age-grouped percentile tables

6. **Resume Manager** (`somatogramm/internal/resume/`)
   - Progress state persistence (reuse from kader-planung)
   - Recovery mechanisms

## Command-Line Interface

### Usage
```bash
./somatogramm [options]
```

### Arguments
- `--output-format`: Output format (csv|json, default: csv)
- `--output-dir`: Output directory (default: current directory)  
- `--concurrency`: Number of concurrent API requests (default: CPU cores)
- `--api-base-url`: Base URL for Portal64 API (default: http://localhost:8080)
- `--timeout`: API request timeout in seconds (default: 30)
- `--verbose`: Enable detailed logging
- `--min-sample-size`: Minimum players per age/gender (default: 100)

## Data Flow Architecture

### Phase 1: Player Data Collection
1. Fetch all players using `/api/v1/players` API with pagination
2. Filter players with valid DWZ ratings (exclude null/zero DWZ)
3. Calculate current age: `current_year - birth_year`  
4. Map gender using existing `MapGeschlechtToGender()` function:
   - 1 = "m" (male)
   - 0 = "w" (female)  
   - 2 = "d" (divers)

### Phase 2: Age-Gender Grouping
1. Group players by age and gender combination
2. Apply minimum sample size filter (100+ players per age/gender group)
3. Log statistics: total players, valid DWZ players, age groups with sufficient data

### Phase 3: Percentile Calculation
For each valid age/gender group:
1. Sort DWZ values in ascending order
2. Calculate percentiles 0, 1, 2, ..., 99, 100 using interpolation method
3. Handle edge cases (duplicate DWZ values, small sample sizes)

### Phase 4: Export Generation
1. Generate separate files for male/female/divers
2. Create combined overview file
3. Include metadata: generation timestamp, sample sizes, age ranges

## Output Schema

### CSV Structure (somatogramm-male-YYYYMMDD-HHMMSS.csv)
```csv
age,p0,p1,p2,p3,...,p98,p99,p100,sample_size,avg_dwz,median_dwz
6,800,820,840,860,...,1980,2000,2100,150,1200,1180
7,850,870,890,910,...,2050,2080,2200,180,1250,1220
...
```

### JSON Structure (somatogramm-male-YYYYMMDD-HHMMSS.json)
```json
{
  "metadata": {
    "generated_at": "2025-09-13T10:30:00Z",
    "gender": "m",
    "total_players": 45230,
    "valid_age_groups": 42,
    "min_sample_size": 100
  },
  "percentiles": {
    "6": {
      "sample_size": 150,
      "avg_dwz": 1200,
      "median_dwz": 1180,
      "percentiles": {
        "0": 800, "1": 820, "2": 840, ..., "100": 2100
      }
    }
  }
}
```

## Percentile Calculation Algorithm

```go
func CalculatePercentiles(dwzValues []int) map[int]int {
    sort.Ints(dwzValues)
    n := len(dwzValues)
    percentiles := make(map[int]int)
    
    for p := 0; p <= 100; p++ {
        if p == 0 {
            percentiles[p] = dwzValues[0]
        } else if p == 100 {
            percentiles[p] = dwzValues[n-1]
        } else {
            // Linear interpolation method
            rank := float64(p) * float64(n-1) / 100.0
            lower := int(rank)
            upper := lower + 1
            
            if upper >= n {
                percentiles[p] = dwzValues[n-1]
            } else {
                weight := rank - float64(lower)
                percentiles[p] = int(float64(dwzValues[lower])*(1-weight) + float64(dwzValues[upper])*weight)
            }
        }
    }
    return percentiles
}
```

## Integration with Main Server

### Service Integration (`internal/services/somatogramm_service.go`)
- Implement `ImportCompleteCallback` interface
- Automatic execution after successful database imports
- Similar architecture to `KaderPlanungService`

### Configuration (`internal/config/config.go`)
```go
type SomatogrammConfig struct {
    Enabled         bool   `env:"SOMATOGRAMM_ENABLED" envDefault:"true"`
    BinaryPath      string `env:"SOMATOGRAMM_BINARY_PATH" envDefault:"somatogramm/bin/somatogramm.exe"`
    OutputDir       string `env:"SOMATOGRAMM_OUTPUT_DIR" envDefault:"internal/static/demo/somatogramm"`
    APIBaseURL      string `env:"SOMATOGRAMM_API_BASE_URL" envDefault:"http://localhost:8080"`
    OutputFormat    string `env:"SOMATOGRAMM_OUTPUT_FORMAT" envDefault:"csv"`
    MinSampleSize   int    `env:"SOMATOGRAMM_MIN_SAMPLE_SIZE" envDefault:"100"`
    Timeout         int    `env:"SOMATOGRAMM_TIMEOUT" envDefault:"30"`
    Concurrency     int    `env:"SOMATOGRAMM_CONCURRENCY" envDefault:"0"`
    Verbose         bool   `env:"SOMATOGRAMM_VERBOSE" envDefault:"false"`
    MaxVersions     int    `env:"SOMATOGRAMM_MAX_VERSIONS" envDefault:"7"`
}
```

### API Endpoints (`internal/api/handlers/somatogramm_handlers.go`)
- `GET /api/v1/somatogramm/status` - Get execution status
- `POST /api/v1/somatogramm/start` - Trigger manual execution  
- `GET /api/v1/somatogramm/files` - List available files
- `GET /api/v1/somatogramm/download/{filename}` - Download specific file

## Demo Page (`internal/static/demo/somatogramm.html`)
- Simple file download interface
- Display generation timestamps and sample sizes
- Links to download CSV/JSON files for male/female/divers
- Basic statistics overview (total players, age ranges covered)

## File Management
- Generate timestamped files: `somatogramm-{gender}-YYYYMMDD-HHMMSS.{ext}`
- Keep configurable number of versions (default: 7)
- Automatic cleanup of old files
- Separate files for each gender plus combined overview

## Error Handling & Logging
- Handle insufficient sample sizes gracefully (skip age group, log warning)
- Robust error handling for API failures
- Comprehensive logging with progress indicators
- Resume capability for large datasets

## Performance Considerations
- Use concurrent API requests for player data fetching
- In-memory processing for percentile calculations
- Efficient sorting algorithms for large datasets
- Progress tracking and checkpoint system

## Statistical Validation
- Validate percentile calculations with known statistical distributions
- Handle edge cases: single DWZ value, identical DWZ values
- Include statistical metadata in output files
- Age range validation (exclude unrealistic ages like < 4 or > 100)

## Deployment Structure
```
somatogramm/
├── cmd/somatogramm/
│   └── main.go
├── internal/
│   ├── api/
│   │   └── client.go
│   ├── models/
│   │   └── models.go
│   ├── processor/
│   │   └── processor.go
│   ├── export/
│   │   └── export.go
│   └── resume/
│       └── manager.go (shared with kader-planung)
├── bin/
│   ├── somatogramm.exe
│   └── somatogramm
└── docs/
    └── somatogramm-design.md (this file)
```

---

*This implementation plan provides a comprehensive blueprint for the Somatogramm feature, following established patterns from kader-planung while addressing the specific requirements of DWZ percentile analysis by age and gender.*