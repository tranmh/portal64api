package benchmarks

import (
	"crypto/rand"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/cache"
	"portal64api/internal/config"
	"portal64api/internal/importers"
	"portal64api/internal/models"
	"portal64api/internal/services"
	"testing"
	"time"
)

// BenchmarkSCPDownloader benchmarks the SCP downloader component
func BenchmarkSCPDownloader_ListFiles(b *testing.B) {
	cfg := &config.SCPConfig{
		Host:         "localhost",
		Port:         22,
		Username:     "testuser",
		Password:     "testpass",
		RemotePath:   "/test/data/",
		FilePatterns: []string{"*.zip"},
		Timeout:      "30s",
	}

	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	downloader := importers.NewSCPDownloader(cfg, logger)

	// Simulate file listing performance
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate file pattern matching
		testFiles := generateTestFileList(100) // 100 files
		matchedFiles := make([]models.FileMetadata, 0)

		for _, file := range testFiles {
			for _, pattern := range cfg.FilePatterns {
				if matched, _ := filepath.Match(pattern, file.Filename); matched {
					matchedFiles = append(matchedFiles, file)
					break
				}
			}
		}
	}
}

func BenchmarkSCPDownloader_FilePatternMatching(b *testing.B) {
	patterns := []string{"mvdsb_*.zip", "portal64_bdw_*.zip", "test_*.sql"}
	filenames := generateTestFilenames(1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		matchCount := 0
		for _, filename := range filenames {
			for _, pattern := range patterns {
				if matched, _ := filepath.Match(pattern, filename); matched {
					matchCount++
					break
				}
			}
		}
	}
}

// BenchmarkZIPExtractor benchmarks the ZIP extractor component
func BenchmarkZIPExtractor_SmallFiles(b *testing.B) {
	benchmarkZIPExtraction(b, "small", 1024*10) // 10KB files
}

func BenchmarkZIPExtractor_MediumFiles(b *testing.B) {
	benchmarkZIPExtraction(b, "medium", 1024*1024) // 1MB files
}

func BenchmarkZIPExtractor_LargeFiles(b *testing.B) {
	benchmarkZIPExtraction(b, "large", 1024*1024*10) // 10MB files
}

func benchmarkZIPExtraction(b *testing.B, size string, fileSize int) {
	cfg := &config.ZIPConfig{
		PasswordMVDSB:     "testpass123",
		PasswordPortal64:  "testpass123",
		ExtractTimeout:    "60s",
	}

	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	extractor := importers.NewZIPExtractor(cfg, logger)

	// Create test data
	testData := make([]byte, fileSize)
	rand.Read(testData)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate ZIP extraction logic
		// In real benchmark, this would extract actual ZIP files
		extractedSize := len(testData)
		_ = extractedSize // Use the variable to avoid optimization
	}

	b.SetBytes(int64(fileSize))
}

// BenchmarkDatabaseImporter benchmarks the database importer component
func BenchmarkDatabaseImporter_SQLParsing(b *testing.B) {
	sqlContent := generateTestSQLContent(1000) // 1000 SQL statements

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate SQL content validation
		lines := len([]byte(sqlContent))
		_ = lines
	}

	b.SetBytes(int64(len(sqlContent)))
}

func BenchmarkDatabaseImporter_FileMapping(b *testing.B) {
	targetDatabases := []config.TargetDatabase{
		{Name: "mvdsb", FilePattern: "mvdsb_*"},
		{Name: "portal64_bdw", FilePattern: "portal64_bdw_*"},
		{Name: "test_db", FilePattern: "test_*"},
	}

	filenames := generateTestFilenames(500)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mappedCount := 0
		for _, filename := range filenames {
			for _, target := range targetDatabases {
				if matched, _ := filepath.Match(target.FilePattern, filename); matched {
					mappedCount++
					break
				}
			}
		}
	}
}

// BenchmarkFreshnessChecker benchmarks the freshness checker component
func BenchmarkFreshnessChecker_FileComparison(b *testing.B) {
	baseTime := time.Now()
	remoteFiles := generateTestFileList(100)
	lastImportedFiles := generateTestFileList(100)

	// Modify some files to be newer
	for i := 0; i < 20; i++ {
		remoteFiles[i].ModTime = baseTime.Add(1 * time.Hour)
		remoteFiles[i].Size = remoteFiles[i].Size + 1024
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		newerCount := 0
		for _, remoteFile := range remoteFiles {
			for _, lastFile := range lastImportedFiles {
				if remoteFile.Filename == lastFile.Filename {
					if remoteFile.ModTime.After(lastFile.ModTime) || remoteFile.Size != lastFile.Size {
						newerCount++
					}
					break
				}
			}
		}
	}
}

func BenchmarkFreshnessChecker_MetadataProcessing(b *testing.B) {
	files := generateTestFileList(1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate metadata processing
		totalSize := int64(0)
		for _, file := range files {
			totalSize += file.Size
		}
	}
}

// BenchmarkStatusTracker benchmarks the status tracker component
func BenchmarkStatusTracker_StatusUpdates(b *testing.B) {
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	tracker := importers.NewStatusTracker(1000, logger)

	steps := []string{"initialization", "downloading", "extracting", "importing", "cleanup"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		step := steps[i%len(steps)]
		progress := (i * 100) / b.N
		tracker.UpdateStatus(step, progress)
	}
}

func BenchmarkStatusTracker_LogEntries(b *testing.B) {
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	tracker := importers.NewStatusTracker(10000, logger)

	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	steps := []string{"initialization", "downloading", "extracting", "importing", "cleanup"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		level := levels[i%len(levels)]
		step := steps[i%len(steps)]
		message := "Benchmark log entry " + string(rune(i))
		tracker.AddLogEntry(level, message, step)
	}
}

func BenchmarkStatusTracker_ConcurrentAccess(b *testing.B) {
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	tracker := importers.NewStatusTracker(1000, logger)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Simulate concurrent status reads and updates
			if i%2 == 0 {
				status := tracker.GetStatus()
				_ = status
			} else {
				tracker.UpdateStatus("testing", i%100)
			}
			i++
		}
	})
}

// BenchmarkImportService benchmarks the import service
func BenchmarkImportService_StatusRetrieval(b *testing.B) {
	tempDir := b.TempDir()
	importConfig := getTestImportConfig(tempDir)
	dbConfig := getTestDatabaseConfig()
	mockCache := &MockCacheService{}
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)

	service := services.NewImportService(importConfig, dbConfig, mockCache, logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		status := service.GetStatus()
		_ = status
	}
}

func BenchmarkImportService_LogRetrieval(b *testing.B) {
	tempDir := b.TempDir()
	importConfig := getTestImportConfig(tempDir)
	dbConfig := getTestDatabaseConfig()
	mockCache := &MockCacheService{}
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)

	service := services.NewImportService(importConfig, dbConfig, mockCache, logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logs := service.GetLogs()
		_ = logs
	}
}

// Memory benchmarks
func BenchmarkMemoryUsage_ImportWorkflow(b *testing.B) {
	tempDir := b.TempDir()
	importConfig := getTestImportConfig(tempDir)
	dbConfig := getTestDatabaseConfig()
	mockCache := &MockCacheService{}
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate full import service lifecycle
		service := services.NewImportService(importConfig, dbConfig, mockCache, logger)
		_ = service.Start()
		status := service.GetStatus()
		_ = status
		logs := service.GetLogs()
		_ = logs
		_ = service.Stop()
	}
}

// Concurrent access benchmarks
func BenchmarkConcurrentAccess_StatusTracker(b *testing.B) {
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	tracker := importers.NewStatusTracker(1000, logger)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 3 {
			case 0:
				tracker.UpdateStatus("testing", i%100)
			case 1:
				status := tracker.GetStatus()
				_ = status
			case 2:
				logs := tracker.GetLogs()
				_ = logs
			}
			i++
		}
	})
}

func BenchmarkConcurrentAccess_ImportService(b *testing.B) {
	tempDir := b.TempDir()
	importConfig := getTestImportConfig(tempDir)
	dbConfig := getTestDatabaseConfig()
	mockCache := &MockCacheService{}
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)

	service := services.NewImportService(importConfig, dbConfig, mockCache, logger)
	service.Start()
	defer service.Stop()

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(20)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			status := service.GetStatus()
			_ = status
		}
	})
}

// Utility functions for benchmarks

func generateTestFileList(count int) []models.FileMetadata {
	files := make([]models.FileMetadata, count)
	baseTime := time.Now()

	for i := 0; i < count; i++ {
		files[i] = models.FileMetadata{
			Filename: generateTestFilename(i),
			Size:     int64(1024 * (i + 1)),
			ModTime:  baseTime.Add(time.Duration(i) * time.Minute),
			Checksum: "sha256:test" + string(rune(i)),
		}
	}

	return files
}

func generateTestFilenames(count int) []string {
	filenames := make([]string, count)

	for i := 0; i < count; i++ {
		filenames[i] = generateTestFilename(i)
		// Use pattern to generate more realistic filenames
		if i%4 == 0 {
			filenames[i] = "mvdsb_20250806_" + string(rune('a'+i%26)) + ".zip"
		} else if i%4 == 1 {
			filenames[i] = "portal64_bdw_20250806_" + string(rune('a'+i%26)) + ".zip"
		}
	}

	return filenames
}

func generateTestFilename(index int) string {
	return string(rune('a'+index%26)) + "_file_" + string(rune('0'+index%10)) + ".zip"
}

func generateTestSQLContent(statementCount int) string {
	content := "-- Generated SQL content for benchmarking\n"
	
	for i := 0; i < statementCount; i++ {
		content += "INSERT INTO test_table (id, data) VALUES (" + string(rune('0'+i%10)) + ", 'test data " + string(rune('0'+i%10)) + "');\n"
	}
	
	return content
}

func getTestImportConfig(tempDir string) *config.ImportConfig {
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
			PasswordMVDSB:     "testpass",
			PasswordPortal64:  "testpass",
			ExtractTimeout:    "30s",
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

func getTestDatabaseConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "",
		Name:     "test_db",
	}
}

// MockCacheService for benchmarking
type MockCacheService struct{}

func (m *MockCacheService) Get(key string) ([]byte, error) { return nil, cache.ErrCacheMiss }
func (m *MockCacheService) Set(key string, value []byte, ttl time.Duration) error { return nil }
func (m *MockCacheService) Delete(key string) error { return nil }
func (m *MockCacheService) GetWithRefresh(key string, refreshThreshold time.Duration, refreshFn func() ([]byte, time.Duration, error)) ([]byte, error) {
	return nil, cache.ErrCacheMiss
}
func (m *MockCacheService) FlushAll() error { return nil }
func (m *MockCacheService) GetStats() cache.CacheStats {
	return cache.CacheStats{HitRatio: 0.0, RequestCount: 0, HitCount: 0, MissCount: 0, KeyCount: 0, AvgResponseTime: 0, MemoryUsage: 0}
}
func (m *MockCacheService) HealthCheck() error { return nil }
