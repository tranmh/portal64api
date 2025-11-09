package services

import (
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKaderPlanungService_ExecuteManually(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.KaderPlanungConfig{
		Enabled:       true,
		BinaryPath:    "../../kader-planung/bin/kader-planung.exe", // Adjust path
		OutputDir:     outputDir,
		APIBaseURL:    "http://localhost:8080",
		Timeout:       30,
		Concurrency:   4,
		MinSampleSize: 10,
		Verbose:       false,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	service := NewKaderPlanungService(cfg, logger)

	// Test 1: Verify ExecuteManually is not a stub
	t.Run("ExecuteManually_Not_A_Stub", func(t *testing.T) {
		params := map[string]interface{}{
			"club_prefix": "C01", // Small subset for testing
			"timeout":     10,
		}

		err := service.ExecuteManually(params)
		require.NoError(t, err, "ExecuteManually should not return error immediately")

		// Wait a bit for execution to start
		time.Sleep(100 * time.Millisecond)

		status := service.GetStatus()
		assert.True(t, status.Running || status.LastExecution.After(time.Now().Add(-5*time.Second)),
			"ExecuteManually should actually start execution, not be a stub")
	})

	// Test 2: Concurrent execution protection
	t.Run("Prevents_Concurrent_Execution", func(t *testing.T) {
		service2 := NewKaderPlanungService(cfg, logger)

		err1 := service2.ExecuteManually(map[string]interface{}{})
		require.NoError(t, err1)

		err2 := service2.ExecuteManually(map[string]interface{}{})
		assert.Error(t, err2, "Should prevent concurrent execution")
		assert.Contains(t, err2.Error(), "already running")

		// Clean up
		time.Sleep(100 * time.Millisecond)
	})
}

func TestKaderPlanungService_BuildCommandArgs(t *testing.T) {
	cfg := &config.KaderPlanungConfig{
		Enabled:       true,
		BinaryPath:    "/path/to/binary",
		OutputDir:     "/output",
		APIBaseURL:    "http://localhost:8080",
		ClubPrefix:    "",
		Timeout:       30,
		Concurrency:   16,
		MinSampleSize: 10,
		Verbose:       false,
	}

	service := NewKaderPlanungService(cfg, logrus.New())

	tests := []struct {
		name     string
		params   map[string]interface{}
		expected []string
	}{
		{
			name:   "Default_Configuration",
			params: map[string]interface{}{},
			expected: []string{
				"--api-base-url", "http://localhost:8080",
				"--output-dir", "/output",
				"--timeout", "30",
				"--concurrency", "16",
				"--min-sample-size", "10",
			},
		},
		{
			name: "With_Club_Prefix",
			params: map[string]interface{}{
				"club_prefix": "C01",
			},
			expected: []string{
				"--api-base-url", "http://localhost:8080",
				"--output-dir", "/output",
				"--club-prefix", "C01",
				"--timeout", "30",
				"--concurrency", "16",
				"--min-sample-size", "10",
			},
		},
		{
			name: "With_Verbose",
			params: map[string]interface{}{
				"verbose": true,
			},
			expected: []string{
				"--api-base-url", "http://localhost:8080",
				"--output-dir", "/output",
				"--timeout", "30",
				"--concurrency", "16",
				"--min-sample-size", "10",
				"--verbose",
			},
		},
		{
			name: "Override_Config",
			params: map[string]interface{}{
				"timeout":     60,
				"concurrency": 8,
			},
			expected: []string{
				"--api-base-url", "http://localhost:8080",
				"--output-dir", "/output",
				"--timeout", "60",
				"--concurrency", "8",
				"--min-sample-size", "10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := service.buildCommandArgs(tt.params)
			assert.Equal(t, tt.expected, args, "Command args should match expected")
		})
	}
}

func TestKaderPlanungService_ListAvailableFiles(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.KaderPlanungConfig{
		OutputDir: outputDir,
	}

	service := NewKaderPlanungService(cfg, logrus.New())

	// Test 1: Empty directory
	t.Run("Empty_Directory", func(t *testing.T) {
		files, err := service.ListAvailableFiles()
		require.NoError(t, err)
		assert.Empty(t, files, "Should return empty list for empty directory")
	})

	// Test 2: With CSV files
	t.Run("With_CSV_Files", func(t *testing.T) {
		// Create test files
		testFile1 := filepath.Join(outputDir, "kader-planung-all-20250101.csv")
		testFile2 := filepath.Join(outputDir, "kader-planung-C01-20250102.csv")
		testFile3 := filepath.Join(outputDir, "ignore.txt")

		require.NoError(t, os.WriteFile(testFile1, []byte("test1"), 0644))
		require.NoError(t, os.WriteFile(testFile2, []byte("test2"), 0644))
		require.NoError(t, os.WriteFile(testFile3, []byte("test3"), 0644))

		files, err := service.ListAvailableFiles()
		require.NoError(t, err)
		assert.Len(t, files, 2, "Should return only CSV files")

		// Verify file info
		for _, file := range files {
			assert.NotEmpty(t, file.Name)
			assert.NotEmpty(t, file.Path)
			assert.Greater(t, file.Size, int64(0))
			assert.False(t, file.ModTime.IsZero())
		}
	})

	// Test 3: Non-existent directory
	t.Run("Non_Existent_Directory", func(t *testing.T) {
		service2 := NewKaderPlanungService(&config.KaderPlanungConfig{
			OutputDir: "/non/existent/path",
		}, logrus.New())

		files, err := service2.ListAvailableFiles()
		require.NoError(t, err, "Should not error on non-existent directory")
		assert.Empty(t, files)
	})
}

func TestKaderPlanungService_ParameterExtraction(t *testing.T) {
	service := NewKaderPlanungService(&config.KaderPlanungConfig{}, logrus.New())

	// Test getStringParam
	t.Run("GetStringParam", func(t *testing.T) {
		params := map[string]interface{}{
			"key1": "value1",
			"key2": 123, // Wrong type
		}

		assert.Equal(t, "value1", service.getStringParam(params, "key1", "default"))
		assert.Equal(t, "default", service.getStringParam(params, "key2", "default"))
		assert.Equal(t, "default", service.getStringParam(params, "missing", "default"))
	})

	// Test getIntParam
	t.Run("GetIntParam", func(t *testing.T) {
		params := map[string]interface{}{
			"int":     42,
			"float":   3.14,
			"string":  "99",
			"invalid": "abc",
		}

		assert.Equal(t, 42, service.getIntParam(params, "int", 0))
		assert.Equal(t, 3, service.getIntParam(params, "float", 0))
		assert.Equal(t, 99, service.getIntParam(params, "string", 0))
		assert.Equal(t, 0, service.getIntParam(params, "invalid", 0))
		assert.Equal(t, 0, service.getIntParam(params, "missing", 0))
	})

	// Test getBoolParam
	t.Run("GetBoolParam", func(t *testing.T) {
		params := map[string]interface{}{
			"true":    true,
			"false":   false,
			"invalid": "yes",
		}

		assert.True(t, service.getBoolParam(params, "true", false))
		assert.False(t, service.getBoolParam(params, "false", true))
		assert.True(t, service.getBoolParam(params, "invalid", true))
		assert.False(t, service.getBoolParam(params, "missing", false))
	})
}

func TestKaderPlanungService_StatusTracking(t *testing.T) {
	cfg := &config.KaderPlanungConfig{
		Enabled: true,
	}
	service := NewKaderPlanungService(cfg, logrus.New())

	// Initial status
	status := service.GetStatus()
	assert.False(t, status.Running)
	assert.True(t, status.StartTime.IsZero())
	assert.True(t, status.LastExecution.IsZero())
}

// Test to ensure no methods return nil without doing work
func TestKaderPlanungService_No_Stub_Methods(t *testing.T) {
	cfg := &config.KaderPlanungConfig{
		Enabled:    true,
		BinaryPath: "test-binary",
		OutputDir:  t.TempDir(),
	}
	service := NewKaderPlanungService(cfg, logrus.New())

	t.Run("ExecuteManually_Does_Something", func(t *testing.T) {
		// This should start execution in background
		err := service.ExecuteManually(map[string]interface{}{})
		assert.NoError(t, err, "Should not error on valid call")

		// Give it time to start
		time.Sleep(50 * time.Millisecond)

		status := service.GetStatus()
		// Should either be running or have executed (and failed due to test-binary)
		assert.True(t,
			status.Running || !status.LastExecution.IsZero() || status.LastError != "",
			"ExecuteManually should change service state, not be a stub returning nil")
	})

	t.Run("ListAvailableFiles_Returns_Real_Data", func(t *testing.T) {
		files, err := service.ListAvailableFiles()
		assert.NoError(t, err)
		assert.NotNil(t, files, "Should return initialized slice, not nil")
	})
}
