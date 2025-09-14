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
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0", 
			ClubIDPrefix3:           "C03",
			ClubName:                "Test Chess Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-297",
			PKZ:                     "PKZ123",               // NEW
			Firstname:               "John",
			Lastname:                "Doe",
			Birthyear:               1990,
			Gender:                  "m",                    // NEW
			CurrentDWZ:              1600,
			ListRanking:             1,                      // NEW
			DWZ12MonthsAgo:          "1550",
			GamesLast12Months:       "8",
			SuccessRateLast12Months: "62,5",
			SomatogramPercentile:    "67,2",                // NEW: Somatogram percentile (Phase 2)
			DWZAgeRelation:          "1500",                // NEW
		},
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0", 
			ClubIDPrefix3:           "C04",
			ClubName:                "Another Club",
			ClubID:                  "C0401",
			PlayerID:                "C0401-123",
			PKZ:                     "PKZ456",               // NEW
			Firstname:               "Jane",
			Lastname:                "Smith",
			Birthyear:               1985,
			Gender:                  "w",                    // NEW
			CurrentDWZ:              1750,
			ListRanking:             2,                      // NEW
			DWZ12MonthsAgo:          models.DataNotAvailable,
			GamesLast12Months:       models.DataNotAvailable,
			SuccessRateLast12Months: models.DataNotAvailable,
			SomatogramPercentile:    models.DataNotAvailable, // NEW: Somatogram percentile (Phase 2)
			DWZAgeRelation:          models.DataNotAvailable, // NEW
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

	// Check header - should start with the new prefix columns and include all new fields
	expectedHeader := "club_id_prefix1;club_id_prefix2;club_id_prefix3;club_name;club_id;player_id;pkz;lastname;firstname;birthyear;gender;current_dwz;list_ranking;dwz_12_months_ago;games_last_12_months;success_rate_last_12_months;somatogram_percentile;dwz_age_relation"
	if !strings.Contains(contentStr, expectedHeader) {
		t.Errorf("CSV header with all new fields not found. Expected: %s", expectedHeader)
	}

	// Check data - should include all field values
	expectedDataRow := "C;C0;C03;Test Chess Club;C0327;C0327-297;PKZ123;Doe;John;1990;m;1600;1;"
	if !strings.Contains(contentStr, expectedDataRow) {
		t.Errorf("First record with all new fields not found in CSV. Looking for: %s", expectedDataRow)
	}

	if !strings.Contains(contentStr, "C;C0;C04;Another Club;C0401;C0401-123") {
		t.Error("Second record with prefix columns not found in CSV")
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
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0", 
			ClubIDPrefix3:           "C03",
			ClubName:                "Test Chess Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-297",
			PKZ:                     "PKZ123",               // NEW
			Firstname:               "John",
			Lastname:                "Doe",
			Birthyear:               1990,
			Gender:                  "m",                    // NEW
			CurrentDWZ:              1600,
			ListRanking:             1,                      // NEW
			DWZ12MonthsAgo:          "1550",
			GamesLast12Months:       "8",
			SuccessRateLast12Months: "62,5",
			SomatogramPercentile:    "67,2",                // NEW: Somatogram percentile (Phase 2)
			DWZAgeRelation:          "1500",                // NEW
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
			ClubIDPrefix1: "C",
			ClubIDPrefix2: "C0", 
			ClubIDPrefix3: "C03",
			ClubName:      "Test Club",
			ClubID:        "C0327",
			PlayerID:      "C0327-" + string(rune(297+i)),
			Firstname:     "Player",
			Lastname:      string(rune(65 + i)), // A, B, C, etc.
			Birthyear:     1990,
			CurrentDWZ:    1600,
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
