package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

// TestKaderPlanungAPI_E2E tests the complete kader-planung API workflow
func TestKaderPlanungAPI_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Step 1: Check API health
	t.Run("Step1_API_Health", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err, "Health check should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Health response: %s", string(body))

		var health map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &health))
		assert.Equal(t, "healthy", health["status"])
	})

	// Step 2: Get initial status
	var initialStatus map[string]interface{}
	t.Run("Step2_Get_Initial_Status", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/kader-planung/status")
		require.NoError(t, err, "Status check should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Initial status response: %s", string(body))

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &result))

		assert.True(t, result["success"].(bool), "Status API should succeed")

		data := result["data"].(map[string]interface{})
		initialStatus = data["status"].(map[string]interface{})

		t.Logf("Running: %v", initialStatus["running"])
		t.Logf("Last error: %v", initialStatus["last_error"])
		t.Logf("Output files: %v", initialStatus["output_files"])
	})

	// Step 3: Trigger execution with minimal parameters
	t.Run("Step3_Trigger_Execution", func(t *testing.T) {
		// Use a small club prefix for faster testing
		payload := `{"club_prefix": "C0101", "timeout": 10, "concurrency": 2, "verbose": true}`

		resp, err := http.Post(
			baseURL+"/api/v1/kader-planung/start",
			"application/json",
			strings.NewReader(payload),
		)
		require.NoError(t, err, "Start execution should not fail")
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Start execution response: %s", string(body))

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &result))
		assert.True(t, result["success"].(bool), "Execution start should succeed")
	})

	// Step 4: Wait and poll status
	t.Run("Step4_Poll_Status_And_Wait", func(t *testing.T) {
		maxWait := 60 * time.Second // Wait up to 60 seconds for test
		pollInterval := 2 * time.Second
		deadline := time.Now().Add(maxWait)

		var finalStatus map[string]interface{}
		executionStarted := false
		executionCompleted := false

		for time.Now().Before(deadline) {
			resp, err := http.Get(baseURL + "/api/v1/kader-planung/status")
			require.NoError(t, err)

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var result map[string]interface{}
			require.NoError(t, json.Unmarshal(body, &result))

			data := result["data"].(map[string]interface{})
			status := data["status"].(map[string]interface{})
			finalStatus = status

			running := status["running"].(bool)
			lastError := status["last_error"]
			outputFiles := status["output_files"]

			t.Logf("[Poll] Running: %v, LastError: '%v', OutputFiles: %v",
				running, lastError, outputFiles)

			if running {
				executionStarted = true
				t.Logf("✓ Execution is running...")
			}

			if executionStarted && !running {
				executionCompleted = true
				t.Logf("✓ Execution completed!")
				break
			}

			if lastError != nil && lastError != "" {
				t.Errorf("❌ Execution error: %v", lastError)
				t.FailNow()
			}

			time.Sleep(pollInterval)
		}

		// Assertions
		assert.True(t, executionStarted,
			"Execution should have started (running=true at some point)")

		assert.True(t, executionCompleted,
			"Execution should have completed within timeout")

		// Check final status
		if finalStatus != nil {
			lastError := finalStatus["last_error"]
			if lastError != nil && lastError != "" {
				t.Errorf("❌ Final status has error: %v", lastError)
			}

			outputFiles := finalStatus["output_files"]
			t.Logf("Final output files: %v", outputFiles)

			if outputFiles != nil {
				files := outputFiles.([]interface{})
				assert.Greater(t, len(files), 0,
					"Should have generated at least one output file")
			}
		}
	})

	// Step 5: List generated files
	t.Run("Step5_List_Generated_Files", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/kader-planung/files")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Files list response: %s", string(body))

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &result))

		assert.True(t, result["success"].(bool))

		files := result["data"].([]interface{})
		t.Logf("Number of files: %d", len(files))

		assert.Greater(t, len(files), 0,
			"Should have at least one CSV file after execution")

		// Verify file structure
		if len(files) > 0 {
			file := files[0].(map[string]interface{})
			assert.NotEmpty(t, file["name"], "File should have name")
			assert.NotEmpty(t, file["path"], "File should have path")
			assert.Greater(t, file["size"], float64(0), "File should have size > 0")

			t.Logf("✓ Generated file: %v (size: %v bytes)",
				file["name"], file["size"])
		}
	})

	// Step 6: Verify we can download a file
	t.Run("Step6_Download_File", func(t *testing.T) {
		// First get file list
		resp, err := http.Get(baseURL + "/api/v1/kader-planung/files")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &result))

		files := result["data"].([]interface{})
		if len(files) == 0 {
			t.Skip("No files to download")
		}

		file := files[0].(map[string]interface{})
		filename := file["name"].(string)

		// Download the file
		downloadURL := baseURL + "/api/v1/kader-planung/download/" + filename
		resp2, err := http.Get(downloadURL)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)
		assert.Contains(t, resp2.Header.Get("Content-Type"), "text/csv")

		downloadedData, _ := io.ReadAll(resp2.Body)
		assert.Greater(t, len(downloadedData), 0, "Downloaded file should not be empty")

		t.Logf("✓ Downloaded file size: %d bytes", len(downloadedData))
	})
}

// TestKaderPlanungAPI_ErrorCases tests error handling
func TestKaderPlanungAPI_ErrorCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("Invalid_Request_Body", func(t *testing.T) {
		resp, err := http.Post(
			baseURL+"/api/v1/kader-planung/start",
			"application/json",
			strings.NewReader(`{invalid json}`),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Download_Nonexistent_File", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/kader-planung/download/nonexistent.csv")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestKaderPlanungAPI_ConcurrentExecution tests concurrent execution protection
func TestKaderPlanungAPI_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Start first execution
	payload := `{"club_prefix": "C0101", "timeout": 30}`

	resp1, err := http.Post(
		baseURL+"/api/v1/kader-planung/start",
		"application/json",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	defer resp1.Body.Close()

	body1, _ := io.ReadAll(resp1.Body)
	t.Logf("First execution: %s", string(body1))

	// Immediately try to start second execution
	resp2, err := http.Post(
		baseURL+"/api/v1/kader-planung/start",
		"application/json",
		strings.NewReader(payload),
	)
	require.NoError(t, err)
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	t.Logf("Second execution: %s", string(body2))

	// Should either get 409 Conflict or success:false
	if resp2.StatusCode == http.StatusConflict {
		t.Log("✓ Got 409 Conflict as expected")
	} else {
		var result map[string]interface{}
		json.Unmarshal(body2, &result)
		if !result["success"].(bool) {
			t.Log("✓ Got success:false as expected")
		}
	}

	// Wait for execution to complete
	time.Sleep(5 * time.Second)
}
