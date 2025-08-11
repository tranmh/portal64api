package export

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/portal64/kader-planung/internal/models"
)

func TestValidateRecords(t *testing.T) {
	exporter := New()

	// Test case 1: Empty records
	t.Run("Empty records", func(t *testing.T) {
		records := []models.KaderPlanungRecord{}
		err := exporter.ValidateRecords(records)
		if err == nil {
			t.Error("Expected error for empty records")
		}
	})

	// Test case 2: Valid records
	t.Run("Valid records", func(t *testing.T) {
		records := []models.KaderPlanungRecord{
			{
				ClubID:    "C0327",
				PlayerID:  "C0327-297",
				Firstname: "John",
				Lastname:  "Doe",
			},
		}
		err := exporter.ValidateRecords(records)
		if err != nil {
			t.Errorf("Unexpected error for valid records: %v", err)
		}
	})

	// Test case 3: Missing club ID
	t.Run("Missing club ID", func(t *testing.T) {
		records := []models.KaderPlanungRecord{
			{
				PlayerID:  "C0327-297",
				Firstname: "John",
				Lastname:  "Doe",
			},
		}
		err := exporter.ValidateRecords(records)
		if err == nil {
			t.Error("Expected error for missing club ID")
		}
	})

	// Test case 4: Missing player ID
	t.Run("Missing player ID", func(t *testing.T) {
		records := []models.KaderPlanungRecord{
			{
				ClubID:    "C0327",
				Firstname: "John",
				Lastname:  "Doe",
			},
		}
		err := exporter.ValidateRecords(records)
		if err == nil {
			t.Error("Expected error for missing player ID")
		}
	})

	// Test case 5: Missing both names
	t.Run("Missing both names", func(t *testing.T) {
		records := []models.KaderPlanungRecord{
			{
				ClubID:   "C0327",
				PlayerID: "C0327-297",
			},
		}
		err := exporter.ValidateRecords(records)
		if err == nil {
			t.Error("Expected error for missing both names")
		}
	})
}

func TestExportCSV(t *testing.T) {
	exporter := New()

	// Create test data
	records := []models.KaderPlanungRecord{
		{
			ClubName:                "Test Chess Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-297",
			Firstname:               "John",
			Lastname:                "Doe",
			Birthyear:               1990,
			CurrentDWZ:              1600,
			DWZ12MonthsAgo:          "1550",
			GamesLast12Months:       "8",
			SuccessRateLast12Months: "62.5",
		},
		{
			ClubName:                "Another Club",
			ClubID:                  "C0401",
			PlayerID:                "C0401-123",
			Firstname:               "Jane",
			Lastname:                "Smith",
			Birthyear:               1985,
			CurrentDWZ:              1750,
			DWZ12MonthsAgo:          models.DataNotAvailable,
			GamesLast12Months:       models.DataNotAvailable,
			SuccessRateLast12Months: models.DataNotAvailable,
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.csv")

	// Export CSV
	err := exporter.exportCSV(records, outputPath)
	if err != nil {
		t.Fatalf("Failed to export CSV: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("CSV file was not created")
	}

	// Read and verify content
	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)

	// Check header
	if !strings.Contains(contentStr, "club_name,club_id,player_id") {
		t.Error("CSV header not found")
	}

	// Check data
	if !strings.Contains(contentStr, "Test Chess Club,C0327,C0327-297") {
		t.Error("First record not found in CSV")
	}

	if !strings.Contains(contentStr, "Another Club,C0401,C0401-123") {
		t.Error("Second record not found in CSV")
	}

	if !strings.Contains(contentStr, models.DataNotAvailable) {
		t.Error("DATA_NOT_AVAILABLE not found in CSV")
	}
}

func TestExportJSON(t *testing.T) {
	exporter := New()

	// Create test data
	records := []models.KaderPlanungRecord{
		{
			ClubName:                "Test Chess Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-297",
			Firstname:               "John",
			Lastname:                "Doe",
			Birthyear:               1990,
			CurrentDWZ:              1600,
			DWZ12MonthsAgo:          "1550",
			GamesLast12Months:       "8",
			SuccessRateLast12Months: "62.5",
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.json")

	// Export JSON
	err := exporter.exportJSON(records, outputPath)
	if err != nil {
		t.Fatalf("Failed to export JSON: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("JSON file was not created")
	}

	// Read and parse JSON
	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var output struct {
		Count   int                         `json:"count"`
		Records []models.KaderPlanungRecord `json:"records"`
	}

	if err := json.Unmarshal(content, &output); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify data
	if output.Count != 1 {
		t.Errorf("Expected count to be 1, got %d", output.Count)
	}

	if len(output.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(output.Records))
	}

	record := output.Records[0]
	if record.ClubName != "Test Chess Club" {
		t.Errorf("Expected club name 'Test Chess Club', got '%s'", record.ClubName)
	}

	if record.PlayerID != "C0327-297" {
		t.Errorf("Expected player ID 'C0327-297', got '%s'", record.PlayerID)
	}
}

func TestExportSample(t *testing.T) {
	exporter := New()

	// Create test data with more records than sample size
	records := make([]models.KaderPlanungRecord, 10)
	for i := 0; i < 10; i++ {
		records[i] = models.KaderPlanungRecord{
			ClubName:   "Test Club",
			ClubID:     "C0327",
			PlayerID:   "C0327-" + string(rune(297+i)),
			Firstname:  "Player",
			Lastname:   string(rune(65 + i)), // A, B, C, etc.
			Birthyear:  1990,
			CurrentDWZ: 1600,
		}
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "sample.csv")

	// Export sample of 5 records
	err := exporter.ExportSample(records, outputPath, "csv", 5)
	if err != nil {
		t.Fatalf("Failed to export sample: %v", err)
	}

	// Read and verify content
	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read sample file: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")

	// Should have header + 5 data lines = 6 total lines
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines (header + 5 records), got %d", len(lines))
	}
}
