package statistics

import (
	"fmt"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
)

// TestStatisticalAnalyzer_ProcessPlayers tests the main statistical processing functionality
func TestStatisticalAnalyzer_ProcessPlayers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise during testing

	t.Run("Empty player list", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(10, logger)
		players := []models.Player{}

		results, err := analyzer.ProcessPlayers(players)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected empty results, got %d genders", len(results))
		}
	})

	t.Run("Small sample below minimum", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(50, logger) // High minimum
		players := createTestPlayers(t, 20, "m", 25, 1500, 100)

		results, err := analyzer.ProcessPlayers(players)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should have no results because sample is too small
		if len(results) != 0 {
			t.Errorf("Expected no results due to small sample, got %d", len(results))
		}
	})

	t.Run("Valid male players statistical analysis", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(10, logger)
		players := createTestPlayers(t, 100, "m", 25, 1500, 200)

		results, err := analyzer.ProcessPlayers(players)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should have results for male players
		if len(results) != 1 {
			t.Errorf("Expected 1 gender result, got %d", len(results))
		}

		maleData, exists := results["m"]
		if !exists {
			t.Error("Expected male data to exist")
		}

		// Check metadata
		if maleData.Metadata.Gender != "m" {
			t.Errorf("Expected gender 'm', got '%s'", maleData.Metadata.Gender)
		}
		if maleData.Metadata.TotalPlayers != 100 {
			t.Errorf("Expected 100 total players, got %d", maleData.Metadata.TotalPlayers)
		}
		if maleData.Metadata.MinSampleSize != 10 {
			t.Errorf("Expected min sample size 10, got %d", maleData.Metadata.MinSampleSize)
		}

		// Should have age group data
		if len(maleData.Percentiles) == 0 {
			t.Error("Expected percentile data for age groups")
		}

		// Check that we have the expected age group
		ageStr := "25" // We created players with age 25

		percentileData, exists := maleData.Percentiles[ageStr]
		if !exists {
			t.Errorf("Expected percentile data for age %s", ageStr)
		}

		// Validate percentile data
		if percentileData.Age != 25 {
			t.Errorf("Expected age 25, got %d", percentileData.Age)
		}
		if percentileData.SampleSize != 100 {
			t.Errorf("Expected sample size 100, got %d", percentileData.SampleSize)
		}
		if percentileData.AvgDWZ < 1300 || percentileData.AvgDWZ > 1700 {
			t.Errorf("Expected average DWZ around 1500, got %.1f", percentileData.AvgDWZ)
		}

		// Check percentile calculations
		if len(percentileData.Percentiles) == 0 {
			t.Error("Expected percentile calculations")
		}

		// Verify some key percentiles exist
		if _, exists := percentileData.Percentiles[50]; !exists {
			t.Error("Expected 50th percentile")
		}
		if _, exists := percentileData.Percentiles[90]; !exists {
			t.Error("Expected 90th percentile")
		}
		if _, exists := percentileData.Percentiles[10]; !exists {
			t.Error("Expected 10th percentile")
		}
	})

	t.Run("Multiple gender analysis", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(5, logger)

		// Create mixed gender players
		malePlayers := createTestPlayers(t, 50, "m", 30, 1600, 100)
		femalePlayers := createTestPlayers(t, 40, "w", 30, 1550, 120)

		allPlayers := append(malePlayers, femalePlayers...)

		results, err := analyzer.ProcessPlayers(allPlayers)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should have results for both genders
		if len(results) != 2 {
			t.Errorf("Expected 2 gender results, got %d", len(results))
		}

		maleData, maleExists := results["m"]
		femaleData, femaleExists := results["w"]

		if !maleExists {
			t.Error("Expected male data to exist")
		}
		if !femaleExists {
			t.Error("Expected female data to exist")
		}

		// Verify gender-specific totals
		if maleData.Metadata.TotalPlayers != 50 {
			t.Errorf("Expected 50 male players, got %d", maleData.Metadata.TotalPlayers)
		}
		if femaleData.Metadata.TotalPlayers != 40 {
			t.Errorf("Expected 40 female players, got %d", femaleData.Metadata.TotalPlayers)
		}
	})

	t.Run("Multiple age groups per gender", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(8, logger)

		var allPlayers []models.Player

		// Create multiple age groups
		ages := []int{20, 25, 30, 35}
		for _, age := range ages {
			players := createTestPlayers(t, 10, "m", age, 1500, 50)
			allPlayers = append(allPlayers, players...)
		}

		results, err := analyzer.ProcessPlayers(allPlayers)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		maleData, exists := results["m"]
		if !exists {
			t.Error("Expected male data to exist")
		}

		// Should have data for all age groups
		expectedAges := len(ages)
		if len(maleData.Percentiles) != expectedAges {
			t.Errorf("Expected %d age groups, got %d", expectedAges, len(maleData.Percentiles))
		}

		// Check ages that should definitely exist
		found := false
		for key := range maleData.Percentiles {
			if key == "20" || key == "25" || key == "30" || key == "35" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find age group data")
		}
	})
}

// TestStatisticalAnalyzer_PercentileCalculations tests percentile accuracy
func TestStatisticalAnalyzer_PercentileCalculations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("Percentile calculation accuracy", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(1, logger)

		// Create players with known DWZ distribution: 1300, 1400, 1500, 1600, 1700
		players := []models.Player{}
		dwzValues := []int{1300, 1400, 1500, 1600, 1700}
		birthYear := 1999

		for i, dwz := range dwzValues {
			player := models.Player{
				ID:         fmt.Sprintf("TEST-%d", i),
				CurrentDWZ: dwz,
				BirthYear:  &birthYear,
				Gender:     "m",
				Status:     "active",
				Active:     true,
			}
			players = append(players, player)
		}

		percentiles := analyzer.calculatePercentiles(players)

		// Test specific percentiles
		if percentiles[0] != 1300 {
			t.Errorf("Expected 0th percentile to be 1300, got %d", percentiles[0])
		}
		if percentiles[100] != 1700 {
			t.Errorf("Expected 100th percentile to be 1700, got %d", percentiles[100])
		}
		if percentiles[50] != 1500 {
			t.Errorf("Expected 50th percentile to be 1500, got %d", percentiles[50])
		}
		if percentiles[25] != 1400 {
			t.Errorf("Expected 25th percentile to be 1400, got %d", percentiles[25])
		}
		if percentiles[75] != 1600 {
			t.Errorf("Expected 75th percentile to be 1600, got %d", percentiles[75])
		}
	})
}

// TestStatisticalAnalyzer_AverageCalculations tests average and median calculations
func TestStatisticalAnalyzer_AverageCalculations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	analyzer := NewStatisticalAnalyzer(1, logger)

	t.Run("Average DWZ calculation", func(t *testing.T) {
		players := []models.Player{
			{CurrentDWZ: 1200},
			{CurrentDWZ: 1400},
			{CurrentDWZ: 1600},
			{CurrentDWZ: 1800},
		}

		avg := analyzer.calculateAverageDWZ(players)
		expectedAvg := 1500.0

		if avg != expectedAvg {
			t.Errorf("Expected average DWZ %.1f, got %.1f", expectedAvg, avg)
		}
	})

	t.Run("Median DWZ calculation - odd count", func(t *testing.T) {
		players := []models.Player{
			{CurrentDWZ: 1200},
			{CurrentDWZ: 1400},
			{CurrentDWZ: 1600},
		}

		median := analyzer.calculateMedianDWZ(players)
		expectedMedian := 1400

		if median != expectedMedian {
			t.Errorf("Expected median DWZ %d, got %d", expectedMedian, median)
		}
	})

	t.Run("Median DWZ calculation - even count", func(t *testing.T) {
		players := []models.Player{
			{CurrentDWZ: 1200},
			{CurrentDWZ: 1400},
			{CurrentDWZ: 1600},
			{CurrentDWZ: 1800},
		}

		median := analyzer.calculateMedianDWZ(players)
		expectedMedian := 1500 // (1400 + 1600) / 2

		if median != expectedMedian {
			t.Errorf("Expected median DWZ %d, got %d", expectedMedian, median)
		}
	})
}

// TestStatisticalAnalyzer_EdgeCases tests edge cases and error conditions
func TestStatisticalAnalyzer_EdgeCases(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	t.Run("Players with nil birth year", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(5, logger)
		players := []models.Player{
			{
				ID:         "TEST-1",
				CurrentDWZ: 1500,
				BirthYear:  nil, // nil birth year
				Gender:     "m",
				Status:     "active",
				Active:     true,
			},
		}

		results, err := analyzer.ProcessPlayers(players)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should have no results because player without birth year is skipped
		if len(results) != 0 {
			t.Errorf("Expected no results for players without birth year, got %d", len(results))
		}
	})

	t.Run("Empty percentile calculation", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(1, logger)
		players := []models.Player{}

		percentiles := analyzer.calculatePercentiles(players)

		if len(percentiles) != 0 {
			t.Errorf("Expected empty percentiles, got %d values", len(percentiles))
		}
	})

	t.Run("Zero average and median", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(1, logger)
		players := []models.Player{}

		avg := analyzer.calculateAverageDWZ(players)
		median := analyzer.calculateMedianDWZ(players)

		if avg != 0 {
			t.Errorf("Expected average 0, got %.1f", avg)
		}
		if median != 0 {
			t.Errorf("Expected median 0, got %d", median)
		}
	})

	t.Run("Min sample size configuration", func(t *testing.T) {
		analyzer := NewStatisticalAnalyzer(25, logger)

		// Test getter
		if analyzer.GetMinSampleSize() != 25 {
			t.Errorf("Expected min sample size 25, got %d", analyzer.GetMinSampleSize())
		}

		// Test setter
		analyzer.SetMinSampleSize(50)
		if analyzer.GetMinSampleSize() != 50 {
			t.Errorf("Expected min sample size 50 after update, got %d", analyzer.GetMinSampleSize())
		}
	})
}

// createTestPlayers is a helper function to create test players with specified parameters
func createTestPlayers(t *testing.T, count int, gender string, age int, baseDWZ int, dwzVariance int) []models.Player {
	t.Helper()

	players := make([]models.Player, count)
	currentYear := time.Now().Year()
	birthYear := currentYear - age

	for i := 0; i < count; i++ {
		// Create some DWZ variance
		dwzOffset := (i % dwzVariance) - (dwzVariance / 2)
		dwz := baseDWZ + dwzOffset

		player := models.Player{
			ID:         fmt.Sprintf("TEST-%s-%d-%d", gender, age, i),
			CurrentDWZ: dwz,
			BirthYear:  &birthYear,
			Gender:     gender,
			Status:     "active",
			Active:     true,
			Name:       fmt.Sprintf("TestPlayer%d", i),
			Firstname:  fmt.Sprintf("Test%d", i),
		}
		players[i] = player
	}

	return players
}

