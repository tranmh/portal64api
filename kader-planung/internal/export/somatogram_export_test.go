package export

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/portal64/kader-planung/internal/models"
)

// TestSomatogramPercentileCSVExport tests CSV export with somatogram percentile column
func TestSomatogramPercentileCSVExport(t *testing.T) {
	exporter := New()

	// Create test records with various somatogram percentile values
	records := []models.KaderPlanungRecord{
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Test Chess Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-297",
			PKZ:                     "PKZ123",
			Firstname:               "John",
			Lastname:                "Doe",
			Birthyear:               1990,
			Gender:                  "m",
			CurrentDWZ:              1600,
			ListRanking:             1,
			DWZ12MonthsAgo:          "1550",
			GamesLast12Months:       "8",
			SuccessRateLast12Months: "62,5",
			SomatogramPercentile:    "67,2",          // Real percentile value
			DWZAgeRelation:          "1500",
		},
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C04",
			ClubName:                "Another Club",
			ClubID:                  "C0401",
			PlayerID:                "C0401-123",
			PKZ:                     "PKZ456",
			Firstname:               "Jane",
			Lastname:                "Smith",
			Birthyear:               1985,
			Gender:                  "w",
			CurrentDWZ:              1750,
			ListRanking:             2,
			DWZ12MonthsAgo:          models.DataNotAvailable,
			GamesLast12Months:       models.DataNotAvailable,
			SuccessRateLast12Months: models.DataNotAvailable,
			SomatogramPercentile:    models.DataNotAvailable, // DATA_NOT_AVAILABLE
			DWZAgeRelation:          models.DataNotAvailable,
		},
		{
			ClubIDPrefix1:           "B",
			ClubIDPrefix2:           "B0",
			ClubIDPrefix3:           "B05",
			ClubName:                "Baden Club",
			ClubID:                  "B0512",
			PlayerID:                "B0512-789",
			PKZ:                     "PKZ789",
			Firstname:               "Max",
			Lastname:                "Mustermann",
			Birthyear:               1995,
			Gender:                  "m",
			CurrentDWZ:              2100,
			ListRanking:             1,
			DWZ12MonthsAgo:          "2050",
			GamesLast12Months:       "12",
			SuccessRateLast12Months: "75,0",
			SomatogramPercentile:    "94,8",          // High percentile
			DWZAgeRelation:          "2000",
		},
	}

	// Export to CSV
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "somatogram_test.csv")

	err := exporter.exportCSV(records, outputPath)
	if err != nil {
		t.Fatalf("Failed to export CSV with somatogram percentiles: %v", err)
	}

	// Read and validate CSV content
	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")

	// Validate header line
	header := lines[0]
	expectedFields := []string{
		"club_id_prefix1", "club_id_prefix2", "club_id_prefix3",
		"club_name", "club_id", "player_id", "pkz", "lastname", "firstname",
		"birthyear", "gender", "current_dwz", "list_ranking",
		"dwz_12_months_ago", "games_last_12_months", "success_rate_last_12_months",
		"somatogram_percentile", "dwz_age_relation",
	}

	for _, field := range expectedFields {
		if !strings.Contains(header, field) {
			t.Errorf("Missing expected field %s in CSV header", field)
		}
	}

	// Validate somatogram_percentile position (should be 16th column, 0-indexed)
	headerFields := strings.Split(header, ";")
	somatogramIndex := -1
	for i, field := range headerFields {
		if field == "somatogram_percentile" {
			somatogramIndex = i
			break
		}
	}

	expectedIndex := 16 // After success_rate_last_12_months (index 15)
	if somatogramIndex != expectedIndex {
		t.Errorf("somatogram_percentile at wrong position: expected index %d, got %d", expectedIndex, somatogramIndex)
	}

	// Validate data rows
	if len(lines) != 4 { // header + 3 records
		t.Errorf("Expected 4 lines (header + 3 records), got %d", len(lines))
	}

	// Check first record with real percentile
	firstRecord := strings.Split(lines[1], ";")
	if len(firstRecord) != len(headerFields) {
		t.Errorf("First record has %d fields, expected %d", len(firstRecord), len(headerFields))
	}
	if firstRecord[somatogramIndex] != "67,2" {
		t.Errorf("First record somatogram_percentile: expected 67,2, got %s", firstRecord[somatogramIndex])
	}

	// Check second record with DATA_NOT_AVAILABLE
	secondRecord := strings.Split(lines[2], ";")
	if secondRecord[somatogramIndex] != models.DataNotAvailable {
		t.Errorf("Second record somatogram_percentile: expected %s, got %s", models.DataNotAvailable, secondRecord[somatogramIndex])
	}

	// Check third record with high percentile
	thirdRecord := strings.Split(lines[3], ";")
	if thirdRecord[somatogramIndex] != "94,8" {
		t.Errorf("Third record somatogram_percentile: expected 94,8, got %s", thirdRecord[somatogramIndex])
	}

	t.Logf("CSV export validation passed: somatogram_percentile in column %d", somatogramIndex+1)
}

// TestSomatogramPercentileValidation tests validation of percentile values
func TestSomatogramPercentileValidation(t *testing.T) {
	// Test CSV parsing to ensure percentile values are correctly formatted

	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "percentile_validation.csv")

	// Create CSV content directly for parsing test
	csvContent := `club_id_prefix1;club_id_prefix2;club_id_prefix3;club_name;club_id;player_id;pkz;lastname;firstname;birthyear;gender;current_dwz;list_ranking;dwz_12_months_ago;games_last_12_months;success_rate_last_12_months;somatogram_percentile;dwz_age_relation
C;C0;C03;Test Club;C0327;C0327-001;PKZ001;Doe;John;1990;m;1500;1;1450;8;62,5;75,3;1400
C;C0;C03;Test Club;C0327;C0327-002;PKZ002;Smith;Jane;1985;w;1600;2;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE;DATA_NOT_AVAILABLE
C;C0;C03;Test Club;C0327;C0327-003;PKZ003;Wilson;Bob;1995;m;1800;1;1750;15;80,0;96,7;1700`

	err := ioutil.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}

	// Parse CSV and validate percentile values
	file, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("Failed to open CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	// Validate header and find somatogram_percentile column
	header := records[0]
	percentileColumn := -1
	for i, field := range header {
		if field == "somatogram_percentile" {
			percentileColumn = i
			break
		}
	}

	if percentileColumn == -1 {
		t.Fatal("somatogram_percentile column not found in CSV")
	}

	// Validate percentile values in data rows
	testCases := []struct {
		row      int
		expected string
		desc     string
	}{
		{1, "75,3", "Normal percentile value"},
		{2, models.DataNotAvailable, "DATA_NOT_AVAILABLE case"},
		{3, "96,7", "High percentile value"},
	}

	for _, tc := range testCases {
		if tc.row >= len(records) {
			t.Errorf("Row %d not found in CSV", tc.row)
			continue
		}

		actual := records[tc.row][percentileColumn]
		if actual != tc.expected {
			t.Errorf("%s: expected %s, got %s", tc.desc, tc.expected, actual)
		}
	}

	t.Log("Percentile value validation passed")
}

// TestSomatogramExportEdgeCases tests edge cases in somatogram export
func TestSomatogramExportEdgeCases(t *testing.T) {
	exporter := New()

	// Test with edge case percentile values
	records := []models.KaderPlanungRecord{
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Edge Case Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-001",
			PKZ:                     "PKZ001",
			Firstname:               "Min",
			Lastname:                "Player",
			Birthyear:               2000,
			Gender:                  "m",
			CurrentDWZ:              800,
			ListRanking:             100,
			DWZ12MonthsAgo:          "750",
			GamesLast12Months:       "3",
			SuccessRateLast12Months: "33,3",
			SomatogramPercentile:    "0,0",             // Minimum percentile
			DWZAgeRelation:          "700",
		},
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Edge Case Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-002",
			PKZ:                     "PKZ002",
			Firstname:               "Max",
			Lastname:                "Player",
			Birthyear:               1980,
			Gender:                  "w",
			CurrentDWZ:              2500,
			ListRanking:             1,
			DWZ12MonthsAgo:          "2480",
			GamesLast12Months:       "25",
			SuccessRateLast12Months: "88,0",
			SomatogramPercentile:    "100,0",           // Maximum percentile
			DWZAgeRelation:          "2400",
		},
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Edge Case Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-003",
			PKZ:                     "PKZ003",
			Firstname:               "Decimal",
			Lastname:                "Player",
			Birthyear:               1992,
			Gender:                  "m",
			CurrentDWZ:              1456,
			ListRanking:             15,
			DWZ12MonthsAgo:          "1432",
			GamesLast12Months:       "7",
			SuccessRateLast12Months: "57,1",
			SomatogramPercentile:    "42,7",            // Decimal percentile
			DWZAgeRelation:          "1400",
		},
		{
			ClubIDPrefix1:           "C",
			ClubIDPrefix2:           "C0",
			ClubIDPrefix3:           "C03",
			ClubName:                "Edge Case Club",
			ClubID:                  "C0327",
			PlayerID:                "C0327-004",
			PKZ:                     "",                 // Empty PKZ
			Firstname:               "",                 // Empty firstname
			Lastname:                "OnlyLastname",
			Birthyear:               1988,
			Gender:                  "",                 // Empty gender
			CurrentDWZ:              0,                  // Zero DWZ
			ListRanking:             0,                  // Zero ranking
			DWZ12MonthsAgo:          models.DataNotAvailable,
			GamesLast12Months:       models.DataNotAvailable,
			SuccessRateLast12Months: models.DataNotAvailable,
			SomatogramPercentile:    models.DataNotAvailable, // Should be DATA_NOT_AVAILABLE due to missing data
			DWZAgeRelation:          models.DataNotAvailable,
		},
	}

	// Export CSV
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "edge_cases.csv")

	err := exporter.exportCSV(records, outputPath)
	if err != nil {
		t.Fatalf("Failed to export CSV with edge cases: %v", err)
	}

	// Read and validate
	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	contentStr := string(content)

	// Test specific edge case values are present
	expectedValues := []string{
		"0,0",                          // Minimum percentile
		"100,0",                        // Maximum percentile
		"42,7",                         // Decimal percentile
		models.DataNotAvailable,        // Missing data case
	}

	for _, value := range expectedValues {
		if !strings.Contains(contentStr, value) {
			t.Errorf("Expected value %s not found in CSV", value)
		}
	}

	// Test that empty fields are handled correctly
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	if len(lines) != 5 { // header + 4 records
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}

	// Check empty fields in last record
	lastRecord := strings.Split(lines[4], ";")
	if lastRecord[6] != "" { // PKZ should be empty
		t.Errorf("Expected empty PKZ, got %s", lastRecord[6])
	}
	if lastRecord[8] != "" { // Firstname should be empty
		t.Errorf("Expected empty firstname, got %s", lastRecord[8])
	}

	t.Log("Edge case validation passed")
}

// TestSomatogramCSVSemicolonSeparator tests German Excel compatibility
func TestSomatogramCSVSemicolonSeparator(t *testing.T) {
	exporter := New()

	record := models.KaderPlanungRecord{
		ClubIDPrefix1:           "C",
		ClubIDPrefix2:           "C0",
		ClubIDPrefix3:           "C03",
		ClubName:                "German Club äöü",   // Test German characters
		ClubID:                  "C0327",
		PlayerID:                "C0327-297",
		PKZ:                     "PKZ123",
		Firstname:               "Hans-Peter",         // Test hyphenated name
		Lastname:                "Müller",             // Test German umlaut
		Birthyear:               1990,
		Gender:                  "m",
		CurrentDWZ:              1600,
		ListRanking:             1,
		DWZ12MonthsAgo:          "1,550",              // Test comma in number
		GamesLast12Months:       "8",
		SuccessRateLast12Months: "62,5",              // Test German decimal format
		SomatogramPercentile:    "67,2",              // Test German decimal format
		DWZAgeRelation:          "1,500",             // Test comma in number
	}

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "german_format.csv")

	err := exporter.exportCSV([]models.KaderPlanungRecord{record}, outputPath)
	if err != nil {
		t.Fatalf("Failed to export German format CSV: %v", err)
	}

	// Read and verify semicolon separation
	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	contentStr := string(content)

	// Count semicolons in header line
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	headerSemicolons := strings.Count(lines[0], ";")
	dataSemicolons := strings.Count(lines[1], ";")

	expectedSemicolons := 17 // 18 fields - 1 = 17 separators
	if headerSemicolons != expectedSemicolons {
		t.Errorf("Header semicolons: expected %d, got %d", expectedSemicolons, headerSemicolons)
	}
	if dataSemicolons != expectedSemicolons {
		t.Errorf("Data semicolons: expected %d, got %d", expectedSemicolons, dataSemicolons)
	}

	// Test that German characters are preserved
	if !strings.Contains(contentStr, "äöü") {
		t.Error("German characters not preserved in club name")
	}
	if !strings.Contains(contentStr, "Müller") {
		t.Error("German umlaut not preserved in lastname")
	}

	// Test that decimal commas are preserved (German format)
	if !strings.Contains(contentStr, "67,2") {
		t.Error("German decimal format not preserved in somatogram percentile")
	}

	t.Log("German CSV format validation passed")
}
