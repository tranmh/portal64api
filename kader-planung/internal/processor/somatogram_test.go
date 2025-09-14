package processor

import (
	"fmt"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
)

// TestGroupPlayersByAgeAndGender tests the age-gender grouping logic
func TestGroupPlayersByAgeAndGender(t *testing.T) {
	processor := &Processor{
		logger: logrus.New(),
	}

	// Create test players with different ages and genders
	birthyear1990 := 1990
	birthyear1985 := 1985
	birthyear2000 := 2000

	players := []models.Player{
		{
			ID:         "C0327-001",
			BirthYear:  &birthyear1990,
			Gender:     "m",
			CurrentDWZ: 1500,
		},
		{
			ID:         "C0327-002",
			BirthYear:  &birthyear1990,
			Gender:     "m",
			CurrentDWZ: 1600,
		},
		{
			ID:         "C0327-003",
			BirthYear:  &birthyear1985,
			Gender:     "w",
			CurrentDWZ: 1550,
		},
		{
			ID:         "C0327-004",
			BirthYear:  &birthyear2000,
			Gender:     "m",
			CurrentDWZ: 1400,
		},
		{
			ID:         "C0327-005",
			BirthYear:  nil, // Missing birth year
			Gender:     "m",
			CurrentDWZ: 1700,
		},
		{
			ID:         "C0327-006",
			BirthYear:  &birthyear1990,
			Gender:     "", // Missing gender
			CurrentDWZ: 1800,
		},
	}

	groups := processor.groupPlayersByAgeAndGender(players)

	// Test group count - should create groups for valid combinations only
	expectedGroups := 3 // (1990,m), (1985,w), (2000,m)
	if len(groups) != expectedGroups {
		t.Logf("Actual groups found:")
		for key, group := range groups {
			t.Logf("  %s: age=%d, gender=%s, size=%d", key, group.Age, group.Gender, group.SampleSize)
		}
		t.Errorf("Expected %d groups, got %d", expectedGroups, len(groups))
	}

	// Test group sizes
	groupSizes := make(map[string]int)
	for _, group := range groups {
		key := fmt.Sprintf("%d-%s", group.Age, group.Gender)
		groupSizes[key] = group.SampleSize
	}

	// Check expected group sizes (calculate current ages dynamically)
	currentYear := time.Now().Year()
	age1990 := currentYear - 1990
	age1985 := currentYear - 1985
	age2000 := currentYear - 2000

	if groupSizes[fmt.Sprintf("%d-m", age1990)] != 2 { // 1990 males
		t.Errorf("Expected 2 players in %d-m group, got %d", age1990, groupSizes[fmt.Sprintf("%d-m", age1990)])
	}
	if groupSizes[fmt.Sprintf("%d-w", age1985)] != 1 { // 1985 female
		t.Errorf("Expected 1 player in %d-w group, got %d", age1985, groupSizes[fmt.Sprintf("%d-w", age1985)])
	}
	if groupSizes[fmt.Sprintf("%d-m", age2000)] != 1 { // 2000 male
		t.Errorf("Expected 1 player in %d-m group, got %d", age2000, groupSizes[fmt.Sprintf("%d-m", age2000)])
	}

	// Test that players with missing data are excluded
	totalPlayersinGroups := 0
	for _, group := range groups {
		totalPlayersinGroups += len(group.Players)
	}
	if totalPlayersinGroups != 4 { // 6 players - 2 with missing data
		t.Errorf("Expected 4 players in valid groups, got %d", totalPlayersinGroups)
	}
}

// TestCalculatePercentilesForGroup tests percentile calculation accuracy
func TestCalculatePercentilesForGroup(t *testing.T) {
	processor := &Processor{
		logger: logrus.New(),
	}

	// Create test players with known DWZ distribution
	players := []models.Player{
		{ID: "P001", CurrentDWZ: 1000},
		{ID: "P002", CurrentDWZ: 1100},
		{ID: "P003", CurrentDWZ: 1200},
		{ID: "P004", CurrentDWZ: 1300},
		{ID: "P005", CurrentDWZ: 1400},
		{ID: "P006", CurrentDWZ: 1500},
		{ID: "P007", CurrentDWZ: 1600},
		{ID: "P008", CurrentDWZ: 1700},
		{ID: "P009", CurrentDWZ: 1800},
		{ID: "P010", CurrentDWZ: 1900},
	}

	percentiles := processor.calculatePercentilesForGroup(players)

	// Test boundary percentiles
	if percentiles[0] != 1000 {
		t.Errorf("Expected 0th percentile to be 1000, got %d", percentiles[0])
	}
	if percentiles[100] != 1900 {
		t.Errorf("Expected 100th percentile to be 1900, got %d", percentiles[100])
	}

	// Test median (50th percentile)
	// expected50th := 1450 // Should be between 1400 and 1500
	if percentiles[50] < 1400 || percentiles[50] > 1500 {
		t.Errorf("Expected 50th percentile to be around 1450, got %d", percentiles[50])
	}

	// Test percentiles are monotonically increasing
	for p := 1; p <= 100; p++ {
		if percentiles[p] < percentiles[p-1] {
			t.Errorf("Percentiles should be monotonically increasing, but %d < %d at percentile %d", 
				percentiles[p], percentiles[p-1], p)
		}
	}

	// Test empty group
	emptyPercentiles := processor.calculatePercentilesForGroup([]models.Player{})
	if len(emptyPercentiles) != 0 {
		t.Error("Empty group should return empty percentile map")
	}
}

// TestFindPercentileForPlayer tests individual player percentile lookup
func TestFindPercentileForPlayer(t *testing.T) {
	processor := &Processor{
		logger: logrus.New(),
	}

	// Create sample percentile distribution
	groupPercentiles := map[int]int{
		0:   1000,
		25:  1250,
		50:  1500,
		75:  1750,
		100: 2000,
	}

	testCases := []struct {
		name     string
		playerDWZ int
		expected float64
	}{
		{
			name:     "Player at 0th percentile",
			playerDWZ: 1000,
			expected: 0.0,
		},
		{
			name:     "Player at 50th percentile",
			playerDWZ: 1500,
			expected: 50.0,
		},
		{
			name:     "Player above 100th percentile",
			playerDWZ: 2200,
			expected: 100.0,
		},
		{
			name:     "Player below 0th percentile",
			playerDWZ: 800,
			expected: 0.0,
		},
		{
			name:     "Player between percentiles",
			playerDWZ: 1400,
			expected: 25.0, // Should find the lower percentile
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			player := models.Player{
				ID:         "TEST",
				CurrentDWZ: tc.playerDWZ,
			}

			result := processor.findPercentileForPlayer(player, groupPercentiles)
			if result != tc.expected {
				t.Errorf("Expected percentile %.1f for DWZ %d, got %.1f", 
					tc.expected, tc.playerDWZ, result)
			}
		})
	}
}

// TestFilterGroupsBySampleSize tests minimum sample size filtering
func TestFilterGroupsBySampleSize(t *testing.T) {
	processor := &Processor{
		logger: logrus.New(),
	}

	// Create test groups with different sample sizes
	groups := map[string]AgeGenderGroup{
		"25-m": {
			Age:        25,
			Gender:     "m",
			SampleSize: 100,
			Players:    make([]models.Player, 100),
		},
		"30-w": {
			Age:        30,
			Gender:     "w",
			SampleSize: 25,
			Players:    make([]models.Player, 25),
		},
		"35-m": {
			Age:        35,
			Gender:     "m",
			SampleSize: 75,
			Players:    make([]models.Player, 75),
		},
		"40-w": {
			Age:        40,
			Gender:     "w",
			SampleSize: 10,
			Players:    make([]models.Player, 10),
		},
	}

	// Test with minimum sample size of 50
	minSampleSize := 50
	validGroups := processor.filterGroupsBySampleSize(groups, minSampleSize)

	// Should keep groups with >= 50 players
	expectedValidGroups := 2 // age 25 (100 players) and age 35 (75 players)
	if len(validGroups) != expectedValidGroups {
		t.Errorf("Expected %d valid groups with min sample size %d, got %d", 
			expectedValidGroups, minSampleSize, len(validGroups))
	}

	// Verify correct groups were kept
	validAges := make(map[int]bool)
	for _, group := range validGroups {
		validAges[group.Age] = true
	}

	if !validAges[25] || !validAges[35] {
		t.Error("Expected to keep groups with ages 25 and 35")
	}

	// Test with minimum sample size of 0 (should keep all)
	allGroups := processor.filterGroupsBySampleSize(groups, 0)
	if len(allGroups) != len(groups) {
		t.Errorf("Expected to keep all %d groups with min sample size 0, got %d", 
			len(groups), len(allGroups))
	}
}

// TestCalculateSomatogramPercentilesIntegration tests the complete percentile calculation flow
func TestCalculateSomatogramPercentilesIntegration(t *testing.T) {
	processor := &Processor{
		logger: logrus.New(),
	}

	// Create realistic test data with multiple age-gender groups
	players := createTestPlayersForPercentileCalculation()

	percentileMap, err := processor.calculateSomatogramPercentiles(players)
	if err != nil {
		t.Fatalf("Failed to calculate percentiles: %v", err)
	}

	// Test that percentiles were calculated for valid players
	expectedPlayers := countPlayersWithValidAgeGender(players)
	
	// Should have fewer players in percentile map due to minimum sample size filtering
	if len(percentileMap) > expectedPlayers {
		t.Errorf("Percentile map has more players (%d) than expected maximum (%d)", 
			len(percentileMap), expectedPlayers)
	}

	// Test percentile value ranges (0-100)
	for playerID, percentile := range percentileMap {
		if percentile < 0 || percentile > 100 {
			t.Errorf("Player %s has invalid percentile %.2f (should be 0-100)", 
				playerID, percentile)
		}
	}

	// Test that higher DWZ generally corresponds to higher percentiles within same group
	testPercentileCorrelationWithDWZ(t, players, percentileMap)
}

// Helper functions

func createTestPlayersForPercentileCalculation() []models.Player {
	players := []models.Player{}

	// Create multiple age-gender groups with sufficient sample sizes
	ages := []int{1990, 1985, 1980, 1975}
	genders := []string{"m", "w"}
	
	playerID := 1
	for _, year := range ages {
		for _, gender := range genders {
			// Create 60 players per group (above minimum sample size of 50)
			for i := 0; i < 60; i++ {
				dwz := 1200 + (i * 10) // DWZ from 1200 to 1790
				players = append(players, models.Player{
					ID:         fmt.Sprintf("P%03d", playerID),
					BirthYear:  &year,
					Gender:     gender,
					CurrentDWZ: dwz,
					ClubID:     "C0327",
					Club:       "Test Club",
				})
				playerID++
			}
		}
	}

	return players
}

func countPlayersWithValidAgeGender(players []models.Player) int {
	count := 0
	for _, player := range players {
		if player.BirthYear != nil && player.Gender != "" {
			count++
		}
	}
	return count
}

func testPercentileCorrelationWithDWZ(t *testing.T, players []models.Player, percentileMap map[string]float64) {
	// Group players by age-gender and test DWZ-percentile correlation within each group
	groups := make(map[string][]models.Player)
	
	for _, player := range players {
		if player.BirthYear != nil && player.Gender != "" {
			key := fmt.Sprintf("%d-%s", *player.BirthYear, player.Gender)
			groups[key] = append(groups[key], player)
		}
	}

	for groupKey, groupPlayers := range groups {
		if len(groupPlayers) < 2 {
			continue
		}

		// Find players in this group that have percentiles
		var playersWithPercentiles []struct {
			DWZ        int
			Percentile float64
		}

		for _, player := range groupPlayers {
			if percentile, exists := percentileMap[player.ID]; exists {
				playersWithPercentiles = append(playersWithPercentiles, struct {
					DWZ        int
					Percentile float64
				}{
					DWZ:        player.CurrentDWZ,
					Percentile: percentile,
				})
			}
		}

		if len(playersWithPercentiles) < 2 {
			continue
		}

		// Test that within each group, higher DWZ generally means higher percentile
		for i := 0; i < len(playersWithPercentiles)-1; i++ {
			for j := i + 1; j < len(playersWithPercentiles); j++ {
				p1 := playersWithPercentiles[i]
				p2 := playersWithPercentiles[j]

				// If DWZ differs significantly, percentiles should correlate
				dwzDiff := p2.DWZ - p1.DWZ
				percentileDiff := p2.Percentile - p1.Percentile

				if dwzDiff > 100 && percentileDiff < -10 {
					t.Errorf("In group %s: Player with DWZ %d has lower percentile (%.1f) than player with DWZ %d (%.1f)", 
						groupKey, p2.DWZ, p2.Percentile, p1.DWZ, p1.Percentile)
				}
			}
		}
	}
}
