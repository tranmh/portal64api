# Feature Request: Enhanced Historical DWZ Tracking

**Requested by:** Michael Meier
**Date:** 2025-12-07
**Status:** Implemented

---

## Summary of Requirements

Based on the email request, the following enhancements are needed for kader-planung:

1. **2-Year Semi-Annual Retrospective**: DWZ values at January 1st and June 30th for the past 2 years plus current year (6 data points total)
2. **Yearly Min/Max DWZ**: Highest and lowest DWZ in the current year
3. **Timing Recommendation**: Generate reports on Thursday/Friday to align with DSB's Wednesday evening data publication

---

## Design Decisions (Confirmed with User)

| Decision | Choice |
|----------|--------|
| Time range | 6 data points: current year + 2 prior years (Jan 1 & June 30 each) |
| Missing data handling | Show `DATA_NOT_AVAILABLE` marker |
| Min/Max scope | Include current DWZ in min/max calculation |
| Scheduling | Manual execution only (no automation) |
| Backward compatibility | Keep existing 12-month columns |
| Column naming | English snake_case with absolute years (e.g., `dwz_2024_01_01`) |

---

## New CSV Columns

### Semi-Annual Historical DWZ (6 columns)

Column names will be dynamically generated based on execution date:

```
dwz_YYYY_01_01   # Jan 1 of year (Y-2)
dwz_YYYY_06_30   # June 30 of year (Y-2)
dwz_YYYY_01_01   # Jan 1 of year (Y-1)
dwz_YYYY_06_30   # June 30 of year (Y-1)
dwz_YYYY_01_01   # Jan 1 of current year
dwz_YYYY_06_30   # June 30 of current year (DATA_NOT_AVAILABLE if before this date)
```

**Example for execution in December 2025:**
- `dwz_2023_01_01`
- `dwz_2023_06_30`
- `dwz_2024_01_01`
- `dwz_2024_06_30`
- `dwz_2025_01_01`
- `dwz_2025_06_30`

### Yearly Min/Max DWZ (2 columns)

```
dwz_max_YYYY     # Highest DWZ in current year (includes current DWZ)
dwz_min_YYYY     # Lowest DWZ in current year (includes current DWZ)
```

**Example for execution in 2025:**
- `dwz_max_2025`
- `dwz_min_2025`

---

## Implementation Plan

### Phase 1: Data Model Updates

**File:** `kader-planung/internal/models/models.go`

1. Add new fields to the player data structure:

```go
type PlayerRecord struct {
    // ... existing fields ...

    // New: Semi-annual historical DWZ (dynamically named in export)
    HistoricalDWZ map[string]string // key: "YYYY_MM_DD", value: DWZ or "DATA_NOT_AVAILABLE"

    // New: Yearly min/max
    DWZMaxCurrentYear string // Highest DWZ in current year
    DWZMinCurrentYear string // Lowest DWZ in current year
}
```

2. Add helper struct for historical analysis configuration:

```go
type HistoricalAnalysisConfig struct {
    TargetDates    []time.Time // Jan 1 and June 30 for past 2 years + current year
    CurrentYear    int
    IncludeMinMax  bool
}
```

### Phase 2: Historical Data Analysis Logic

**File:** `kader-planung/internal/processor/unified_processor.go`

1. **Create new function `AnalyzeExtendedHistoricalData`:**

```go
func AnalyzeExtendedHistoricalData(ratingHistory []models.RatingPoint, config HistoricalAnalysisConfig) map[string]string {
    results := make(map[string]string)

    for _, targetDate := range config.TargetDates {
        // Find DWZ closest to but not after targetDate
        dwz := findDWZAtDate(ratingHistory, targetDate)
        key := fmt.Sprintf("%d_%02d_%02d", targetDate.Year(), targetDate.Month(), targetDate.Day())
        results[key] = dwz // either numeric string or "DATA_NOT_AVAILABLE"
    }

    return results
}
```

2. **Create new function `CalculateYearlyMinMax`:**

```go
func CalculateYearlyMinMax(ratingHistory []models.RatingPoint, currentDWZ int, year int) (min, max string) {
    // Start with current DWZ as both min and max (per user requirement)
    minDWZ := currentDWZ
    maxDWZ := currentDWZ

    yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
    yearEnd := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

    for _, point := range ratingHistory {
        if point.Date.After(yearStart) && point.Date.Before(yearEnd) {
            if point.DWZ > maxDWZ {
                maxDWZ = point.DWZ
            }
            if point.DWZ < minDWZ {
                minDWZ = point.DWZ
            }
        }
    }

    return strconv.Itoa(minDWZ), strconv.Itoa(maxDWZ)
}
```

3. **Helper function `findDWZAtDate`:**

```go
func findDWZAtDate(history []models.RatingPoint, targetDate time.Time) string {
    // Find the most recent evaluation on or before targetDate
    var closestPoint *models.RatingPoint

    for i := range history {
        if history[i].Date != nil && !history[i].Date.After(targetDate) {
            if closestPoint == nil || history[i].Date.After(*closestPoint.Date) {
                closestPoint = &history[i]
            }
        }
    }

    if closestPoint == nil {
        return "DATA_NOT_AVAILABLE"
    }
    return strconv.Itoa(closestPoint.DWZ)
}
```

### Phase 3: Target Date Generation

**File:** `kader-planung/internal/processor/unified_processor.go`

```go
func GenerateTargetDates() []time.Time {
    now := time.Now()
    currentYear := now.Year()

    dates := make([]time.Time, 0, 6)

    // Generate for current year and 2 prior years
    for yearOffset := -2; yearOffset <= 0; yearOffset++ {
        year := currentYear + yearOffset

        // January 1st
        jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
        dates = append(dates, jan1)

        // June 30th
        jun30 := time.Date(year, 6, 30, 0, 0, 0, 0, time.UTC)
        dates = append(dates, jun30)
    }

    return dates
}
```

### Phase 4: Export Updates

**File:** `kader-planung/internal/export/unified_export.go`

1. **Update CSV header generation to include dynamic column names:**

```go
func generateCSVHeader(currentYear int) []string {
    header := []string{
        // ... existing columns ...
        "dwz_12_months_ago",      // Keep for backward compatibility
        "games_last_12_months",   // Keep for backward compatibility
        "success_rate_last_12_months", // Keep for backward compatibility
    }

    // Add semi-annual historical columns (6 columns)
    for yearOffset := -2; yearOffset <= 0; yearOffset++ {
        year := currentYear + yearOffset
        header = append(header, fmt.Sprintf("dwz_%d_01_01", year))
        header = append(header, fmt.Sprintf("dwz_%d_06_30", year))
    }

    // Add yearly min/max columns
    header = append(header, fmt.Sprintf("dwz_min_%d", currentYear))
    header = append(header, fmt.Sprintf("dwz_max_%d", currentYear))

    // ... remaining existing columns (somatogram_percentile, dwz_age_relation) ...

    return header
}
```

2. **Update record export to include new fields:**

```go
func recordToCSVRow(record *PlayerRecord, currentYear int) []string {
    row := []string{
        // ... existing fields ...
    }

    // Add semi-annual historical DWZ values
    for yearOffset := -2; yearOffset <= 0; yearOffset++ {
        year := currentYear + yearOffset
        jan1Key := fmt.Sprintf("%d_01_01", year)
        jun30Key := fmt.Sprintf("%d_06_30", year)

        row = append(row, getHistoricalDWZ(record, jan1Key))
        row = append(row, getHistoricalDWZ(record, jun30Key))
    }

    // Add yearly min/max
    row = append(row, record.DWZMinCurrentYear)
    row = append(row, record.DWZMaxCurrentYear)

    return row
}
```

### Phase 5: API Data Retrieval Optimization

**File:** `kader-planung/internal/api/client.go`

The existing `GetPlayerRatingHistory` API endpoint already returns complete rating history with tournament dates. No API changes needed - we just need to process the data differently.

However, for efficiency, we should ensure:
1. Rating history is sorted by date
2. All evaluations from the past 2+ years are fetched (currently fetches all history, so this is already covered)

### Phase 6: Integration into Processing Pipeline

**File:** `kader-planung/internal/processor/unified_processor.go`

Modify `ProcessPlayer` or equivalent function to:

```go
func (p *UnifiedProcessor) ProcessPlayer(player *models.Player, ratingHistory []models.RatingPoint) *PlayerRecord {
    record := &PlayerRecord{
        // ... existing field mapping ...
    }

    // Existing: 12-month analysis (keep for backward compatibility)
    record.DWZ12MonthsAgo, record.GamesLast12Months, record.SuccessRate =
        AnalyzeHistoricalData(ratingHistory)

    // New: Extended historical analysis
    config := HistoricalAnalysisConfig{
        TargetDates:   GenerateTargetDates(),
        CurrentYear:   time.Now().Year(),
        IncludeMinMax: true,
    }
    record.HistoricalDWZ = AnalyzeExtendedHistoricalData(ratingHistory, config)
    record.DWZMinCurrentYear, record.DWZMaxCurrentYear =
        CalculateYearlyMinMax(ratingHistory, player.CurrentDWZ, config.CurrentYear)

    return record
}
```

---

## Updated CSV Output Format

### Complete Column List (with new columns highlighted)

| # | Column Name | Description | Source |
|---|-------------|-------------|--------|
| 1 | club_id_prefix_1 | First char of club ID | Existing |
| 2 | club_id_prefix_2 | First 2 chars | Existing |
| 3 | club_id_prefix_3 | First 3 chars | Existing |
| 4 | club_name | Club name | Existing |
| 5 | club_id | Club ID (C0327) | Existing |
| 6 | player_id | Player ID | Existing |
| 7 | pkz | Player code number | Existing |
| 8 | lastname | Last name | Existing |
| 9 | firstname | First name | Existing |
| 10 | birthyear | Birth year | Existing |
| 11 | gender | m/w/d | Existing |
| 12 | current_dwz | Current DWZ | Existing |
| 13 | list_ranking | Rank by DWZ | Existing |
| 14 | dwz_12_months_ago | DWZ 12 months ago | Existing (kept) |
| 15 | games_last_12_months | Games count | Existing (kept) |
| 16 | success_rate_last_12_months | Win rate | Existing (kept) |
| 17 | **dwz_YYYY_01_01** | DWZ Jan 1 (Y-2) | **NEW** |
| 18 | **dwz_YYYY_06_30** | DWZ June 30 (Y-2) | **NEW** |
| 19 | **dwz_YYYY_01_01** | DWZ Jan 1 (Y-1) | **NEW** |
| 20 | **dwz_YYYY_06_30** | DWZ June 30 (Y-1) | **NEW** |
| 21 | **dwz_YYYY_01_01** | DWZ Jan 1 (current) | **NEW** |
| 22 | **dwz_YYYY_06_30** | DWZ June 30 (current) | **NEW** |
| 23 | **dwz_min_YYYY** | Lowest DWZ this year | **NEW** |
| 24 | **dwz_max_YYYY** | Highest DWZ this year | **NEW** |
| 25 | somatogram_percentile | Germany-wide percentile | Existing |
| 26 | dwz_age_relation | DWZ vs age expectation | Existing |

**Total: 26 columns** (vs. current 17 columns, +8 new columns with 1 being dynamic year suffix)

---

## Edge Cases and Special Handling

### 1. Future Dates

If the current date is before June 30, the `dwz_YYYY_06_30` column for the current year should show `DATA_NOT_AVAILABLE`.

```go
func (p *UnifiedProcessor) shouldIncludeDate(targetDate time.Time) bool {
    return !targetDate.After(time.Now())
}
```

### 2. New Players Without History

Players who registered recently may not have data for older dates:
- All missing historical dates: `DATA_NOT_AVAILABLE`
- Min/Max: Use only current DWZ if no tournament history exists

### 3. Players Without Current DWZ

If `CurrentDWZ == 0`:
- Historical DWZ: Still calculate from rating history
- Min/Max: If no valid DWZ exists, show `DATA_NOT_AVAILABLE`

### 4. Inactive Players

- Include in output with appropriate flags
- Historical data still calculated normally
- Min/Max still calculated normally

---

## Testing Plan

### Unit Tests

**File:** `kader-planung/internal/processor/unified_processor_test.go`

1. **Test `findDWZAtDate`:**
   - Player with history spanning target date
   - Player with no history before target date
   - Player with exact match on target date
   - Empty history

2. **Test `CalculateYearlyMinMax`:**
   - Player with multiple evaluations in year
   - Player with only current DWZ (no evaluations)
   - Player with evaluations only before current year
   - Current DWZ is the min
   - Current DWZ is the max

3. **Test `GenerateTargetDates`:**
   - Verify 6 dates generated
   - Verify correct years
   - Verify Jan 1 and June 30 dates

4. **Test `AnalyzeExtendedHistoricalData`:**
   - Complete history
   - Partial history
   - No history

### Integration Tests

**File:** `kader-planung/internal/processor/unified_processor_integration_test.go`

1. End-to-end test with sample player data
2. CSV export verification with new columns
3. Column header dynamic naming verification

---

## Performance Considerations

### No Additional API Calls Required

The existing `GetPlayerRatingHistory` API returns complete history. The new analysis only requires:
- Additional in-memory processing (O(n) where n = evaluations per player)
- No extra database queries
- No extra network requests

### Memory Impact

- Additional ~200 bytes per player for new fields
- For 50,000 players: ~10 MB additional memory
- Acceptable overhead

### Processing Time

- Additional processing: negligible (simple date comparisons and min/max)
- Estimated impact: < 1% increase in total processing time

---

## File Changes Summary

| File | Change Type |
|------|-------------|
| `kader-planung/internal/models/models.go` | Add new fields |
| `kader-planung/internal/processor/unified_processor.go` | Add analysis functions |
| `kader-planung/internal/export/unified_export.go` | Update CSV export |
| `kader-planung/internal/processor/unified_processor_test.go` | Add unit tests |

---

## Rollout Plan

1. **Development**: Implement changes in feature branch
2. **Testing**: Run against sample data, verify CSV output
3. **Review**: Code review with focus on edge cases
4. **Staging**: Test with full German dataset
5. **Release**: Merge to master, update documentation

---

## Documentation Updates

After implementation, update:
- `kader-planung/README.md`: Document new columns
- `kader-planung/docs/kader-planung-design.md`: Add historical analysis section
- `CLAUDE.md`: Update CSV column documentation

---

## Notes for Manual Execution

Per the email recommendation:

> Generate the report on **Thursday or Friday** to align with DSB's Wednesday evening data publication.

This ensures the kader-planung output matches the DWZ data used in nomination committee meetings.

No automation is required - users should manually execute:
```bash
./kader-planung --club-prefix=C --output-dir=/path/to/output
```

---

## Implementation Notes (2025-12-07)

### Files Modified

1. **kader-planung/internal/models/models.go**
   - Added `HistoricalDWZ` map field to `KaderPlanungRecord`
   - Added `DWZMinCurrentYear` and `DWZMaxCurrentYear` fields
   - Added new functions:
     - `GenerateTargetDates()` - generates 6 semi-annual target dates
     - `FindDWZAtDate()` - finds DWZ closest to but not after a target date
     - `CalculateYearlyMinMax()` - calculates min/max DWZ for current year
     - `AnalyzeExtendedHistoricalData()` - analyzes history for all target dates
     - `GetHistoricalDWZColumnHeaders()` - generates dynamic column headers
     - `GetMinMaxColumnHeaders()` - generates dynamic min/max column headers

2. **kader-planung/internal/processor/unified_processor.go**
   - Updated `convertPlayersToRecords()` to populate new fields during historical analysis

3. **kader-planung/internal/processor/processor.go**
   - Updated `createKaderPlanungRecord()` to initialize new fields
   - Updated `generateRecordWithPercentile()` to populate extended historical data

4. **kader-planung/internal/export/unified_export.go**
   - Updated `exportCSVDetailed()` to include dynamic column headers and values

5. **kader-planung/internal/models/models_test.go**
   - Added comprehensive unit tests for all new functions

6. **kader-planung/internal/processor/processor_test.go**
   - Fixed test initialization (added logger to prevent nil pointer issues)

### Test Results

All new unit tests pass:
- `TestGenerateTargetDates` - verifies 6 correct dates are generated
- `TestFindDWZAtDate` - tests various edge cases for historical lookup
- `TestCalculateYearlyMinMax` - tests min/max calculation scenarios
- `TestAnalyzeExtendedHistoricalData` - tests complete historical analysis
- `TestGetHistoricalDWZColumnHeaders` - verifies dynamic header generation
- `TestGetMinMaxColumnHeaders` - verifies min/max header generation

---

## Appendix: Sample Output

For a player "Max Mustermann" in December 2025:

```csv
...;dwz_2023_01_01;dwz_2023_06_30;dwz_2024_01_01;dwz_2024_06_30;dwz_2025_01_01;dwz_2025_06_30;dwz_min_2025;dwz_max_2025;...
...;1456;1478;1502;1534;1567;1589;1534;1612;...
```

For a new player registered in March 2025:

```csv
...;dwz_2023_01_01;dwz_2023_06_30;dwz_2024_01_01;dwz_2024_06_30;dwz_2025_01_01;dwz_2025_06_30;dwz_min_2025;dwz_max_2025;...
...;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;1423;1400;1450;...
```
