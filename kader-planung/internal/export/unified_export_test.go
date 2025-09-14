package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/statistics"
	"github.com/sirupsen/logrus"
)

// TestUnifiedExporter_StatisticalFormats tests statistical data export formats
func TestUnifiedExporter_StatisticalFormats(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Create test statistical data
	statisticalData := createTestStatisticalData()

	t.Run("CSV Statistical Export", func(t *testing.T) {
		config := &UnifiedExportConfig{
			OutputDir:         tempDir,
			Format:           CSVStatistical,
			IncludeStatistics: true,
			Timestamp:        false, // For predictable filenames in tests
		}

		exporter := NewUnifiedExporter(config, logger)
		result, err := exporter.Export(nil, statisticalData, "statistical")

		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		// Should have created files for each gender
		expectedFiles := 2 // male and female
		if len(result.Files) != expectedFiles {
			t.Errorf("Expected %d files, got %d: %v", expectedFiles, len(result.Files), result.Files)
		}

		// Verify male file
		maleFile := ""
		femaleFile := ""
		for _, file := range result.Files {
			if strings.Contains(file, "male") && !strings.Contains(file, "female") {
				maleFile = file
			} else if strings.Contains(file, "female") {
				femaleFile = file
			}
		}

		if maleFile == "" {
			t.Errorf("Expected male statistical file in: %v", result.Files)
		}
		if femaleFile == "" {
			t.Errorf("Expected female statistical file in: %v", result.Files)
		}

		// Verify male file content
		if maleFile != "" {
			verifyStatisticalCSV(t, tempDir, maleFile, "male")
		}
	})

	t.Run("JSON Statistical Export", func(t *testing.T) {
		config := &UnifiedExportConfig{
			OutputDir:         tempDir,
			Format:           JSONStatistical,
			IncludeStatistics: true,
			Timestamp:        false,
		}

		exporter := NewUnifiedExporter(config, logger)
		result, err := exporter.Export(nil, statisticalData, "statistical")

		if err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		// Should have created JSON files for each gender
		expectedFiles := 2
		if len(result.Files) != expectedFiles {
			t.Errorf("Expected %d files, got %d", expectedFiles, len(result.Files))
		}

		// Verify JSON file content
		for _, file := range result.Files {
			if strings.Contains(file, "male") && !strings.Contains(file, "female") {
				verifyStatisticalJSON(t, tempDir, file, "m")
			} else if strings.Contains(file, "female") {
				verifyStatisticalJSON(t, tempDir, file, "w")
			}
		}
	})
}

// TestUnifiedExporter_CombinedFormats tests combined export formats
func TestUnifiedExporter_CombinedFormats(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Create test data
	records := createTestRecords()
	statisticalData := createTestStatisticalData()

	t.Run("CSV Combined Export", func(t *testing.T) {
		config := &UnifiedExportConfig{
			OutputDir:         tempDir,
			Format:           CSVCombined,
			IncludeStatistics: true,
			Timestamp:        false,
		}

		exporter := NewUnifiedExporter(config, logger)
		result, err := exporter.Export(records, statisticalData, "hybrid")

		if err != nil {
			t.Fatalf("Combined export failed: %v", err)
		}

		// Should have created both detailed and statistical files
		expectedMinFiles := 3 // 1 detailed + 2 statistical (male/female)
		if len(result.Files) < expectedMinFiles {
			t.Errorf("Expected at least %d files, got %d", expectedMinFiles, len(result.Files))
		}

		// Verify we have both types
		hasDetailed := false
		hasStatistical := false
		for _, file := range result.Files {
			if strings.Contains(file, "statistical") {
				hasStatistical = true
			} else if strings.Contains(file, "hybrid") {
				hasDetailed = true
			}
		}

		if !hasDetailed {
			t.Error("Expected detailed file in combined export")
		}
		if !hasStatistical {
			t.Error("Expected statistical files in combined export")
		}
	})

	t.Run("JSON Combined Export", func(t *testing.T) {
		config := &UnifiedExportConfig{
			OutputDir:         tempDir,
			Format:           JSONCombined,
			IncludeStatistics: true,
			Timestamp:        false,
		}

		exporter := NewUnifiedExporter(config, logger)
		result, err := exporter.Export(records, statisticalData, "hybrid")

		if err != nil {
			t.Fatalf("Combined JSON export failed: %v", err)
		}

		// Should have created one combined JSON file
		expectedFiles := 1
		if len(result.Files) != expectedFiles {
			t.Errorf("Expected %d file, got %d", expectedFiles, len(result.Files))
		}

		// Verify combined JSON content
		verifyJSONCombined(t, tempDir, result.Files[0], len(records))
	})
}

// TestUnifiedExporter_OutputFormatParsing tests output format parsing
func TestUnifiedExporter_OutputFormatParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected OutputFormat
	}{
		{"csv", CSVDetailed},
		{"csv-detailed", CSVDetailed},
		{"json", JSONDetailed},
		{"json-detailed", JSONDetailed},
		{"csv-statistical", CSVStatistical},
		{"json-statistical", JSONStatistical},
		{"csv-combined", CSVCombined},
		{"json-combined", JSONCombined},
		{"invalid", CSVDetailed}, // Default fallback
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseOutputFormat(tc.input)
			if result != tc.expected {
				t.Errorf("Expected format %s for input '%s', got %s",
					tc.expected.String(), tc.input, result.String())
			}
		})
	}
}

// TestUnifiedExporter_TimestampGeneration tests filename generation with timestamps
func TestUnifiedExporter_TimestampGeneration(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &UnifiedExportConfig{
		OutputDir: tempDir,
		Format:   CSVDetailed,
		Timestamp: true,
	}

	exporter := NewUnifiedExporter(config, logger)

	// Test filename generation
	filename1 := exporter.generateFilename("test", "csv")

	// Should have timestamps
	if !strings.Contains(filename1, "test-") {
		t.Errorf("Expected timestamp in filename: %s", filename1)
	}
	if !strings.HasSuffix(filename1, ".csv") {
		t.Errorf("Expected .csv suffix: %s", filename1)
	}

	// Test without timestamp
	config.Timestamp = false
	filename3 := exporter.generateFilename("test", "csv")
	expected := "test.csv"
	if filename3 != expected {
		t.Errorf("Expected '%s', got '%s'", expected, filename3)
	}
}

// TestUnifiedExporter_EmptyData tests handling of empty data sets
func TestUnifiedExporter_EmptyData(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &UnifiedExportConfig{
		OutputDir: tempDir,
		Format:   CSVStatistical,
		Timestamp: false,
	}

	exporter := NewUnifiedExporter(config, logger)

	t.Run("Empty statistical data", func(t *testing.T) {
		result, err := exporter.Export(nil, nil, "statistical")

		if err != nil {
			t.Fatalf("Export should handle empty data: %v", err)
		}

		// Should have no files for empty data
		if len(result.Files) != 0 {
			t.Errorf("Expected no files for empty data, got %d", len(result.Files))
		}
	})

	t.Run("Empty records", func(t *testing.T) {
		config.Format = CSVDetailed
		result, err := exporter.Export(nil, nil, "detailed")

		if err != nil {
			t.Fatalf("Export should handle empty records: %v", err)
		}

		// Should have no files for empty data
		if len(result.Files) != 0 {
			t.Errorf("Expected no files for empty records, got %d", len(result.Files))
		}
	})
}

// Helper functions

func createTestRecords() []models.KaderPlanungRecord {
	return []models.KaderPlanungRecord{
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Test Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-001",
			PKZ:                     "12345",
			Lastname:                "TestPlayer",
			Firstname:               "Test1",
			Birthyear:               1995,
			Gender:                  "m",
			CurrentDWZ:              1600,
			ListRanking:             1,
			DWZ12MonthsAgo:          "1550",
			GamesLast12Months:       "8",
			SuccessRateLast12Months: "65,0",
			DWZAgeRelation:          "average",
		},
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Test Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-002",
			PKZ:                     "12346",
			Lastname:                "TestPlayer",
			Firstname:               "Test2",
			Birthyear:               1990,
			Gender:                  "w",
			CurrentDWZ:              1550,
			ListRanking:             2,
			DWZ12MonthsAgo:          models.DataNotAvailable,
			GamesLast12Months:       models.DataNotAvailable,
			SuccessRateLast12Months: models.DataNotAvailable,
			DWZAgeRelation:          "above_average",
		},
	}
}

func createTestStatisticalData() map[string]statistics.SomatogrammData {
	return map[string]statistics.SomatogrammData{
		"m": {
			Metadata: statistics.SomatogrammMetadata{
				GeneratedAt:    time.Now(),
				Gender:         "m",
				TotalPlayers:   100,
				ValidAgeGroups: 3,
				MinSampleSize:  10,
			},
			Percentiles: map[string]statistics.PercentileData{
				"25": {
					Age:        25,
					SampleSize: 30,
					AvgDWZ:     1520.5,
					MedianDWZ:  1510,
					Percentiles: map[int]int{
						0:   1200,
						25:  1400,
						50:  1510,
						75:  1650,
						100: 1800,
					},
				},
				"30": {
					Age:        30,
					SampleSize: 45,
					AvgDWZ:     1580.2,
					MedianDWZ:  1570,
					Percentiles: map[int]int{
						0:   1250,
						25:  1450,
						50:  1570,
						75:  1700,
						100: 1900,
					},
				},
			},
		},
		"w": {
			Metadata: statistics.SomatogrammMetadata{
				GeneratedAt:    time.Now(),
				Gender:         "w",
				TotalPlayers:   75,
				ValidAgeGroups: 2,
				MinSampleSize:  10,
			},
			Percentiles: map[string]statistics.PercentileData{
				"25": {
					Age:        25,
					SampleSize: 25,
					AvgDWZ:     1450.8,
					MedianDWZ:  1440,
					Percentiles: map[int]int{
						0:   1150,
						25:  1300,
						50:  1440,
						75:  1580,
						100: 1720,
					},
				},
			},
		},
	}
}

func verifyStatisticalCSV(t *testing.T, baseDir, filename, expectedGender string) {
	t.Helper()

	filepath := filepath.Join(baseDir, filename)
	file, err := os.Open(filepath)
	if err != nil {
		t.Fatalf("Could not open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Could not read CSV: %v", err)
	}

	if len(records) < 8 { // Metadata rows + header + at least 1 data row
		t.Errorf("Expected at least 8 rows, got %d", len(records))
	}

	// Check metadata rows
	found := false
	for _, record := range records[:7] { // Check first 7 rows for metadata
		if len(record) > 1 && strings.Contains(record[1], expectedGender) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find gender '%s' in metadata", expectedGender)
	}

	// Find header row (should have 'age' as first column)
	headerRowIdx := -1
	for i, record := range records {
		if len(record) > 0 && record[0] == "age" {
			headerRowIdx = i
			break
		}
	}

	if headerRowIdx == -1 {
		t.Error("Expected to find header row with 'age'")
	} else {
		// Check that we have data rows after header
		dataRows := records[headerRowIdx+1:]
		if len(dataRows) == 0 {
			t.Error("Expected data rows after header")
		}

		// Check first data row has numeric age
		if len(dataRows) > 0 && len(dataRows[0]) > 0 {
			age := dataRows[0][0]
			if age != "25" && age != "30" { // Expected ages from test data
				t.Errorf("Expected age '25' or '30', got '%s'", age)
			}
		}
	}
}

func verifyStatisticalJSON(t *testing.T, baseDir, filename, expectedGender string) {
	t.Helper()

	filepath := filepath.Join(baseDir, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Could not read JSON file: %v", err)
	}

	var data statistics.SomatogrammData
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Could not parse JSON: %v", err)
	}

	if data.Metadata.Gender != expectedGender {
		t.Errorf("Expected gender '%s', got '%s'", expectedGender, data.Metadata.Gender)
	}

	if len(data.Percentiles) == 0 {
		t.Error("Expected percentile data in JSON")
	}

	// Check that percentiles have the expected structure
	for age, percentileData := range data.Percentiles {
		if percentileData.Age == 0 {
			t.Errorf("Expected non-zero age for age group %s", age)
		}
		if len(percentileData.Percentiles) == 0 {
			t.Errorf("Expected percentile values for age group %s", age)
		}
	}
}

func verifyJSONCombined(t *testing.T, baseDir, filename string, expectedRecordCount int) {
	t.Helper()

	filepath := filepath.Join(baseDir, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Could not read combined JSON file: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Could not parse combined JSON: %v", err)
	}

	// Check metadata
	_, exists := data["metadata"]
	if !exists {
		t.Error("Expected metadata in combined JSON")
	}

	// Check records
	records, exists := data["records"]
	if !exists {
		t.Error("Expected records in combined JSON")
	} else {
		recordSlice, ok := records.([]interface{})
		if !ok {
			t.Error("Expected records to be an array")
		} else if len(recordSlice) != expectedRecordCount {
			t.Errorf("Expected %d records, got %d", expectedRecordCount, len(recordSlice))
		}
	}

	// Check statistical data
	statisticalData, exists := data["statistical_data"]
	if !exists {
		t.Error("Expected statistical_data in combined JSON")
	} else {
		statMap, ok := statisticalData.(map[string]interface{})
		if !ok {
			t.Error("Expected statistical_data to be an object")
		} else if len(statMap) == 0 {
			t.Error("Expected non-empty statistical data")
		}
	}
}