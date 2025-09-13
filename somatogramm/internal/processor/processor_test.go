package processor

import (
	"strconv"
	"testing"
	"time"

	"somatogramm/internal/models"
)

func TestNewProcessor(t *testing.T) {
	processor := NewProcessor(100, true)

	if processor.MinSampleSize != 100 {
		t.Errorf("expected MinSampleSize 100, got %d", processor.MinSampleSize)
	}

	if !processor.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestGroupPlayersByAgeAndGender(t *testing.T) {
	processor := NewProcessor(10, false)

	birthYear1990 := 1990
	birthYear2000 := 2000

	players := []models.Player{
		{ID: "TEST-1", BirthYear: &birthYear1990, Gender: "m", CurrentDWZ: 1500},
		{ID: "TEST-2", BirthYear: &birthYear1990, Gender: "m", CurrentDWZ: 1600},
		{ID: "TEST-3", BirthYear: &birthYear1990, Gender: "w", CurrentDWZ: 1400},
		{ID: "TEST-4", BirthYear: &birthYear2000, Gender: "m", CurrentDWZ: 1200},
		{ID: "TEST-5", BirthYear: nil, Gender: "m", CurrentDWZ: 1300}, // Should be skipped
	}

	groups := processor.groupPlayersByAgeAndGender(players)

	// Ages are calculated in the grouping function, not used directly in this test

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	// Check that players with nil birth year are skipped
	totalPlayersInGroups := 0
	for _, group := range groups {
		totalPlayersInGroups += len(group.Players)
	}

	if totalPlayersInGroups != 4 { // 5 players - 1 with nil birth year
		t.Errorf("expected 4 players in groups, got %d", totalPlayersInGroups)
	}
}

func TestFilterGroupsBySampleSize(t *testing.T) {
	processor := NewProcessor(2, false) // MinSampleSize = 2

	groups := map[string]models.AgeGenderGroup{
		"m-25": {
			Age:        25,
			Gender:     "m",
			Players:    make([]models.Player, 3), // 3 players - meets requirement
			SampleSize: 3,
		},
		"w-25": {
			Age:        25,
			Gender:     "w",
			Players:    make([]models.Player, 1), // 1 player - below requirement
			SampleSize: 1,
		},
		"m-30": {
			Age:        30,
			Gender:     "m",
			Players:    make([]models.Player, 5), // 5 players - meets requirement
			SampleSize: 5,
		},
	}

	validGroups := processor.filterGroupsBySampleSize(groups)

	if len(validGroups) != 2 {
		t.Errorf("expected 2 valid groups, got %d", len(validGroups))
	}

	// Check sorting: groups should be sorted by gender, then age
	if len(validGroups) >= 2 {
		if validGroups[0].Gender > validGroups[1].Gender {
			t.Error("groups are not sorted by gender")
		}
		if validGroups[0].Gender == validGroups[1].Gender && validGroups[0].Age > validGroups[1].Age {
			t.Error("groups are not sorted by age within same gender")
		}
	}
}

func TestCalculatePercentiles(t *testing.T) {
	processor := NewProcessor(10, false)

	// Test with simple dataset: [100, 200, 300, 400, 500]
	players := []models.Player{
		{CurrentDWZ: 100},
		{CurrentDWZ: 200},
		{CurrentDWZ: 300},
		{CurrentDWZ: 400},
		{CurrentDWZ: 500},
	}

	percentiles := processor.calculatePercentiles(players)

	// Test key percentiles
	if percentiles[0] != 100 {
		t.Errorf("expected 0th percentile to be 100, got %d", percentiles[0])
	}

	if percentiles[50] != 300 { // median should be middle value
		t.Errorf("expected 50th percentile to be 300, got %d", percentiles[50])
	}

	if percentiles[100] != 500 {
		t.Errorf("expected 100th percentile to be 500, got %d", percentiles[100])
	}

	// Check that all percentiles from 0-100 exist
	for p := 0; p <= 100; p++ {
		if _, exists := percentiles[p]; !exists {
			t.Errorf("percentile %d is missing", p)
		}
	}
}

func TestCalculatePercentilesEmptySlice(t *testing.T) {
	processor := NewProcessor(10, false)

	players := []models.Player{}
	percentiles := processor.calculatePercentiles(players)

	if len(percentiles) != 0 {
		t.Errorf("expected empty percentiles map, got %d entries", len(percentiles))
	}
}

func TestCalculatePercentilesSingleValue(t *testing.T) {
	processor := NewProcessor(10, false)

	players := []models.Player{
		{CurrentDWZ: 1500},
	}

	percentiles := processor.calculatePercentiles(players)

	// All percentiles should be the same value
	for p := 0; p <= 100; p++ {
		if percentiles[p] != 1500 {
			t.Errorf("expected percentile %d to be 1500, got %d", p, percentiles[p])
		}
	}
}

func TestCalculateAverageDWZ(t *testing.T) {
	processor := NewProcessor(10, false)

	players := []models.Player{
		{CurrentDWZ: 1000},
		{CurrentDWZ: 1500},
		{CurrentDWZ: 2000},
	}

	avg := processor.calculateAverageDWZ(players)
	expected := 1500.0 // (1000 + 1500 + 2000) / 3

	if avg != expected {
		t.Errorf("expected average DWZ %.2f, got %.2f", expected, avg)
	}
}

func TestCalculateAverageDWZEmptySlice(t *testing.T) {
	processor := NewProcessor(10, false)

	players := []models.Player{}
	avg := processor.calculateAverageDWZ(players)

	if avg != 0 {
		t.Errorf("expected average DWZ 0 for empty slice, got %.2f", avg)
	}
}

func TestCalculateMedianDWZ(t *testing.T) {
	processor := NewProcessor(10, false)

	// Test odd number of players
	playersOdd := []models.Player{
		{CurrentDWZ: 1000},
		{CurrentDWZ: 1500},
		{CurrentDWZ: 2000},
	}

	medianOdd := processor.calculateMedianDWZ(playersOdd)
	if medianOdd != 1500 {
		t.Errorf("expected median DWZ 1500 for odd count, got %d", medianOdd)
	}

	// Test even number of players
	playersEven := []models.Player{
		{CurrentDWZ: 1000},
		{CurrentDWZ: 1500},
		{CurrentDWZ: 2000},
		{CurrentDWZ: 2500},
	}

	medianEven := processor.calculateMedianDWZ(playersEven)
	expected := (1500 + 2000) / 2 // 1750
	if medianEven != expected {
		t.Errorf("expected median DWZ %d for even count, got %d", expected, medianEven)
	}
}

func TestCalculateMedianDWZEmptySlice(t *testing.T) {
	processor := NewProcessor(10, false)

	players := []models.Player{}
	median := processor.calculateMedianDWZ(players)

	if median != 0 {
		t.Errorf("expected median DWZ 0 for empty slice, got %d", median)
	}
}

func TestProcessPlayers(t *testing.T) {
	processor := NewProcessor(2, false) // MinSampleSize = 2

	currentYear := time.Now().Year()
	birthYear1990 := 1990

	players := []models.Player{
		{ID: "TEST-1", BirthYear: &birthYear1990, Gender: "m", CurrentDWZ: 1500},
		{ID: "TEST-2", BirthYear: &birthYear1990, Gender: "m", CurrentDWZ: 1600},
		{ID: "TEST-3", BirthYear: &birthYear1990, Gender: "w", CurrentDWZ: 1400}, // Only 1 female player - below min sample size
	}

	processedData, err := processor.ProcessPlayers(players)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should have data for males (2 players) but not females (1 player < minSampleSize)
	if len(processedData) != 1 {
		t.Errorf("expected 1 gender in processed data, got %d", len(processedData))
	}

	if _, exists := processedData["m"]; !exists {
		t.Error("expected processed data for males")
	}

	if _, exists := processedData["w"]; exists {
		t.Error("did not expect processed data for females (below min sample size)")
	}

	// Check metadata
	maleData := processedData["m"]
	if maleData.Metadata.Gender != "m" {
		t.Errorf("expected gender 'm', got %s", maleData.Metadata.Gender)
	}

	if maleData.Metadata.TotalPlayers != 2 {
		t.Errorf("expected 2 total players, got %d", maleData.Metadata.TotalPlayers)
	}

	expectedAge := currentYear - 1990
	ageKey := string(rune(expectedAge + '0'))
	if len(ageKey) == 1 {
		ageKey = string(rune(expectedAge/10+'0')) + string(rune(expectedAge%10+'0'))
	}

	// For proper age key construction
	ageKeyStr := ""
	age := expectedAge
	for age > 0 {
		ageKeyStr = string(rune(age%10+'0')) + ageKeyStr
		age /= 10
	}

	// Use string conversion instead
	ageKeyStr = strconv.Itoa(expectedAge)

	if _, exists := maleData.Percentiles[ageKeyStr]; !exists {
		t.Errorf("expected percentile data for age %s", ageKeyStr)
	}
}

func TestVerboseLogging(t *testing.T) {
	processor := NewProcessor(10, true)

	// Test that verbose logging doesn't panic
	processor.log("test message")

	if !processor.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestNonVerboseLogging(t *testing.T) {
	processor := NewProcessor(10, false)

	// Test that non-verbose logging doesn't panic
	processor.log("test message")

	if processor.Verbose {
		t.Error("expected Verbose to be false")
	}
}

func TestFilterGroupsBySampleSizeEdgeCases(t *testing.T) {
	processor := NewProcessor(5, true) // MinSampleSize = 5, verbose = true

	groups := map[string]models.AgeGenderGroup{
		"m-25": {
			Age:        25,
			Gender:     "m",
			Players:    make([]models.Player, 5), // Exactly at minimum
			SampleSize: 5,
		},
		"w-30": {
			Age:        30,
			Gender:     "w",
			Players:    make([]models.Player, 4), // Below minimum
			SampleSize: 4,
		},
		"d-35": {
			Age:        35,
			Gender:     "d",
			Players:    make([]models.Player, 10), // Above minimum
			SampleSize: 10,
		},
	}

	validGroups := processor.filterGroupsBySampleSize(groups)

	// Should include groups with 5 and 10 players, exclude group with 4 players
	if len(validGroups) != 2 {
		t.Errorf("expected 2 valid groups, got %d", len(validGroups))
	}

	// Check that groups are properly sorted
	found25 := false
	found35 := false
	for _, group := range validGroups {
		if group.Age == 25 && group.Gender == "m" {
			found25 = true
		}
		if group.Age == 35 && group.Gender == "d" {
			found35 = true
		}
	}

	if !found25 {
		t.Error("expected to find age 25 male group")
	}
	if !found35 {
		t.Error("expected to find age 35 divers group")
	}
}