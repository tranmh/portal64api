package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
)

// Client represents the Portal64 API client
type Client struct {
	baseURL     string
	httpClient  *http.Client
	logger      *logrus.Logger
	concurrency int // Number of concurrent workers for efficient bulk operations
}

// NewClient creates a new API client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     60 * time.Second,
			},
		},
		logger:      logrus.StandardLogger(),
		concurrency: 8, // Default concurrency level
	}
}

// NewClientWithConcurrency creates a new API client with specified concurrency
func NewClientWithConcurrency(baseURL string, timeout time.Duration, concurrency int) *Client {
	client := NewClient(baseURL, timeout)
	client.concurrency = concurrency
	return client
}

// SearchClubs retrieves all clubs, optionally filtered by prefix
func (c *Client) SearchClubs(prefix string, limit, offset int) (*models.ClubSearchResult, error) {
	params := url.Values{}
	if prefix != "" {
		params.Set("query", prefix)
	}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	params.Set("sort_by", "name")
	params.Set("sort_order", "asc")

	endpoint := "/api/v1/clubs?" + params.Encode()
	
	var response models.APIResponse
	if err := c.makeRequest("GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to search clubs: %w", err)
	}

	return &response.Data, nil
}
// GetClubProfile retrieves detailed information about a specific club
func (c *Client) GetClubProfile(clubID string) (*models.Club, error) {
	endpoint := fmt.Sprintf("/api/v1/clubs/%s", clubID)
	
	var club models.Club
	if err := c.makeRequest("GET", endpoint, nil, &club); err != nil {
		return nil, fmt.Errorf("failed to get club profile for %s: %w", clubID, err)
	}

	return &club, nil
}

// GetClubPlayers retrieves all players belonging to a specific club
func (c *Client) GetClubPlayers(clubID string, active bool, limit, offset int) (*models.PlayerSearchResult, error) {
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	if active {
		params.Set("active", "true")
	}

	endpoint := fmt.Sprintf("/api/v1/clubs/%s/players?%s", clubID, params.Encode())
	
	var response models.PlayerAPIResponse
	if err := c.makeRequest("GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get players for club %s: %w", clubID, err)
	}

	return &response.Data, nil
}
// GetTournamentDetails retrieves detailed information about a specific tournament
func (c *Client) GetTournamentDetails(tournamentID string) (*models.Tournament, error) {
	endpoint := fmt.Sprintf("/api/v1/tournaments/%s", tournamentID)
	
	var tournament models.Tournament
	if err := c.makeRequest("GET", endpoint, nil, &tournament); err != nil {
		c.logger.Debugf("Failed to get tournament details for %s: %v", tournamentID, err)
		return nil, fmt.Errorf("failed to get tournament details for %s: %w", tournamentID, err)
	}

	return &tournament, nil
}

// GetPlayerProfile retrieves detailed information about a specific player
func (c *Client) GetPlayerProfile(playerID string) (*models.Player, error) {
	endpoint := fmt.Sprintf("/api/v1/players/%s", playerID)
	
	var player models.Player
	if err := c.makeRequest("GET", endpoint, nil, &player); err != nil {
		return nil, fmt.Errorf("failed to get player profile for %s: %w", playerID, err)
	}

	return &player, nil
}

// GetPlayerRatingHistory retrieves the complete rating history for a player
func (c *Client) GetPlayerRatingHistory(playerID string) (*models.RatingHistory, error) {
	endpoint := fmt.Sprintf("/api/v1/players/%s/rating-history", playerID)
	
	var response models.RatingHistoryResponse
	if err := c.makeRequest("GET", endpoint, nil, &response); err != nil {
		c.logger.Debugf("API request failed for player %s: %v", playerID, err)
		return nil, fmt.Errorf("failed to get rating history for %s: %w", playerID, err)
	}

	if !response.Success {
		c.logger.Debugf("API returned unsuccessful response for player %s", playerID)
		return nil, fmt.Errorf("API returned unsuccessful response for player %s", playerID)
	}
	
	c.logger.Debugf("API returned %d tournament results for player %s", len(response.Data), playerID)

	// Convert the API response to the expected format
	history := &models.RatingHistory{
		PlayerID: playerID,
		Points:   make([]models.RatingPoint, 0, len(response.Data)),
	}

	// Convert tournament results to rating points
	for _, result := range response.Data {
		// Use pre-computed tournament date from optimized API - no more API calls needed!
		var tournamentDate time.Time
		if result.TournamentDate != nil {
			tournamentDate = *result.TournamentDate
		} else {
			// Fallback to estimation only if date not available from API
			c.logger.Debugf("Tournament %s has no pre-computed date, using estimation fallback", result.TournamentID)
			tournamentDate = c.estimateTournamentDate(result.TournamentID)
		}

		point := models.RatingPoint{
			Date:       tournamentDate,        // Use pre-computed date - PERFORMANCE OPTIMIZED!
			DWZ:        result.DWZNew,         // Use the new DWZ after the tournament
			Games:      result.Games,
			Points:     result.Points,         // Use points directly from API
			Tournament: result.TournamentName, // Use tournament name instead of ID for better display
		}

		history.Points = append(history.Points, point)
	}

	c.logger.Debugf("Converted %d tournament results to rating points for player %s", 
		len(response.Data), playerID)

	return history, nil
}

// GetAllClubs retrieves all clubs by making multiple paginated requests
func (c *Client) GetAllClubs(prefix string) ([]models.Club, error) {
	const batchSize = 500
	var allClubs []models.Club
	offset := 0

	c.logger.Debug("Starting to fetch all clubs...")
	for {
		result, err := c.SearchClubs(prefix, batchSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch clubs batch (offset %d): %w", offset, err)
		}

		allClubs = append(allClubs, result.Data...)

		c.logger.Debugf("Fetched %d clubs (total so far: %d)", len(result.Data), len(allClubs))

		// Check if we've got all clubs
		if len(result.Data) < batchSize || len(allClubs) >= result.Meta.Total {
			break
		}

		offset += batchSize
	}

	// Filter clubs by prefix if specified
	if prefix != "" {
		filteredClubs := make([]models.Club, 0)
		for _, club := range allClubs {
			if len(club.ID) >= len(prefix) && club.ID[:len(prefix)] == prefix {
				filteredClubs = append(filteredClubs, club)
			}
		}
		allClubs = filteredClubs
	}

	c.logger.Infof("Found %d clubs matching criteria", len(allClubs))
	return allClubs, nil
}
// GetAllClubPlayers retrieves all players for a specific club
func (c *Client) GetAllClubPlayers(clubID string, activeOnly bool) ([]models.Player, error) {
	const batchSize = 500
	var allPlayers []models.Player
	offset := 0

	c.logger.Debugf("Fetching all players for club %s...", clubID)

	for {
		result, err := c.GetClubPlayers(clubID, activeOnly, batchSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch players batch for club %s (offset %d): %w", clubID, offset, err)
		}

		allPlayers = append(allPlayers, result.Data...)

		c.logger.Debugf("Fetched %d players for club %s (total so far: %d)", 
			len(result.Data), clubID, len(allPlayers))

		// Check if we've got all players
		if len(result.Data) < batchSize || len(allPlayers) >= result.Meta.Total {
			break
		}

		offset += batchSize
	}

	c.logger.Debugf("Found %d players for club %s", len(allPlayers), clubID)
	return allPlayers, nil
}
// CheckHealth checks if the API is accessible
func (c *Client) CheckHealth() error {
	endpoint := "/health"
	
	var response map[string]interface{}
	if err := c.makeRequest("GET", endpoint, nil, &response); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// makeRequest performs an HTTP request and handles the response
func (c *Client) makeRequest(method, endpoint string, body io.Reader, result interface{}) error {
	url := c.baseURL + endpoint
	
	c.logger.Debugf("Making %s request to %s", method, url)
	
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debugf("Response status: %d, body length: %d", resp.StatusCode, len(responseBody))

	if resp.StatusCode >= 400 {
		var apiErr models.APIError
		if json.Unmarshal(responseBody, &apiErr) == nil {
			return apiErr
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	if result != nil {
		if err := json.Unmarshal(responseBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// SetTimeout updates the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetLogger sets a custom logger for the client
func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

// getTournamentDate fetches tournament details and returns the latest available date
// from start_date, end_date, finished_on, computed_on fields
// 
// DEPRECATED: This method is no longer used due to API optimization.
// The Portal64 API now returns pre-computed tournament dates in rating history responses,
// eliminating the need for additional API calls. This method is kept for reference only.
func (c *Client) getTournamentDate(tournamentID string) time.Time {
	// Try to fetch tournament details
	tournament, err := c.GetTournamentDetails(tournamentID)
	if err != nil {
		c.logger.Debugf("Could not fetch tournament details for %s, trying to estimate date from ID: %v", tournamentID, err)
		// Use estimation as fallback instead of generic date
		estimatedDate := c.estimateTournamentDate(tournamentID)
		c.logger.Debugf("Tournament %s: estimated date %v from tournament ID", tournamentID, estimatedDate)
		return estimatedDate
	}

	// Collect all available dates
	var availableDates []time.Time
	
	if tournament.StartDate != nil {
		availableDates = append(availableDates, *tournament.StartDate)
		c.logger.Debugf("Tournament %s has start_date: %v", tournamentID, *tournament.StartDate)
	}
	
	if tournament.EndDate != nil {
		availableDates = append(availableDates, *tournament.EndDate)
		c.logger.Debugf("Tournament %s has end_date: %v", tournamentID, *tournament.EndDate)
	}
	
	if tournament.FinishedOn != nil {
		availableDates = append(availableDates, *tournament.FinishedOn)
		c.logger.Debugf("Tournament %s has finished_on: %v", tournamentID, *tournament.FinishedOn)
	}
	
	if tournament.ComputedOn != nil {
		availableDates = append(availableDates, *tournament.ComputedOn)
		c.logger.Debugf("Tournament %s has computed_on: %v", tournamentID, *tournament.ComputedOn)
	}

	// If no dates are available, use estimation as fallback
	if len(availableDates) == 0 {
		c.logger.Debugf("Tournament %s has no available dates, estimating date from ID", tournamentID)
		estimatedDate := c.estimateTournamentDate(tournamentID)
		c.logger.Debugf("Tournament %s: estimated date %v from tournament ID", tournamentID, estimatedDate)
		return estimatedDate
	}

	// Find the latest date
	latestDate := availableDates[0]
	for _, date := range availableDates[1:] {
		if date.After(latestDate) {
			latestDate = date
		}
	}

	c.logger.Debugf("Tournament %s: selected latest date %v from %d available dates", 
		tournamentID, latestDate, len(availableDates))
	
	return latestDate
}

// estimateTournamentDate attempts to estimate tournament date from tournament ID
// Tournament IDs follow patterns like: B914-550-P4P, C413-612-DSV, T117893
// Format appears to be: [LETTER][YY][WW]-[OTHER]-[SUFFIX]
// Where YY = year (14 = 2014, 13 = 2013, etc.) and WW = week of year
// 
// DEPRECATED: This function is kept for fallback purposes only.
// Use getTournamentDate() instead which fetches real tournament data.
func (c *Client) estimateTournamentDate(tournamentID string) time.Time {
	// Default fallback date if parsing fails
	fallbackDate := time.Now().AddDate(-1, 0, 0) // 1 year ago
	
	if len(tournamentID) < 4 {
		return fallbackDate
	}

	// Try to parse tournament ID patterns
	if len(tournamentID) >= 4 && (tournamentID[0] == 'B' || tournamentID[0] == 'C') {
		// Pattern: B914-550-P4P or C413-612-DSV
		yearStr := tournamentID[1:3]
		if year, err := strconv.Atoi(yearStr); err == nil {
			// Convert 2-digit year to 4-digit year
			// Assume years 00-30 are 2000-2030, years 31-99 are 1931-1999
			fullYear := 2000 + year
			if year >= 31 {
				fullYear = 1900 + year
			}
			
			// Try to extract week number
			if len(tournamentID) >= 6 {
				weekStr := tournamentID[3:5]
				if week, err := strconv.Atoi(weekStr); err == nil && week >= 1 && week <= 53 {
					// Calculate date from year and week
					jan1 := time.Date(fullYear, 1, 1, 0, 0, 0, 0, time.UTC)
					_, week1 := jan1.ISOWeek()
					weekDiff := week - week1
					tournamentDate := jan1.AddDate(0, 0, weekDiff*7)
					return tournamentDate
				}
			}
			
			// If we can't parse week, just use mid-year
			return time.Date(fullYear, 6, 15, 0, 0, 0, 0, time.UTC)
		}
	} else if strings.HasPrefix(tournamentID, "T") {
		// Pattern: T117893 - might be some other encoding
		// For now, just use fallback
		return fallbackDate
	}

	c.logger.Debugf("Could not parse tournament date from ID: %s, using fallback", tournamentID)
	return fallbackDate
}

// ========================================
// ENHANCED EFFICIENT BULK OPERATIONS
// Inspired by Somatogramm's concurrent approach
// ========================================

// FetchAllPlayersEfficient retrieves all players via concurrent club processing
// This is the high-performance method that reduces API calls from 1+N+2NP to just 1+N
func (c *Client) FetchAllPlayersEfficient(clubPrefix string) ([]models.Player, error) {
	c.logger.Debug("Starting efficient bulk player fetch...")

	// First, fetch all clubs (same as before, but we'll use all data efficiently)
	clubs, err := c.GetAllClubs(clubPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clubs: %w", err)
	}

	c.logger.Debugf("Found %d clubs, fetching players concurrently with %d workers...", len(clubs), c.concurrency)

	// Create channels and worker pool for concurrent processing
	clubChan := make(chan models.Club, len(clubs))
	resultChan := make(chan []models.Player, len(clubs))
	errorChan := make(chan error, len(clubs))

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for club := range clubChan {
				c.logger.Debugf("Worker %d: Fetching players for club %s (%s)", workerID, club.ID, club.Name)

				clubPlayers, err := c.GetAllClubPlayers(club.ID, false) // Include all players
				if err != nil {
					c.logger.Warnf("Worker %d: failed to fetch players for club %s: %v", workerID, club.ID, err)
					errorChan <- err
					resultChan <- nil
					continue
				}

				c.logger.Debugf("Worker %d: Found %d players for club %s", workerID, len(clubPlayers), club.ID)

				// Set club information for each player (like Somatogramm does)
				for i := range clubPlayers {
					clubPlayers[i].ClubID = club.ID
					clubPlayers[i].Club = club.Name
				}

				resultChan <- clubPlayers
			}
		}(i)
	}

	// Send clubs to workers
	for _, club := range clubs {
		clubChan <- club
	}
	close(clubChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var allPlayers []models.Player
	processedClubs := 0
	errors := 0

	for result := range resultChan {
		processedClubs++
		if result != nil {
			allPlayers = append(allPlayers, result...)
		} else {
			errors++
		}
	}

	// Log any errors that occurred
	for err := range errorChan {
		if err != nil {
			c.logger.Debugf("Error occurred during bulk fetch: %v", err)
		}
	}

	c.logger.Infof("Efficient bulk fetch completed: %d total players from %d clubs (%d successful, %d errors)",
		len(allPlayers), len(clubs), processedClubs-errors, errors)

	return allPlayers, nil
}

// FilterValidPlayersForStatistics filters players for statistical analysis (like Somatogramm)
func (c *Client) FilterValidPlayersForStatistics(players []models.Player) []models.Player {
	var validPlayers []models.Player
	currentYear := time.Now().Year()

	for _, player := range players {
		// Skip players with invalid DWZ
		if player.CurrentDWZ <= 0 {
			continue
		}

		// Skip players without birth year
		if player.BirthYear == nil {
			continue
		}

		// Skip players with unrealistic ages
		age := currentYear - *player.BirthYear
		if age < 4 || age > 100 {
			continue
		}

		// Normalize gender field (like Somatogramm does)
		player.Gender = c.mapGenderToString(player.Gender)
		validPlayers = append(validPlayers, player)
	}

	c.logger.Debugf("Filtered %d valid players from %d total for statistical analysis", len(validPlayers), len(players))
	return validPlayers
}

// mapGenderToString normalizes gender values (from Somatogramm)
func (c *Client) mapGenderToString(gender string) string {
	switch strings.ToLower(gender) {
	case "1", "m", "male":
		return "m"
	case "0", "w", "female":
		return "w"
	case "2", "d", "divers":
		return "d"
	default:
		return "m" // Default to male if unknown
	}
}

// SetConcurrency updates the concurrency level for bulk operations
func (c *Client) SetConcurrency(concurrency int) {
	c.concurrency = concurrency
	c.logger.Debugf("Updated API client concurrency to %d", concurrency)
}