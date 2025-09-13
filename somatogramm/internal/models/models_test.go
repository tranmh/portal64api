package models

import (
	"testing"
	"time"
)

func TestPlayerValidation(t *testing.T) {
	tests := []struct {
		name     string
		player   Player
		expected bool
	}{
		{
			name: "valid player",
			player: Player{
				ID:         "TEST-1",
				PKZ:        "12345",
				Name:       "Doe",
				Firstname:  "John",
				BirthYear:  intPtr(1990),
				Gender:     "m",
				CurrentDWZ: 1500,
			},
			expected: true,
		},
		{
			name: "player with nil birth year",
			player: Player{
				ID:         "TEST-2",
				CurrentDWZ: 1500,
				BirthYear:  nil,
			},
			expected: false,
		},
		{
			name: "player with zero DWZ",
			player: Player{
				ID:         "TEST-3",
				BirthYear:  intPtr(1990),
				CurrentDWZ: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.player.BirthYear != nil && tt.player.CurrentDWZ > 0
			if valid != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, valid)
			}
		})
	}
}

func TestSomatogrammMetadata(t *testing.T) {
	metadata := SomatogrammMetadata{
		GeneratedAt:    time.Now(),
		Gender:         "m",
		TotalPlayers:   1000,
		ValidAgeGroups: 20,
		MinSampleSize:  100,
	}

	if metadata.Gender != "m" {
		t.Errorf("expected gender 'm', got %s", metadata.Gender)
	}

	if metadata.TotalPlayers != 1000 {
		t.Errorf("expected 1000 players, got %d", metadata.TotalPlayers)
	}

	if metadata.ValidAgeGroups != 20 {
		t.Errorf("expected 20 age groups, got %d", metadata.ValidAgeGroups)
	}
}

func TestPercentileData(t *testing.T) {
	percentiles := map[int]int{
		0:   800,
		50:  1500,
		100: 2200,
	}

	data := PercentileData{
		Age:         25,
		SampleSize:  150,
		AvgDWZ:      1400.5,
		MedianDWZ:   1450,
		Percentiles: percentiles,
	}

	if data.Age != 25 {
		t.Errorf("expected age 25, got %d", data.Age)
	}

	if data.Percentiles[50] != 1500 {
		t.Errorf("expected 50th percentile 1500, got %d", data.Percentiles[50])
	}
}

func intPtr(i int) *int {
	return &i
}