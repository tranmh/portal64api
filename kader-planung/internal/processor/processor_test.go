package processor

import (
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/models"
)

func TestAnalyzeHistoricalData(t *testing.T) {
	processor := &Processor{}

	// Test case 1: No history data
	t.Run("No history data", func(t *testing.T) {
		analysis := processor.AnalyzeHistoricalData(nil)

		if analysis.HasHistoricalData {
			t.Error("Expected HasHistoricalData to be false")
		}
		if analysis.DWZ12MonthsAgo != models.DataNotAvailable {
			t.Errorf("Expected DWZ12MonthsAgo to be %s, got %s", models.DataNotAvailable, analysis.DWZ12MonthsAgo)
		}
	})

	// Test case 2: Empty history
	t.Run("Empty history", func(t *testing.T) {
		history := &models.RatingHistory{
			PlayerID: "C0327-297",
			Points:   []models.RatingPoint{},
		}

		analysis := processor.AnalyzeHistoricalData(history)

		if analysis.HasHistoricalData {
			t.Error("Expected HasHistoricalData to be false for empty history")
		}
	})

	// Test case 3: History with data points
	t.Run("History with data points", func(t *testing.T) {
		now := time.Now()
		twoYearsAgo := now.AddDate(-2, 0, 0)
		oneYearAgo := now.AddDate(-1, 0, 0)
		sixMonthsAgo := now.AddDate(0, -6, 0)

		history := &models.RatingHistory{
			PlayerID: "C0327-297",
			Points: []models.RatingPoint{
				{
					Date:       twoYearsAgo,
					DWZ:        1500,
					Games:      5,
					Points:     3.5,
					Tournament: "Test Tournament 1",
				},
				{
					Date:       oneYearAgo,
					DWZ:        1520,
					Games:      3,
					Points:     2.0,
					Tournament: "Test Tournament 2",
				},
				{
					Date:       sixMonthsAgo,
					DWZ:        1540,
					Games:      4,
					Points:     3.0,
					Tournament: "Test Tournament 3",
				},
			},
		}

		analysis := processor.AnalyzeHistoricalData(history)

		if !analysis.HasHistoricalData {
			t.Error("Expected HasHistoricalData to be true")
		}

		// Should find DWZ from 1 year ago (closest to 12 months ago target)
		if analysis.DWZ12MonthsAgo != "1520" {
			t.Errorf("Expected DWZ12MonthsAgo to be 1520, got %s", analysis.DWZ12MonthsAgo)
		}

		// Should count games from last 12 months (only the 6-month-ago entry)
		if analysis.GamesLast12Months != 4 {
			t.Errorf("Expected GamesLast12Months to be 4, got %d", analysis.GamesLast12Months)
		}

		// Success rate should be (2 + 0.5*2) / 4 * 100 = 75%
		expectedRate := 75.0
		if analysis.SuccessRateLast12Months != expectedRate {
			t.Errorf("Expected success rate to be %.1f, got %.1f", expectedRate, analysis.SuccessRateLast12Months)
		}
	})
}

func TestCreateKaderPlanungRecord(t *testing.T) {
	processor := &Processor{}

	club := models.Club{
		ID:   "C0327",
		Name: "Test Chess Club",
	}

	birthyear := 1990
	player := models.Player{
		ID:         "C0327-297",
		Firstname:  "John",
		Name:       "Doe",
		BirthYear:  &birthyear,
		CurrentDWZ: 1600,
	}

	analysis := &models.HistoricalAnalysis{
		DWZ12MonthsAgo:          "1550",
		GamesLast12Months:       8,
		SuccessRateLast12Months: 62.5,
		HasHistoricalData:       true,
	}

	record := processor.createKaderPlanungRecord(club, player, analysis)

	// Test basic fields
	if record.ClubName != club.Name {
		t.Errorf("Expected club name %s, got %s", club.Name, record.ClubName)
	}
	if record.ClubID != club.ID {
		t.Errorf("Expected club ID %s, got %s", club.ID, record.ClubID)
	}
	if record.PlayerID != player.ID {
		t.Errorf("Expected player ID %s, got %s", player.ID, record.PlayerID)
	}
	if record.Firstname != player.Firstname {
		t.Errorf("Expected firstname %s, got %s", player.Firstname, record.Firstname)
	}
	if record.Lastname != player.Name {
		t.Errorf("Expected lastname %s, got %s", player.Name, record.Lastname)
	}
	if record.Birthyear != *player.BirthYear {
		t.Errorf("Expected birthyear %d, got %d", *player.BirthYear, record.Birthyear)
	}
	if record.CurrentDWZ != player.CurrentDWZ {
		t.Errorf("Expected current DWZ %d, got %d", player.CurrentDWZ, record.CurrentDWZ)
	}

	// Test analysis fields
	if record.DWZ12MonthsAgo != "1550" {
		t.Errorf("Expected DWZ 12 months ago to be 1550, got %s", record.DWZ12MonthsAgo)
	}
	if record.GamesLast12Months != "8" {
		t.Errorf("Expected games last 12 months to be 8, got %s", record.GamesLast12Months)
	}
	if record.SuccessRateLast12Months != "62,5" {
		t.Errorf("Expected success rate to be 62.5, got %s", record.SuccessRateLast12Months)
	}
}

func TestCreateKaderPlanungRecordNoData(t *testing.T) {
	processor := &Processor{}

	club := models.Club{
		ID:   "C0327",
		Name: "Test Chess Club",
	}

	birthyear := 1990
	player := models.Player{
		ID:         "C0327-297",
		Firstname:  "John",
		Name:       "Doe",
		BirthYear:  &birthyear,
		CurrentDWZ: 1600,
	}

	// Test with no analysis (nil)
	record := processor.createKaderPlanungRecord(club, player, nil)

	if record.DWZ12MonthsAgo != models.DataNotAvailable {
		t.Errorf("Expected DWZ 12 months ago to be %s, got %s", models.DataNotAvailable, record.DWZ12MonthsAgo)
	}
	if record.GamesLast12Months != models.DataNotAvailable {
		t.Errorf("Expected games last 12 months to be %s, got %s", models.DataNotAvailable, record.GamesLast12Months)
	}
	if record.SuccessRateLast12Months != models.DataNotAvailable {
		t.Errorf("Expected success rate to be %s, got %s", models.DataNotAvailable, record.SuccessRateLast12Months)
	}

	// Test with analysis but no historical data
	analysis := &models.HistoricalAnalysis{
		HasHistoricalData: false,
	}

	record = processor.createKaderPlanungRecord(club, player, analysis)

	if record.DWZ12MonthsAgo != models.DataNotAvailable {
		t.Errorf("Expected DWZ 12 months ago to be %s, got %s", models.DataNotAvailable, record.DWZ12MonthsAgo)
	}
	if record.GamesLast12Months != models.DataNotAvailable {
		t.Errorf("Expected games last 12 months to be %s, got %s", models.DataNotAvailable, record.GamesLast12Months)
	}
	if record.SuccessRateLast12Months != models.DataNotAvailable {
		t.Errorf("Expected success rate to be %s, got %s", models.DataNotAvailable, record.SuccessRateLast12Months)
	}
}
