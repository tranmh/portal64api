package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"somatogramm/internal/models"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Verbose    bool
}

func NewClient(baseURL string, timeout time.Duration, verbose bool) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Verbose: verbose,
	}
}

func (c *Client) log(message string) {
	if c.Verbose {
		fmt.Printf("[API] %s\n", message)
	}
}

type Club struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ClubSearchResult struct {
	Data []Club `json:"data"`
	Meta struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Count  int `json:"count"`
	} `json:"meta"`
}

func (c *Client) FetchAllPlayers() ([]models.Player, error) {
	c.log("Starting to fetch all players via clubs...")

	// First, fetch all clubs
	clubs, err := c.fetchAllClubs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clubs: %w", err)
	}

	c.log(fmt.Sprintf("Found %d clubs, fetching players for each club...", len(clubs)))

	var allPlayers []models.Player
	for _, club := range clubs {
		c.log(fmt.Sprintf("Fetching players for club %s (%s)", club.ID, club.Name))

		clubPlayers, err := c.fetchClubPlayers(club)
		if err != nil {
			c.log(fmt.Sprintf("Warning: failed to fetch players for club %s: %v", club.ID, err))
			continue
		}

		c.log(fmt.Sprintf("Found %d players for club %s", len(clubPlayers), club.ID))
		validPlayers := c.filterValidPlayers(clubPlayers)
		allPlayers = append(allPlayers, validPlayers...)
	}

	c.log(fmt.Sprintf("Fetched %d total valid players from %d clubs", len(allPlayers), len(clubs)))
	return allPlayers, nil
}

func (c *Client) fetchAllClubs() ([]Club, error) {
	const batchSize = 500
	var allClubs []Club
	offset := 0

	for {
		c.log(fmt.Sprintf("Fetching clubs: offset=%d, limit=%d", offset, batchSize))

		url := fmt.Sprintf("%s/api/v1/clubs?limit=%d&offset=%d&sort_by=name&sort_order=asc", c.BaseURL, batchSize, offset)

		resp, err := c.HTTPClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch clubs (offset %d): %w", offset, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var apiResp models.APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
		}

		if !apiResp.Success {
			return nil, fmt.Errorf("API error: %s", apiResp.Error)
		}

		var clubsResp ClubSearchResult
		dataBytes, err := json.Marshal(apiResp.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		if err := json.Unmarshal(dataBytes, &clubsResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal clubs response: %w", err)
		}

		c.log(fmt.Sprintf("Received %d clubs in this batch", len(clubsResp.Data)))
		allClubs = append(allClubs, clubsResp.Data...)

		if len(clubsResp.Data) < batchSize {
			break
		}

		offset += batchSize
	}

	return allClubs, nil
}

func (c *Client) fetchClubPlayers(club Club) ([]models.Player, error) {
	const batchSize = 500
	var allPlayers []models.Player
	offset := 0

	for {
		url := fmt.Sprintf("%s/api/v1/clubs/%s/players?limit=%d&offset=%d", c.BaseURL, club.ID, batchSize, offset)

		resp, err := c.HTTPClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch players for club %s (offset %d): %w", club.ID, offset, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var apiResp models.APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
		}

		if !apiResp.Success {
			return nil, fmt.Errorf("API error: %s", apiResp.Error)
		}

		// The API returns data in format: { "data": [...players...], "meta": {...} }
		// We need to extract the players from the "data" field
		dataMap, ok := apiResp.Data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected data format: expected object")
		}

		playersData, ok := dataMap["data"]
		if !ok {
			return nil, fmt.Errorf("no 'data' field found in response")
		}

		// Convert players data to Player structs
		dataBytes, err := json.Marshal(playersData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal players data: %w", err)
		}

		var players []models.Player
		if err := json.Unmarshal(dataBytes, &players); err != nil {
			return nil, fmt.Errorf("failed to unmarshal players: %w", err)
		}

		// Set club information for each player
		for i := range players {
			players[i].ClubID = club.ID
			players[i].ClubName = club.Name
		}

		allPlayers = append(allPlayers, players...)

		if len(players) < batchSize {
			break
		}

		offset += batchSize
	}

	return allPlayers, nil
}

func (c *Client) filterValidPlayers(players []models.Player) []models.Player {
	var validPlayers []models.Player

	currentYear := time.Now().Year()

	for _, player := range players {
		if player.CurrentDWZ <= 0 {
			continue
		}

		if player.BirthYear == nil {
			continue
		}

		age := currentYear - *player.BirthYear
		if age < 4 || age > 100 {
			continue
		}

		player.Gender = c.mapGenderToString(player.Gender)
		validPlayers = append(validPlayers, player)
	}

	return validPlayers
}

func (c *Client) mapGenderToString(gender string) string {
	switch strings.ToLower(gender) {
	case "1", "m", "male":
		return "m"
	case "0", "w", "female":
		return "w"
	case "2", "d", "divers":
		return "d"
	default:
		return "m"
	}
}

func (c *Client) GetPlayerAge(birthYear int) int {
	return time.Now().Year() - birthYear
}

func MapGeschlechtToGender(geschlecht int) string {
	switch geschlecht {
	case 1:
		return "m"
	case 0:
		return "w"
	case 2:
		return "d"
	default:
		return "m"
	}
}