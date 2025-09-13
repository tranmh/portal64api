package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"somatogramm/internal/api"
	"somatogramm/internal/export"
	"somatogramm/internal/models"
	"somatogramm/internal/processor"
)

// Integration test that tests the main workflow without calling main()
func TestSomatogrammWorkflow(t *testing.T) {
	// Create a mock server that handles both clubs and players endpoints
	birthYear := 1990
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var response models.APIResponse

		if strings.Contains(r.URL.Path, "/api/v1/clubs") && !strings.Contains(r.URL.Path, "/players") {
			// Mock clubs response - return one club
			response = models.APIResponse{
				Success: true,
				Data: api.ClubSearchResult{
					Data: []api.Club{
						{ID: "TEST001", Name: "Test Club"},
					},
				},
			}
		} else if strings.Contains(r.URL.Path, "/players") {
			// Mock players response for the club - return 150 players
			players := make([]models.Player, 150)
			for i := 0; i < 150; i++ {
				players[i] = models.Player{
					ID:         fmt.Sprintf("TEST001-%d", i+1),
					PKZ:        "PKZ123",
					Name:       "TestPlayer",
					Firstname:  "Test",
					BirthYear:  &birthYear,
					Gender:     "m",
					CurrentDWZ: 1000 + i*10,
				}
			}

			response = models.APIResponse{
				Success: true,
				Data: map[string]interface{}{
					"data": players,
					"meta": map[string]interface{}{
						"total":  len(players),
						"limit":  500,
						"offset": 0,
						"count":  len(players),
					},
				},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create temporary directory for output
	tempDir, err := os.MkdirTemp("", "somatogramm_integration")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test the workflow components

	// Create API client
	client := api.NewClient(server.URL, 30*time.Second, false)

	// Fetch players
	players, err := client.FetchAllPlayers()
	if err != nil {
		t.Fatalf("failed to fetch players: %v", err)
	}

	if len(players) != 150 {
		t.Errorf("expected 150 players, got %d", len(players))
	}

	// Process players
	proc := processor.NewProcessor(50, false) // Lower sample size for testing
	processedData, err := proc.ProcessPlayers(players)
	if err != nil {
		t.Fatalf("failed to process players: %v", err)
	}

	if len(processedData) == 0 {
		t.Error("expected processed data, got none")
	}

	// Export data
	exporter := export.NewExporter(tempDir, "csv", 7, false)
	if err := exporter.ExportData(processedData); err != nil {
		t.Fatalf("failed to export data: %v", err)
	}

	// Check that files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("expected output files to be created")
	}
}