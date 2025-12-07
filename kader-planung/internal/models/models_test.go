package models

import (
	"testing"
	"time"
)

func TestCalculateClubIDPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		clubID   string
		expected1 string
		expected2 string
		expected3 string
	}{
		{
			name:     "Normal club ID",
			clubID:   "C0327",
			expected1: "C",
			expected2: "C0",
			expected3: "C03",
		},
		{
			name:     "Short club ID (2 chars)",
			clubID:   "B5", 
			expected1: "B",
			expected2: "B5",
			expected3: "",
		},
		{
			name:     "Single character club ID",
			clubID:   "A",
			expected1: "A",
			expected2: "",
			expected3: "",
		},
		{
			name:     "Empty club ID",
			clubID:   "",
			expected1: "",
			expected2: "",
			expected3: "",
		},
		{
			name:     "Long club ID",
			clubID:   "C0327123",
			expected1: "C",
			expected2: "C0",
			expected3: "C03",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix1, prefix2, prefix3 := CalculateClubIDPrefixes(tt.clubID)
			
			if prefix1 != tt.expected1 {
				t.Errorf("CalculateClubIDPrefixes(%q) prefix1 = %q, expected %q", tt.clubID, prefix1, tt.expected1)
			}
			if prefix2 != tt.expected2 {
				t.Errorf("CalculateClubIDPrefixes(%q) prefix2 = %q, expected %q", tt.clubID, prefix2, tt.expected2)
			}
			if prefix3 != tt.expected3 {
				t.Errorf("CalculateClubIDPrefixes(%q) prefix3 = %q, expected %q", tt.clubID, prefix3, tt.expected3)
			}
		})
	}
}

func TestGenerateTargetDates(t *testing.T) {
	dates := GenerateTargetDates()

	// Should generate 6 dates
	if len(dates) != 6 {
		t.Errorf("GenerateTargetDates() returned %d dates, expected 6", len(dates))
	}

	currentYear := time.Now().Year()

	// Verify dates are in correct order and have correct values
	expectedDates := []struct {
		year  int
		month time.Month
		day   int
	}{
		{currentYear - 2, time.January, 1},
		{currentYear - 2, time.June, 30},
		{currentYear - 1, time.January, 1},
		{currentYear - 1, time.June, 30},
		{currentYear, time.January, 1},
		{currentYear, time.June, 30},
	}

	for i, expected := range expectedDates {
		if dates[i].Year() != expected.year {
			t.Errorf("dates[%d] year = %d, expected %d", i, dates[i].Year(), expected.year)
		}
		if dates[i].Month() != expected.month {
			t.Errorf("dates[%d] month = %v, expected %v", i, dates[i].Month(), expected.month)
		}
		if dates[i].Day() != expected.day {
			t.Errorf("dates[%d] day = %d, expected %d", i, dates[i].Day(), expected.day)
		}
	}
}

func TestFindDWZAtDate(t *testing.T) {
	tests := []struct {
		name       string
		history    []RatingPoint
		targetDate time.Time
		expected   string
	}{
		{
			name:       "Empty history",
			history:    []RatingPoint{},
			targetDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   DataNotAvailable,
		},
		{
			name: "Exact match on target date",
			history: []RatingPoint{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), DWZ: 1500},
			},
			targetDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   "1500",
		},
		{
			name: "Target date before all history",
			history: []RatingPoint{
				{Date: time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC), DWZ: 1600},
			},
			targetDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   DataNotAvailable,
		},
		{
			name: "Target date after all history",
			history: []RatingPoint{
				{Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), DWZ: 1400},
				{Date: time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC), DWZ: 1450},
			},
			targetDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   "1450", // Should return the most recent before target
		},
		{
			name: "Multiple history points - find closest before target",
			history: []RatingPoint{
				{Date: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), DWZ: 1400},
				{Date: time.Date(2023, 9, 20, 0, 0, 0, 0, time.UTC), DWZ: 1450},
				{Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), DWZ: 1500},
			},
			targetDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected:   "1450", // Should return Sept 2023 value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindDWZAtDate(tt.history, tt.targetDate)
			if result != tt.expected {
				t.Errorf("FindDWZAtDate() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestCalculateYearlyMinMax(t *testing.T) {
	tests := []struct {
		name        string
		history     []RatingPoint
		currentDWZ  int
		year        int
		expectedMin string
		expectedMax string
	}{
		{
			name:        "No current DWZ",
			history:     []RatingPoint{},
			currentDWZ:  0,
			year:        2024,
			expectedMin: DataNotAvailable,
			expectedMax: DataNotAvailable,
		},
		{
			name:        "Only current DWZ (no history in year)",
			history:     []RatingPoint{},
			currentDWZ:  1500,
			year:        2024,
			expectedMin: "1500",
			expectedMax: "1500",
		},
		{
			name: "History with min lower than current",
			history: []RatingPoint{
				{Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), DWZ: 1400},
				{Date: time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC), DWZ: 1450},
			},
			currentDWZ:  1500,
			year:        2024,
			expectedMin: "1400",
			expectedMax: "1500",
		},
		{
			name: "History with max higher than current",
			history: []RatingPoint{
				{Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), DWZ: 1600},
				{Date: time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC), DWZ: 1550},
			},
			currentDWZ:  1500,
			year:        2024,
			expectedMin: "1500",
			expectedMax: "1600",
		},
		{
			name: "History from different year should be ignored",
			history: []RatingPoint{
				{Date: time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), DWZ: 1200},
				{Date: time.Date(2023, 9, 20, 0, 0, 0, 0, time.UTC), DWZ: 1800},
			},
			currentDWZ:  1500,
			year:        2024,
			expectedMin: "1500",
			expectedMax: "1500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minDWZ, maxDWZ := CalculateYearlyMinMax(tt.history, tt.currentDWZ, tt.year)
			if minDWZ != tt.expectedMin {
				t.Errorf("CalculateYearlyMinMax() min = %q, expected %q", minDWZ, tt.expectedMin)
			}
			if maxDWZ != tt.expectedMax {
				t.Errorf("CalculateYearlyMinMax() max = %q, expected %q", maxDWZ, tt.expectedMax)
			}
		})
	}
}

func TestAnalyzeExtendedHistoricalData(t *testing.T) {
	// Create target dates manually for testing
	targetDates := []time.Time{
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
	}

	history := []RatingPoint{
		{Date: time.Date(2022, 11, 15, 0, 0, 0, 0, time.UTC), DWZ: 1400},
		{Date: time.Date(2023, 3, 20, 0, 0, 0, 0, time.UTC), DWZ: 1450},
		{Date: time.Date(2023, 9, 10, 0, 0, 0, 0, time.UTC), DWZ: 1500},
		{Date: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), DWZ: 1550},
	}

	results := AnalyzeExtendedHistoricalData(history, targetDates)

	// Verify expected values
	expectedResults := map[string]string{
		"2023_01_01": "1400", // Uses Nov 2022 value (closest before Jan 2023)
		"2023_06_30": "1450", // Uses March 2023 value (closest before June 2023)
		"2024_01_01": "1500", // Uses Sept 2023 value (closest before Jan 2024)
		"2024_06_30": "1550", // Uses Feb 2024 value (closest before June 2024)
	}

	for key, expected := range expectedResults {
		if results[key] != expected {
			t.Errorf("AnalyzeExtendedHistoricalData()[%q] = %q, expected %q", key, results[key], expected)
		}
	}
}

func TestGetHistoricalDWZColumnHeaders(t *testing.T) {
	headers := GetHistoricalDWZColumnHeaders(2025)

	expected := []string{
		"dwz_2023_01_01",
		"dwz_2023_06_30",
		"dwz_2024_01_01",
		"dwz_2024_06_30",
		"dwz_2025_01_01",
		"dwz_2025_06_30",
	}

	if len(headers) != len(expected) {
		t.Errorf("GetHistoricalDWZColumnHeaders() returned %d headers, expected %d", len(headers), len(expected))
	}

	for i, exp := range expected {
		if headers[i] != exp {
			t.Errorf("GetHistoricalDWZColumnHeaders()[%d] = %q, expected %q", i, headers[i], exp)
		}
	}
}

func TestGetMinMaxColumnHeaders(t *testing.T) {
	headers := GetMinMaxColumnHeaders(2025)

	expected := []string{
		"dwz_min_2025",
		"dwz_max_2025",
	}

	if len(headers) != len(expected) {
		t.Errorf("GetMinMaxColumnHeaders() returned %d headers, expected %d", len(headers), len(expected))
	}

	for i, exp := range expected {
		if headers[i] != exp {
			t.Errorf("GetMinMaxColumnHeaders()[%d] = %q, expected %q", i, headers[i], exp)
		}
	}
}
