package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"portal64api/internal/api"
	"portal64api/internal/cache"
	"portal64api/internal/config"
	"portal64api/internal/database"
	"portal64api/internal/models"
	"portal64api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite defines the test suite for integration tests
type IntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
	dbs    *database.Databases
}

// SetupSuite runs once before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Load test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			MVDSB: config.DatabaseConnection{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "",
				Database: "mvdsb",
				Charset:  "utf8mb4",
			},
			Portal64BDW: config.DatabaseConnection{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "",
				Database: "portal64_bdw",
				Charset:  "utf8mb4",
			},
		},
	}

	// Connect to test databases
	dbs, err := database.Connect(cfg)
	if err != nil {
		suite.T().Skipf("Skipping integration tests: failed to connect to test databases: %v", err)
		return
	}

	suite.dbs = dbs

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup routes
	// Create mock cache service for integration tests
	mockCacheService := &cache.MockCacheService{}
	
	// Create nil import service for integration tests (not needed for basic API tests)
	var importService *services.ImportService = nil
	
	suite.router = api.SetupRoutes(dbs, mockCacheService, importService)
}

// TearDownSuite runs once after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.dbs != nil {
		suite.dbs.Close()
	}
}

// TestHealthEndpoint tests the health check endpoint
func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Equal(suite.T(), "1.0.0", response["version"])
}

// TestPlayersEndpoint tests the players search endpoint
func (suite *IntegrationTestSuite) TestPlayersEndpoint() {
	if suite.dbs == nil {
		suite.T().Skip("Database not available, skipping test")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/players?limit=5", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

// TestPlayerByIDEndpoint tests getting a specific player
func (suite *IntegrationTestSuite) TestPlayerByIDEndpoint() {
	if suite.dbs == nil {
		suite.T().Skip("Database not available, skipping test")
	}

	// Test with a potentially valid player ID format
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/players/C0101-1014", nil)
	suite.router.ServeHTTP(w, req)

	// Should return either 200 (found) or 404 (not found), but not 400 (bad request)
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusNotFound)

	// Test with invalid player ID format
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/players/invalid-id", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestClubsEndpoint tests the clubs search endpoint
func (suite *IntegrationTestSuite) TestClubsEndpoint() {
	if suite.dbs == nil {
		suite.T().Skip("Database not available, skipping test")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/clubs?limit=5", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

// TestClubByIDEndpoint tests getting a specific club
func (suite *IntegrationTestSuite) TestClubByIDEndpoint() {
	if suite.dbs == nil {
		suite.T().Skip("Database not available, skipping test")
	}

	// Test with a potentially valid club ID format
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/clubs/C0101", nil)
	suite.router.ServeHTTP(w, req)

	// Should return either 200 (found) or 404 (not found)
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}

// TestTournamentsEndpoint tests the tournaments search endpoint
func (suite *IntegrationTestSuite) TestTournamentsEndpoint() {
	if suite.dbs == nil {
		suite.T().Skip("Database not available, skipping test")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tournaments?limit=5", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response.Success)
}

// TestCSVFormat tests CSV response format
func (suite *IntegrationTestSuite) TestCSVFormat() {
	if suite.dbs == nil {
		suite.T().Skip("Database not available, skipping test")
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/clubs?format=csv&limit=1", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(suite.T(), w.Header().Get("Content-Disposition"), "attachment")
}

// TestCORSHeaders tests CORS headers are present
func (suite *IntegrationTestSuite) TestCORSHeaders() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/players", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestInvalidEndpoint tests 404 handling
func (suite *IntegrationTestSuite) TestInvalidEndpoint() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestSwaggerDocumentation tests that Swagger documentation is available
func (suite *IntegrationTestSuite) TestSwaggerDocumentation() {
	// Try different swagger endpoints that gin-swagger typically serves
	endpoints := []string{"/swagger/", "/swagger/index.html", "/swagger/doc.json"}
	
	foundWorking := false
	for _, endpoint := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", endpoint, nil)
		suite.router.ServeHTTP(w, req)
		
		// If we get 200, redirect, or even a structured response, consider it working
		if w.Code == http.StatusOK || (w.Code >= 300 && w.Code < 400) {
			foundWorking = true
			break
		}
	}
	
	assert.True(suite.T(), foundWorking)
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
