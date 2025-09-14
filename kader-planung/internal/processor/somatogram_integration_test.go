package processor

import (
	"fmt"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
)

// TestSomatogramIntegrationPipeline tests the complete pipeline with somatogram percentiles
func TestSomatogramIntegrationPipeline(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test validates the integration logic without actual API calls
	// We test the percentile calculation components directly
	
	processor := &Processor{
		logger: logrus.New(),
	}

	// Test the core percentile calculation with mock data
	players := createLargeDatasetForPerformanceTest(1000) // Sufficient dataset for testing (50+ per group)

	startTime := time.Now()
	
	// Test percentile calculation directly
	percentileMap, err := processor.calculateSomatogramPercentiles(players)
	if err != nil {
		t.Fatalf("Percentile calculation failed: %v", err)
	}

	duration := time.Since(startTime)
	t.Logf("Integration test completed in %v", duration)

	// Validate results
	if len(percentileMap) == 0 {
		t.Error("Expected at least some percentiles from pipeline")
	}

	// Test that all percentiles are in valid range
	for playerID, percentile := range percentileMap {
		if percentile < 0 || percentile > 100 {
			t.Errorf("Percentile out of range for player %s: %.2f", playerID, percentile)
		}
	}

	t.Logf("Results: %d players with calculated percentiles", len(percentileMap))

	// Test performance expectations
	if duration > 10*time.Second {
		t.Errorf("Pipeline took too long: %v (expected < 10s for integration test)", duration)
	}
}

// TestSomatogramAccuracyValidation tests percentile accuracy against known distributions
func TestSomatogramAccuracyValidation(t *testing.T) {
	processor := &Processor{
		logger: logrus.New(),
	}

	// Create test data with known statistical properties
	players := createStatisticalTestData()

	// Calculate percentiles
	percentileMap, err := processor.calculateSomatogramPercentiles(players)
	if err != nil {
		t.Fatalf("Failed to calculate percentiles: %v", err)
	}

	// Test statistical properties
	testPercentileDistribution(t, players, percentileMap)
	testPercentileAccuracy(t, players, percentileMap)
}

// TestSomatogramPerformanceWithLargeDataset tests performance with realistic dataset size
func TestSomatogramPerformanceWithLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	processor := &Processor{
		logger: logrus.New(),
	}

	// Create large dataset (simulate Germany-wide data)
	startTime := time.Now()
	players := createLargeDatasetForPerformanceTest(10000) // 10k players
	dataCreationTime := time.Since(startTime)

	t.Logf("Created %d test players in %v", len(players), dataCreationTime)

	// Test percentile calculation performance
	startTime = time.Now()
	percentileMap, err := processor.calculateSomatogramPercentiles(players)
	calculationTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Failed to calculate percentiles for large dataset: %v", err)
	}

	t.Logf("Calculated percentiles for %d players in %v", len(percentileMap), calculationTime)

	// Performance expectations
	if calculationTime > 10*time.Second {
		t.Errorf("Percentile calculation too slow: %v (expected < 10s)", calculationTime)
	}

	// Memory efficiency test
	if len(percentileMap) == 0 {
		t.Error("No percentiles calculated - possible minimum sample size issue")
	}

	// Validate results are reasonable
	percentileSum := 0.0
	for _, percentile := range percentileMap {
		percentileSum += percentile
	}
	avgPercentile := percentileSum / float64(len(percentileMap))

	// Average percentile should be around 50
	if avgPercentile < 30 || avgPercentile > 70 {
		t.Errorf("Average percentile suspicious: %.2f (expected ~50)", avgPercentile)
	}
}

// Helper functions for integration tests

func createStatisticalTestData() []models.Player {
	players := []models.Player{}

	// Create two age-gender groups with known distributions
	// Group 1: 30-year-old males with normal DWZ distribution
	birthyear := 1994
	for i := 0; i < 100; i++ {
		dwz := 1400 + int(float64(i)*6) // Linear distribution 1400-1994
		players = append(players, models.Player{
			ID:         fmt.Sprintf("M30-%03d", i),
			BirthYear:  &birthyear,
			Gender:     "m",
			CurrentDWZ: dwz,
		})
	}

	// Group 2: 25-year-old females with different distribution
	birthyear2 := 1999
	for i := 0; i < 80; i++ {
		dwz := 1300 + int(float64(i)*7) // Linear distribution 1300-1859
		players = append(players, models.Player{
			ID:         fmt.Sprintf("W25-%03d", i),
			BirthYear:  &birthyear2,
			Gender:     "w",
			CurrentDWZ: dwz,
		})
	}

	return players
}

func createLargeDatasetForPerformanceTest(size int) []models.Player {
	players := make([]models.Player, 0, size)
	
	ages := []int{1970, 1975, 1980, 1985, 1990, 1995, 2000}
	genders := []string{"m", "w"}
	
	for i := 0; i < size; i++ {
		birthyear := ages[i%len(ages)]
		gender := genders[i%len(genders)]
		dwz := 1000 + (i % 1500) // DWZ range 1000-2499
		
		players = append(players, models.Player{
			ID:         fmt.Sprintf("PERF-%05d", i),
			BirthYear:  &birthyear,
			Gender:     gender,
			CurrentDWZ: dwz,
		})
	}

	return players
}

func validateRecordCompleteness(t *testing.T, record models.KaderPlanungRecord) {
	// Test required fields are present
	if record.ClubID == "" {
		t.Errorf("Missing ClubID for player %s", record.PlayerID)
	}
	if record.PlayerID == "" {
		t.Error("Missing PlayerID in record")
	}
	if record.Firstname == "" && record.Lastname == "" {
		t.Errorf("Missing both names for player %s", record.PlayerID)
	}

	// Test prefix fields are calculated
	if record.ClubID != "" {
		expectedPrefix1 := string(record.ClubID[0])
		if record.ClubIDPrefix1 != expectedPrefix1 {
			t.Errorf("Incorrect ClubIDPrefix1 for %s: expected %s, got %s", 
				record.ClubID, expectedPrefix1, record.ClubIDPrefix1)
		}
	}

	// Test somatogram percentile field exists
	if record.SomatogramPercentile == "" {
		t.Errorf("Missing SomatogramPercentile for player %s", record.PlayerID)
	}
}

func testPercentileDistribution(t *testing.T, players []models.Player, percentileMap map[string]float64) {
	if len(percentileMap) < 50 {
		t.Skip("Not enough players for distribution testing")
	}

	// Collect all percentiles
	percentiles := make([]float64, 0, len(percentileMap))
	for _, p := range percentileMap {
		percentiles = append(percentiles, p)
	}

	// Test percentile distribution properties
	minP := percentiles[0]
	maxP := percentiles[0]
	sum := 0.0

	for _, p := range percentiles {
		if p < minP {
			minP = p
		}
		if p > maxP {
			maxP = p
		}
		sum += p
	}

	avg := sum / float64(len(percentiles))

	// Test statistical properties
	if minP < 0 || maxP > 100 {
		t.Errorf("Percentiles out of valid range: %.2f - %.2f", minP, maxP)
	}

	// Average should be reasonable (not testing exact 50 due to filtering)
	if avg < 10 || avg > 90 {
		t.Errorf("Average percentile suspicious: %.2f", avg)
	}

	t.Logf("Percentile distribution: min=%.1f, max=%.1f, avg=%.1f", minP, maxP, avg)
}

func testPercentileAccuracy(t *testing.T, players []models.Player, percentileMap map[string]float64) {
	// Group players by age-gender for accuracy testing
	groups := make(map[string][]models.Player)
	
	for _, player := range players {
		if player.BirthYear != nil && player.Gender != "" {
			// Use birth year as age for simplicity in test
			key := fmt.Sprintf("%d-%s", *player.BirthYear, player.Gender)
			groups[key] = append(groups[key], player)
		}
	}

	// Test accuracy within each group
	for groupKey, groupPlayers := range groups {
		if len(groupPlayers) < 10 {
			continue // Skip small groups
		}

		// Find min and max DWZ in group
		minDWZ := groupPlayers[0].CurrentDWZ
		maxDWZ := groupPlayers[0].CurrentDWZ
		
		for _, player := range groupPlayers {
			if player.CurrentDWZ < minDWZ {
				minDWZ = player.CurrentDWZ
			}
			if player.CurrentDWZ > maxDWZ {
				maxDWZ = player.CurrentDWZ
			}
		}

		// Test that player with lowest DWZ has low percentile
		var minPlayer, maxPlayer models.Player
		for _, player := range groupPlayers {
			if player.CurrentDWZ == minDWZ {
				minPlayer = player
			}
			if player.CurrentDWZ == maxDWZ {
				maxPlayer = player
			}
		}

		minPercentile, minExists := percentileMap[minPlayer.ID]
		maxPercentile, maxExists := percentileMap[maxPlayer.ID]

		if minExists && maxExists && minDWZ != maxDWZ {
			if minPercentile > maxPercentile {
				t.Errorf("In group %s: min DWZ player has higher percentile (%.1f) than max DWZ player (%.1f)", 
					groupKey, minPercentile, maxPercentile)
			}
		}
	}
}
