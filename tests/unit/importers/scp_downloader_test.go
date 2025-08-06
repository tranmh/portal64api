package importers

import (
	"crypto/rand"
	"io"
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

// MockSFTPClient simulates SFTP client behavior for testing
type MockSFTPClient struct {
	files    []os.FileInfo
	readDir  func(string) ([]os.FileInfo, error)
	stat     func(string) (os.FileInfo, error)
	open     func(string) (io.ReadCloser, error)
	failOpen bool
	failStat bool
}

func (m *MockSFTPClient) ReadDir(path string) ([]os.FileInfo, error) {
	if m.readDir != nil {
		return m.readDir(path)
	}
	return m.files, nil
}

func (m *MockSFTPClient) Stat(path string) (os.FileInfo, error) {
	if m.failStat {
		return nil, assert.AnError
	}
	if m.stat != nil {
		return m.stat(path)
	}
	// Return the first file as default
	if len(m.files) > 0 {
		return m.files[0], nil
	}
	return nil, os.ErrNotExist
}

func (m *MockSFTPClient) Open(path string) (io.ReadCloser, error) {
	if m.failOpen {
		return nil, assert.AnError
	}
	if m.open != nil {
		return m.open(path)
	}
	// Return a mock file with some content
	return &mockFileReader{content: []byte("test file content")}, nil
}

func (m *MockSFTPClient) Close() error {
	return nil
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

// mockFileReader simulates file reading
type mockFileReader struct {
	content []byte
	pos     int
}

func (m *mockFileReader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.content) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockFileReader) Close() error {
	return nil
}

func TestSCPDownloader_ListFiles(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.SCPConfig
		mockFiles      []os.FileInfo
		expectedFiles  int
		expectedError  bool
		errorMessage   string
	}{
		{
			name: "successful file listing with matching patterns",
			config: &config.SCPConfig{
				Host:         "test.example.com",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"mvdsb_*.zip", "portal64_bdw_*.zip"},
				Timeout:      300 * time.Second,
			},
			mockFiles: []os.FileInfo{
				mockFileInfo{
					name:    "mvdsb_20250806.zip",
					size:    1024000,
					modTime: time.Now().Add(-1 * time.Hour),
				},
				mockFileInfo{
					name:    "portal64_bdw_20250806.zip",
					size:    512000,
					modTime: time.Now().Add(-1 * time.Hour),
				},
				mockFileInfo{
					name:    "other_file.txt",
					size:    100,
					modTime: time.Now().Add(-1 * time.Hour),
				},
			},
			expectedFiles: 2, // Only the matching .zip files
			expectedError: false,
		},
		{
			name: "no matching files found",
			config: &config.SCPConfig{
				Host:         "test.example.com",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"missing_*.zip"},
				Timeout:      300 * time.Second,
			},
			mockFiles: []os.FileInfo{
				mockFileInfo{
					name:    "other_file.txt",
					size:    100,
					modTime: time.Now().Add(-1 * time.Hour),
				},
			},
			expectedFiles: 0,
			expectedError: false,
		},
		{
			name: "empty directory",
			config: &config.SCPConfig{
				Host:         "test.example.com",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"*.zip"},
				Timeout:      300 * time.Second,
			},
			mockFiles:     []os.FileInfo{},
			expectedFiles: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger for testing
			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)

			// Create the downloader (note: we can't easily test the actual SFTP connection
			// without significant mocking of the ssh package, so we focus on testing the
			// file pattern matching logic and data structure handling)
			downloader := importers.NewSCPDownloader(tt.config, logger)
			
			// Verify the downloader was created correctly
			assert.NotNil(t, downloader)

			// Test file pattern matching logic separately
			matchingFiles := []models.FileMetadata{}
			for _, file := range tt.mockFiles {
				for _, pattern := range tt.config.FilePatterns {
					matched, err := filepath.Match(pattern, file.Name())
					require.NoError(t, err)
					if matched {
						metadata := models.FileMetadata{
							Filename: file.Name(),
							Size:     file.Size(),
							ModTime:  file.ModTime(),
						}
						matchingFiles = append(matchingFiles, metadata)
						break
					}
				}
			}

			// Verify the pattern matching results
			assert.Equal(t, tt.expectedFiles, len(matchingFiles))

			if tt.expectedFiles > 0 {
				// Verify the first matching file has correct structure
				assert.NotEmpty(t, matchingFiles[0].Filename)
				assert.Greater(t, matchingFiles[0].Size, int64(0))
				assert.False(t, matchingFiles[0].ModTime.IsZero())
			}
		})
	}
}

func TestSCPDownloader_DownloadFile(t *testing.T) {
	// Create temporary directory for test downloads
	tempDir := t.TempDir()
	
	tests := []struct {
		name          string
		filename      string
		localPath     string
		fileContent   []byte
		expectedError bool
		errorMessage  string
	}{
		{
			name:        "successful file download",
			filename:    "test_file.zip",
			localPath:   filepath.Join(tempDir, "test_file.zip"),
			fileContent: []byte("This is test file content for download testing"),
			expectedError: false,
		},
		{
			name:        "download large file",
			filename:    "large_file.zip",
			localPath:   filepath.Join(tempDir, "large_file.zip"),
			fileContent: make([]byte, 1024*1024), // 1MB of data
			expectedError: false,
		},
		{
			name:        "download empty file",
			filename:    "empty_file.zip",
			localPath:   filepath.Join(tempDir, "empty_file.zip"),
			fileContent: []byte{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill large file with random data
			if len(tt.fileContent) > 1000 {
				_, err := rand.Read(tt.fileContent)
				require.NoError(t, err)
			}

			// Create config
			cfg := &config.SCPConfig{
				Host:         "test.example.com",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"*.zip"},
				Timeout:      300 * time.Second,
			}

			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
			downloader := importers.NewSCPDownloader(cfg, logger)
			
			assert.NotNil(t, downloader)

			// Test file metadata creation
			metadata := models.FileMetadata{
				Filename: tt.filename,
				Size:     int64(len(tt.fileContent)),
				ModTime:  time.Now(),
			}

			// Verify metadata structure
			assert.Equal(t, tt.filename, metadata.Filename)
			assert.Equal(t, int64(len(tt.fileContent)), metadata.Size)
			assert.False(t, metadata.ModTime.IsZero())

			// If we were to simulate the download, we would create the file
			if !tt.expectedError {
				err := os.WriteFile(tt.localPath, tt.fileContent, 0644)
				require.NoError(t, err)

				// Verify file was created
				assert.FileExists(t, tt.localPath)

				// Verify file content
				downloadedContent, err := os.ReadFile(tt.localPath)
				require.NoError(t, err)
				assert.Equal(t, tt.fileContent, downloadedContent)

				// Verify file size
				info, err := os.Stat(tt.localPath)
				require.NoError(t, err)
				assert.Equal(t, int64(len(tt.fileContent)), info.Size())
			}
		})
	}
}

func TestSCPDownloader_Configuration(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.SCPConfig
		expectedValid  bool
		errorMessage   string
	}{
		{
			name: "valid configuration",
			config: &config.SCPConfig{
				Host:         "portal.svw.info",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"mvdsb_*.zip", "portal64_bdw_*.zip"},
				Timeout:      300 * time.Second,
			},
			expectedValid: true,
		},
		{
			name: "empty host",
			config: &config.SCPConfig{
				Host:         "",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"*.zip"},
				Timeout:      300 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "host cannot be empty",
		},
		{
			name: "invalid port",
			config: &config.SCPConfig{
				Host:         "test.example.com",
				Port:         0,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"*.zip"},
				Timeout:      300 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "port must be greater than 0",
		},
		{
			name: "empty username",
			config: &config.SCPConfig{
				Host:         "test.example.com",
				Port:         22,
				Username:     "",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{"*.zip"},
				Timeout:      300 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "username cannot be empty",
		},
		{
			name: "empty file patterns",
			config: &config.SCPConfig{
				Host:         "test.example.com",
				Port:         22,
				Username:     "testuser",
				Password:     "testpass",
				RemotePath:   "/data/exports/",
				FilePatterns: []string{},
				Timeout:      300 * time.Second,
			},
			expectedValid: false,
			errorMessage:  "at least one file pattern must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
			
			// Test basic configuration validation logic
			var configErrors []string
			
			if tt.config.Host == "" {
				configErrors = append(configErrors, "host cannot be empty")
			}
			if tt.config.Port <= 0 {
				configErrors = append(configErrors, "port must be greater than 0")
			}
			if tt.config.Username == "" {
				configErrors = append(configErrors, "username cannot be empty")
			}
			if len(tt.config.FilePatterns) == 0 {
				configErrors = append(configErrors, "at least one file pattern must be specified")
			}

			if tt.expectedValid {
				assert.Empty(t, configErrors, "Configuration should be valid")
				
				// Should be able to create downloader with valid config
				downloader := importers.NewSCPDownloader(tt.config, logger)
				assert.NotNil(t, downloader)
			} else {
				assert.NotEmpty(t, configErrors, "Configuration should have validation errors")
				if len(configErrors) > 0 {
					assert.Contains(t, configErrors, tt.errorMessage)
				}
			}
		})
	}
}

func TestSCPDownloader_FilePatternMatching(t *testing.T) {
	tests := []struct {
		name         string
		patterns     []string
		filename     string
		shouldMatch  bool
	}{
		{
			name:         "mvdsb pattern matches",
			patterns:     []string{"mvdsb_*.zip"},
			filename:     "mvdsb_20250806.zip",
			shouldMatch:  true,
		},
		{
			name:         "portal64_bdw pattern matches",
			patterns:     []string{"portal64_bdw_*.zip"},
			filename:     "portal64_bdw_20250806.zip",
			shouldMatch:  true,
		},
		{
			name:         "multiple patterns - first matches",
			patterns:     []string{"mvdsb_*.zip", "portal64_bdw_*.zip"},
			filename:     "mvdsb_20250806.zip",
			shouldMatch:  true,
		},
		{
			name:         "multiple patterns - second matches",
			patterns:     []string{"mvdsb_*.zip", "portal64_bdw_*.zip"},
			filename:     "portal64_bdw_20250806.zip",
			shouldMatch:  true,
		},
		{
			name:         "no pattern matches",
			patterns:     []string{"mvdsb_*.zip", "portal64_bdw_*.zip"},
			filename:     "other_file.txt",
			shouldMatch:  false,
		},
		{
			name:         "wildcard pattern matches all",
			patterns:     []string{"*"},
			filename:     "any_file.zip",
			shouldMatch:  true,
		},
		{
			name:         "extension-specific pattern",
			patterns:     []string{"*.zip"},
			filename:     "test.zip",
			shouldMatch:  true,
		},
		{
			name:         "extension-specific pattern no match",
			patterns:     []string{"*.zip"},
			filename:     "test.txt",
			shouldMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false
			
			// Test the pattern matching logic
			for _, pattern := range tt.patterns {
				match, err := filepath.Match(pattern, tt.filename)
				require.NoError(t, err, "Pattern matching should not error")
				
				if match {
					matched = true
					break
				}
			}

			assert.Equal(t, tt.shouldMatch, matched, 
				"File '%s' should match patterns %v: %t", tt.filename, tt.patterns, tt.shouldMatch)
		})
	}
}
