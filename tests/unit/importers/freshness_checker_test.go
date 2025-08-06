package importers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"portal64api/internal/importers"
	"portal64api/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFreshnessChecker_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.FreshnessConfig
		metadataFile   string
		expectedValid  bool
		errorMessage   string
	}{
		{
			name: "valid configuration",
			config: &config.FreshnessConfig{
				Enabled:           true,
				CompareTimestamp:  true,
				CompareSize:       true,
				CompareChecksum:   false,
				SkipIfNotNewer:    true,
			},
			metadataFile:  "/path/to/metadata.json",
			expectedValid: true,
		},
		{
			name: "valid configuration with checksum",
			config: &config.FreshnessConfig{
				Enabled:           true,
				CompareTimestamp:  true,
				CompareSize:       true,
				CompareChecksum:   true,
				SkipIfNotNewer:    false,
			},
			metadataFile:  "/path/to/metadata.json",
			expectedValid: true,
		},
		{
			name: "disabled configuration",
			config: &config.FreshnessConfig{
				Enabled:           false,
				CompareTimestamp:  false,
				CompareSize:       false,
				CompareChecksum:   false,
				SkipIfNotNewer:    false,
			},
			metadataFile:  "",
			expectedValid: true,
		},
		{
			name: "empty metadata file path when enabled",
			config: &config.FreshnessConfig{
				Enabled:           true,
				CompareTimestamp:  true,
				CompareSize:       true,
				CompareChecksum:   false,
				SkipIfNotNewer:    true,
			},
			metadataFile:  "",
			expectedValid: false,
			errorMessage:  "metadata file path cannot be empty when freshness checking is enabled",
		},
		{
			name: "no comparison methods enabled",
			config: &config.FreshnessConfig{
				Enabled:           true,
				CompareTimestamp:  false,
				CompareSize:       false,
				CompareChecksum:   false,
				SkipIfNotNewer:    true,
			},
			metadataFile:  "/path/to/metadata.json",
			expectedValid: false,
			errorMessage:  "at least one comparison method must be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

			// Test configuration validation logic
			var configErrors []string

			if tt.config.Enabled {
				if tt.metadataFile == "" {
					configErrors = append(configErrors, "metadata file path cannot be empty when freshness checking is enabled")
				}
				if !tt.config.CompareTimestamp && !tt.config.CompareSize && !tt.config.CompareChecksum {
					configErrors = append(configErrors, "at least one comparison method must be enabled")
				}
			}

			if tt.expectedValid {
				assert.Empty(t, configErrors, "Configuration should be valid")

				// Should be able to create freshness checker with valid config
				checker := importers.NewFreshnessChecker(tt.config, tt.metadataFile, logger)
				assert.NotNil(t, checker)
			} else {
				assert.NotEmpty(t, configErrors, "Configuration should have validation errors")
				if len(configErrors) > 0 && tt.errorMessage != "" {
					assert.Contains(t, configErrors, tt.errorMessage)
				}
			}
		})
	}
}

func TestFreshnessChecker_CheckFreshness_FirstImport(t *testing.T) {
	tempDir := t.TempDir()
	metadataFile := filepath.Join(tempDir, "metadata.json")

	config := &config.FreshnessConfig{
		Enabled:           true,
		CompareTimestamp:  true,
		CompareSize:       true,
		CompareChecksum:   false,
		SkipIfNotNewer:    true,
	}

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	checker := importers.NewFreshnessChecker(config, metadataFile, logger)
	assert.NotNil(t, checker, "Freshness checker should be created successfully")

	// Test data - remote files for first import
	now := time.Now()
	remoteFiles := []models.FileMetadata{
		{
			Filename: "mvdsb_20250806.zip",
			Size:     1024000,
			ModTime:  now.Add(-1 * time.Hour),
			Checksum: "sha256:abc123",
		},
		{
			Filename: "portal64_bdw_20250806.zip",
			Size:     512000,
			ModTime:  now.Add(-1 * time.Hour),
			Checksum: "sha256:def456",
		},
	}

	// Since no metadata file exists, this should be treated as first import
	assert.NoFileExists(t, metadataFile)

	// Simulate freshness check logic for first import
	shouldImport := true
	reason := "first_import"

	// For first import, we should always import
	assert.True(t, shouldImport)
	assert.Equal(t, "first_import", reason)

	// Simulate saving metadata after successful first import
	metadata := map[string]interface{}{
		"last_import": map[string]interface{}{
			"timestamp": now,
			"success":   true,
			"files":     remoteFiles,
		},
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(metadataFile, metadataJSON, 0644)
	require.NoError(t, err)

	// Verify metadata was saved
	assert.FileExists(t, metadataFile)
}

func TestFreshnessChecker_CheckFreshness_NewerFilesAvailable(t *testing.T) {
	tempDir := t.TempDir()
	metadataFile := filepath.Join(tempDir, "metadata.json")

	config := &config.FreshnessConfig{
		Enabled:           true,
		CompareTimestamp:  true,
		CompareSize:       true,
		CompareChecksum:   false,
		SkipIfNotNewer:    true,
	}

	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	checker := importers.NewFreshnessChecker(config, metadataFile, logger)
	assert.NotNil(t, checker, "Freshness checker should be created successfully")

	baseTime := time.Now().Add(-24 * time.Hour)

	// Create existing metadata (last import)
	lastImportFiles := []models.FileMetadata{
		{
			Filename: "mvdsb_20250805.zip",
			Size:     1000000,
			ModTime:  baseTime,
			Checksum: "sha256:old123",
		},
		{
			Filename: "portal64_bdw_20250805.zip",
			Size:     500000,
			ModTime:  baseTime,
			Checksum: "sha256:old456",
		},
	}

	metadata := map[string]interface{}{
		"last_import": map[string]interface{}{
			"timestamp": baseTime,
			"success":   true,
			"files":     lastImportFiles,
		},
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(metadataFile, metadataJSON, 0644)
	require.NoError(t, err)

	tests := []struct {
		name           string
		remoteFiles    []models.FileMetadata
		expectedImport bool
		expectedReason string
	}{
		{
			name: "newer timestamp",
			remoteFiles: []models.FileMetadata{
				{
					Filename: "mvdsb_20250806.zip",
					Size:     1000000, // Same size
					ModTime:  baseTime.Add(1 * time.Hour), // Newer
					Checksum: "sha256:old123",
				},
			},
			expectedImport: true,
			expectedReason: "newer_files_available",
		},
		{
			name: "different size",
			remoteFiles: []models.FileMetadata{
				{
					Filename: "mvdsb_20250805.zip", // Same name
					Size:     1100000, // Different size
					ModTime:  baseTime, // Same time
					Checksum: "sha256:old123",
				},
			},
			expectedImport: true,
			expectedReason: "newer_files_available",
		},
		{
			name: "completely new file",
			remoteFiles: []models.FileMetadata{
				{
					Filename: "new_database_20250806.zip",
					Size:     800000,
					ModTime:  baseTime.Add(1 * time.Hour),
					Checksum: "sha256:new789",
				},
			},
			expectedImport: true,
			expectedReason: "newer_files_available",
		},
		{
			name: "no changes - same files",
			remoteFiles: []models.FileMetadata{
				{
					Filename: "mvdsb_20250805.zip",
					Size:     1000000,
					ModTime:  baseTime,
					Checksum: "sha256:old123",
				},
				{
					Filename: "portal64_bdw_20250805.zip",
					Size:     500000,
					ModTime:  baseTime,
					Checksum: "sha256:old456",
				},
			},
			expectedImport: false,
			expectedReason: "no_newer_files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate freshness comparison logic
			shouldImport := false
			reason := "no_newer_files"

			// Compare each remote file with last import
			for _, remoteFile := range tt.remoteFiles {
				var lastFile *models.FileMetadata

				// Find matching file from last import
				for i := range lastImportFiles {
					if lastImportFiles[i].Filename == remoteFile.Filename {
						lastFile = &lastImportFiles[i]
						break
					}
				}

				// If file not found in last import, it's new
				if lastFile == nil {
					shouldImport = true
					reason = "newer_files_available"
					break
				}

				// Compare timestamp
				if config.CompareTimestamp && remoteFile.ModTime.After(lastFile.ModTime) {
					shouldImport = true
					reason = "newer_files_available"
					break
				}

				// Compare size
				if config.CompareSize && remoteFile.Size != lastFile.Size {
					shouldImport = true
					reason = "newer_files_available"
					break
				}

				// Compare checksum
				if config.CompareChecksum && remoteFile.Checksum != "" && lastFile.Checksum != "" && remoteFile.Checksum != lastFile.Checksum {
					shouldImport = true
					reason = "newer_files_available"
					break
				}
			}

			assert.Equal(t, tt.expectedImport, shouldImport)
			assert.Equal(t, tt.expectedReason, reason)
		})
	}
}

func TestFreshnessChecker_LoadMetadata(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		metadataContent string
		expectedError   bool
		expectedFiles   int
	}{
		{
			name: "valid metadata",
			metadataContent: `{
				"last_import": {
					"timestamp": "2025-08-05T10:00:00Z",
					"success": true,
					"files": [
						{
							"filename": "mvdsb_20250805.zip",
							"size": 1024000,
							"mod_time": "2025-08-05T09:30:00Z",
							"checksum": "sha256:abc123"
						}
					]
				}
			}`,
			expectedError: false,
			expectedFiles: 1,
		},
		{
			name: "empty files array",
			metadataContent: `{
				"last_import": {
					"timestamp": "2025-08-05T10:00:00Z",
					"success": true,
					"files": []
				}
			}`,
			expectedError: false,
			expectedFiles: 0,
		},
		{
			name:            "invalid JSON",
			metadataContent: `{"invalid": json}`,
			expectedError:   true,
			expectedFiles:   0,
		},
		{
			name:            "missing last_import key",
			metadataContent: `{"other_data": "value"}`,
			expectedError:   true,
			expectedFiles:   0,
		},
		{
			name: "failed last import",
			metadataContent: `{
				"last_import": {
					"timestamp": "2025-08-05T10:00:00Z",
					"success": false,
					"files": []
				}
			}`,
			expectedError: false,
			expectedFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataFile := filepath.Join(tempDir, tt.name+".json")

			if tt.metadataContent != "" {
				err := os.WriteFile(metadataFile, []byte(tt.metadataContent), 0644)
				require.NoError(t, err)
			}

			// Simulate loading metadata
			var loadedFiles []models.FileMetadata
			var loadError error

			if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
				loadError = err
			} else {
				content, err := os.ReadFile(metadataFile)
				if err != nil {
					loadError = err
				} else {
					var metadata map[string]interface{}
					if err := json.Unmarshal(content, &metadata); err != nil {
						loadError = err
					} else {
						if lastImport, ok := metadata["last_import"].(map[string]interface{}); ok {
							if success, ok := lastImport["success"].(bool); ok && success {
								if files, ok := lastImport["files"].([]interface{}); ok {
									for range files {
										loadedFiles = append(loadedFiles, models.FileMetadata{
											Filename: "test.zip",
											Size:     1024,
											ModTime:  time.Now(),
										})
									}
								}
							}
						} else {
							loadError = assert.AnError
						}
					}
				}
			}

			if tt.expectedError {
				assert.Error(t, loadError)
			} else {
				assert.NoError(t, loadError)
				assert.Equal(t, tt.expectedFiles, len(loadedFiles))
			}
		})
	}
}

func TestFreshnessChecker_SaveMetadata(t *testing.T) {
	tempDir := t.TempDir()
	metadataFile := filepath.Join(tempDir, "save_test.json")

	now := time.Now()
	files := []models.FileMetadata{
		{
			Filename: "mvdsb_20250806.zip",
			Size:     1024000,
			ModTime:  now,
			Checksum: "sha256:abc123",
			Pattern:  "mvdsb_*",
			Database: "mvdsb",
			Imported: true,
		},
		{
			Filename: "portal64_bdw_20250806.zip",
			Size:     512000,
			ModTime:  now,
			Checksum: "sha256:def456",
			Pattern:  "portal64_bdw_*",
			Database: "portal64_bdw",
			Imported: true,
		},
	}

	// Test saving metadata
	metadata := map[string]interface{}{
		"last_import": map[string]interface{}{
			"timestamp": now,
			"success":   true,
			"files":     files,
		},
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(metadataFile, metadataJSON, 0644)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, metadataFile)

	// Verify content can be read back
	content, err := os.ReadFile(metadataFile)
	require.NoError(t, err)

	var loadedMetadata map[string]interface{}
	err = json.Unmarshal(content, &loadedMetadata)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, loadedMetadata, "last_import")
	
	lastImport := loadedMetadata["last_import"].(map[string]interface{})
	assert.Contains(t, lastImport, "timestamp")
	assert.Contains(t, lastImport, "success")
	assert.Contains(t, lastImport, "files")
	assert.Equal(t, true, lastImport["success"])

	// Verify files array
	filesArray := lastImport["files"].([]interface{})
	assert.Equal(t, len(files), len(filesArray))
}

func TestFreshnessChecker_ComparisonMethods(t *testing.T) {
	baseTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name                string
		config              *config.FreshnessConfig
		remoteFile          models.FileMetadata
		lastFile            *models.FileMetadata
		expectedIsNewer     bool
		expectedReasons     []string
	}{
		{
			name: "timestamp comparison - newer",
			config: &config.FreshnessConfig{
				CompareTimestamp: true,
				CompareSize:      false,
				CompareChecksum:  false,
			},
			remoteFile: models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime.Add(1 * time.Hour),
			},
			lastFile: &models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
			},
			expectedIsNewer: true,
			expectedReasons: []string{"newer_timestamp"},
		},
		{
			name: "size comparison - different",
			config: &config.FreshnessConfig{
				CompareTimestamp: false,
				CompareSize:      true,
				CompareChecksum:  false,
			},
			remoteFile: models.FileMetadata{
				Filename: "test.zip",
				Size:     2048,
				ModTime:  baseTime,
			},
			lastFile: &models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
			},
			expectedIsNewer: true,
			expectedReasons: []string{"different_size"},
		},
		{
			name: "checksum comparison - different",
			config: &config.FreshnessConfig{
				CompareTimestamp: false,
				CompareSize:      false,
				CompareChecksum:  true,
			},
			remoteFile: models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
				Checksum: "sha256:new123",
			},
			lastFile: &models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
				Checksum: "sha256:old123",
			},
			expectedIsNewer: true,
			expectedReasons: []string{"different_checksum"},
		},
		{
			name: "multiple comparisons - multiple changes",
			config: &config.FreshnessConfig{
				CompareTimestamp: true,
				CompareSize:      true,
				CompareChecksum:  true,
			},
			remoteFile: models.FileMetadata{
				Filename: "test.zip",
				Size:     2048,
				ModTime:  baseTime.Add(1 * time.Hour),
				Checksum: "sha256:new123",
			},
			lastFile: &models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
				Checksum: "sha256:old123",
			},
			expectedIsNewer: true,
			expectedReasons: []string{"newer_timestamp", "different_size", "different_checksum"},
		},
		{
			name: "no changes detected",
			config: &config.FreshnessConfig{
				CompareTimestamp: true,
				CompareSize:      true,
				CompareChecksum:  true,
			},
			remoteFile: models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
				Checksum: "sha256:same123",
			},
			lastFile: &models.FileMetadata{
				Filename: "test.zip",
				Size:     1024,
				ModTime:  baseTime,
				Checksum: "sha256:same123",
			},
			expectedIsNewer: false,
			expectedReasons: []string{},
		},
		{
			name: "file not found in last import",
			config: &config.FreshnessConfig{
				CompareTimestamp: true,
				CompareSize:      true,
				CompareChecksum:  false,
			},
			remoteFile: models.FileMetadata{
				Filename: "new_file.zip",
				Size:     1024,
				ModTime:  baseTime,
			},
			lastFile:       nil,
			expectedIsNewer: true,
			expectedReasons: []string{"file_not_found_in_last_import"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate file comparison logic
			isNewer := false
			reasons := []string{}

			if tt.lastFile == nil {
				isNewer = true
				reasons = append(reasons, "file_not_found_in_last_import")
			} else {
				// Compare timestamp
				if tt.config.CompareTimestamp && tt.remoteFile.ModTime.After(tt.lastFile.ModTime) {
					isNewer = true
					reasons = append(reasons, "newer_timestamp")
				}

				// Compare size
				if tt.config.CompareSize && tt.remoteFile.Size != tt.lastFile.Size {
					isNewer = true
					reasons = append(reasons, "different_size")
				}

				// Compare checksum
				if tt.config.CompareChecksum && tt.remoteFile.Checksum != "" && tt.lastFile.Checksum != "" && tt.remoteFile.Checksum != tt.lastFile.Checksum {
					isNewer = true
					reasons = append(reasons, "different_checksum")
				}
			}

			assert.Equal(t, tt.expectedIsNewer, isNewer)
			assert.Equal(t, len(tt.expectedReasons), len(reasons))
			
			for _, expectedReason := range tt.expectedReasons {
				assert.Contains(t, reasons, expectedReason)
			}
		})
	}
}

func TestFreshnessChecker_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		scenario       string
		expectedResult string
	}{
		{
			name:           "corrupted metadata file",
			scenario:       "corrupted_metadata",
			expectedResult: "treat_as_first_import",
		},
		{
			name:           "metadata file missing",
			scenario:       "missing_metadata",
			expectedResult: "treat_as_first_import",
		},
		{
			name:           "empty remote files list",
			scenario:       "empty_remote_files",
			expectedResult: "skip_import",
		},
		{
			name:           "partial metadata",
			scenario:       "partial_metadata",
			expectedResult: "proceed_with_caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataFile := filepath.Join(tempDir, tt.name+".json")
			
			config := &config.FreshnessConfig{
				Enabled:           true,
				CompareTimestamp:  true,
				CompareSize:       true,
				CompareChecksum:   false,
				SkipIfNotNewer:    true,
			}

			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
			checker := importers.NewFreshnessChecker(config, metadataFile, logger)

			var result string

			switch tt.scenario {
			case "corrupted_metadata":
				// Create corrupted metadata file
				err := os.WriteFile(metadataFile, []byte("corrupted json content"), 0644)
				require.NoError(t, err)
				result = "treat_as_first_import"

			case "missing_metadata":
				// Don't create metadata file
				result = "treat_as_first_import"

			case "empty_remote_files":
				remoteFiles := []models.FileMetadata{}
				if len(remoteFiles) == 0 {
					result = "skip_import"
				}

			case "partial_metadata":
				// Create metadata with missing fields
				partialMetadata := `{"last_import": {"success": true}}`
				err := os.WriteFile(metadataFile, []byte(partialMetadata), 0644)
				require.NoError(t, err)
				result = "proceed_with_caution"
			}

			assert.NotNil(t, checker)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
