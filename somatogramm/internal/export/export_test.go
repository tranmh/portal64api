package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"somatogramm/internal/models"
)

func TestNewExporter(t *testing.T) {
	exporter := NewExporter("/tmp/test", "csv", 5, true)

	if exporter.OutputDir != "/tmp/test" {
		t.Errorf("expected OutputDir '/tmp/test', got %s", exporter.OutputDir)
	}

	if exporter.OutputFormat != "csv" {
		t.Errorf("expected OutputFormat 'csv', got %s", exporter.OutputFormat)
	}

	if exporter.MaxVersions != 5 {
		t.Errorf("expected MaxVersions 5, got %d", exporter.MaxVersions)
	}

	if !exporter.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestGetGenderName(t *testing.T) {
	exporter := NewExporter("/tmp", "csv", 5, false)

	tests := []struct {
		input    string
		expected string
	}{
		{"m", "male"},
		{"w", "female"},
		{"d", "divers"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := exporter.getGenderName(tt.input)
			if result != tt.expected {
				t.Errorf("getGenderName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExportJSON(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "json", 5, false)

	// Create test data
	testData := models.SomatogrammData{
		Metadata: models.SomatogrammMetadata{
			GeneratedAt:    time.Now(),
			Gender:         "m",
			TotalPlayers:   100,
			ValidAgeGroups: 5,
			MinSampleSize:  10,
		},
		Percentiles: map[string]models.PercentileData{
			"25": {
				Age:        25,
				SampleSize: 50,
				AvgDWZ:     1500.5,
				MedianDWZ:  1480,
				Percentiles: map[int]int{
					0:   800,
					50:  1480,
					100: 2200,
				},
			},
		},
	}

	timestamp := "20240101-120000"
	err = exporter.exportJSON("male", testData, timestamp)

	if err != nil {
		t.Errorf("exportJSON failed: %v", err)
	}

	// Check if file was created
	expectedFilename := "somatogramm-male-20240101-120000.json"
	filepath := filepath.Join(tempDir, expectedFilename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("expected file %s was not created", expectedFilename)
	}

	// Read and validate JSON content
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Errorf("failed to read exported file: %v", err)
	}

	var exported models.SomatogrammData
	if err := json.Unmarshal(content, &exported); err != nil {
		t.Errorf("failed to unmarshal exported JSON: %v", err)
	}

	if exported.Metadata.Gender != "m" {
		t.Errorf("expected gender 'm', got %s", exported.Metadata.Gender)
	}

	if exported.Metadata.TotalPlayers != 100 {
		t.Errorf("expected 100 total players, got %d", exported.Metadata.TotalPlayers)
	}
}

func TestExportCSV(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "csv", 5, false)

	// Create test data
	testData := models.SomatogrammData{
		Metadata: models.SomatogrammMetadata{
			GeneratedAt:    time.Now(),
			Gender:         "w",
			TotalPlayers:   150,
			ValidAgeGroups: 3,
			MinSampleSize:  20,
		},
		Percentiles: map[string]models.PercentileData{
			"20": {
				Age:        20,
				SampleSize: 30,
				AvgDWZ:     1200.5,
				MedianDWZ:  1180,
				Percentiles: map[int]int{
					0:   600,
					25:  1000,
					50:  1180,
					75:  1400,
					100: 1800,
				},
			},
			"25": {
				Age:        25,
				SampleSize: 50,
				AvgDWZ:     1350.0,
				MedianDWZ:  1320,
				Percentiles: map[int]int{
					0:   700,
					25:  1100,
					50:  1320,
					75:  1600,
					100: 2000,
				},
			},
		},
	}

	timestamp := "20240101-120000"
	err = exporter.exportCSV("female", testData, timestamp)

	if err != nil {
		t.Errorf("exportCSV failed: %v", err)
	}

	// Check if file was created
	expectedFilename := "somatogramm-female-20240101-120000.csv"
	filepath := filepath.Join(tempDir, expectedFilename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("expected file %s was not created", expectedFilename)
	}

	// Read and validate CSV content
	file, err := os.Open(filepath)
	if err != nil {
		t.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Errorf("failed to read CSV records: %v", err)
	}

	// Check header row
	if len(records) < 1 {
		t.Error("CSV file should have at least a header row")
	}

	header := records[0]
	expectedHeaderStart := []string{"age", "sample_size", "avg_dwz", "median_dwz"}

	for i, expectedCol := range expectedHeaderStart {
		if len(header) <= i || header[i] != expectedCol {
			t.Errorf("expected header column %d to be %s, got %s", i, expectedCol, header[i])
		}
	}

	// Check that we have percentile columns p0 through p100
	if len(header) != 4+101 { // 4 metadata columns + 101 percentile columns (0-100)
		t.Errorf("expected %d header columns, got %d", 4+101, len(header))
	}

	// Check data rows (should be 2: age 20 and age 25, sorted by age)
	if len(records) != 3 { // header + 2 data rows
		t.Errorf("expected 3 rows (header + 2 data), got %d", len(records))
	}

	// First data row should be age 20
	if len(records) > 1 && records[1][0] != "20" {
		t.Errorf("expected first data row to be age 20, got %s", records[1][0])
	}

	// Second data row should be age 25
	if len(records) > 2 && records[2][0] != "25" {
		t.Errorf("expected second data row to be age 25, got %s", records[2][0])
	}
}

func TestExportDataCreateDirectory(t *testing.T) {
	// Create temporary base directory
	tempBase, err := os.MkdirTemp("", "somatogramm_base")
	if err != nil {
		t.Fatalf("failed to create temp base dir: %v", err)
	}
	defer os.RemoveAll(tempBase)

	// Use a subdirectory that doesn't exist yet
	testDir := filepath.Join(tempBase, "somatogramm_output")
	exporter := NewExporter(testDir, "json", 5, false)

	testData := map[string]models.SomatogrammData{
		"m": {
			Metadata: models.SomatogrammMetadata{
				GeneratedAt:    time.Now(),
				Gender:         "m",
				TotalPlayers:   50,
				ValidAgeGroups: 2,
				MinSampleSize:  10,
			},
			Percentiles: map[string]models.PercentileData{
				"30": {
					Age:        30,
					SampleSize: 25,
					AvgDWZ:     1400.0,
					MedianDWZ:  1380,
					Percentiles: map[int]int{
						0:   800,
						50:  1380,
						100: 2000,
					},
				},
			},
		},
	}

	err = exporter.ExportData(testData)

	if err != nil {
		t.Errorf("ExportData failed: %v", err)
	}

	// Check if directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to be created", testDir)
	}

	// Check if file was created
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Errorf("failed to read output directory: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file in output directory, got %d", len(files))
	}

	if len(files) > 0 && !files[0].Type().IsRegular() {
		t.Error("expected a regular file in output directory")
	}
}

func TestListFiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "csv", 5, false)

	// Create some test files
	testFiles := []string{
		"somatogramm-male-20240101-120000.csv",
		"somatogramm-female-20240101-120000.json",
		"somatogramm-divers-20240101-130000.csv",
		"other-file.txt", // Should be ignored
	}

	for _, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filepath, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	fileInfos, err := exporter.ListFiles()
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}

	// Should return 3 files (excluding other-file.txt)
	if len(fileInfos) != 3 {
		t.Errorf("expected 3 files, got %d", len(fileInfos))
	}

	// Check that files are sorted by modification time (newest first)
	// Since we created them in sequence, the last one should be first
	if len(fileInfos) > 0 && fileInfos[0].Name != "somatogramm-divers-20240101-130000.csv" {
		t.Errorf("expected first file to be somatogramm-divers-20240101-130000.csv, got %s", fileInfos[0].Name)
	}

	// Check file properties
	for _, info := range fileInfos {
		if info.Size == 0 {
			t.Errorf("file %s should have non-zero size", info.Name)
		}

		if info.Name == "somatogramm-female-20240101-120000.json" && !info.IsJSON {
			t.Errorf("file %s should be marked as JSON", info.Name)
		}

		if (info.Name == "somatogramm-male-20240101-120000.csv" || info.Name == "somatogramm-divers-20240101-130000.csv") && info.IsJSON {
			t.Errorf("file %s should not be marked as JSON", info.Name)
		}
	}
}

func TestExportDataUnsupportedFormat(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "xml", 5, false) // unsupported format

	testData := map[string]models.SomatogrammData{
		"m": {
			Metadata: models.SomatogrammMetadata{
				GeneratedAt:    time.Now(),
				Gender:         "m",
				TotalPlayers:   50,
				ValidAgeGroups: 1,
				MinSampleSize:  10,
			},
			Percentiles: map[string]models.PercentileData{
				"30": {
					Age:        30,
					SampleSize: 25,
					AvgDWZ:     1400.0,
					MedianDWZ:  1380,
					Percentiles: map[int]int{
						0:   800,
						50:  1380,
						100: 2000,
					},
				},
			},
		},
	}

	err = exporter.ExportData(testData)

	if err == nil {
		t.Error("expected error for unsupported format, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("expected 'unsupported output format' error, got: %v", err)
	}
}

func TestExportJSONError(t *testing.T) {
	// Use a directory that doesn't exist and can't be created
	exporter := NewExporter("/invalid/path/that/does/not/exist", "json", 5, false)

	testData := models.SomatogrammData{
		Metadata: models.SomatogrammMetadata{
			GeneratedAt:    time.Now(),
			Gender:         "m",
			TotalPlayers:   50,
			ValidAgeGroups: 1,
			MinSampleSize:  10,
		},
		Percentiles: map[string]models.PercentileData{
			"30": {
				Age:        30,
				SampleSize: 25,
				AvgDWZ:     1400.0,
				MedianDWZ:  1380,
				Percentiles: map[int]int{
					0:   800,
					50:  1380,
					100: 2000,
				},
			},
		},
	}

	timestamp := "20240101-120000"
	err := exporter.exportJSON("male", testData, timestamp)

	if err == nil {
		t.Error("expected error when writing to invalid path, got nil")
	}
}

func TestExportCSVError(t *testing.T) {
	// Use a directory that doesn't exist and can't be created
	exporter := NewExporter("/invalid/path/that/does/not/exist", "csv", 5, false)

	testData := models.SomatogrammData{
		Metadata: models.SomatogrammMetadata{
			GeneratedAt:    time.Now(),
			Gender:         "w",
			TotalPlayers:   50,
			ValidAgeGroups: 1,
			MinSampleSize:  10,
		},
		Percentiles: map[string]models.PercentileData{
			"25": {
				Age:        25,
				SampleSize: 30,
				AvgDWZ:     1300.0,
				MedianDWZ:  1280,
				Percentiles: map[int]int{
					0:   700,
					50:  1280,
					100: 1900,
				},
			},
		},
	}

	timestamp := "20240101-120000"
	err := exporter.exportCSV("female", testData, timestamp)

	if err == nil {
		t.Error("expected error when writing to invalid path, got nil")
	}
}

func TestCleanupOldFilesMaxVersionsZero(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "csv", 0, false) // MaxVersions = 0

	// Create some test files
	testFiles := []string{
		"somatogramm-male-20240101-120000.csv",
		"somatogramm-male-20240102-120000.csv",
	}

	for _, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filepath, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	// Should not delete anything when MaxVersions is 0
	err = exporter.cleanupOldFiles()
	if err != nil {
		t.Errorf("cleanupOldFiles failed: %v", err)
	}

	// Check that files still exist
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files to remain, got %d", len(files))
	}
}

func TestCleanupOldFilesWithDeletion(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "csv", 1, true) // Keep only 1 version

	// Create multiple test files for same gender with different timestamps
	testFiles := []string{
		"somatogramm-male-20240101-120000.csv",
		"somatogramm-male-20240102-120000.csv",
		"somatogramm-male-20240103-120000.csv",
	}

	for i, filename := range testFiles {
		filepath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filepath, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
		// Add slight delay to ensure different modification times
		time.Sleep(10 * time.Millisecond)
		// Update modification time to simulate chronological order
		newTime := time.Now().Add(time.Duration(i) * time.Minute)
		os.Chtimes(filepath, newTime, newTime)
	}

	err = exporter.cleanupOldFiles()
	if err != nil {
		t.Errorf("cleanupOldFiles failed: %v", err)
	}

	// Check that only 1 file remains (the newest one)
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file to remain after cleanup, got %d", len(files))
	}

	if len(files) > 0 && files[0].Name() != "somatogramm-male-20240103-120000.csv" {
		t.Errorf("expected newest file to remain, got %s", files[0].Name())
	}
}

func TestListFilesError(t *testing.T) {
	exporter := NewExporter("/invalid/path/that/does/not/exist", "csv", 5, false)

	fileInfos, err := exporter.ListFiles()

	if err == nil {
		t.Error("expected error when reading invalid directory, got nil")
	}

	if fileInfos != nil {
		t.Errorf("expected nil fileInfos on error, got %v", fileInfos)
	}
}

func TestVerboseLogging(t *testing.T) {
	exporter := NewExporter("/tmp", "csv", 5, true) // verbose = true

	// This test mainly ensures the log function doesn't panic when verbose is true
	// The actual log output would go to stdout and is difficult to capture in unit tests
	exporter.log("test message")

	// Test that exporter was created with verbose flag
	if !exporter.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestExportCSVWithMissingPercentiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "csv", 5, false)

	// Create test data with missing percentiles
	testData := models.SomatogrammData{
		Metadata: models.SomatogrammMetadata{
			GeneratedAt:    time.Now(),
			Gender:         "m",
			TotalPlayers:   50,
			ValidAgeGroups: 1,
			MinSampleSize:  10,
		},
		Percentiles: map[string]models.PercentileData{
			"25": {
				Age:        25,
				SampleSize: 30,
				AvgDWZ:     1300.0,
				MedianDWZ:  1280,
				Percentiles: map[int]int{
					0:  700,
					50: 1280,
					// Missing some percentiles
				},
			},
		},
	}

	timestamp := "20240101-120000"
	err = exporter.exportCSV("male", testData, timestamp)

	if err != nil {
		t.Errorf("exportCSV should handle missing percentiles gracefully, got error: %v", err)
	}

	// Verify file was created
	expectedFilename := "somatogramm-male-20240101-120000.csv"
	filepath := filepath.Join(tempDir, expectedFilename)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("expected file %s was not created", expectedFilename)
	}
}

func TestExportCSVWithInvalidAge(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "somatogramm_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exporter := NewExporter(tempDir, "csv", 5, false)

	// Create test data with invalid age key (non-numeric)
	testData := models.SomatogrammData{
		Metadata: models.SomatogrammMetadata{
			GeneratedAt:    time.Now(),
			Gender:         "w",
			TotalPlayers:   50,
			ValidAgeGroups: 1,
			MinSampleSize:  10,
		},
		Percentiles: map[string]models.PercentileData{
			"invalid": {
				Age:        25,
				SampleSize: 30,
				AvgDWZ:     1300.0,
				MedianDWZ:  1280,
				Percentiles: map[int]int{
					0:   700,
					50:  1280,
					100: 1900,
				},
			},
			"30": {
				Age:        30,
				SampleSize: 40,
				AvgDWZ:     1400.0,
				MedianDWZ:  1380,
				Percentiles: map[int]int{
					0:   800,
					50:  1380,
					100: 2000,
				},
			},
		},
	}

	timestamp := "20240101-120000"
	err = exporter.exportCSV("female", testData, timestamp)

	if err != nil {
		t.Errorf("exportCSV should handle invalid age keys gracefully, got error: %v", err)
	}

	// Check if file was created and only contains valid age (30)
	expectedFilename := "somatogramm-female-20240101-120000.csv"
	filepath := filepath.Join(tempDir, expectedFilename)

	file, err := os.Open(filepath)
	if err != nil {
		t.Fatalf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV records: %v", err)
	}

	// Should have header + 1 data row (only age 30, invalid age skipped)
	if len(records) != 2 {
		t.Errorf("expected 2 rows (header + 1 data), got %d", len(records))
	}

	if len(records) > 1 && records[1][0] != "30" {
		t.Errorf("expected first data row to be age 30, got %s", records[1][0])
	}
}