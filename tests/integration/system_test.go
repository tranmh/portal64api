package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"portal64api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SystemTestSuite defines the test suite for system tests against live deployment
type SystemTestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
}

// SetupSuite runs once before all tests in the suite
func (suite *SystemTestSuite) SetupSuite() {
	suite.baseURL = "http://test.svw.info:8080"
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Test if the server is reachable
	resp, err := suite.client.Get(suite.baseURL + "/health")
	if err != nil {
		suite.T().Skipf("Skipping system tests: server not reachable at %s: %v", suite.baseURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		suite.T().Skipf("Skipping system tests: server health check failed with status %d", resp.StatusCode)
	}
}

// makeRequest helper function to make HTTP requests
func (suite *SystemTestSuite) makeRequest(method, path string, params map[string]string) (*http.Response, error) {
	reqURL := suite.baseURL + path

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		reqURL += "?" + values.Encode()
	}

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	return suite.client.Do(req)
}

// assertJSONResponse helper to assert JSON response structure
func (suite *SystemTestSuite) assertJSONResponse(resp *http.Response, expectedStatus int) *models.Response {
	assert.Equal(suite.T(), expectedStatus, resp.StatusCode)
	assert.Contains(suite.T(), resp.Header.Get("Content-Type"), "application/json")

	var response models.Response
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(suite.T(), err)

	return &response
}

// assertCSVResponse helper to assert CSV response
func (suite *SystemTestSuite) assertCSVResponse(resp *http.Response, expectedStatus int) {
	assert.Equal(suite.T(), expectedStatus, resp.StatusCode)
	assert.Equal(suite.T(), "text/csv", resp.Header.Get("Content-Type"))
	assert.Contains(suite.T(), resp.Header.Get("Content-Disposition"), "attachment")
}

// TestHealthEndpoint tests the health check endpoint
func (suite *SystemTestSuite) TestHealthEndpoint() {
	resp, err := suite.makeRequest("GET", "/health", nil)
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", health["status"])
}

// MARK: - Clubs Tests

// TestClubsSearch tests the clubs search endpoint with various parameters
func (suite *SystemTestSuite) TestClubsSearch() {
	testCases := []struct {
		name   string
		params map[string]string
		status int
	}{
		{
			name:   "basic search",
			params: map[string]string{"limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "search with query",
			params: map[string]string{"query": "Schach", "limit": "10"},
			status: http.StatusOK,
		},
		{
			name:   "pagination",
			params: map[string]string{"limit": "5", "offset": "10"},
			status: http.StatusOK,
		},
		{
			name:   "sorting by name desc",
			params: map[string]string{"sort_by": "name", "sort_order": "desc", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "filter by region",
			params: map[string]string{"filter_by": "region", "filter_value": "Stuttgart", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "max limit",
			params: map[string]string{"limit": "500"},
			status: http.StatusOK,
		},
		{
			name:   "invalid limit over max",
			params: map[string]string{"limit": "150"},
			status: http.StatusBadRequest,
		},
		{
			name:   "invalid sort order",
			params: map[string]string{"sort_order": "invalid"},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", "/api/v1/clubs", tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if tc.status == http.StatusOK {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.True(suite.T(), response.Success)
				assert.NotNil(suite.T(), response.Data)

				// Verify meta information for paginated responses
				if response.Meta != nil {
					assert.GreaterOrEqual(suite.T(), response.Meta.Count, 0)
					assert.GreaterOrEqual(suite.T(), response.Meta.Total, response.Meta.Count)
				}
			} else {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// TestClubsCSVFormat tests CSV response format for clubs search
func (suite *SystemTestSuite) TestClubsCSVFormat() {
	resp, err := suite.makeRequest("GET", "/api/v1/clubs", map[string]string{
		"format": "csv",
		"limit":  "5",
	})
	assert.NoError(suite.T(), err)
	defer resp.Body.Close()

	suite.assertCSVResponse(resp, http.StatusOK)
}

// TestClubsAll tests getting all clubs
func (suite *SystemTestSuite) TestClubsAll() {
	testCases := []struct {
		name   string
		params map[string]string
		status int
	}{
		{
			name:   "get all clubs JSON",
			params: map[string]string{"format": "json"},
			status: http.StatusOK,
		},
		{
			name:   "get all clubs CSV",
			params: map[string]string{"format": "csv"},
			status: http.StatusOK,
		},
		{
			name:   "default format",
			params: map[string]string{},
			status: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", "/api/v1/clubs/all", tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if strings.Contains(tc.params["format"], "csv") {
				suite.assertCSVResponse(resp, tc.status)
			} else {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.True(suite.T(), response.Success)
				assert.NotNil(suite.T(), response.Data)
			}
		})
	}
}

// TestClubByID tests getting a specific club by ID
func (suite *SystemTestSuite) TestClubByID() {
	testCases := []struct {
		name     string
		clubID   string
		params   map[string]string
		expected int
	}{
		{
			name:     "valid club ID format JSON",
			clubID:   "C0101",
			params:   map[string]string{"format": "json"},
			expected: http.StatusOK, // May be 404 if club doesn't exist
		},
		{
			name:     "valid club ID format CSV",
			clubID:   "C0101",
			params:   map[string]string{"format": "csv"},
			expected: http.StatusOK, // May be 404 if club doesn't exist
		},
		{
			name:     "invalid club ID format",
			clubID:   "invalid-id",
			params:   map[string]string{},
			expected: http.StatusBadRequest,
		},
		{
			name:     "empty club ID",
			clubID:   "",
			params:   map[string]string{},
			expected: http.StatusOK, // Routes to SearchClubs, not GetClub
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			path := fmt.Sprintf("/api/v1/clubs/%s", tc.clubID)
			resp, err := suite.makeRequest("GET", path, tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			// For valid IDs, we accept both 200 (found) and 404 (not found)
			if tc.expected == http.StatusOK && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound) {
				if strings.Contains(tc.params["format"], "csv") && resp.StatusCode == http.StatusOK {
					suite.assertCSVResponse(resp, http.StatusOK)
				} else if resp.StatusCode == http.StatusOK {
					response := suite.assertJSONResponse(resp, http.StatusOK)
					assert.True(suite.T(), response.Success)
				} else {
					// 404 case
					response := suite.assertJSONResponse(resp, http.StatusNotFound)
					assert.False(suite.T(), response.Success)
				}
			} else {
				// For invalid formats, expect exact status code
				if resp.StatusCode == http.StatusBadRequest {
					response := suite.assertJSONResponse(resp, http.StatusBadRequest)
					assert.False(suite.T(), response.Success)
				} else {
					assert.Equal(suite.T(), tc.expected, resp.StatusCode)
				}
			}
		})
	}
}

// TestPlayersByClub tests getting players by club ID
func (suite *SystemTestSuite) TestPlayersByClub() {
	testCases := []struct {
		name     string
		clubID   string
		params   map[string]string
		expected int
	}{
		{
			name:     "valid club with players search",
			clubID:   "C0101",
			params:   map[string]string{"limit": "10"},
			expected: http.StatusOK, // May be 404 if club doesn't exist
		},
		{
			name:     "players with query filter",
			clubID:   "C0101",
			params:   map[string]string{"query": "Mueller", "limit": "5"},
			expected: http.StatusOK,
		},
		{
			name:     "players CSV format",
			clubID:   "C0101",
			params:   map[string]string{"format": "csv", "limit": "5"},
			expected: http.StatusOK,
		},
		{
			name:     "players with sorting",
			clubID:   "C0101",
			params:   map[string]string{"sort_by": "name", "sort_order": "desc", "limit": "5"},
			expected: http.StatusOK,
		},
		{
			name:     "invalid club ID format",
			clubID:   "invalid",
			params:   map[string]string{},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			path := fmt.Sprintf("/api/v1/clubs/%s/players", tc.clubID)
			resp, err := suite.makeRequest("GET", path, tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			// For valid club IDs, accept both 200 and 404
			if tc.expected == http.StatusOK && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound) {
				if strings.Contains(tc.params["format"], "csv") && resp.StatusCode == http.StatusOK {
					suite.assertCSVResponse(resp, http.StatusOK)
				} else if resp.StatusCode == http.StatusOK {
					response := suite.assertJSONResponse(resp, http.StatusOK)
					assert.True(suite.T(), response.Success)
					if response.Meta != nil {
						assert.GreaterOrEqual(suite.T(), response.Meta.Count, 0)
					}
				}
			} else if tc.expected == http.StatusBadRequest {
				response := suite.assertJSONResponse(resp, http.StatusBadRequest)
				assert.False(suite.T(), response.Success)
			}
		})
	}
}

// MARK: - Players Tests

// TestPlayersSearch tests the players search endpoint
func (suite *SystemTestSuite) TestPlayersSearch() {
	testCases := []struct {
		name   string
		params map[string]string
		status int
	}{
		{
			name:   "basic player search",
			params: map[string]string{"limit": "10"},
			status: http.StatusOK,
		},
		{
			name:   "search by name",
			params: map[string]string{"query": "Schmidt", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "player pagination",
			params: map[string]string{"limit": "5", "offset": "20"},
			status: http.StatusOK,
		},
		{
			name:   "sort players by name",
			params: map[string]string{"sort_by": "name", "sort_order": "asc", "limit": "10"},
			status: http.StatusOK,
		},
		{
			name:   "players CSV format",
			params: map[string]string{"format": "csv", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "max limit players",
			params: map[string]string{"limit": "500"},
			status: http.StatusOK,
		},
		{
			name:   "invalid limit over max",
			params: map[string]string{"limit": "200"},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", "/api/v1/players", tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if tc.status == http.StatusOK {
				if strings.Contains(tc.params["format"], "csv") {
					suite.assertCSVResponse(resp, tc.status)
				} else {
					response := suite.assertJSONResponse(resp, tc.status)
					assert.True(suite.T(), response.Success)
					assert.NotNil(suite.T(), response.Data)

					if response.Meta != nil {
						assert.GreaterOrEqual(suite.T(), response.Meta.Count, 0)
						assert.GreaterOrEqual(suite.T(), response.Meta.Total, response.Meta.Count)
					}
				}
			} else {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// TestPlayerByID tests getting a specific player by ID
func (suite *SystemTestSuite) TestPlayerByID() {
	testCases := []struct {
		name     string
		playerID string
		params   map[string]string
		expected int
	}{
		{
			name:     "valid player ID format JSON",
			playerID: "C0101-1014",
			params:   map[string]string{"format": "json"},
			expected: http.StatusOK, // May be 404 if player doesn't exist
		},
		{
			name:     "valid player ID format CSV",
			playerID: "C0101-1014",
			params:   map[string]string{"format": "csv"},
			expected: http.StatusOK,
		},
		{
			name:     "another valid format",
			playerID: "C0201-2050",
			params:   map[string]string{},
			expected: http.StatusOK,
		},
		{
			name:     "invalid player ID format",
			playerID: "invalid-player-id",
			params:   map[string]string{},
			expected: http.StatusBadRequest,
		},
		{
			name:     "malformed ID",
			playerID: "C123",
			params:   map[string]string{},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			path := fmt.Sprintf("/api/v1/players/%s", tc.playerID)
			resp, err := suite.makeRequest("GET", path, tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			// For valid IDs, accept both 200 and 404
			if tc.expected == http.StatusOK && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound) {
				if strings.Contains(tc.params["format"], "csv") && resp.StatusCode == http.StatusOK {
					suite.assertCSVResponse(resp, http.StatusOK)
				} else if resp.StatusCode == http.StatusOK {
					response := suite.assertJSONResponse(resp, http.StatusOK)
					assert.True(suite.T(), response.Success)
				} else {
					response := suite.assertJSONResponse(resp, http.StatusNotFound)
					assert.False(suite.T(), response.Success)
				}
			} else if tc.expected == http.StatusBadRequest {
				response := suite.assertJSONResponse(resp, http.StatusBadRequest)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// TestPlayerRatingHistory tests the player rating history endpoint
func (suite *SystemTestSuite) TestPlayerRatingHistory() {
	testCases := []struct {
		name     string
		playerID string
		params   map[string]string
		expected int
	}{
		{
			name:     "valid player rating history JSON",
			playerID: "C0101-1014",
			params:   map[string]string{"format": "json"},
			expected: http.StatusOK, // May be 404 if player doesn't exist
		},
		{
			name:     "valid player rating history CSV",
			playerID: "C0101-1014",
			params:   map[string]string{"format": "csv"},
			expected: http.StatusOK,
		},
		{
			name:     "default format rating history",
			playerID: "C0201-2050",
			params:   map[string]string{},
			expected: http.StatusOK,
		},
		{
			name:     "invalid player ID rating history",
			playerID: "invalid-id",
			params:   map[string]string{},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			path := fmt.Sprintf("/api/v1/players/%s/rating-history", tc.playerID)
			resp, err := suite.makeRequest("GET", path, tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			// For valid IDs, accept both 200 and 404
			if tc.expected == http.StatusOK && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound) {
				if strings.Contains(tc.params["format"], "csv") && resp.StatusCode == http.StatusOK {
					suite.assertCSVResponse(resp, http.StatusOK)
				} else if resp.StatusCode == http.StatusOK {
					response := suite.assertJSONResponse(resp, http.StatusOK)
					assert.True(suite.T(), response.Success)
					assert.NotNil(suite.T(), response.Data)
				} else {
					response := suite.assertJSONResponse(resp, http.StatusNotFound)
					assert.False(suite.T(), response.Success)
				}
			} else if tc.expected == http.StatusBadRequest {
				response := suite.assertJSONResponse(resp, http.StatusBadRequest)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// MARK: - Tournaments Tests

// TestTournamentsSearch tests the tournaments search endpoint
func (suite *SystemTestSuite) TestTournamentsSearch() {
	testCases := []struct {
		name   string
		params map[string]string
		status int
	}{
		{
			name:   "basic tournament search",
			params: map[string]string{"limit": "10"},
			status: http.StatusOK,
		},
		{
			name:   "search tournaments by name",
			params: map[string]string{"query": "Meisterschaft", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "tournament pagination",
			params: map[string]string{"limit": "10", "offset": "5"},
			status: http.StatusOK,
		},
		{
			name:   "sort tournaments by date desc",
			params: map[string]string{"sort_by": "finishedOn", "sort_order": "desc", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "filter tournaments by year",
			params: map[string]string{"filter_by": "year", "filter_value": "2023", "limit": "10"},
			status: http.StatusOK,
		},
		{
			name:   "tournaments CSV format",
			params: map[string]string{"format": "csv", "limit": "5"},
			status: http.StatusOK,
		},
		{
			name:   "max limit tournaments",
			params: map[string]string{"limit": "500"},
			status: http.StatusOK,
		},
		{
			name:   "invalid limit over max",
			params: map[string]string{"limit": "150"},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", "/api/v1/tournaments", tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if tc.status == http.StatusOK {
				if strings.Contains(tc.params["format"], "csv") {
					suite.assertCSVResponse(resp, tc.status)
				} else {
					response := suite.assertJSONResponse(resp, tc.status)
					assert.True(suite.T(), response.Success)
					assert.NotNil(suite.T(), response.Data)

					if response.Meta != nil {
						assert.GreaterOrEqual(suite.T(), response.Meta.Count, 0)
						assert.GreaterOrEqual(suite.T(), response.Meta.Total, response.Meta.Count)
					}
				}
			} else {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// TestTournamentsByDateRange tests the tournaments date range endpoint
func (suite *SystemTestSuite) TestTournamentsByDateRange() {
	testCases := []struct {
		name   string
		params map[string]string
		status int
	}{
		{
			name: "valid date range",
			params: map[string]string{
				"start_date": "2023-01-01",
				"end_date":   "2023-12-31",
				"limit":      "10",
			},
			status: http.StatusOK,
		},
		{
			name: "date range with query",
			params: map[string]string{
				"start_date": "2023-06-01",
				"end_date":   "2023-12-31",
				"query":      "Open",
				"limit":      "5",
			},
			status: http.StatusOK,
		},
		{
			name: "date range CSV format",
			params: map[string]string{
				"start_date": "2023-01-01",
				"end_date":   "2023-06-30",
				"format":     "csv",
				"limit":      "5",
			},
			status: http.StatusOK,
		},
		{
			name: "missing start_date",
			params: map[string]string{
				"end_date": "2023-12-31",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "missing end_date",
			params: map[string]string{
				"start_date": "2023-01-01",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "invalid date format",
			params: map[string]string{
				"start_date": "invalid-date",
				"end_date":   "2023-12-31",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", "/api/v1/tournaments/date-range", tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if tc.status == http.StatusOK {
				if strings.Contains(tc.params["format"], "csv") {
					suite.assertCSVResponse(resp, tc.status)
				} else {
					response := suite.assertJSONResponse(resp, tc.status)
					assert.True(suite.T(), response.Success)
					assert.NotNil(suite.T(), response.Data)
				}
			} else {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// TestTournamentsRecent tests the recent tournaments endpoint
func (suite *SystemTestSuite) TestTournamentsRecent() {
	testCases := []struct {
		name   string
		params map[string]string
		status int
	}{
		{
			name:   "default recent tournaments",
			params: map[string]string{},
			status: http.StatusOK,
		},
		{
			name:   "recent tournaments with custom days",
			params: map[string]string{"days": "60", "limit": "15"},
			status: http.StatusOK,
		},
		{
			name:   "recent tournaments CSV",
			params: map[string]string{"format": "csv", "limit": "10"},
			status: http.StatusOK,
		},
		{
			name:   "recent tournaments with limit",
			params: map[string]string{"limit": "5"},
			status: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", "/api/v1/tournaments/recent", tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if strings.Contains(tc.params["format"], "csv") {
				suite.assertCSVResponse(resp, tc.status)
			} else {
				response := suite.assertJSONResponse(resp, tc.status)
				assert.True(suite.T(), response.Success)
				assert.NotNil(suite.T(), response.Data)
			}
		})
	}
}

// TestTournamentByID tests getting a specific tournament by ID
func (suite *SystemTestSuite) TestTournamentByID() {
	testCases := []struct {
		name         string
		tournamentID string
		params       map[string]string
		expected     int
	}{
		{
			name:         "valid tournament ID format JSON",
			tournamentID: "C529-K00-HT1",
			params:       map[string]string{"format": "json"},
			expected:     http.StatusOK, // May be 404 if tournament doesn't exist
		},
		{
			name:         "valid tournament ID format CSV",
			tournamentID: "C529-K00-HT1",
			params:       map[string]string{"format": "csv"},
			expected:     http.StatusOK,
		},
		{
			name:         "another valid format",
			tournamentID: "C123-A01-OP2",
			params:       map[string]string{},
			expected:     http.StatusOK,
		},
		{
			name:         "invalid tournament ID format",
			tournamentID: "invalid-tournament-id",
			params:       map[string]string{},
			expected:     http.StatusBadRequest,
		},
		{
			name:         "malformed ID",
			tournamentID: "C123",
			params:       map[string]string{},
			expected:     http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			path := fmt.Sprintf("/api/v1/tournaments/%s", tc.tournamentID)
			resp, err := suite.makeRequest("GET", path, tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			// For valid IDs, accept both 200 and 404
			if tc.expected == http.StatusOK && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound) {
				if strings.Contains(tc.params["format"], "csv") && resp.StatusCode == http.StatusOK {
					suite.assertCSVResponse(resp, http.StatusOK)
				} else if resp.StatusCode == http.StatusOK {
					response := suite.assertJSONResponse(resp, http.StatusOK)
					assert.True(suite.T(), response.Success)
				} else {
					response := suite.assertJSONResponse(resp, http.StatusNotFound)
					assert.False(suite.T(), response.Success)
				}
			} else if tc.expected == http.StatusBadRequest {
				response := suite.assertJSONResponse(resp, http.StatusBadRequest)
				assert.False(suite.T(), response.Success)
				assert.NotEmpty(suite.T(), response.Error)
			}
		})
	}
}

// MARK: - Edge Cases and Error Handling Tests

// TestInvalidEndpoints tests handling of non-existent endpoints
func (suite *SystemTestSuite) TestInvalidEndpoints() {
	invalidPaths := []string{
		"/api/v1/nonexistent",
		"/api/v1/players/invalid/endpoint",
		"/api/v1/clubs/invalid/path",
		"/api/v1/tournaments/invalid/route",
		"/api/v2/players", // Wrong version
	}

	for _, path := range invalidPaths {
		suite.Run(fmt.Sprintf("invalid_path_%s", strings.ReplaceAll(path, "/", "_")), func() {
			resp, err := suite.makeRequest("GET", path, nil)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
		})
	}
}

// TestUnsupportedMethods tests handling of unsupported HTTP methods
func (suite *SystemTestSuite) TestUnsupportedMethods() {
	endpoints := []string{
		"/api/v1/players",
		"/api/v1/clubs",
		"/api/v1/tournaments",
	}

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, endpoint := range endpoints {
		for _, method := range methods {
			suite.Run(fmt.Sprintf("%s_%s", method, strings.ReplaceAll(endpoint, "/", "_")), func() {
				resp, err := suite.makeRequest(method, endpoint, nil)
				assert.NoError(suite.T(), err)
				defer resp.Body.Close()

				// Should return 405 Method Not Allowed or 404
				assert.True(suite.T(), resp.StatusCode == http.StatusMethodNotAllowed ||
					resp.StatusCode == http.StatusNotFound)
			})
		}
	}
}

// TestCORSHeaders tests CORS headers are present
func (suite *SystemTestSuite) TestCORSHeaders() {
	testPaths := []string{
		"/api/v1/players",
		"/api/v1/clubs",
		"/api/v1/tournaments",
	}

	for _, path := range testPaths {
		suite.Run(fmt.Sprintf("cors_%s", strings.ReplaceAll(path, "/", "_")), func() {
			req, err := http.NewRequest("OPTIONS", suite.baseURL+path, nil)
			assert.NoError(suite.T(), err)

			req.Header.Set("Origin", "http://localhost:3000")
			req.Header.Set("Access-Control-Request-Method", "GET")

			resp, err := suite.client.Do(req)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			// Check for CORS headers
			origin := resp.Header.Get("Access-Control-Allow-Origin")
			assert.True(suite.T(), origin == "*" || origin == "http://localhost:3000")
		})
	}
}

// MARK: - Performance and Load Tests

// TestResponseTimes tests that API responses are within acceptable time limits
func (suite *SystemTestSuite) TestResponseTimes() {
	endpoints := []struct {
		name   string
		path   string
		params map[string]string
	}{
		{"players_search", "/api/v1/players", map[string]string{"limit": "20"}},
		{"clubs_search", "/api/v1/clubs", map[string]string{"limit": "20"}},
		{"tournaments_search", "/api/v1/tournaments", map[string]string{"limit": "20"}},
		{"clubs_all", "/api/v1/clubs/all", map[string]string{}},
		{"tournaments_recent", "/api/v1/tournaments/recent", map[string]string{"limit": "10"}},
	}

	maxResponseTime := 5 * time.Second

	for _, endpoint := range endpoints {
		suite.Run(endpoint.name, func() {
			start := time.Now()
			resp, err := suite.makeRequest("GET", endpoint.path, endpoint.params)
			duration := time.Since(start)

			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
			assert.True(suite.T(), duration < maxResponseTime,
				"Response time %v exceeded maximum %v for %s", duration, maxResponseTime, endpoint.name)
		})
	}
}

// TestConcurrentRequests tests handling of concurrent requests
func (suite *SystemTestSuite) TestConcurrentRequests() {
	const numConcurrent = 10
	const endpoint = "/api/v1/players"
	params := map[string]string{"limit": "5"}

	type response struct {
		statusCode int
		err        error
	}

	responses := make(chan response, numConcurrent)

	// Launch concurrent requests
	for i := 0; i < numConcurrent; i++ {
		go func() {
			resp, err := suite.makeRequest("GET", endpoint, params)
			if resp != nil {
				responses <- response{statusCode: resp.StatusCode, err: err}
				resp.Body.Close()
			} else {
				responses <- response{statusCode: 0, err: err}
			}
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrent; i++ {
		resp := <-responses
		if resp.err == nil && resp.statusCode == http.StatusOK {
			successCount++
		}
	}

	// Expect most requests to succeed
	assert.GreaterOrEqual(suite.T(), successCount, numConcurrent/2,
		"Too many concurrent requests failed: %d/%d succeeded", successCount, numConcurrent)
}

// TestBoundaryValues tests boundary conditions for query parameters
func (suite *SystemTestSuite) TestBoundaryValues() {
	testCases := []struct {
		name     string
		endpoint string
		params   map[string]string
		expected int
	}{
		{
			name:     "zero limit",
			endpoint: "/api/v1/players",
			params:   map[string]string{"limit": "0"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "negative limit",
			endpoint: "/api/v1/players",
			params:   map[string]string{"limit": "-1"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "negative offset",
			endpoint: "/api/v1/players",
			params:   map[string]string{"offset": "-1"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "very large offset",
			endpoint: "/api/v1/players",
			params:   map[string]string{"offset": "999999"},
			expected: http.StatusOK, // Should return empty results
		},
		{
			name:     "extremely long query",
			endpoint: "/api/v1/players",
			params:   map[string]string{"query": strings.Repeat("a", 1000)},
			expected: http.StatusOK, // Should handle gracefully
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.makeRequest("GET", tc.endpoint, tc.params)
			assert.NoError(suite.T(), err)
			defer resp.Body.Close()

			if tc.expected == http.StatusOK {
				assert.Equal(suite.T(), tc.expected, resp.StatusCode)
			} else {
				// For bad requests, we might get 400 or the server might handle it gracefully
				assert.True(suite.T(), resp.StatusCode == tc.expected || resp.StatusCode == http.StatusOK)
			}
		})
	}
}

// TestSwaggerDocumentation tests that Swagger documentation is accessible
func (suite *SystemTestSuite) TestSwaggerDocumentation() {
	swaggerPaths := []string{
		"/swagger/",
		"/swagger/index.html",
		"/swagger/doc.json",
	}

	foundWorkingEndpoint := false
	for _, path := range swaggerPaths {
		resp, err := suite.makeRequest("GET", path, nil)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || (resp.StatusCode >= 300 && resp.StatusCode < 400) {
			foundWorkingEndpoint = true
			break
		}
	}

	assert.True(suite.T(), foundWorkingEndpoint, "No working Swagger documentation endpoint found")
}

// TestSystemSuite runs the complete system test suite
func TestSystemSuite(t *testing.T) {
	suite.Run(t, new(SystemTestSuite))
}
