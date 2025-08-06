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
			Timeout:      "30s",
		},
		ZIP: config.ZIPConfig{
			Password:       "testzip123",
			ExtractTimeout: "30s",
		},
		Database: config.DatabaseImportConfig{
			ImportTimeout: "600s",
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
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "",
		Name:     "test_portal64api",
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
			Timeout:      "30s",
		},
		ZIP: config.ZIPConfig{
			Password:       "testpass",
			ExtractTimeout: "30s",
		},
		Database: config.DatabaseImportConfig{
			ImportTimeout: "60s",
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
