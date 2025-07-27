package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"portal64api/internal/models"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCSVResponseFormatting tests that CSV response formatting works correctly
func TestCSVResponseFormatting(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("CSV format with club data", func(t *testing.T) {
		// Create test data
		clubs := []models.ClubResponse{
			{
				ID:          "C0101",
				Name:        "Test Club 1",
				ShortName:   "TC1",
				Region:      "Baden",
				District:    "Stuttgart",
				MemberCount: 50,
				AverageDWZ:  1500,
				Status:      "active",
			},
			{
				ID:          "C0102",
				Name:        "Test Club 2",
				ShortName:   "TC2",
				Region:      "Württemberg",
				District:    "Tübingen",
				MemberCount: 75,
				AverageDWZ:  1600,
				Status:      "active",
			},
		}

		// Create a test router
		router := gin.New()
		router.GET("/test-csv", func(c *gin.Context) {
			utils.SendCSVResponse(c, "test_clubs.csv", clubs)
		})

		// Create a request with CSV format
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-csv", nil)
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "test_clubs.csv")

		// Verify CSV content
		body := w.Body.String()
		lines := strings.Split(body, "\n")
		
		// Should have header + 2 data rows + possible empty line at end
		assert.GreaterOrEqual(t, len(lines), 3)
		
		// Check header line
		headerLine := lines[0]
		assert.Contains(t, headerLine, "id")
		assert.Contains(t, headerLine, "name")
		assert.Contains(t, headerLine, "member_count")
		
		// Check first data line
		dataLine := lines[1]
		assert.Contains(t, dataLine, "C0101")
		assert.Contains(t, dataLine, "Test Club 1")
		assert.Contains(t, dataLine, "50")
		
		t.Logf("CSV output:\n%s", body)
	})
}

// TestHandleResponseFormat tests the HandleResponse function format detection
func TestHandleResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CSV format via query parameter", func(t *testing.T) {
		clubs := []models.ClubResponse{
			{
				ID:          "C0101",
				Name:        "Test Club",
				ShortName:   "TC",
				MemberCount: 50,
				Status:      "active",
			},
		}

		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			utils.HandleResponse(c, clubs, "clubs.csv")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?format=csv", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	})

	t.Run("JSON format by default", func(t *testing.T) {
		clubs := []models.ClubResponse{
			{
				ID:          "C0101",
				Name:        "Test Club",
				ShortName:   "TC",
				MemberCount: 50,
				Status:      "active",
			},
		}

		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			utils.HandleResponse(c, clubs, "clubs.csv")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})
}
