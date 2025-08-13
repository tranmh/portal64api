// Test configuration for SCP Import Feature tests
package testconfig

import (
	"path/filepath"
	"portal64api/internal/config"
	"time"
)

// GetTestImportConfig returns a configuration suitable for testing
func GetTestImportConfig(tempDir string) *config.ImportConfig {
	return &config.ImportConfig{
		Enabled:  true,
		Schedule: "0 2 * * *",
		SCP: config.SCPConfig{
			Host:         "localhost", // Use localhost for testing
			Port:         22,
			Username:     "testuser",
			Password:     "testpass123",
			RemotePath:   "/test/data/exports/",
			FilePatterns: []string{"test_mvdsb_*.zip", "test_portal64_bdw_*.zip"},
			Timeout:      30 * time.Second,
		},
		ZIP: config.ZIPConfig{
			PasswordMVDSB:     "testzip123",
			PasswordPortal64:  "testzip123",
			ExtractTimeout:    30 * time.Second,
		},
		Database: config.ImportDBConfig{
			ImportTimeout: 600 * time.Second,
			TargetDatabases: []config.TargetDatabase{
				{Name: "test_mvdsb", FilePattern: "test_mvdsb_*"},
				{Name: "test_portal64_bdw", FilePattern: "test_portal64_bdw_*"},
			},
		},
		Storage: config.StorageConfig{
			TempDir:          filepath.Join(tempDir, "import", "temp"),
			MetadataFile:     filepath.Join(tempDir, "import", "metadata.json"),
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
	}
}

// GetTestDatabaseConfig returns a database configuration for testing
func GetTestDatabaseConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
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
}

// GetMinimalImportConfig returns minimal configuration for basic testing
func GetMinimalImportConfig(tempDir string) *config.ImportConfig {
	return &config.ImportConfig{
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
			PasswordMVDSB:    "testpass",
			PasswordPortal64: "testpass",
			ExtractTimeout:   30 * time.Second,
		},
		Database: config.ImportDBConfig{
			ImportTimeout: 60 * time.Second,
			TargetDatabases: []config.TargetDatabase{
				{Name: "test_db", FilePattern: "*"},
			},
		},
		Storage: config.StorageConfig{
			TempDir:      filepath.Join(tempDir, "temp"),
			MetadataFile: filepath.Join(tempDir, "metadata.json"),
		},
	}
}