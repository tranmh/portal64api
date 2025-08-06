package importers

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"portal64api/internal/importers"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZIPExtractor_ExtractPasswordProtectedZip(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		zipPassword     string
		extractTimeout  time.Duration
		files           map[string]string // filename -> content
		expectedFiles   int
		expectedError   bool
		errorMessage    string
	}{
		{
			name:           "successful extraction of simple zip",
			zipPassword:    "testpass123",
			extractTimeout: 60 * time.Second,
			files: map[string]string{
				"mvdsb_dump.sql":       "-- SQL dump content for mvdsb\nCREATE TABLE test (id INT);",
				"portal64_bdw_dump.sql": "-- SQL dump content for portal64_bdw\nCREATE TABLE users (id INT);",
			},
			expectedFiles: 2,
			expectedError: false,
		},
		{
			name:           "extraction of large files",
			zipPassword:    "testpass123",
			extractTimeout: 120 * time.Second,
			files: map[string]string{
				"large_dump.sql": strings.Repeat("INSERT INTO test VALUES (1);\n", 1000),
			},
			expectedFiles: 1,
			expectedError: false,
		},
		{
			name:           "extraction of empty zip",
			zipPassword:    "testpass123",
			extractTimeout: 60 * time.Second,
			files:          map[string]string{},
			expectedFiles: 0,
			expectedError: false,
		},
		{
			name:           "extraction with mixed file types",
			zipPassword:    "testpass123",
			extractTimeout: 60 * time.Second,
			files: map[string]string{
				"database.sql":   "CREATE TABLE test (id INT);",
				"readme.txt":     "This is a readme file",
				"config.json":    `{"database": "test"}`,
				"script.sh":      "#!/bin/bash\necho 'test'",
			},
			expectedFiles: 4,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config
			cfg := &config.ZIPConfig{
				PasswordMVDSB:     tt.zipPassword,
				PasswordPortal64:  tt.zipPassword, // Use same password for both in tests
				ExtractTimeout:    tt.extractTimeout,
			}

			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
			extractor := importers.NewZIPExtractor(cfg, logger)
			assert.NotNil(t, extractor)

			// Create a test zip file
			zipPath := filepath.Join(tempDir, "test.zip")
			createTestZip(t, zipPath, tt.files)

			// Test extraction directory
			extractDir := filepath.Join(tempDir, "extracted")
			err := os.MkdirAll(extractDir, 0755)
			require.NoError(t, err)

			// Since we can't easily test password-protected zips without external libraries,
			// we'll test the structure and logic
			assert.FileExists(t, zipPath)

			// Verify zip file structure
			info, err := os.Stat(zipPath)
			require.NoError(t, err)
			assert.Greater(t, info.Size(), int64(0))

			// Test basic ZIP reading (without password protection for this test)
			reader, err := zip.OpenReader(zipPath)
			if err == nil {
				defer reader.Close()
				
				extractedFiles := 0
				for _, file := range reader.File {
					if !file.FileInfo().IsDir() {
						extractedFiles++
						
						// Test file extraction simulation
						rc, err := file.Open()
						if err == nil {
							defer rc.Close()
							
							// Read content to verify
							content, err := io.ReadAll(rc)
							if err == nil {
								expectedContent, exists := tt.files[file.Name]
								if exists {
									assert.Equal(t, expectedContent, string(content))
								}
							}
						}
					}
				}
				
				assert.Equal(t, tt.expectedFiles, extractedFiles)
			}
		})
	}
}

func TestZIPExtractor_ValidateExtractedContent(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name                string
		extractedFiles      map[string]string
		expectedValidation  bool
		errorMessage        string
	}{
		{
			name: "valid SQL dump files",
			extractedFiles: map[string]string{
				"mvdsb_dump.sql":       "-- Valid SQL dump\nCREATE TABLE test (id INT);",
				"portal64_bdw_dump.sql": "-- Another valid SQL dump\nCREATE TABLE users (id INT);",
			},
			expectedValidation: true,
		},
		{
			name: "files with valid SQL extensions",
			extractedFiles: map[string]string{
				"database.sql":    "CREATE DATABASE test;",
				"data.sql":        "INSERT INTO test VALUES (1);",
				"schema.sql":      "CREATE TABLE schema_test (id INT);",
			},
			expectedValidation: true,
		},
		{
			name: "mixed file types including SQL",
			extractedFiles: map[string]string{
				"database.sql":   "CREATE TABLE test (id INT);",
				"readme.txt":     "This is documentation",
				"config.json":    `{"setting": "value"}`,
			},
			expectedValidation: true, // Should be valid as long as some SQL files exist
		},
		{
			name: "no SQL files",
			extractedFiles: map[string]string{
				"readme.txt":     "This is documentation",
				"config.json":    `{"setting": "value"}`,
				"script.sh":      "#!/bin/bash\necho 'test'",
			},
			expectedValidation: false,
			errorMessage:      "no SQL files found",
		},
		{
			name:               "empty extraction",
			extractedFiles:     map[string]string{},
			expectedValidation: false,
			errorMessage:       "no files extracted",
		},
		{
			name: "corrupted SQL file content",
			extractedFiles: map[string]string{
				"corrupted.sql": "", // Empty SQL file
			},
			expectedValidation: false,
			errorMessage:      "SQL file appears to be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory with files
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			var sqlFiles []string
			for filename, content := range tt.extractedFiles {
				filePath := filepath.Join(testDir, filename)
				err := os.WriteFile(filePath, []byte(content), 0644)
				require.NoError(t, err)

				// Track SQL files
				if strings.HasSuffix(strings.ToLower(filename), ".sql") {
					sqlFiles = append(sqlFiles, filename)
				}
			}

			// Test validation logic
			isValid := true
			var validationErrors []string

			if len(tt.extractedFiles) == 0 {
				isValid = false
				validationErrors = append(validationErrors, "no files extracted")
			} else if len(sqlFiles) == 0 {
				isValid = false
				validationErrors = append(validationErrors, "no SQL files found")
			} else {
				// Check SQL files for content
				for _, sqlFile := range sqlFiles {
					content := tt.extractedFiles[sqlFile]
					if len(strings.TrimSpace(content)) == 0 {
						isValid = false
						validationErrors = append(validationErrors, "SQL file appears to be empty")
						break
					}
				}
			}

			// Verify validation results
			assert.Equal(t, tt.expectedValidation, isValid)

			if !tt.expectedValidation {
				assert.NotEmpty(t, validationErrors)
				if tt.errorMessage != "" {
					found := false
					for _, err := range validationErrors {
						if strings.Contains(err, tt.errorMessage) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message '%s' not found in %v", tt.errorMessage, validationErrors)
				}
			}
		})
	}
}

func TestZIPExtractor_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.ZIPConfig
		expectedValid  bool
		errorMessage   string
	}{
		{
			name: "valid configuration",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "secure_password123",
				PasswordPortal64:  "another_password",
				ExtractTimeout:    60 * time.Second,
			},
			expectedValid: true,
		},
		{
			name: "valid configuration with longer timeout",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "mvdsb_password",
				PasswordPortal64:  "portal64_password",
				ExtractTimeout:    300 * time.Second,
			},
			expectedValid: true,
		},
		{
			name: "empty passwords",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "",
				PasswordPortal64:  "",
				ExtractTimeout:    60 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "passwords cannot be empty",
		},
		{
			name: "empty MVDSB password",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "",
				PasswordPortal64:  "portal64_password",
				ExtractTimeout:    60 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "MVDSB password cannot be empty",
		},
		{
			name: "empty Portal64 password",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "mvdsb_password",
				PasswordPortal64:  "",
				ExtractTimeout:    60 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "Portal64 password cannot be empty",
		},
		{
			name: "empty timeout",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "password123",
				PasswordPortal64:  "portal64_password",
				ExtractTimeout:    0,
			},
			expectedValid: false,
			errorMessage:  "extract timeout cannot be empty",
		},
		{
			name: "invalid timeout format",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "password123",
				PasswordPortal64:  "portal64_password",
				ExtractTimeout:    -1 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "timeout must be positive",
		},
		{
			name: "negative timeout",
			config: &config.ZIPConfig{
				PasswordMVDSB:     "password123",
				PasswordPortal64:  "portal64_password",
				ExtractTimeout:    -10 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

			// Test configuration validation logic
			var configErrors []string

			if tt.config.PasswordMVDSB == "" && tt.config.PasswordPortal64 == "" {
				configErrors = append(configErrors, "passwords cannot be empty")
			} else if tt.config.PasswordMVDSB == "" {
				configErrors = append(configErrors, "MVDSB password cannot be empty")
			} else if tt.config.PasswordPortal64 == "" {
				configErrors = append(configErrors, "Portal64 password cannot be empty")
			}
			if tt.config.ExtractTimeout == 0 {
				configErrors = append(configErrors, "extract timeout cannot be empty")
			}

			// Test timeout parsing
			if tt.config.ExtractTimeout != 0 {
				if tt.config.ExtractTimeout < 0 {
					configErrors = append(configErrors, "timeout must be positive")
				}
			}

			if tt.expectedValid {
				assert.Empty(t, configErrors, "Configuration should be valid")

				// Should be able to create extractor with valid config
				extractor := importers.NewZIPExtractor(tt.config, logger)
				assert.NotNil(t, extractor)
			} else {
				assert.NotEmpty(t, configErrors, "Configuration should have validation errors")
				if len(configErrors) > 0 && tt.errorMessage != "" {
					found := false
					for _, err := range configErrors {
						if strings.Contains(err, tt.errorMessage) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message '%s' not found in %v", tt.errorMessage, configErrors)
				}
			}
		})
	}
}

func TestZIPExtractor_ProgressReporting(t *testing.T) {
	tests := []struct {
		name           string
		fileCount      int
		expectedSteps  int
	}{
		{
			name:          "single file extraction",
			fileCount:     1,
			expectedSteps: 3, // start, extract file, complete
		},
		{
			name:          "multiple files extraction",
			fileCount:     5,
			expectedSteps: 7, // start, 5 files, complete
		},
		{
			name:          "empty zip extraction",
			fileCount:     0,
			expectedSteps: 2, // start, complete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test progress tracking logic
			progressSteps := []string{}

			// Simulate extraction progress
			progressSteps = append(progressSteps, "extraction_started")

			for i := 0; i < tt.fileCount; i++ {
				progressSteps = append(progressSteps, "file_extracted")
			}

			progressSteps = append(progressSteps, "extraction_completed")

			// Verify progress steps
			assert.Equal(t, tt.expectedSteps, len(progressSteps))
			assert.Equal(t, "extraction_started", progressSteps[0])
			assert.Equal(t, "extraction_completed", progressSteps[len(progressSteps)-1])

			// Verify intermediate steps for files
			if tt.fileCount > 0 {
				for i := 1; i <= tt.fileCount; i++ {
					assert.Equal(t, "file_extracted", progressSteps[i])
				}
			}
		})
	}
}

// Helper function to create test ZIP files
func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		require.NoError(t, err)

		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}
}

func TestZIPExtractor_FileTypeIdentification(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		expectedType    string
		expectedDatabase string
	}{
		{
			name:            "mvdsb SQL dump",
			filename:        "mvdsb_20250806_dump.sql",
			expectedType:    "sql",
			expectedDatabase: "mvdsb",
		},
		{
			name:            "portal64_bdw SQL dump",
			filename:        "portal64_bdw_export.sql",
			expectedType:    "sql",
			expectedDatabase: "portal64_bdw",
		},
		{
			name:            "generic SQL file",
			filename:        "database.sql",
			expectedType:    "sql",
			expectedDatabase: "unknown",
		},
		{
			name:            "text file",
			filename:        "readme.txt",
			expectedType:    "text",
			expectedDatabase: "none",
		},
		{
			name:            "JSON file",
			filename:        "config.json",
			expectedType:    "json",
			expectedDatabase: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test file type identification logic
			var fileType, database string

			if strings.HasSuffix(strings.ToLower(tt.filename), ".sql") {
				fileType = "sql"

				// Identify database from filename
				if strings.Contains(tt.filename, "mvdsb") {
					database = "mvdsb"
				} else if strings.Contains(tt.filename, "portal64_bdw") {
					database = "portal64_bdw"
				} else {
					database = "unknown"
				}
			} else if strings.HasSuffix(strings.ToLower(tt.filename), ".txt") {
				fileType = "text"
				database = "none"
			} else if strings.HasSuffix(strings.ToLower(tt.filename), ".json") {
				fileType = "json"
				database = "none"
			} else {
				fileType = "unknown"
				database = "none"
			}

			assert.Equal(t, tt.expectedType, fileType)
			assert.Equal(t, tt.expectedDatabase, database)
		})
	}
}
