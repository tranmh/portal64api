package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"portal64api/internal/api"
	"portal64api/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestSwaggerEndpoints tests Swagger endpoints without database dependency
func TestSwaggerEndpoints(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create empty database struct - Swagger doesn't use it
	dbs := &database.Databases{}

	// Setup routes
	router := api.SetupRoutes(dbs)

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		description    string
	}{
		{
			name:           "Swagger base redirect",
			endpoint:       "/swagger",
			expectedStatus: http.StatusMovedPermanently,
			description:    "Base /swagger should redirect to /swagger/",
		},
		{
			name:           "Swagger index",
			endpoint:       "/swagger/",
			expectedStatus: http.StatusOK,
			description:    "Swagger index should be accessible",
		},
		{
			name:           "Swagger index.html",
			endpoint:       "/swagger/index.html",
			expectedStatus: http.StatusMovedPermanently,
			description:    "Swagger index.html should redirect to /swagger/",
		},
		{
			name:           "Swagger JSON docs",
			endpoint:       "/swagger/doc.json",
			expectedStatus: http.StatusOK,
			description:    "Swagger JSON docs should be accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.endpoint, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code,
				"Expected status %d for %s, got %d", tt.expectedStatus, tt.endpoint, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// For successful responses, check that we got some content
				assert.Greater(t, len(w.Body.String()), 0,
					"Expected non-empty response body for %s", tt.endpoint)
			}

			if tt.expectedStatus == http.StatusMovedPermanently {
				// For redirects, check the Location header
				location := w.Header().Get("Location")
				assert.Equal(t, "/swagger/", location,
					"Expected redirect to /swagger/, got %s", location)
			}
		})
	}
}
