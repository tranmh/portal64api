package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"somatogramm/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080", 30*time.Second, true)

	if client.BaseURL != "http://localhost:8080" {
		t.Errorf("expected BaseURL 'http://localhost:8080', got %s", client.BaseURL)
	}

	if !client.Verbose {
		t.Error("expected verbose to be true")
	}

	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", client.HTTPClient.Timeout)
	}
}

func TestNewClientTrimsSlash(t *testing.T) {
	client := NewClient("http://localhost:8080/", 30*time.Second, false)

	if client.BaseURL != "http://localhost:8080" {
		t.Errorf("expected BaseURL without trailing slash, got %s", client.BaseURL)
	}
}

func TestFilterValidPlayers(t *testing.T) {
	client := NewClient("http://localhost:8080", 30*time.Second, false)

	birthYear1990 := 1990
	birthYear2022 := 2022  // Too young (age < 4)
	birthYear1900 := 1900  // Too old (age > 100)

	players := []models.Player{
		{
			ID:         "TEST-1",
			CurrentDWZ: 1500,
			BirthYear:  &birthYear1990,
			Gender:     "m",
		},
		{
			ID:         "TEST-2",
			CurrentDWZ: 0, // Invalid - zero DWZ
			BirthYear:  &birthYear1990,
			Gender:     "m",
		},
		{
			ID:         "TEST-3",
			CurrentDWZ: 1200,
			BirthYear:  nil, // Invalid - nil birth year
			Gender:     "w",
		},
		{
			ID:         "TEST-4",
			CurrentDWZ: 1800,
			BirthYear:  &birthYear2022, // Invalid - too young (age < 4)
			Gender:     "m",
		},
		{
			ID:         "TEST-5",
			CurrentDWZ: 1600,
			BirthYear:  &birthYear1900, // Invalid - too old (age > 100)
			Gender:     "w",
		},
	}

	validPlayers := client.filterValidPlayers(players)

	// Only the first player should be valid
	if len(validPlayers) != 1 {
		t.Errorf("expected 1 valid player, got %d", len(validPlayers))
	}

	if len(validPlayers) > 0 && validPlayers[0].ID != "TEST-1" {
		t.Errorf("expected first valid player to have ID TEST-1, got %s", validPlayers[0].ID)
	}
}

func TestMapGenderToString(t *testing.T) {
	client := NewClient("http://localhost:8080", 30*time.Second, false)

	tests := []struct {
		input    string
		expected string
	}{
		{"1", "m"},
		{"m", "m"},
		{"male", "m"},
		{"0", "w"},
		{"w", "w"},
		{"female", "w"},
		{"2", "d"},
		{"d", "d"},
		{"divers", "d"},
		{"unknown", "m"}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := client.mapGenderToString(tt.input)
			if result != tt.expected {
				t.Errorf("mapGenderToString(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPlayerAge(t *testing.T) {
	client := NewClient("http://localhost:8080", 30*time.Second, false)

	currentYear := time.Now().Year()
	birthYear := 1990
	expectedAge := currentYear - birthYear

	age := client.GetPlayerAge(birthYear)
	if age != expectedAge {
		t.Errorf("expected age %d, got %d", expectedAge, age)
	}
}

func TestMapGeschlechtToGender(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "m"},
		{0, "w"},
		{2, "d"},
		{99, "m"}, // default case
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := MapGeschlechtToGender(tt.input)
			if result != tt.expected {
				t.Errorf("MapGeschlechtToGender(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFetchAllPlayersHTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, false)
	players, err := client.FetchAllPlayers()

	if err == nil {
		t.Error("expected an error, got nil")
	}

	if players != nil {
		t.Errorf("expected nil players, got %v", players)
	}
}

func TestFetchAllPlayersSuccess(t *testing.T) {
	birthYear := 1990

	// Create mock responses for the new approach (clubs first, then players)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var mockResponse models.APIResponse

		if strings.Contains(r.URL.Path, "/api/v1/clubs") && !strings.Contains(r.URL.Path, "/players") {
			// Mock clubs response
			mockResponse = models.APIResponse{
				Success: true,
				Data: ClubSearchResult{
					Data: []Club{
						{ID: "CLUB001", Name: "Test Club"},
					},
				},
			}
		} else if strings.Contains(r.URL.Path, "/players") {
			// Mock players response for a club (real API format)
			mockResponse = models.APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"data": []models.Player{
						{
							ID:         "CLUB001-1",
							PKZ:        "12345",
							Name:       "Doe",
							Firstname:  "John",
							BirthYear:  &birthYear,
							Gender:     "m",
							CurrentDWZ: 1500,
						},
					},
					"meta": map[string]interface{}{
						"total":  1,
						"limit":  500,
						"offset": 0,
						"count":  1,
					},
				},
			}
		}

		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, false)
	players, err := client.FetchAllPlayers()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(players) != 1 {
		t.Errorf("expected 1 player, got %d", len(players))
	}

	if len(players) > 0 && players[0].CurrentDWZ != 1500 {
		t.Errorf("expected DWZ 1500, got %d", players[0].CurrentDWZ)
	}
}

func TestFetchAllPlayersAPIError(t *testing.T) {
	// Create mock response with API error (for clubs request)
	mockResponse := models.APIResponse{
		Success: false,
		Error:   "Database connection failed",
		Data:    nil,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, false)
	players, err := client.FetchAllPlayers()

	if err == nil {
		t.Error("expected error for API failure, got nil")
	}

	if players != nil {
		t.Errorf("expected nil players on API error, got %v", players)
	}

	if err != nil && !strings.Contains(err.Error(), "Database connection failed") {
		t.Errorf("expected error message to contain 'Database connection failed', got: %v", err)
	}
}

func TestFetchAllPlayersInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json}"))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, false)
	players, err := client.FetchAllPlayers()

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	if players != nil {
		t.Errorf("expected nil players on JSON error, got %v", players)
	}
}

func TestFetchAllPlayersTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewClient(server.URL, 10*time.Millisecond, false)
	players, err := client.FetchAllPlayers()

	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	if players != nil {
		t.Errorf("expected nil players on timeout, got %v", players)
	}
}

func TestVerboseLogging(t *testing.T) {
	client := NewClient("http://localhost:8080", 30*time.Second, true)

	// Test that verbose logging doesn't panic
	// The actual log output would go to stdout and is difficult to capture
	client.log("test message")

	if !client.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestNonVerboseLogging(t *testing.T) {
	client := NewClient("http://localhost:8080", 30*time.Second, false)

	// Test that non-verbose logging doesn't panic
	client.log("test message")

	if client.Verbose {
		t.Error("expected Verbose to be false")
	}
}

func TestFetchAllPlayersWithMultipleClubs(t *testing.T) {
	birthYear := 1990

	// Create mock response for multiple clubs test
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var mockResponse models.APIResponse

		if strings.Contains(r.URL.Path, "/api/v1/clubs") && !strings.Contains(r.URL.Path, "/players") {
			// Mock clubs response with 2 clubs
			mockResponse = models.APIResponse{
				Success: true,
				Data: ClubSearchResult{
					Data: []Club{
						{ID: "CLUB001", Name: "Test Club 1"},
						{ID: "CLUB002", Name: "Test Club 2"},
					},
				},
			}
		} else if strings.Contains(r.URL.Path, "CLUB001/players") {
			// Mock players response for club 1 (real API format)
			mockResponse = models.APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"data": []models.Player{
						{
							ID:         "CLUB001-1",
							PKZ:        "PKZ001",
							Name:       "Player1",
							Firstname:  "Test",
							BirthYear:  &birthYear,
							Gender:     "m",
							CurrentDWZ: 1500,
						},
					},
					"meta": map[string]interface{}{
						"total":  1,
						"limit":  500,
						"offset": 0,
						"count":  1,
					},
				},
			}
		} else if strings.Contains(r.URL.Path, "CLUB002/players") {
			// Mock players response for club 2 (real API format)
			mockResponse = models.APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"data": []models.Player{
						{
							ID:         "CLUB002-1",
							PKZ:        "PKZ002",
							Name:       "Player2",
							Firstname:  "Test",
							BirthYear:  &birthYear,
							Gender:     "w",
							CurrentDWZ: 1200,
						},
					},
					"meta": map[string]interface{}{
						"total":  1,
						"limit":  500,
						"offset": 0,
						"count":  1,
					},
				},
			}
		}

		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, true) // verbose = true
	players, err := client.FetchAllPlayers()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should have fetched players from both clubs
	if len(players) != 2 {
		t.Errorf("expected 2 players, got %d", len(players))
	}
}

func TestFetchAllPlayersEmptyResponse(t *testing.T) {
	// Create mock response with empty clubs array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var mockResponse models.APIResponse
		if strings.Contains(r.URL.Path, "/api/v1/clubs") && !strings.Contains(r.URL.Path, "/players") {
			// Empty clubs response
			mockResponse = models.APIResponse{
				Success: true,
				Data: ClubSearchResult{
					Data: []Club{},
				},
			}
		}

		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second, false)
	players, err := client.FetchAllPlayers()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(players) != 0 {
		t.Errorf("expected 0 players, got %d", len(players))
	}
}