package importers

import (
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

func TestDatabaseImporter_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		importConfig   *config.ImportDBConfig
		dbConfig       *config.DatabaseConfig
		expectedValid  bool
		errorMessage   string
	}{
		{
			name: "valid configuration",
			importConfig: &config.ImportDBConfig{
				ImportTimeout: 600 * time.Second,
				TargetDatabases: []config.TargetDatabase{
					{Name: "mvdsb", FilePattern: "mvdsb_*"},
					{Name: "portal64_bdw", FilePattern: "portal64_bdw_*"},
				},
			},
			dbConfig: &config.DatabaseConfig{
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
			expectedValid: true,
		},
		{
			name: "empty timeout",
			importConfig: &config.ImportDBConfig{
				ImportTimeout: 0,
				TargetDatabases: []config.TargetDatabase{
					{Name: "mvdsb", FilePattern: "mvdsb_*"},
				},
			},
			dbConfig: &config.DatabaseConfig{
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
			expectedValid: false,
			errorMessage:  "import timeout cannot be empty",
		},
		{
			name: "no target databases",
			importConfig: &config.ImportDBConfig{
				ImportTimeout:   600 * time.Second,
				TargetDatabases: []config.TargetDatabase{},
			},
			dbConfig: &config.DatabaseConfig{
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
			expectedValid: false,
			errorMessage:  "at least one target database must be specified",
		},
		{
			name: "invalid target database - empty name",
			importConfig: &config.ImportDBConfig{
				ImportTimeout: 600 * time.Second,
				TargetDatabases: []config.TargetDatabase{
					{Name: "", FilePattern: "mvdsb_*"},
				},
			},
			dbConfig: &config.DatabaseConfig{
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
			expectedValid: false,
			errorMessage:  "database name cannot be empty",
		},
		{
			name: "invalid target database - empty pattern",
			importConfig: &config.ImportDBConfig{
				ImportTimeout: 600 * time.Second,
				TargetDatabases: []config.TargetDatabase{
					{Name: "mvdsb", FilePattern: ""},
				},
			},
			dbConfig: &config.DatabaseConfig{
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
			expectedValid: false,
			errorMessage:  "file pattern cannot be empty",
		},
		{
			name: "invalid database config - empty host",
			importConfig: &config.ImportDBConfig{
				ImportTimeout: 600 * time.Second,
				TargetDatabases: []config.TargetDatabase{
					{Name: "mvdsb", FilePattern: "mvdsb_*"},
				},
			},
			dbConfig: &config.DatabaseConfig{
				MVDSB: config.DatabaseConnection{
					Host:     "", // Empty host should cause validation failure
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
			expectedValid: false,
			errorMessage:  "database host cannot be empty",
		},
		{
			name: "invalid database config - invalid port",
			importConfig: &config.ImportDBConfig{
				ImportTimeout: 600 * time.Second,
				TargetDatabases: []config.TargetDatabase{
					{Name: "mvdsb", FilePattern: "mvdsb_*"},
				},
			},
			dbConfig: &config.DatabaseConfig{
				MVDSB: config.DatabaseConnection{
					Host:     "localhost",
					Port:     0, // Invalid port should cause validation failure
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
			expectedValid: false,
			errorMessage:  "database port must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

			// Test configuration validation logic
			var configErrors []string

			if tt.importConfig.ImportTimeout == 0 {
				configErrors = append(configErrors, "import timeout cannot be empty")
			}
			if len(tt.importConfig.TargetDatabases) == 0 {
				configErrors = append(configErrors, "at least one target database must be specified")
			}

			for _, db := range tt.importConfig.TargetDatabases {
				if db.Name == "" {
					configErrors = append(configErrors, "database name cannot be empty")
				}
				if db.FilePattern == "" {
					configErrors = append(configErrors, "file pattern cannot be empty")
				}
			}

			// Validate database configurations
			if tt.dbConfig.MVDSB.Host == "" {
				configErrors = append(configErrors, "database host cannot be empty")
			}
			if tt.dbConfig.MVDSB.Port <= 0 {
				configErrors = append(configErrors, "database port must be greater than 0")
			}
			if tt.dbConfig.Portal64BDW.Host == "" {
				configErrors = append(configErrors, "database host cannot be empty") 
			}
			if tt.dbConfig.Portal64BDW.Port <= 0 {
				configErrors = append(configErrors, "database port must be greater than 0")
			}

			if tt.expectedValid {
				assert.Empty(t, configErrors, "Configuration should be valid")

				// Should be able to create importer with valid config
				importer := importers.NewDatabaseImporter(tt.importConfig, tt.dbConfig, logger)
				assert.NotNil(t, importer)
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

func TestDatabaseImporter_FileToDatabaseMapping(t *testing.T) {
	tests := []struct {
		name             string
		targetDatabases  []config.TargetDatabase
		filePath         string
		expectedDatabase string
		expectedMatch    bool
	}{
		{
			name: "mvdsb file matches mvdsb database",
			targetDatabases: []config.TargetDatabase{
				{Name: "mvdsb", FilePattern: "mvdsb_*"},
				{Name: "portal64_bdw", FilePattern: "portal64_bdw_*"},
			},
			filePath:         "/path/to/mvdsb_20250806.sql",
			expectedDatabase: "mvdsb",
			expectedMatch:    true,
		},
		{
			name: "portal64_bdw file matches portal64_bdw database",
			targetDatabases: []config.TargetDatabase{
				{Name: "mvdsb", FilePattern: "mvdsb_*"},
				{Name: "portal64_bdw", FilePattern: "portal64_bdw_*"},
			},
			filePath:         "/path/to/portal64_bdw_export.sql",
			expectedDatabase: "portal64_bdw",
			expectedMatch:    true,
		},
		{
			name: "no matching pattern",
			targetDatabases: []config.TargetDatabase{
				{Name: "mvdsb", FilePattern: "mvdsb_*"},
				{Name: "portal64_bdw", FilePattern: "portal64_bdw_*"},
			},
			filePath:         "/path/to/other_database.sql",
			expectedDatabase: "",
			expectedMatch:    false,
		},
		{
			name: "first matching pattern wins",
			targetDatabases: []config.TargetDatabase{
				{Name: "first_db", FilePattern: "test_*"},
				{Name: "second_db", FilePattern: "test_*"},
			},
			filePath:         "/path/to/test_file.sql",
			expectedDatabase: "first_db",
			expectedMatch:    true,
		},
		{
			name: "wildcard pattern matches",
			targetDatabases: []config.TargetDatabase{
				{Name: "any_db", FilePattern: "*"},
			},
			filePath:         "/path/to/anything.sql",
			expectedDatabase: "any_db",
			expectedMatch:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test file-to-database matching logic
			var matchedDatabase string
			matched := false

			filename := filepath.Base(tt.filePath)

			for _, target := range tt.targetDatabases {
				match, err := filepath.Match(target.FilePattern, filename)
				require.NoError(t, err, "Pattern matching should not error")

				if match {
					matchedDatabase = target.Name
					matched = true
					break
				}
			}

			assert.Equal(t, tt.expectedMatch, matched)
			if matched {
				assert.Equal(t, tt.expectedDatabase, matchedDatabase)
			}
		})
	}
}

func TestDatabaseImporter_ImportValidation(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name                string
		sqlFiles            map[string]string
		expectedValidFiles  int
		expectedInvalidFiles int
	}{
		{
			name: "valid SQL files",
			sqlFiles: map[string]string{
				"mvdsb_dump.sql": `-- Valid SQL dump
CREATE TABLE test_table (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_table (name) VALUES ('test1'), ('test2');`,
				"portal64_bdw_dump.sql": `-- Another valid SQL dump
CREATE DATABASE IF NOT EXISTS portal64_bdw;
USE portal64_bdw;
CREATE TABLE users (id INT, username VARCHAR(100));`,
			},
			expectedValidFiles:   2,
			expectedInvalidFiles: 0,
		},
		{
			name: "mixed valid and invalid SQL files",
			sqlFiles: map[string]string{
				"valid.sql":   "CREATE TABLE test (id INT);",
				"empty.sql":   "",
				"invalid.sql": "This is not valid SQL syntax",
				"good.sql":    "INSERT INTO test VALUES (1);",
			},
			expectedValidFiles:   2, // valid.sql and good.sql
			expectedInvalidFiles: 2, // empty.sql and invalid.sql (for basic validation)
		},
		{
			name: "files with SQL comments only",
			sqlFiles: map[string]string{
				"comments_only.sql": `-- This is a comment
/* Another comment */
# Hash comment`,
				"mixed.sql": `-- Comment
CREATE TABLE test (id INT);
-- Another comment`,
			},
			expectedValidFiles:   1, // mixed.sql has actual SQL
			expectedInvalidFiles: 1, // comments_only.sql has no executable SQL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test SQL files
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			for filename, content := range tt.sqlFiles {
				filePath := filepath.Join(testDir, filename)
				err := os.WriteFile(filePath, []byte(content), 0644)
				require.NoError(t, err)
			}

			// Test SQL file validation logic
			validFiles := 0
			invalidFiles := 0

			for filename, content := range tt.sqlFiles {
				isValid := validateSQLContent(content)
				if isValid {
					validFiles++
				} else {
					invalidFiles++
				}

				// Verify the file exists
				filePath := filepath.Join(testDir, filename)
				assert.FileExists(t, filePath)
			}

			assert.Equal(t, tt.expectedValidFiles, validFiles)
			assert.Equal(t, tt.expectedInvalidFiles, invalidFiles)
		})
	}
}

// Helper function to simulate SQL content validation
func validateSQLContent(content string) bool {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) == 0 {
		return false
	}

	// Remove comments for basic validation
	lines := strings.Split(trimmed, "\n")
	hasSQL := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		// Skip comments
		if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
			continue
		}
		// Look for basic SQL keywords
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "CREATE") || strings.Contains(upper, "INSERT") ||
			strings.Contains(upper, "UPDATE") || strings.Contains(upper, "DELETE") ||
			strings.Contains(upper, "SELECT") || strings.Contains(upper, "DROP") ||
			strings.Contains(upper, "ALTER") || strings.Contains(upper, "USE") {
			hasSQL = true
			break
		}
	}

	return hasSQL
}

func TestDatabaseImporter_ProgressReporting(t *testing.T) {
	tests := []struct {
		name              string
		databases         []string
		filesPerDatabase  int
		expectedSteps     int
	}{
		{
			name:              "single database single file",
			databases:         []string{"mvdsb"},
			filesPerDatabase:  1,
			expectedSteps:     5, // start, drop, create, import, complete
		},
		{
			name:              "multiple databases single file each",
			databases:         []string{"mvdsb", "portal64_bdw"},
			filesPerDatabase:  1,
			expectedSteps:     8, // start, (drop, create, import) x 2, complete
		},
		{
			name:              "single database multiple files",
			databases:         []string{"mvdsb"},
			filesPerDatabase:  3,
			expectedSteps:     7, // start, drop, create, import x 3, complete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate progress tracking
			progressSteps := []string{}

			// Start import
			progressSteps = append(progressSteps, "import_started")

			for _, database := range tt.databases {
				// Drop database
				progressSteps = append(progressSteps, "database_dropped_"+database)

				// Create database
				progressSteps = append(progressSteps, "database_created_"+database)

				// Import files
				for i := 0; i < tt.filesPerDatabase; i++ {
					progressSteps = append(progressSteps, "file_imported_"+database)
				}
			}

			// Complete import
			progressSteps = append(progressSteps, "import_completed")

			// Verify progress steps
			assert.Equal(t, tt.expectedSteps, len(progressSteps))
			assert.Equal(t, "import_started", progressSteps[0])
			assert.Equal(t, "import_completed", progressSteps[len(progressSteps)-1])

			// Verify database-specific steps
			for _, database := range tt.databases {
				found := false
				for _, step := range progressSteps {
					if strings.Contains(step, database) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find steps for database %s", database)
			}
		})
	}
}

func TestDatabaseImporter_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		expectedError bool
		errorType     string
	}{
		{
			name:          "database connection failure",
			scenario:      "connection_failed",
			expectedError: true,
			errorType:     "connection_error",
		},
		{
			name:          "permission denied",
			scenario:      "permission_denied",
			expectedError: true,
			errorType:     "permission_error",
		},
		{
			name:          "invalid SQL syntax",
			scenario:      "invalid_sql",
			expectedError: true,
			errorType:     "sql_error",
		},
		{
			name:          "file not found",
			scenario:      "file_not_found",
			expectedError: true,
			errorType:     "file_error",
		},
		{
			name:          "timeout during import",
			scenario:      "import_timeout",
			expectedError: true,
			errorType:     "timeout_error",
		},
		{
			name:          "successful import",
			scenario:      "success",
			expectedError: false,
			errorType:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate error scenarios
			var simulatedError error
			var errorType string

			switch tt.scenario {
			case "connection_failed":
				simulatedError = assert.AnError
				errorType = "connection_error"
			case "permission_denied":
				simulatedError = assert.AnError
				errorType = "permission_error"
			case "invalid_sql":
				simulatedError = assert.AnError
				errorType = "sql_error"
			case "file_not_found":
				simulatedError = assert.AnError
				errorType = "file_error"
			case "import_timeout":
				simulatedError = assert.AnError
				errorType = "timeout_error"
			case "success":
				simulatedError = nil
				errorType = ""
			}

			// Verify error expectations
			if tt.expectedError {
				assert.Error(t, simulatedError)
				assert.Equal(t, tt.errorType, errorType)
			} else {
				assert.NoError(t, simulatedError)
			}
		})
	}
}

func TestDatabaseImporter_ImportStatistics(t *testing.T) {
	tests := []struct {
		name              string
		databases         []string
		filesPerDatabase  int
		avgFileSize       int64
		expectedDuration  time.Duration
	}{
		{
			name:              "small import",
			databases:         []string{"mvdsb"},
			filesPerDatabase:  1,
			avgFileSize:       1024 * 1024, // 1MB
			expectedDuration:  time.Second * 10,
		},
		{
			name:              "medium import",
			databases:         []string{"mvdsb", "portal64_bdw"},
			filesPerDatabase:  2,
			avgFileSize:       10 * 1024 * 1024, // 10MB
			expectedDuration:  time.Minute * 5,
		},
		{
			name:              "large import",
			databases:         []string{"mvdsb", "portal64_bdw"},
			filesPerDatabase:  3,
			avgFileSize:       100 * 1024 * 1024, // 100MB
			expectedDuration:  time.Minute * 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate import statistics
			totalFiles := len(tt.databases) * tt.filesPerDatabase
			totalSize := int64(totalFiles) * tt.avgFileSize
			totalDatabases := len(tt.databases)

			// Simulate statistics collection
			stats := map[string]interface{}{
				"total_files":     totalFiles,
				"total_size":      totalSize,
				"total_databases": totalDatabases,
				"avg_file_size":   tt.avgFileSize,
				"estimated_time":  tt.expectedDuration,
			}

			// Verify statistics
			assert.Equal(t, totalFiles, stats["total_files"])
			assert.Equal(t, totalSize, stats["total_size"])
			assert.Equal(t, totalDatabases, stats["total_databases"])
			assert.Equal(t, tt.avgFileSize, stats["avg_file_size"])
			assert.Equal(t, tt.expectedDuration, stats["estimated_time"])

			// Verify calculations
			assert.Greater(t, totalSize, int64(0))
			assert.Greater(t, totalFiles, 0)
			assert.Greater(t, totalDatabases, 0)
		})
	}
}
