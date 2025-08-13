package models

import (
	"time"
)

// Club represents a chess club
type Club struct {
	ID           string     `json:"id"`              // Format: C0101
	Name         string     `json:"name"`
	ShortName    string     `json:"short_name"`
	Region       string     `json:"region"`
	District     string     `json:"district"`
	FoundingDate *time.Time `json:"founding_date"`
	MemberCount  int        `json:"member_count"`
	AverageDWZ   float64    `json:"average_dwz"`
	Status       string     `json:"status"`
}

// Player represents a chess player
type Player struct {
	ID         string    `json:"id"`          // Format: C0101-123
	Name       string    `json:"name"`        // Last name
	Firstname  string    `json:"firstname"`   
	Club       string    `json:"club"`        // Club name
	ClubID     string    `json:"club_id"`     // Format: C0101
	BirthYear  *int      `json:"birth_year"`  // GDPR compliant: only birth year
	Gender     string    `json:"gender"`
	Nation     string    `json:"nation"`
	FideID     uint      `json:"fide_id"`
	CurrentDWZ int       `json:"current_dwz"`
	DWZIndex   int       `json:"dwz_index"`
	Status     string    `json:"status"`
	Active     bool      `json:"-"`           // Derived field for backward compatibility
	LastUpdate time.Time `json:"-"`           // Not available in API response
}
// TournamentResult represents a single tournament result from the API
type TournamentResult struct {
	ID             int        `json:"id"`
	TournamentID   string     `json:"tournament_id"`
	TournamentName string     `json:"tournament_name"` // NEW: Tournament name from optimized API
	TournamentDate *time.Time `json:"tournament_date"` // NEW: Pre-computed tournament date
	ECoefficient   int        `json:"e_coefficient"`
	We             float64    `json:"we"`
	Achievement    int        `json:"achievement"`
	Level          int        `json:"level"`
	Games          int        `json:"games"`
	UnratedGames   int        `json:"unrated_games"`
	Points         float64    `json:"points"`
	DWZOld         int        `json:"dwz_old"`
	DWZOldIndex    int        `json:"dwz_old_index"`
	DWZNew         int        `json:"dwz_new"`
	DWZNewIndex    int        `json:"dwz_new_index"`
}

// RatingHistoryResponse represents the API response for rating history
type RatingHistoryResponse struct {
	Success bool               `json:"success"`
	Data    []TournamentResult `json:"data"`
}

// RatingPoint represents a single point in a player's rating history (legacy format)
type RatingPoint struct {
	Date       time.Time `json:"date"`
	DWZ        int       `json:"dwz"`
	Games      int       `json:"games"`
	Points     float64   `json:"points"`     // Points scored in tournament
	Tournament string    `json:"tournament"`
}

// RatingHistory represents a player's complete rating history (legacy format)
type RatingHistory struct {
	PlayerID string        `json:"player_id"`
	Points   []RatingPoint `json:"points"`
}

// KaderPlanungRecord represents a single row in the final export
type KaderPlanungRecord struct {
	ClubIDPrefix1              string `json:"club_id_prefix1" csv:"club_id_prefix1"`
	ClubIDPrefix2              string `json:"club_id_prefix2" csv:"club_id_prefix2"`
	ClubIDPrefix3              string `json:"club_id_prefix3" csv:"club_id_prefix3"`
	ClubName                   string `json:"club_name" csv:"club_name"`
	ClubID                     string `json:"club_id" csv:"club_id"`
	PlayerID                   string `json:"player_id" csv:"player_id"`
	Lastname                   string `json:"lastname" csv:"lastname"`
	Firstname                  string `json:"firstname" csv:"firstname"`
	Birthyear                  int    `json:"birthyear" csv:"birthyear"`
	CurrentDWZ                 int    `json:"current_dwz" csv:"current_dwz"`
	DWZ12MonthsAgo             string `json:"dwz_12_months_ago" csv:"dwz_12_months_ago"`
	GamesLast12Months          string `json:"games_last_12_months" csv:"games_last_12_months"`
	SuccessRateLast12Months    string `json:"success_rate_last_12_months" csv:"success_rate_last_12_months"`
}
// APIResponse represents the wrapper response from Portal64 API
type APIResponse struct {
	Success bool                `json:"success"`
	Data    ClubSearchResult    `json:"data"`
}

// ClubSearchResult represents the result of searching for clubs
type ClubSearchResult struct {
	Data []Club `json:"data"`
	Meta Meta   `json:"meta"`
}

// Meta represents response metadata
type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

// PlayerAPIResponse represents the wrapper response from Portal64 API for players
type PlayerAPIResponse struct {
	Success bool                `json:"success"`
	Data    PlayerSearchResult  `json:"data"`
}

// PlayerSearchResult represents the result of searching for players
type PlayerSearchResult struct {
	Data []Player `json:"data"`
	Meta Meta     `json:"meta"`
}

// HistoricalAnalysis contains calculated historical statistics
type HistoricalAnalysis struct {
	DWZ12MonthsAgo        string
	GamesLast12Months     int
	SuccessRateLast12Months float64
	HasHistoricalData     bool
}

// ProcessingStats tracks processing statistics
type ProcessingStats struct {
	TotalClubs          int
	ProcessedClubs      int
	TotalPlayers        int
	ProcessedPlayers    int
	PlayersWithHistory  int
	PlayersWithoutHistory int
	Errors              int
	StartTime           time.Time
	EstimatedEndTime    time.Time
}

// APIError represents an API error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

// Tournament represents a tournament with date information
type Tournament struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	Type         string     `json:"type"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	FinishedOn   *time.Time `json:"finished_on"`
	ComputedOn   *time.Time `json:"computed_on"`
	Status       string     `json:"status"`
}

// Constants for data availability
const (
	DataNotAvailable = "DATA_NOT_AVAILABLE"
)

// CalculateClubIDPrefixes calculates the three club_id prefixes from a club_id
// For example: "C0327" -> {"C", "C0", "C03"}
func CalculateClubIDPrefixes(clubID string) (string, string, string) {
	if clubID == "" {
		return "", "", ""
	}
	
	// Get prefixes of different lengths
	prefix1 := ""
	prefix2 := ""
	prefix3 := ""
	
	if len(clubID) >= 1 {
		prefix1 = clubID[:1]
	}
	if len(clubID) >= 2 {
		prefix2 = clubID[:2]
	}
	if len(clubID) >= 3 {
		prefix3 = clubID[:3]
	}
	
	return prefix1, prefix2, prefix3
}