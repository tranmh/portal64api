package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"portal64api/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImportEndToEnd tests the complete import feature end-to-end
func TestImportEndToEnd(t *testing.T) {
	// Skip E2E tests if not explicitly enabled
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("E2E tests disabled. Set RUN_E2E_TESTS=true to enable.")
	}

	// Get base URL from environment or use default
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	tests := []struct {
		name              string
		scenario          string
		expectedFinalStatus string
		timeout           time.Duration
	}{
		{
			name:              "manual_import_trigger",
			scenario:          "trigger_and_wait",
			expectedFinalStatus: "success", // or "failed" depending on environment
			timeout:           5 * time.Minute,
		},
		{
			name:              "status_monitoring",
			scenario:          "status_only",
			expectedFinalStatus: "idle",
			timeout:           30 * time.Second,
		},
		{
			name:              "log_retrieval",
			scenario:          "logs_only",
			expectedFinalStatus: "", // Not applicable
			timeout:           30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.scenario {
			case "trigger_and_wait":
				testManualImportTrigger(t, baseURL, tt.expectedFinalStatus, tt.timeout)
			case "status_only":
				testStatusMonitoring(t, baseURL)
			case "logs_only":
				testLogRetrieval(t, baseURL)
			}
		})
	}
}

func testManualImportTrigger(t *testing.T, baseURL, expectedFinalStatus string, timeout time.Duration) {
	// Step 1: Get initial status
	initialStatus := getImportStatus(t, baseURL)
	require.NotNil(t, initialStatus)
	
	t.Logf("Initial status: %s", initialStatus.Status)
	
	// Don't trigger if already running
	if initialStatus.Status == "running" {
		t.Log("Import already running, waiting for completion...")
	} else {
		// Step 2: Trigger manual import
		triggered := triggerManualImport(t, baseURL)
		if !triggered {
			t.Skip("Manual import could not be triggered (service may be disabled or busy)")
		}
		
		t.Log("Manual import triggered successfully")
	}
	
	// Step 3: Monitor progress until completion
	finalStatus := waitForImportCompletion(t, baseURL, timeout)
	require.NotNil(t, finalStatus)
	
	t.Logf("Final status: %s", finalStatus.Status)
	
	// Step 4: Verify final state
	assert.Contains(t, []string{"success", "failed", "skipped"}, finalStatus.Status)
	assert.Equal(t, 100, finalStatus.Progress)
	
	if finalStatus.Status == "success" {
		assert.Empty(t, finalStatus.Error)
		assert.NotNil(t, finalStatus.LastSuccess)
		assert.NotNil(t, finalStatus.FilesInfo)
		
		// Verify files were processed
		if finalStatus.FilesInfo != nil {
			t.Logf("Remote files: %d", len(finalStatus.FilesInfo.RemoteFiles))
			t.Logf("Downloaded: %d", len(finalStatus.FilesInfo.Downloaded))
			t.Logf("Extracted: %d", len(finalStatus.FilesInfo.Extracted))
			t.Logf("Imported: %d", len(finalStatus.FilesInfo.Imported))
		}
	}
	
	if finalStatus.Status == "failed" {
		assert.NotEmpty(t, finalStatus.Error)
		t.Logf("Import failed: %s", finalStatus.Error)
	}
	
	if finalStatus.Status == "skipped" {
		assert.NotEmpty(t, finalStatus.SkipReason)
		t.Logf("Import skipped: %s", finalStatus.SkipReason)
	}
	
	// Step 5: Verify logs were created
	logs := getImportLogs(t, baseURL)
	require.NotNil(t, logs)
	assert.Greater(t, len(logs), 0, "Expected log entries to be created")
	
	// Verify log structure
	for _, logEntry := range logs {
		assert.NotEmpty(t, logEntry.Level)
		assert.NotEmpty(t, logEntry.Message)
		assert.NotEmpty(t, logEntry.Step)
		assert.False(t, logEntry.Timestamp.IsZero())
	}
	
	// Print log summary
	t.Logf("Generated %d log entries", len(logs))
	if len(logs) > 0 {
		t.Logf("First log: [%s] %s", logs[0].Level, logs[0].Message)
		t.Logf("Last log: [%s] %s", logs[len(logs)-1].Level, logs[len(logs)-1].Message)
	}
}

func testStatusMonitoring(t *testing.T, baseURL string) {
	// Test multiple status requests to ensure consistency
	for i := 0; i < 5; i++ {
		status := getImportStatus(t, baseURL)
		require.NotNil(t, status)
		
		// Verify status structure
		assert.Contains(t, []string{"idle", "running", "success", "failed", "skipped"}, status.Status)
		assert.GreaterOrEqual(t, status.Progress, 0)
		assert.LessOrEqual(t, status.Progress, 100)
		assert.GreaterOrEqual(t, status.RetryCount, 0)
		assert.GreaterOrEqual(t, status.MaxRetries, 0)
		
		// Verify time fields are consistent
		if status.StartedAt != nil && status.CompletedAt != nil {
			assert.True(t, status.CompletedAt.After(*status.StartedAt) ||
				status.CompletedAt.Equal(*status.StartedAt))
		}
		
		if status.NextScheduled != nil {
			assert.True(t, status.NextScheduled.After(time.Now().Add(-24*time.Hour)),
				"Next scheduled time should be reasonable")
		}
		
		t.Logf("Status check %d: %s (progress: %d%%)", i+1, status.Status, status.Progress)
		
		time.Sleep(100 * time.Millisecond)
	}
}

func testLogRetrieval(t *testing.T, baseURL string) {
	// Test log retrieval multiple times
	for i := 0; i < 3; i++ {
		logs := getImportLogs(t, baseURL)
		require.NotNil(t, logs)
		
		t.Logf("Log retrieval %d: found %d entries", i+1, len(logs))
		
		// If there are logs, verify their structure
		for j, logEntry := range logs {
			assert.NotEmpty(t, logEntry.Level, "Log entry %d should have level", j)
			assert.NotEmpty(t, logEntry.Message, "Log entry %d should have message", j)
			assert.NotEmpty(t, logEntry.Step, "Log entry %d should have step", j)
			assert.False(t, logEntry.Timestamp.IsZero(), "Log entry %d should have timestamp", j)
			
			// Verify log level is valid
			assert.Contains(t, []string{"INFO", "WARN", "ERROR", "DEBUG"}, logEntry.Level)
		}
		
		// Verify logs are sorted by timestamp (newest first or oldest first)
		if len(logs) > 1 {
			isChronological := true
			isReverseChronological := true
			
			for j := 1; j < len(logs); j++ {
				if logs[j].Timestamp.Before(logs[j-1].Timestamp) {
					isChronological = false
				}
				if logs[j].Timestamp.After(logs[j-1].Timestamp) {
					isReverseChronological = false
				}
			}
			
			assert.True(t, isChronological || isReverseChronological,
				"Logs should be sorted chronologically")
		}
		
		time.Sleep(500 * time.Millisecond)
	}
}

// Helper functions

func getImportStatus(t *testing.T, baseURL string) *models.ImportStatus {
	url := fmt.Sprintf("%s/api/v1/import/status", baseURL)
	
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode, "Status endpoint should return 200")
	
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	
	var status models.ImportStatus
	err = json.Unmarshal(body, &status)
	require.NoError(t, err, "Status response should be valid JSON")
	
	return &status
}

func triggerManualImport(t *testing.T, baseURL string) bool {
	url := fmt.Sprintf("%s/api/v1/import/start", baseURL)
	
	requestBody := bytes.NewBufferString("{}")
	resp, err := http.Post(url, "application/json", requestBody)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	
	// Handle different response codes
	switch resp.StatusCode {
	case http.StatusOK:
		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "started_at")
		return true
		
	case http.StatusConflict:
		// Import already running
		t.Log("Import already in progress")
		return false
		
	case http.StatusServiceUnavailable:
		// Import service disabled
		t.Log("Import service is not available")
		return false
		
	default:
		var errorResponse map[string]interface{}
		err = json.Unmarshal(body, &errorResponse)
		if err == nil && errorResponse["error"] != nil {
			t.Logf("Import trigger failed: %s", errorResponse["error"])
		}
		require.Failf(t, "Unexpected response code", "Got %d, expected 200, 409, or 503", resp.StatusCode)
		return false
	}
}

func waitForImportCompletion(t *testing.T, baseURL string, timeout time.Duration) *models.ImportStatus {
	timeoutChan := time.After(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeoutChan:
			t.Fatalf("Import did not complete within timeout of %v", timeout)
			return nil
			
		case <-ticker.C:
			status := getImportStatus(t, baseURL)
			
			t.Logf("Import status: %s (progress: %d%%)", status.Status, status.Progress)
			
			// Check if import is complete
			if status.Status != "running" {
				return status
			}
			
			// Log current step if available
			if status.CurrentStep != "" {
				t.Logf("Current step: %s", status.CurrentStep)
			}
		}
	}
}

func getImportLogs(t *testing.T, baseURL string) []models.ImportLogEntry {
	url := fmt.Sprintf("%s/api/v1/import/logs", baseURL)
	
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	require.Equal(t, http.StatusOK, resp.StatusCode, "Logs endpoint should return 200")
	
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err, "Logs response should be valid JSON")
	
	require.Contains(t, response, "logs")
	
	logsData, err := json.Marshal(response["logs"])
	require.NoError(t, err)
	
	var logs []models.ImportLogEntry
	err = json.Unmarshal(logsData, &logs)
	require.NoError(t, err, "Logs data should be valid array")
	
	return logs
}

// TestImportStressTest performs stress testing on import endpoints
func TestImportStressTest(t *testing.T) {
	if os.Getenv("RUN_E2E_STRESS_TESTS") != "true" {
		t.Skip("E2E stress tests disabled. Set RUN_E2E_STRESS_TESTS=true to enable.")
	}

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	t.Run("concurrent_status_requests", func(t *testing.T) {
		const numGoroutines = 20
		const requestsPerGoroutine = 50

		results := make(chan error, numGoroutines*requestsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < requestsPerGoroutine; j++ {
					_, err := http.Get(fmt.Sprintf("%s/api/v1/import/status", baseURL))
					results <- err
				}
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
			select {
			case err := <-results:
				assert.NoError(t, err, "Status request should not fail")
			case <-time.After(30 * time.Second):
				t.Fatal("Stress test timed out")
			}
		}
	})

	t.Run("concurrent_log_requests", func(t *testing.T) {
		const numGoroutines = 10
		const requestsPerGoroutine = 30

		results := make(chan error, numGoroutines*requestsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < requestsPerGoroutine; j++ {
					_, err := http.Get(fmt.Sprintf("%s/api/v1/import/logs", baseURL))
					results <- err
				}
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
			select {
			case err := <-results:
				assert.NoError(t, err, "Log request should not fail")
			case <-time.After(20 * time.Second):
				t.Fatal("Log stress test timed out")
			}
		}
	})

	t.Run("rapid_import_triggers", func(t *testing.T) {
		const numAttempts = 10

		successCount := 0
		conflictCount := 0

		for i := 0; i < numAttempts; i++ {
			requestBody := bytes.NewBufferString("{}")
			resp, err := http.Post(fmt.Sprintf("%s/api/v1/import/start", baseURL),
				"application/json", requestBody)
			require.NoError(t, err)
			resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				successCount++
			case http.StatusConflict:
				conflictCount++
			case http.StatusServiceUnavailable:
				// Service disabled, acceptable
			default:
				t.Logf("Unexpected status code: %d", resp.StatusCode)
			}

			// Small delay between attempts
			time.Sleep(100 * time.Millisecond)
		}

		t.Logf("Import trigger results: %d success, %d conflicts", successCount, conflictCount)

		// At most one import should succeed initially
		assert.LessOrEqual(t, successCount, 3, "Should have limited successful triggers")
		assert.GreaterOrEqual(t, conflictCount, numAttempts/2, "Most triggers should result in conflicts")
	})
}

// TestImportConfiguration tests configuration-related functionality
func TestImportConfiguration(t *testing.T) {
	if os.Getenv("RUN_E2E_CONFIG_TESTS") != "true" {
		t.Skip("E2E config tests disabled. Set RUN_E2E_CONFIG_TESTS=true to enable.")
	}

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	t.Run("verify_scheduled_import_time", func(t *testing.T) {
		status := getImportStatus(t, baseURL)
		require.NotNil(t, status)

		if status.NextScheduled != nil {
			nextScheduled := *status.NextScheduled
			now := time.Now()

			// Next scheduled should be in the future
			assert.True(t, nextScheduled.After(now),
				"Next scheduled time should be in the future")

			// Next scheduled should be within reasonable bounds (24-48 hours)
			maxExpected := now.Add(48 * time.Hour)
			assert.True(t, nextScheduled.Before(maxExpected),
				"Next scheduled time should be within 48 hours")

			t.Logf("Next scheduled import: %s", nextScheduled.Format(time.RFC3339))
		}
	})

	t.Run("verify_retry_configuration", func(t *testing.T) {
		status := getImportStatus(t, baseURL)
		require.NotNil(t, status)

		// Verify retry configuration is reasonable
		assert.GreaterOrEqual(t, status.MaxRetries, 0)
		assert.LessOrEqual(t, status.MaxRetries, 10) // Reasonable upper bound
		assert.GreaterOrEqual(t, status.RetryCount, 0)
		assert.LessOrEqual(t, status.RetryCount, status.MaxRetries)

		t.Logf("Retry configuration: %d/%d", status.RetryCount, status.MaxRetries)
	})

	t.Run("verify_service_availability", func(t *testing.T) {
		// Test that import endpoints are available (not returning 503)
		endpoints := []string{
			"/api/v1/import/status",
			"/api/v1/import/logs",
		}

		for _, endpoint := range endpoints {
			resp, err := http.Get(fmt.Sprintf("%s%s", baseURL, endpoint))
			require.NoError(t, err)
			resp.Body.Close()

			assert.NotEqual(t, http.StatusServiceUnavailable, resp.StatusCode,
				"Endpoint %s should not return 503 if import service is enabled", endpoint)
		}
	})
}

// TestImportDocumentation tests that endpoints match documentation
func TestImportDocumentation(t *testing.T) {
	if os.Getenv("RUN_E2E_DOC_TESTS") != "true" {
		t.Skip("E2E documentation tests disabled. Set RUN_E2E_DOC_TESTS=true to enable.")
	}

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	t.Run("status_endpoint_response_format", func(t *testing.T) {
		status := getImportStatus(t, baseURL)
		require.NotNil(t, status)

		// Verify all documented fields are present
		assert.Contains(t, []string{"idle", "running", "success", "failed", "skipped"}, status.Status)
		assert.GreaterOrEqual(t, status.Progress, 0)
		assert.LessOrEqual(t, status.Progress, 100)
		// CurrentStep can be empty
		// StartedAt, CompletedAt, LastSuccess can be nil
		// NextScheduled should be present for scheduled imports
		assert.GreaterOrEqual(t, status.RetryCount, 0)
		assert.GreaterOrEqual(t, status.MaxRetries, 0)
		// Error and SkipReason can be empty
		// FilesInfo can be nil
	})

	t.Run("logs_endpoint_response_format", func(t *testing.T) {
		logs := getImportLogs(t, baseURL)
		require.NotNil(t, logs)

		// Verify log entry format
		for _, logEntry := range logs {
			assert.Contains(t, []string{"INFO", "WARN", "ERROR", "DEBUG"}, logEntry.Level)
			assert.NotEmpty(t, logEntry.Message)
			assert.NotEmpty(t, logEntry.Step)
			assert.False(t, logEntry.Timestamp.IsZero())
		}
	})

	t.Run("start_endpoint_response_format", func(t *testing.T) {
		// Try to start import (might fail, but response format should be correct)
		requestBody := bytes.NewBufferString("{}")
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/import/start", baseURL),
			"application/json", requestBody)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		require.NoError(t, err, "Response should be valid JSON")

		switch resp.StatusCode {
		case http.StatusOK:
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "started_at")
		case http.StatusConflict, http.StatusServiceUnavailable, http.StatusInternalServerError:
			assert.Contains(t, response, "error")
		default:
			t.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
	})
}

// TestImportRecovery tests error recovery scenarios
func TestImportRecovery(t *testing.T) {
	if os.Getenv("RUN_E2E_RECOVERY_TESTS") != "true" {
		t.Skip("E2E recovery tests disabled. Set RUN_E2E_RECOVERY_TESTS=true to enable.")
	}

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	t.Run("service_restart_recovery", func(t *testing.T) {
		// This test would verify that the service can recover after restart
		// In a real scenario, this would involve:
		// 1. Getting current status
		// 2. Restarting the service (external process)
		// 3. Verifying status is restored correctly
		// 4. Verifying next scheduled import is maintained

		// For now, just verify that status is consistent
		status1 := getImportStatus(t, baseURL)
		time.Sleep(1 * time.Second)
		status2 := getImportStatus(t, baseURL)

		// Next scheduled should be consistent (within reason)
		if status1.NextScheduled != nil && status2.NextScheduled != nil {
			timeDiff := status2.NextScheduled.Sub(*status1.NextScheduled)
			assert.LessOrEqual(t, timeDiff.Abs(), 2*time.Second,
				"Next scheduled time should be stable")
		}
	})

	t.Run("graceful_error_handling", func(t *testing.T) {
		// Test that the service handles various error scenarios gracefully
		// This includes testing with invalid requests, etc.

		// Test with invalid JSON
		invalidJSON := bytes.NewBufferString("{invalid json}")
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/import/start", baseURL),
			"application/json", invalidJSON)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should handle invalid JSON gracefully
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusOK, http.StatusConflict, http.StatusServiceUnavailable},
			resp.StatusCode)
	})
}
