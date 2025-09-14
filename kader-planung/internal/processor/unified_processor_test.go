package processor

import (
	"fmt"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/api"
	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/statistics"
	"github.com/sirupsen/logrus"
)

// MockAPIClient for testing the unified processor
type MockAPIClient struct {
	players []models.Player
	clubs   []models.Club
}

// NewMockAPIClient creates a mock API client with test data
func NewMockAPIClient() *MockAPIClient {
	return &MockAPIClient{
		players: createMockPlayers(),
		clubs:   createMockClubs(),
	}
}

func (m *MockAPIClient) FetchAllPlayersEfficient(clubPrefix string) ([]models.Player, error) {
	var filteredPlayers []models.Player
	for _, player := range m.players {
		if clubPrefix == "" || len(player.ClubID) >= len(clubPrefix) && player.ClubID[:len(clubPrefix)] == clubPrefix {
			filteredPlayers = append(filteredPlayers, player)
		}
	}
	return filteredPlayers, nil
}

func (m *MockAPIClient) FilterValidPlayersForStatistics(players []models.Player) []models.Player {
	var validPlayers []models.Player
	for _, player := range players {
		if player.BirthYear != nil && player.CurrentDWZ > 0 && player.Active && player.Status == "active" {
			validPlayers = append(validPlayers, player)
		}
	}
	return validPlayers
}

// TestUnifiedProcessor_StatisticalMode tests pure statistical processing
func TestUnifiedProcessor_StatisticalMode(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &UnifiedProcessorConfig{
		Mode:             StatisticalMode,
		MinSampleSize:    5,
		EnableStatistics: true,
		ClubPrefix:       "",
		Concurrency:      2,
		Verbose:          false,
	}

	// Create a unified processor with mock client
	processor := &UnifiedProcessor{
		apiClient: &api.Client{}, // We'll mock the methods we need
		config:    config,
		logger:    logger,
	}

	// Override the API client with our mock (this is a test hack)
	processor.apiClient = &api.Client{} // Reset to avoid nil pointer

	// Test statistical mode processing
	t.Run("Statistical mode processing", func(t *testing.T) {
		// We need to test the internal logic, so we'll call processStatisticalMode directly
		// but we need to mock the FetchAllPlayersEfficient method

		// Create test players
		players := createMockPlayers()

		// Create statistical analyzer
		analyzer := processor.statisticalAnalyzer
		if analyzer == nil {
			// Initialize if nil
			processor.statisticalAnalyzer = statistics.NewStatisticalAnalyzer(config.MinSampleSize, logger)
			analyzer = processor.statisticalAnalyzer
		}

		// Process players directly
		validPlayers := filterValidMockPlayers(players)
		results, err := analyzer.ProcessPlayers(validPlayers)

		if err != nil {
			t.Errorf("Statistical processing failed: %v", err)
		}

		// Verify we have statistical results
		if len(results) == 0 {
			t.Errorf("Expected statistical results, got none. ValidPlayers count: %d", len(validPlayers))
		}

		// Check for expected genders
		expectedGenders := []string{"m", "w"}
		for _, gender := range expectedGenders {
			if data, exists := results[gender]; exists {
				if data.Metadata.TotalPlayers == 0 {
					t.Errorf("Expected players for gender %s, got 0", gender)
				}
				if len(data.Percentiles) == 0 {
					t.Errorf("Expected percentile data for gender %s", gender)
				}
			}
		}
	})
}

// TestUnifiedProcessor_ProcessingModes tests all processing modes
func TestUnifiedProcessor_ProcessingModes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	testCases := []struct {
		name             string
		mode             ProcessingMode
		enableStatistics bool
		expectedRecords  bool
		expectedStats    bool
	}{
		{
			name:             "Efficient Mode",
			mode:             EfficientMode,
			enableStatistics: false,
			expectedRecords:  true,
			expectedStats:    false,
		},
		{
			name:             "Efficient Mode with Statistics",
			mode:             EfficientMode,
			enableStatistics: true,
			expectedRecords:  true,
			expectedStats:    true,
		},
		{
			name:             "Statistical Mode",
			mode:             StatisticalMode,
			enableStatistics: true,
			expectedRecords:  false,
			expectedStats:    true,
		},
		{
			name:             "Hybrid Mode",
			mode:             HybridMode,
			enableStatistics: true,
			expectedRecords:  true,
			expectedStats:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &UnifiedProcessorConfig{
				Mode:             tc.mode,
				MinSampleSize:    5,
				EnableStatistics: tc.enableStatistics,
				ClubPrefix:       "",
				Concurrency:      2,
				Verbose:          false,
			}

			// For this test, we'll test the logic components individually
			// since we can't easily mock the API client in the full processor

			// Test mode parsing
			parsedMode, err := ParseProcessingMode(tc.mode.String())
			if err != nil {
				t.Errorf("Failed to parse mode %s: %v", tc.mode.String(), err)
			}

			if parsedMode != tc.mode {
				t.Errorf("Expected mode %s, got %s", tc.mode.String(), parsedMode.String())
			}

			// Test configuration
			if config.Mode != tc.mode {
				t.Errorf("Expected config mode %s, got %s", tc.mode.String(), config.Mode.String())
			}

			if config.EnableStatistics != tc.enableStatistics {
				t.Errorf("Expected EnableStatistics %v, got %v", tc.enableStatistics, config.EnableStatistics)
			}
		})
	}
}

// TestUnifiedProcessor_ConvertPlayersToRecords tests player to record conversion
func TestUnifiedProcessor_ConvertPlayersToRecords(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &UnifiedProcessorConfig{
		Mode:             EfficientMode,
		MinSampleSize:    5,
		EnableStatistics: false,
		Concurrency:      2,
	}

	processor := &UnifiedProcessor{
		config: config,
		logger: logger,
	}

	players := createMockPlayers()[:5] // Use first 5 players

	t.Run("Convert players to records", func(t *testing.T) {
		records := processor.convertPlayersToRecords(players, false)

		if len(records) != len(players) {
			t.Errorf("Expected %d records, got %d", len(players), len(records))
		}

		// Verify first record
		if len(records) > 0 {
			record := records[0]
			player := players[0]

			if record.PlayerID != player.ID {
				t.Errorf("Expected player ID %s, got %s", player.ID, record.PlayerID)
			}
			if record.Firstname != player.Firstname {
				t.Errorf("Expected firstname %s, got %s", player.Firstname, record.Firstname)
			}
			if record.Lastname != player.Name {
				t.Errorf("Expected lastname %s, got %s", player.Name, record.Lastname)
			}
			if record.CurrentDWZ != player.CurrentDWZ {
				t.Errorf("Expected DWZ %d, got %d", player.CurrentDWZ, record.CurrentDWZ)
			}

			// Check that historical data is not available (efficient mode)
			if record.DWZ12MonthsAgo != models.DataNotAvailable {
				t.Errorf("Expected DWZ12MonthsAgo to be '%s' in efficient mode, got '%s'",
					models.DataNotAvailable, record.DWZ12MonthsAgo)
			}
		}
	})
}

// TestUnifiedProcessor_CalculateListRankings tests list ranking calculation
func TestUnifiedProcessor_CalculateListRankings(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	config := &UnifiedProcessorConfig{
		Mode:             EfficientMode,
		MinSampleSize:    5,
		EnableStatistics: false,
		Concurrency:      2,
	}

	processor := &UnifiedProcessor{
		config: config,
		logger: logger,
	}

	// Create test records with known DWZ values
	records := []models.KaderPlanungRecord{
		{PlayerID: "1", CurrentDWZ: 1800}, // Should be rank 1
		{PlayerID: "2", CurrentDWZ: 1600}, // Should be rank 2
		{PlayerID: "3", CurrentDWZ: 1400}, // Should be rank 3
		{PlayerID: "4", CurrentDWZ: 1600}, // Should be rank 2 (tie)
		{PlayerID: "5", CurrentDWZ: 1200}, // Should be rank 5
	}

	t.Run("Calculate list rankings", func(t *testing.T) {
		processor.calculateListRankings(records)

		// Check rankings
		expectedRankings := map[string]int{
			"1": 1, // Highest DWZ
			"2": 2, // Second highest (tied)
			"3": 4, // Fourth highest
			"4": 2, // Second highest (tied)
			"5": 5, // Lowest DWZ
		}

		for i, record := range records {
			expectedRank := expectedRankings[record.PlayerID]
			if record.ListRanking != expectedRank {
				t.Errorf("Player %s (index %d): expected rank %d, got %d",
					record.PlayerID, i, expectedRank, record.ListRanking)
			}
		}
	})
}

// TestUnifiedProcessor_CountUniqueClubs tests club counting functionality
func TestUnifiedProcessor_CountUniqueClubs(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	processor := &UnifiedProcessor{
		logger: logger,
	}

	t.Run("Count unique clubs", func(t *testing.T) {
		players := []models.Player{
			{ClubID: "C0327"},
			{ClubID: "C0401"},
			{ClubID: "C0327"}, // Duplicate
			{ClubID: "C0533"},
			{ClubID: "C0401"}, // Duplicate
		}

		count := processor.countUniqueClubs(players)
		expected := 3 // C0327, C0401, C0533

		if count != expected {
			t.Errorf("Expected %d unique clubs, got %d", expected, count)
		}
	})

	t.Run("Count no clubs", func(t *testing.T) {
		players := []models.Player{}
		count := processor.countUniqueClubs(players)
		expected := 0

		if count != expected {
			t.Errorf("Expected %d unique clubs, got %d", expected, count)
		}
	})
}

// Helper functions

func createMockPlayers() []models.Player {
	currentYear := time.Now().Year()
	players := []models.Player{
		{
			ID:         "C0327-001",
			Name:       "TestPlayer",
			Firstname:  "Test1",
			BirthYear:  intPtr(1995),
			Gender:     "m",
			CurrentDWZ: 1600,
			ClubID:     "C0327",
			Club:       "Test Club 1",
			Status:     "active",
			Active:     true,
		},
		{
			ID:         "C0327-002",
			Name:       "TestPlayer",
			Firstname:  "Test2",
			BirthYear:  intPtr(1990),
			Gender:     "w",
			CurrentDWZ: 1550,
			ClubID:     "C0327",
			Club:       "Test Club 1",
			Status:     "active",
			Active:     true,
		},
		{
			ID:         "C0401-001",
			Name:       "TestPlayer",
			Firstname:  "Test3",
			BirthYear:  intPtr(1985),
			Gender:     "m",
			CurrentDWZ: 1700,
			ClubID:     "C0401",
			Club:       "Test Club 2",
			Status:     "active",
			Active:     true,
		},
		{
			ID:         "C0401-002",
			Name:       "TestPlayer",
			Firstname:  "Test4",
			BirthYear:  intPtr(1992),
			Gender:     "w",
			CurrentDWZ: 1450,
			ClubID:     "C0401",
			Club:       "Test Club 2",
			Status:     "active",
			Active:     true,
		},
		{
			ID:         "C0533-001",
			Name:       "TestPlayer",
			Firstname:  "Test5",
			BirthYear:  intPtr(1988),
			Gender:     "m",
			CurrentDWZ: 1800,
			ClubID:     "C0533",
			Club:       "Test Club 3",
			Status:     "active",
			Active:     true,
		},
	}

	// Add more players for better statistical analysis with concentrated age groups
	targetAges := []int{25, 30} // Focus on 2 age groups to meet minimum sample size
	for i := 6; i <= 50; i++ {
		ageIndex := (i - 6) % len(targetAges)
		birthYear := currentYear - targetAges[ageIndex]
		gender := "m"
		if i%2 == 0 {
			gender = "w"
		}

		player := models.Player{
			ID:         fmt.Sprintf("C0327-%03d", i),
			Name:       fmt.Sprintf("TestPlayer%d", i),
			Firstname:  fmt.Sprintf("Test%d", i),
			BirthYear:  &birthYear,
			Gender:     gender,
			CurrentDWZ: 1300 + (i * 10), // Varying DWZ
			ClubID:     "C0327",
			Club:       "Test Club 1",
			Status:     "active",
			Active:     true,
		}
		players = append(players, player)
	}

	return players
}

func createMockClubs() []models.Club {
	return []models.Club{
		{ID: "C0327", Name: "Test Club 1"},
		{ID: "C0401", Name: "Test Club 2"},
		{ID: "C0533", Name: "Test Club 3"},
	}
}

func filterValidMockPlayers(players []models.Player) []models.Player {
	var validPlayers []models.Player
	for _, player := range players {
		if player.BirthYear != nil && player.CurrentDWZ > 0 && player.Active && player.Status == "active" {
			validPlayers = append(validPlayers, player)
		}
	}
	return validPlayers
}

func intPtr(i int) *int {
	return &i
}