package models

import (
	"time"
)

type Player struct {
	ID        string    `json:"id"`
	PKZ       string    `json:"pkz"`
	Name      string    `json:"name"`
	Firstname string    `json:"firstname"`
	BirthYear *int      `json:"birth_year"`
	Gender    string    `json:"gender"`
	CurrentDWZ int      `json:"current_dwz"`
	ClubID    string    `json:"club_id,omitempty"`    // Added for debug output
	ClubName  string    `json:"club_name,omitempty"`  // Added for debug output
}

type DebugPlayerRecord struct {
	PlayerID       string `json:"player_id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ClubID         string `json:"club_id"`
	ClubName       string `json:"club_name"`
	BirthYear      int    `json:"birth_year"`
	CalculatedAge  int    `json:"calculated_age"`
	CurrentDWZ     int    `json:"current_dwz"`
	Gender         string `json:"gender"`
}

type AgeGenderGroup struct {
	Age        int     `json:"age"`
	Gender     string  `json:"gender"`
	Players    []Player `json:"players"`
	SampleSize int     `json:"sample_size"`
	AvgDWZ     float64 `json:"avg_dwz"`
	MedianDWZ  int     `json:"median_dwz"`
}

type PercentileData struct {
	Age         int            `json:"age"`
	SampleSize  int            `json:"sample_size"`
	AvgDWZ      float64        `json:"avg_dwz"`
	MedianDWZ   int            `json:"median_dwz"`
	Percentiles map[int]int    `json:"percentiles"`
}

type SomatogrammMetadata struct {
	GeneratedAt      time.Time `json:"generated_at"`
	Gender           string    `json:"gender"`
	TotalPlayers     int       `json:"total_players"`
	ValidAgeGroups   int       `json:"valid_age_groups"`
	MinSampleSize    int       `json:"min_sample_size"`
}

type SomatogrammData struct {
	Metadata    SomatogrammMetadata       `json:"metadata"`
	Percentiles map[string]PercentileData `json:"percentiles"`
}

type Config struct {
	OutputFormat   string
	OutputDir      string
	Concurrency    int
	APIBaseURL     string
	Timeout        int
	Verbose        bool
	MinSampleSize  int
	// Debug mode fields
	DebugAge       int
	DebugGender    string
	DebugOutput    string
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

type PlayersResponse struct {
	Players []Player `json:"players"`
	Meta    struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Count  int `json:"count"`
	} `json:"meta"`
}