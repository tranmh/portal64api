package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"portal64api/internal/api/handlers"
	"portal64api/internal/cache"
	"portal64api/internal/config"
	"portal64api/internal/models"
	"portal64api/internal/services"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImportWorkflow_Integration tests the complete import workflow
func TestImportWorkflow_Integration(t *testing.T) {
	// Skip this test in CI/CD environments where SCP server is not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	tempDir := t.TempDir()

	// Create test configuration
	importConfig := &config.ImportConfig{
		Enabled:  true,
		Schedule: "0 2 * * *",
		SCP: config.SCPConfig{
			Host:         "localhost", // For testing, use localhost (would need mock SCP server)
			Port:         2222,        // Non-standard port for testing
			Username:     "testuser",
			Password:     "testpass",
			RemotePath:   "/test/data/",
			FilePatterns: []string{"test_*.zip"},
			Timeout:      30 * time.Second,
		},
		ZIP: config.ZIPConfig{
			PasswordMVDSB:     "testzip123",
			PasswordPortal64:  "testzip123",
			ExtractTimeout:    30 * time.Second,
		},
		Database: config.ImportDBConfig{
			ImportTimeout: 60 * time.Second,
			TargetDatabases: []config.TargetDatabase{
				{Name: "test_mvdsb", FilePattern: "test_mvdsb_*"},
				{Name: "test_portal64_bdw", FilePattern: "test_portal64_bdw_*"},
			},
		},
		Storage: config.StorageConfig{
			TempDir:          filepath.Join(tempDir, "temp"),
			MetadataFile:     filepath.Join(tempDir, "metadata.json"),
			CleanupOnSuccess: true,
			KeepFailedFiles:  true,
		},
		Freshness: config.FreshnessConfig{
			Enabled:           true,
			CompareTimestamp:  true,
			CompareSize:       true,
			CompareChecksum:   false,
			SkipIfNotNewer:    true,
		},
		LoadCheck: config.LoadCheckConfig{
			Enabled:       true,
			DelayDuration: 1 * time.Hour,
			MaxDelays:     3,
			LoadThreshold: 100,
		},
		Retry: config.RetryConfig{
			Enabled:     true,
			MaxAttempts: 2,
			RetryDelay:  5 * time.Minute,
			FailFast:    true,
		},
	}

	dbConfig := &config.DatabaseConfig{
		MVDSB: config.DatabaseConnection{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "test_mvdsb",
			Charset:  "utf8mb4",
		},
		Portal64BDW: config.DatabaseConnection{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "test_portal64_bdw",
			Charset:  "utf8mb4",
		},
	}

	// Create directories
	err := os.MkdirAll(importConfig.Storage.TempDir, 0755)
	require.NoError(t, err)

	// Create mock cache service
	mockCache := cache.NewMockCacheService(true)

	logger := log.New(os.Stdout, "INTEGRATION_TEST: ", log.LstdFlags)

	tests := []struct {
		name                string
		setupMockSCP        func() error
		expectedStatus      string
		expectedLogEntries  int
		skipReason          string
		errorContains       string
	}{
		{
			name: "successful_first_import",
			setupMockSCP: func() error {
				// Would set up mock SCP server with test files
				// For this integration test, we simulate the workflow
				return nil
			},
			expectedStatus:     "success",
			expectedLogEntries: 5, // Multiple log entries for complete workflow
		},
		{
			name: "no_newer_files_skip",
			setupMockSCP: func() error {
				// Create existing metadata to simulate previous successful import
				metadata := map[string]interface{}{
					"last_import": map[string]interface{}{
						"timestamp": time.Now().Add(-1 * time.Hour),
						"success":   true,
						"files": []models.FileMetadata{
							{
								Filename: "test_mvdsb_20250806.zip",
								Size:     1024000,
								ModTime:  time.Now().Add(-2 * time.Hour),
								Checksum: "sha256:test123",
							},
						},
					},
				}
				metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
				return ioutil.WriteFile(importConfig.Storage.MetadataFile, metadataJSON, 0644)
			},
			expectedStatus:     "skipped",
			skipReason:         "no_newer_files_available",
			expectedLogEntries: 3, // Fewer logs when skipped
		},
		{
			name: "scp_connection_failure",
			setupMockSCP: func() error {
				// Don't set up SCP server - will cause connection failure
				return nil
			},
			expectedStatus:     "failed",
			errorContains:      "connection",
			expectedLogEntries: 2, // Start + error log
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up from previous test
			os.RemoveAll(importConfig.Storage.TempDir)
			os.Remove(importConfig.Storage.MetadataFile)
			os.MkdirAll(importConfig.Storage.TempDir, 0755)

			// Setup mock environment
			err := tt.setupMockSCP()
			require.NoError(t, err)

			// Create import service
			importService := services.NewImportService(importConfig, dbConfig, mockCache, logger)
			require.NotNil(t, importService)

			// Start service
			err = importService.Start()
			require.NoError(t, err)
			defer importService.Stop()

			// Trigger manual import
			err = importService.TriggerManualImport()

			// Wait for import to complete (with timeout)
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()

			var finalStatus *models.ImportStatus

		statusCheck:
			for {
				select {
				case <-timeout:
					t.Fatal("Import did not complete within timeout")
				case <-ticker.C:
					status := importService.GetStatus()
					if status.Status != "running" {
						finalStatus = status
						break statusCheck
					}
				}
			}

			// Verify final status
			require.NotNil(t, finalStatus)
			assert.Equal(t, tt.expectedStatus, finalStatus.Status)

			// Verify skip reason if applicable
			if tt.skipReason != "" {
				assert.Equal(t, tt.skipReason, finalStatus.SkipReason)
			}

			// Verify error if applicable
			if tt.errorContains != "" {
				assert.Contains(t, finalStatus.Error, tt.errorContains)
			}

			// Verify log entries
			logs := importService.GetLogs(100)
			assert.GreaterOrEqual(t, len(logs), tt.expectedLogEntries)

			// Verify log structure
			if len(logs) > 0 {
				for _, logEntry := range logs {
					assert.NotEmpty(t, logEntry.Level)
					assert.NotEmpty(t, logEntry.Message)
					assert.NotEmpty(t, logEntry.Step)
					assert.False(t, logEntry.Timestamp.IsZero())
				}
			}

			// Verify completion times
			if finalStatus.Status != "running" {
				assert.NotNil(t, finalStatus.CompletedAt)
				if finalStatus.StartedAt != nil {
					assert.True(t, finalStatus.CompletedAt.After(*finalStatus.StartedAt) ||
						finalStatus.CompletedAt.Equal(*finalStatus.StartedAt))
				}
			}

			// Verify success-specific conditions
			if finalStatus.Status == "success" {
				assert.Equal(t, 100, finalStatus.Progress)
				assert.Empty(t, finalStatus.Error)
				assert.NotNil(t, finalStatus.LastSuccess)
				
				// Verify metadata file was updated
				assert.FileExists(t, importConfig.Storage.MetadataFile)
				
				// Verify temp directory cleanup if configured
				if importConfig.Storage.CleanupOnSuccess {
					files, _ := ioutil.ReadDir(importConfig.Storage.TempDir)
					assert.Empty(t, files, "Temp directory should be cleaned on success")
				}
			}

			// Verify failure-specific conditions
			if finalStatus.Status == "failed" {
				assert.NotEmpty(t, finalStatus.Error)
				assert.Greater(t, finalStatus.RetryCount, 0)
				
				// Verify failed files are kept if configured
				if importConfig.Storage.KeepFailedFiles {
					// Check temp directory still exists
					_, err := os.Stat(importConfig.Storage.TempDir)
					assert.NoError(t, err)
				}
			}

			// Verify skipped-specific conditions
			if finalStatus.Status == "skipped" {
				assert.Equal(t, 100, finalStatus.Progress)
				assert.Empty(t, finalStatus.Error)
				assert.NotEmpty(t, finalStatus.SkipReason)
			}
		})
	}
}

// TestImportAPI_Integration tests the HTTP API endpoints
func TestImportAPI_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()

	// Create minimal import configuration
	importConfig := &config.ImportConfig{
		Enabled:  true,
		Schedule: "0 2 * * *",
		SCP: config.SCPConfig{
			Host:         "localhost",
			Port:         22,
			Username:     "testuser",
			Password:     "testpass",
			RemotePath:   "/test/",
			FilePatterns: []string{"*.zip"},
			Timeout:      30 * time.Second,
		},
		ZIP: config.ZIPConfig{
			PasswordMVDSB:     "testpass",
			PasswordPortal64:  "testpass",
			ExtractTimeout:    30 * time.Second,
		},
		Database: config.ImportDBConfig{
			ImportTimeout: 60 * time.Second,
			TargetDatabases: []config.TargetDatabase{
				{Name: "test_db", FilePattern: "test_*"},
			},
		},
		Storage: config.StorageConfig{
			TempDir:      filepath.Join(tempDir, "temp"),
			MetadataFile: filepath.Join(tempDir, "metadata.json"),
		},
	}

	dbConfig := &config.DatabaseConfig{
		MVDSB: config.DatabaseConnection{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "test_mvdsb",
			Charset:  "utf8mb4",
		},
		Portal64BDW: config.DatabaseConnection{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "test_portal64_bdw",
			Charset:  "utf8mb4",
		},
	}

	mockCache := cache.NewMockCacheService(true)
	logger := log.New(os.Stdout, "API_TEST: ", log.LstdFlags)

	// Create directories
	err := os.MkdirAll(importConfig.Storage.TempDir, 0755)
	require.NoError(t, err)

	// Create import service
	importService := services.NewImportService(importConfig, dbConfig, mockCache, logger)
	require.NotNil(t, importService)

	err = importService.Start()
	require.NoError(t, err)
	defer importService.Stop()

	// Create handler
	handler := handlers.NewImportHandler(importService)

	// Setup router
	router := gin.New()
	router.GET("/api/v1/import/status", handler.GetImportStatus)
	router.POST("/api/v1/import/start", handler.StartManualImport)
	router.GET("/api/v1/import/logs", handler.GetImportLogs)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		body           string
		expectedCode   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "get_initial_status",
			endpoint:     "/api/v1/import/status",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var status models.ImportStatus
				err := json.Unmarshal(w.Body.Bytes(), &status)
				require.NoError(t, err)
				
				assert.Equal(t, "idle", status.Status)
				assert.Equal(t, 0, status.Progress)
				assert.Empty(t, status.CurrentStep)
				assert.NotNil(t, status.NextScheduled)
				assert.True(t, status.NextScheduled.After(time.Now()))
			},
		},
		{
			name:         "get_empty_logs",
			endpoint:     "/api/v1/import/logs",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Contains(t, response, "logs")
				logs := response["logs"].([]interface{})
				assert.Empty(t, logs)
			},
		},
		{
			name:         "trigger_manual_import",
			endpoint:     "/api/v1/import/start",
			method:       http.MethodPost,
			body:         "{}",
			expectedCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				// Note: In integration test, this might fail due to SCP connection
				// but the API should still respond with appropriate message
				if w.Code == http.StatusOK {
					assert.Contains(t, response, "message")
					assert.Contains(t, response, "started_at")
				} else {
					assert.Contains(t, response, "error")
				}
			},
		},
		{
			name:         "status_after_trigger",
			endpoint:     "/api/v1/import/status",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var status models.ImportStatus
				err := json.Unmarshal(w.Body.Bytes(), &status)
				require.NoError(t, err)
				
				// Status might be running, failed, or completed depending on timing
				assert.Contains(t, []string{"idle", "running", "success", "failed", "skipped"}, status.Status)
				
				// If running, should have started time
				if status.Status == "running" {
					assert.NotNil(t, status.StartedAt)
					assert.Nil(t, status.CompletedAt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.endpoint, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.endpoint, nil)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Basic assertions
			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

			// Custom validation
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}

			// Small delay between requests to avoid race conditions
			time.Sleep(100 * time.Millisecond)
		})
	}
}

// TestImportConcurrency_Integration tests concurrent access to import service
func TestImportConcurrency_Integration(t *testing.T) {
	tempDir := t.TempDir()

	importConfig := &config.ImportConfig{
		Enabled:  true,
		Schedule: "0 2 * * *",
		SCP: config.SCPConfig{
			Host:         "localhost",
			Port:         22,
			Username:     "testuser",
			Password:     "testpass",
			RemotePath:   "/test/",
			FilePatterns: []string{"*.zip"},
			Timeout:      10 * time.Second,
		},
		ZIP: config.ZIPConfig{
			PasswordMVDSB:     "testpass",
			PasswordPortal64:  "testpass",
			ExtractTimeout:    10 * time.Second,
		},
		Database: config.ImportDBConfig{
			ImportTimeout: 30 * time.Second,
			TargetDatabases: []config.TargetDatabase{
				{Name: "test_db", FilePattern: "test_*"},
			},
		},
		Storage: config.StorageConfig{
			TempDir:      filepath.Join(tempDir, "temp"),
			MetadataFile: filepath.Join(tempDir, "metadata.json"),
		},
	}

	dbConfig := &config.DatabaseConfig{
		MVDSB: config.DatabaseConnection{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "test_mvdsb",
			Charset:  "utf8mb4",
		},
		Portal64BDW: config.DatabaseConnection{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "test_portal64_bdw",
			Charset:  "utf8mb4",
		},
	}

	mockCache := cache.NewMockCacheService(true)
	logger := log.New(os.Stdout, "CONCURRENCY_TEST: ", log.LstdFlags)

	os.MkdirAll(importConfig.Storage.TempDir, 0755)

	importService := services.NewImportService(importConfig, dbConfig, mockCache, logger)
	require.NotNil(t, importService)

	err := importService.Start()
	require.NoError(t, err)
	defer importService.Stop()

	// Test concurrent status checks
	t.Run("concurrent_status_checks", func(t *testing.T) {
		const numGoroutines = 10
		const numRequests = 50

		results := make(chan *models.ImportStatus, numGoroutines*numRequests)
		errors := make(chan error, numGoroutines*numRequests)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < numRequests; j++ {
					status := importService.GetStatus()
					if status != nil {
						results <- status
					} else {
						errors <- fmt.Errorf("got nil status")
					}
					time.Sleep(time.Millisecond) // Small delay
				}
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines*numRequests; i++ {
			select {
			case status := <-results:
				assert.NotNil(t, status)
				assert.Contains(t, []string{"idle", "running", "success", "failed", "skipped"}, status.Status)
			case err := <-errors:
				t.Errorf("Concurrent access error: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests to complete")
			}
		}
	})

	// Test concurrent log access
	t.Run("concurrent_log_access", func(t *testing.T) {
		const numGoroutines = 5
		const numRequests = 20

		results := make(chan []models.ImportLogEntry, numGoroutines*numRequests)
		errors := make(chan error, numGoroutines*numRequests)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < numRequests; j++ {
					logs := importService.GetLogs(100)
					results <- logs
					time.Sleep(time.Millisecond)
				}
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines*numRequests; i++ {
			select {
			case logs := <-results:
				assert.NotNil(t, logs)
				// Logs might be empty initially, that's OK
			case err := <-errors:
				t.Errorf("Concurrent log access error: %v", err)
			case <-time.After(3 * time.Second):
				t.Fatal("Timeout waiting for concurrent log requests to complete")
			}
		}
	})

	// Test multiple import triggers (should be rejected if one is running)
	t.Run("multiple_import_triggers", func(t *testing.T) {
		const numAttempts = 5

		results := make(chan error, numAttempts)

		for i := 0; i < numAttempts; i++ {
			go func() {
				err := importService.TriggerManualImport()
				results <- err
			}()
		}

		successCount := 0
		rejectionCount := 0

		for i := 0; i < numAttempts; i++ {
			select {
			case err := <-results:
				if err == nil {
					successCount++
				} else {
					rejectionCount++
					// Should be rejected with appropriate error
					assert.Contains(t, err.Error(), "import") // Generic check
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for import trigger results")
			}
		}

		// At most one import should succeed, others should be rejected
		assert.LessOrEqual(t, successCount, 1, "Only one import should succeed")
		assert.GreaterOrEqual(t, rejectionCount, numAttempts-1, "Others should be rejected")
	})
}
