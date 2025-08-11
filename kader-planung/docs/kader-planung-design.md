# Kader-Planung Application - High Level Design

## Overview
A standalone Golang application that generates comprehensive player roster reports by consuming the Portal64 REST API. The application produces detailed CSV/JSON/Excel reports with historical player data for club management purposes.

## Application Architecture

### Core Components

1. **CLI Interface** (`cmd/kader-planung/`)
   - Command-line argument parsing
   - Configuration validation
   - Application lifecycle management

2. **API Client** (`internal/api/`)
   - REST API wrapper for Portal64 endpoints
   - HTTP client with timeout handling
   - Response parsing and error handling

3. **Data Models** (`internal/models/`)
   - Club, Player, Tournament structures
   - Rating history data structures
   - Export data structures

4. **Processing Engine** (`internal/processor/`)
   - Concurrent data fetching
   - Historical data analysis
   - Statistics calculations

5. **Export Engine** (`internal/export/`)
   - CSV, JSON, Excel format writers
   - Data formatting and validation
6. **Resume Manager** (`internal/resume/`)
   - Progress state persistence
   - Resume checkpoint management
   - Recovery mechanisms

## Command-Line Interface

### Usage
```bash
./kader-planung [options]
```

### Arguments
- `--club-prefix`: Filter clubs by ID prefix (e.g., "C0" for C0327, C0401, etc.)
- `--output-format`: Output format (csv|json|excel, default: csv)
- `--output-dir`: Output directory (default: current directory)
- `--concurrency`: Number of concurrent API requests (default: 1)
- `--resume`: Resume from previous run using checkpoint file
- `--checkpoint-file`: Custom checkpoint file path
- `--api-base-url`: Base URL for Portal64 API
- `--timeout`: API request timeout in seconds (default: 30)
- `--verbose`: Enable detailed logging

## Data Flow Architecture

### Phase 1: Club Discovery
1. Fetch all clubs using `search_clubs` API
2. Filter clubs by prefix if specified
3. Log total number of clubs to process

### Phase 2: Player Data Collection (Concurrent)
1. For each club:
   - Fetch club details (`get_club_profile`)
   - Fetch all club players (`get_club_players`)
   - For each player:
     - Fetch player profile (`get_player_profile`)
     - Fetch rating history (`get_player_rating_history`)

### Phase 3: Historical Analysis
1. Calculate DWZ 12 months ago (closest available point)
2. Count games in last 12 months from rating history
3. Calculate success rate from recent games
4. Handle missing data with "DATA_NOT_AVAILABLE"
### Phase 4: Export Generation
1. Aggregate all processed data
2. Generate output in requested format
3. Save with timestamp-based filename

## Output Schema

### CSV Columns
| Column | Description | Data Source |
|--------|-------------|-------------|
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

### Filename Format
`kader-planung-{club-prefix}-{timestamp}.{extension}`

Examples:
- `kader-planung-C0-20240811-143022.csv`
- `kader-planung-all-20240811-143022.xlsx`

## Concurrent Processing Design

### Worker Pool Pattern
```go
type WorkerPool struct {
    concurrency int
    jobs        chan Job
    results     chan Result
    errors      chan error
}
```

### Error Handling Strategy
- Individual player failures: Log error, continue processing
- Critical API failures: Pause processing, retry with backoff
- Timeout errors: Log and mark as "DATA_NOT_AVAILABLE"
## Resume Functionality

### Checkpoint File Structure
```json
{
    "timestamp": "2024-08-11T14:30:22Z",
    "config": {
        "club_prefix": "C0",
        "output_format": "csv",
        "concurrency": 5
    },
    "progress": {
        "total_clubs": 150,
        "processed_clubs": 75,
        "current_phase": "player_processing"
    },
    "processed_items": [
        {"type": "club", "id": "C0327", "status": "completed"},
        {"type": "player", "id": "C0327-297", "status": "completed"}
    ],
    "partial_data": []
}
```

## Historical Data Analysis

### DWZ 12 Months Ago Algorithm
```go
func FindDWZ12MonthsAgo(history []RatingPoint) string {
    target := time.Now().AddDate(0, -12, 0)
    
    var closest *RatingPoint
    var minDiff time.Duration = math.MaxInt64
    
    for _, point := range history {
        if point.Date.Before(target) {
            diff := target.Sub(point.Date)
            if diff < minDiff {
                closest = &point
                minDiff = diff
            }
        }
    }
    
    if closest == nil {
        return "DATA_NOT_AVAILABLE"
    }
    return strconv.Itoa(closest.DWZ)
}
```

### Games & Success Rate Calculation
- Count all rated games in the last 12 months from rating history
- Success rate = (Wins + 0.5*Draws) / Total Games * 100
- Handle division by zero (no games = "DATA_NOT_AVAILABLE")

---

*This design document serves as the blueprint for implementing the Kader-Planung application. All components are designed for maintainability, extensibility, and robust error handling.*