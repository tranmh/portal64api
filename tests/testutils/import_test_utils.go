// Test utilities for SCP Import Feature tests
package testutils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"portal64api/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CreateTestZipFile creates a password-protected ZIP file with SQL content for testing
func CreateTestZipFile(t *testing.T, zipPath, password string, files map[string]string) {
	// Create a buffer to write our archive to
	buf := new(bytes.Buffer)

	// Create a new zip archive
	zipWriter := zip.NewWriter(buf)

	// Add files to the archive
	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		require.NoError(t, err)

		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
	}

	// Close the zip writer
	err := zipWriter.Close()
	require.NoError(t, err)

	// Write the ZIP file to disk
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	require.NoError(t, err)

	t.Logf("Created test ZIP file: %s with %d files", zipPath, len(files))
}

// CreateTestSQLFile creates a SQL file with test data
func CreateTestSQLFile(t *testing.T, sqlPath, dbName string) {
	content := GetTestSQLContent(dbName)
	
	// Ensure directory exists
	dir := filepath.Dir(sqlPath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(sqlPath, []byte(content), 0644)
	require.NoError(t, err)

	t.Logf("Created test SQL file: %s for database %s", sqlPath, dbName)
}

// GetTestSQLContent returns test SQL content for a given database
func GetTestSQLContent(dbName string) string {
	switch strings.ToLower(dbName) {
	case "mvdsb":
		return `-- Test SQL dump for mvdsb database
DROP DATABASE IF EXISTS test_mvdsb;
CREATE DATABASE test_mvdsb;
USE test_mvdsb;

CREATE TABLE test_players (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    dwz_rating INT DEFAULT 1000,
    club_id VARCHAR(10),
    status TINYINT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test_clubs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    vkz VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(10),
    status TINYINT DEFAULT 1
);

INSERT INTO test_clubs (vkz, name, region, status) VALUES 
('C0327', 'Test Club Muehlacker', 'C', 1),
('C0101', 'Test Post-SV Ulm', 'C', 1);

INSERT INTO test_players (name, dwz_rating, club_id, status) VALUES 
('Test Player 1', 1500, 'C0327', 1),
('Test Player 2', 1800, 'C0327', 1),
('Test Player 3', 2000, 'C0101', 1),
('Test Player 4', 1600, 'C0101', 1);`

	case "portal64_bdw":
		return `-- Test SQL dump for portal64_bdw database
DROP DATABASE IF EXISTS test_portal64_bdw;
CREATE DATABASE test_portal64_bdw;
USE test_portal64_bdw;

CREATE TABLE test_tournaments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tournament_id VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date DATE,
    end_date DATE,
    status VARCHAR(50) DEFAULT 'scheduled',
    region VARCHAR(10)
);

CREATE TABLE test_tournament_results (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tournament_id VARCHAR(20),
    player_name VARCHAR(255),
    old_dwz INT,
    new_dwz INT,
    performance INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_tournaments (tournament_id, name, start_date, end_date, status, region) VALUES 
('C531-634-S25', 'Test Stadtmeisterschaft Stuttgart', '2025-09-01', '2025-09-02', 'completed', 'C'),
('C529-K00-HT1', 'Test Herbstturnier', '2025-10-15', '2025-10-16', 'scheduled', 'C'),
('B718-A08-BEL', 'Test Belgien Open', '2025-08-01', '2025-08-03', 'completed', 'B');

INSERT INTO test_tournament_results (tournament_id, player_name, old_dwz, new_dwz, performance) VALUES 
('C531-634-S25', 'Test Player 1', 1500, 1520, 1550),
('C531-634-S25', 'Test Player 2', 1800, 1785, 1760),
('B718-A08-BEL', 'Test Player 3', 2000, 2015, 2050);`

	default:
		return `-- Generic test SQL dump
DROP DATABASE IF EXISTS test_database;
CREATE DATABASE test_database;
USE test_database;

CREATE TABLE test_table (
    id INT AUTO_INCREMENT PRIMARY KEY,
    data VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_table (data) VALUES 
('Test data 1'),
('Test data 2'),
('Test data 3');`
	}
}

// CreateTestMetadataFile creates a metadata file simulating previous import
func CreateTestMetadataFile(t *testing.T, metadataPath string, lastImport models.ImportStatus) {
	// Ensure directory exists
	dir := filepath.Dir(metadataPath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	// Create metadata structure
	metadata := map[string]interface{}{
		"last_import": map[string]interface{}{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"success":   lastImport.Status == "success",
			"files":     []models.FileMetadata{},
		},
	}

	if lastImport.FilesInfo != nil {
		metadata["last_import"].(map[string]interface{})["files"] = lastImport.FilesInfo.LastImported
	}

	// Write to file
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(metadataPath, data, 0644)
	require.NoError(t, err)

	t.Logf("Created test metadata file: %s", metadataPath)
}

// WaitForCondition waits for a condition to be met with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration, message string) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	
	t.Fatalf("Condition not met within timeout: %s", message)
}

// AssertImportStatus verifies import status meets expectations
func AssertImportStatus(t *testing.T, status *models.ImportStatus, expectedStatus string, expectedProgress int) {
	require.NotNil(t, status)
	require.Equal(t, expectedStatus, status.Status)
	
	if expectedProgress >= 0 {
		require.Equal(t, expectedProgress, status.Progress)
	}
	
	// Verify progress is within valid range
	require.GreaterOrEqual(t, status.Progress, 0)
	require.LessOrEqual(t, status.Progress, 100)
	
	// Verify retry count is valid
	require.GreaterOrEqual(t, status.RetryCount, 0)
	require.LessOrEqual(t, status.RetryCount, status.MaxRetries)
	
	// Verify time consistency
	if status.StartedAt != nil && status.CompletedAt != nil {
		require.True(t, status.CompletedAt.After(*status.StartedAt) || status.CompletedAt.Equal(*status.StartedAt))
	}
	
	// Status-specific assertions
	switch expectedStatus {
	case "success":
		require.Empty(t, status.Error)
		require.Equal(t, 100, status.Progress)
		require.NotNil(t, status.LastSuccess)
	case "failed":
		require.NotEmpty(t, status.Error)
		require.NotNil(t, status.CompletedAt)
	case "skipped":
		require.NotEmpty(t, status.SkipReason)
		require.Equal(t, 100, status.Progress)
		require.Empty(t, status.Error)
	case "running":
		require.NotNil(t, status.StartedAt)
		require.Nil(t, status.CompletedAt)
		require.Empty(t, status.Error)
	}
}

// AssertLogEntries verifies log entries structure and content
func AssertLogEntries(t *testing.T, logs []models.ImportLogEntry, minCount int) {
	require.GreaterOrEqual(t, len(logs), minCount, "Expected at least %d log entries", minCount)
	
	for i, logEntry := range logs {
		require.NotEmpty(t, logEntry.Level, "Log entry %d should have level", i)
		require.NotEmpty(t, logEntry.Message, "Log entry %d should have message", i)
		require.NotEmpty(t, logEntry.Step, "Log entry %d should have step", i)
		require.False(t, logEntry.Timestamp.IsZero(), "Log entry %d should have timestamp", i)
		
		// Verify valid log levels
		require.Contains(t, []string{"INFO", "WARN", "ERROR", "DEBUG"}, logEntry.Level)
	}
}

// CleanupTestDir removes test directory and all contents
func CleanupTestDir(t *testing.T, testDir string) {
	if testDir == "" || testDir == "/" || testDir == "C:\\" {
		t.Fatal("Refusing to cleanup dangerous directory path")
	}
	
	err := os.RemoveAll(testDir)
	if err != nil {
		t.Logf("Warning: failed to cleanup test directory %s: %v", testDir, err)
	}
}

// SetupTestDirectories creates necessary test directories
func SetupTestDirectories(t *testing.T, baseDir string) map[string]string {
	dirs := map[string]string{
		"temp":     filepath.Join(baseDir, "import", "temp"),
		"metadata": filepath.Join(baseDir, "import"),
		"logs":     filepath.Join(baseDir, "logs"),
		"test_data": filepath.Join(baseDir, "test_data"),
	}
	
	for name, path := range dirs {
		err := os.MkdirAll(path, 0755)
		require.NoError(t, err, "Failed to create %s directory: %s", name, path)
	}
	
	return dirs
}

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m MockFileInfo) Name() string       { return m.name }
func (m MockFileInfo) Size() int64        { return m.size }
func (m MockFileInfo) Mode() os.FileMode  { return m.mode }
func (m MockFileInfo) ModTime() time.Time { return m.modTime }
func (m MockFileInfo) IsDir() bool        { return m.isDir }
func (m MockFileInfo) Sys() interface{}   { return nil }

// NewMockFileInfo creates a mock file info for testing
func NewMockFileInfo(name string, size int64, modTime time.Time) MockFileInfo {
	return MockFileInfo{
		name:    name,
		size:    size,
		mode:    0644,
		modTime: modTime,
		isDir:   false,
	}
}

// MockReadCloser implements io.ReadCloser for testing
type MockReadCloser struct {
	*bytes.Reader
	closed bool
}

func (m *MockReadCloser) Close() error {
	m.closed = true
	return nil
}

// NewMockReadCloser creates a mock ReadCloser with given content
func NewMockReadCloser(content []byte) *MockReadCloser {
	return &MockReadCloser{
		Reader: bytes.NewReader(content),
		closed: false,
	}
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadJSONFile reads and parses a JSON file
func ReadJSONFile(t *testing.T, path string, target interface{}) {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	
	err = json.Unmarshal(data, target)
	require.NoError(t, err)
}

// WriteJSONFile writes an object as JSON to a file
func WriteJSONFile(t *testing.T, path string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	
	err = os.WriteFile(path, jsonData, 0644)
	require.NoError(t, err)
}

// CompareFiles compares two files and returns true if they are identical
func CompareFiles(t *testing.T, path1, path2 string) bool {
	data1, err1 := os.ReadFile(path1)
	data2, err2 := os.ReadFile(path2)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return bytes.Equal(data1, data2)
}

// GetFileSize returns the size of a file
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// LogTestStart logs the start of a test with context
func LogTestStart(t *testing.T, testName, description string) {
	t.Logf("=== Starting Test: %s ===", testName)
	if description != "" {
		t.Logf("Description: %s", description)
	}
	t.Logf("Time: %s", time.Now().Format(time.RFC3339))
}

// LogTestEnd logs the end of a test with duration
func LogTestEnd(t *testing.T, testName string, startTime time.Time) {
	duration := time.Since(startTime)
	t.Logf("=== Completed Test: %s ===", testName)
	t.Logf("Duration: %v", duration)
}

// CaptureLogs captures logs for testing (would need to be integrated with actual logger)
type LogCapture struct {
	entries []string
}

func (lc *LogCapture) Write(p []byte) (n int, err error) {
	lc.entries = append(lc.entries, string(p))
	return len(p), nil
}

func (lc *LogCapture) GetEntries() []string {
	return lc.entries
}

func (lc *LogCapture) Clear() {
	lc.entries = nil
}

// NewLogCapture creates a new log capture for testing
func NewLogCapture() *LogCapture {
	return &LogCapture{
		entries: make([]string, 0),
	}
}

// SkipIfCI skips a test if running in CI environment
func SkipIfCI(t *testing.T, reason string) {
	if os.Getenv("CI") != "" {
		t.Skipf("Skipping in CI environment: %s", reason)
	}
}

// SkipIfNotCI skips a test if NOT running in CI environment
func SkipIfNotCI(t *testing.T, reason string) {
	if os.Getenv("CI") == "" {
		t.Skipf("Skipping outside CI environment: %s", reason)
	}
}

// RequireEnvVar requires an environment variable to be set
func RequireEnvVar(t *testing.T, envVar, purpose string) string {
	value := os.Getenv(envVar)
	if value == "" {
		t.Fatalf("Environment variable %s is required for %s", envVar, purpose)
	}
	return value
}

// SetupTempEnv sets up temporary environment variables for testing
func SetupTempEnv(t *testing.T, vars map[string]string) func() {
	original := make(map[string]string)
	
	for key, value := range vars {
		original[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	
	return func() {
		for key, value := range original {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}
}

// TimeoutAfter returns a channel that will receive after the specified duration
func TimeoutAfter(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

// Eventually waits for a condition to be true within a timeout
func Eventually(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	WaitForCondition(t, condition, timeout, 100*time.Millisecond, message)
}
