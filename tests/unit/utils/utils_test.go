package utils

import (
	"net/http"
	"net/url"
	"testing"

	"portal64api/internal/models"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseSearchParams(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    string
		expected models.SearchRequest
	}{
		{
			name:  "Default parameters",
			query: "",
			expected: models.SearchRequest{
				Query:     "",
				Limit:     20,
				Offset:    0,
				SortBy:    "name",
				SortOrder: "asc",
			},
		},
		{
			name:  "Custom parameters",
			query: "query=test&limit=50&offset=10&sort_by=dwz&sort_order=desc",
			expected: models.SearchRequest{
				Query:     "test",
				Limit:     50,
				Offset:    10,
				SortBy:    "dwz",
				SortOrder: "desc",
			},
		},
		{
			name:  "Limit too high",
			query: "limit=200",
			expected: models.SearchRequest{
				Query:     "",
				Limit:     100, // Should be capped at 100
				Offset:    0,
				SortBy:    "name",
				SortOrder: "asc",
			},
		},
		{
			name:  "Invalid limit",
			query: "limit=invalid",
			expected: models.SearchRequest{
				Query:     "",
				Limit:     20, // Should use default
				Offset:    0,
				SortBy:    "name",
				SortOrder: "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context with query parameters
			c, _ := gin.CreateTestContext(nil)
			req := &http.Request{}
			// Parse query string
			if tt.query != "" {
				parsedURL, _ := url.Parse("http://example.com?" + tt.query)
				req.URL = parsedURL
			} else {
				req.URL = &url.URL{}
			}
			c.Request = req

			// Test
			result := utils.ParseSearchParams(c)

			// Assert
			assert.Equal(t, tt.expected.Query, result.Query)
			assert.Equal(t, tt.expected.Limit, result.Limit)
			assert.Equal(t, tt.expected.Offset, result.Offset)
			assert.Equal(t, tt.expected.SortBy, result.SortBy)
			assert.Equal(t, tt.expected.SortOrder, result.SortOrder)
		})
	}
}

func TestGeneratePlayerID(t *testing.T) {
	tests := []struct {
		name     string
		vkz      string
		personID uint
		expected string
	}{
		{
			name:     "Standard player ID",
			vkz:      "C0101",
			personID: 1014,
			expected: "C0101-1014",
		},
		{
			name:     "Different club",
			vkz:      "C0205",
			personID: 5678,
			expected: "C0205-5678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GeneratePlayerID(tt.vkz, tt.personID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePlayerID(t *testing.T) {
	tests := []struct {
		name          string
		playerID      string
		expectedVKZ   string
		expectedID    uint
		expectedError bool
	}{
		{
			name:          "Valid player ID",
			playerID:      "C0101-1014",
			expectedVKZ:   "C0101",
			expectedID:    1014,
			expectedError: false,
		},
		{
			name:          "Invalid format - no dash",
			playerID:      "C01011014",
			expectedVKZ:   "",
			expectedID:    0,
			expectedError: true,
		},
		{
			name:          "Invalid format - multiple dashes",
			playerID:      "C0101-10-14",
			expectedVKZ:   "",
			expectedID:    0,
			expectedError: true,
		},
		{
			name:          "Invalid person ID",
			playerID:      "C0101-abc",
			expectedVKZ:   "",
			expectedID:    0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vkz, personID, err := utils.ParsePlayerID(tt.playerID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVKZ, vkz)
				assert.Equal(t, tt.expectedID, personID)
			}
		})
	}
}
