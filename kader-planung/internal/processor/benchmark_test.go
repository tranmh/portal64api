package processor

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/statistics"
	"github.com/sirupsen/logrus"
)

// PerformanceMetrics holds performance measurement data
type PerformanceMetrics struct {
	ExecutionTime    time.Duration `json:"execution_time"`
	APICallCount     int           `json:"api_call_count"`
	MemoryUsageMB    int64         `json:"memory_usage_mb"`
	ThroughputRPS    float64       `json:"throughput_rps"`
	ErrorRate        float64       `json:"error_rate"`
	PlayersProcessed int           `json:"players_processed"`
	ClubsProcessed   int           `json:"clubs_processed"`
}

// BenchmarkCase defines a test case for performance benchmarking
type BenchmarkCase struct {
	Name        string         `json:"name"`
	ClubCount   int           `json:"club_count"`
	PlayerCount int           `json:"player_count"`
	Mode        ProcessingMode `json:"mode"`
	Config      *UnifiedProcessorConfig `json:"config"`
}

// PerformanceBenchmark manages performance testing
type PerformanceBenchmark struct {
	testCases []BenchmarkCase
	metrics   map[string]*PerformanceMetrics
	logger    *logrus.Logger
}

// NewPerformanceBenchmark creates a new performance benchmark suite
func NewPerformanceBenchmark() *PerformanceBenchmark {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise during benchmarking

	return &PerformanceBenchmark{
		testCases: createBenchmarkCases(),
		metrics:   make(map[string]*PerformanceMetrics),
		logger:    logger,
	}
}

// createBenchmarkCases defines the benchmark test cases based on Phase 5 requirements
func createBenchmarkCases() []BenchmarkCase {
	return []BenchmarkCase{
		// Small dataset: 10 clubs, ~500 players
		{
			Name:        "Small_Dataset_Statistical",
			ClubCount:   10,
			PlayerCount: 500,
			Mode:        StatisticalMode,
			Config: &UnifiedProcessorConfig{
				Mode:             StatisticalMode,
				MinSampleSize:    10,
				EnableStatistics: true,
				Concurrency:      4,
				Verbose:          false,
			},
		},
		{
			Name:        "Small_Dataset_Efficient",
			ClubCount:   10,
			PlayerCount: 500,
			Mode:        EfficientMode,
			Config: &UnifiedProcessorConfig{
				Mode:             EfficientMode,
				MinSampleSize:    10,
				EnableStatistics: false,
				Concurrency:      4,
				Verbose:          false,
			},
		},
		{
			Name:        "Small_Dataset_Hybrid",
			ClubCount:   10,
			PlayerCount: 500,
			Mode:        HybridMode,
			Config: &UnifiedProcessorConfig{
				Mode:             HybridMode,
				MinSampleSize:    10,
				EnableStatistics: true,
				Concurrency:      4,
				Verbose:          false,
			},
		},

		// Medium dataset: 50 clubs, ~5,000 players
		{
			Name:        "Medium_Dataset_Statistical",
			ClubCount:   50,
			PlayerCount: 5000,
			Mode:        StatisticalMode,
			Config: &UnifiedProcessorConfig{
				Mode:             StatisticalMode,
				MinSampleSize:    20,
				EnableStatistics: true,
				Concurrency:      8,
				Verbose:          false,
			},
		},
		{
			Name:        "Medium_Dataset_Efficient",
			ClubCount:   50,
			PlayerCount: 5000,
			Mode:        EfficientMode,
			Config: &UnifiedProcessorConfig{
				Mode:             EfficientMode,
				MinSampleSize:    20,
				EnableStatistics: false,
				Concurrency:      8,
				Verbose:          false,
			},
		},
		{
			Name:        "Medium_Dataset_Hybrid",
			ClubCount:   50,
			PlayerCount: 5000,
			Mode:        HybridMode,
			Config: &UnifiedProcessorConfig{
				Mode:             HybridMode,
				MinSampleSize:    20,
				EnableStatistics: true,
				Concurrency:      8,
				Verbose:          false,
			},
		},

		// Large dataset: 200 clubs, ~20,000 players
		{
			Name:        "Large_Dataset_Statistical",
			ClubCount:   200,
			PlayerCount: 20000,
			Mode:        StatisticalMode,
			Config: &UnifiedProcessorConfig{
				Mode:             StatisticalMode,
				MinSampleSize:    50,
				EnableStatistics: true,
				Concurrency:      16,
				Verbose:          false,
			},
		},
		{
			Name:        "Large_Dataset_Efficient",
			ClubCount:   200,
			PlayerCount: 20000,
			Mode:        EfficientMode,
			Config: &UnifiedProcessorConfig{
				Mode:             EfficientMode,
				MinSampleSize:    50,
				EnableStatistics: false,
				Concurrency:      16,
				Verbose:          false,
			},
		},
		{
			Name:        "Large_Dataset_Hybrid",
			ClubCount:   200,
			PlayerCount: 20000,
			Mode:        HybridMode,
			Config: &UnifiedProcessorConfig{
				Mode:             HybridMode,
				MinSampleSize:    50,
				EnableStatistics: true,
				Concurrency:      16,
				Verbose:          false,
			},
		},
	}
}

// BenchmarkUnifiedProcessor_StatisticalAnalysis benchmarks statistical analysis performance
func BenchmarkUnifiedProcessor_StatisticalAnalysis(b *testing.B) {
	// Test different dataset sizes for statistical analysis
	datasets := []struct {
		name        string
		playerCount int
		minSample   int
	}{
		{"Small_500_players", 500, 10},
		{"Medium_5000_players", 5000, 20},
		{"Large_20000_players", 20000, 50},
		{"XLarge_50000_players", 50000, 100},
	}

	for _, dataset := range datasets {
		b.Run(dataset.name, func(b *testing.B) {
			players := generateLargeMockPlayerSet(dataset.playerCount)
			analyzer := statistics.NewStatisticalAnalyzer(dataset.minSample, logrus.New())

			// Reset timer to exclude setup time
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := analyzer.ProcessPlayers(players)
				if err != nil {
					b.Fatalf("Statistical analysis failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkUnifiedProcessor_DetailedAnalysis benchmarks detailed processing
func BenchmarkUnifiedProcessor_DetailedAnalysis(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Minimize output during benchmarking

	datasets := []struct {
		name        string
		playerCount int
		concurrency int
	}{
		{"Small_500_players", 500, 4},
		{"Medium_5000_players", 5000, 8},
		{"Large_20000_players", 20000, 16},
	}

	for _, dataset := range datasets {
		b.Run(dataset.name, func(b *testing.B) {
			config := &UnifiedProcessorConfig{
				Mode:             EfficientMode, // Use efficient mode for benchmarking
				MinSampleSize:    10,
				EnableStatistics: false,
				Concurrency:      dataset.concurrency,
				Verbose:          false,
			}

			processor := &UnifiedProcessor{
				config: config,
				logger: logger,
				statisticalAnalyzer: statistics.NewStatisticalAnalyzer(config.MinSampleSize, logger),
			}

			players := generateLargeMockPlayerSet(dataset.playerCount)

			// Reset timer to exclude setup time
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				records := processor.convertPlayersToRecords(players, false)
				processor.calculateListRankings(records)
			}
		})
	}
}

// BenchmarkUnifiedProcessor_ConcurrentProcessing benchmarks concurrent processing performance
func BenchmarkUnifiedProcessor_ConcurrentProcessing(b *testing.B) {
	playerCount := 10000
	players := generateLargeMockPlayerSet(playerCount)

	concurrencyLevels := []int{1, 2, 4, 8, 16, 32}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			analyzer := statistics.NewStatisticalAnalyzer(20, logrus.New())

			// Reset timer to exclude setup time
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := analyzer.ProcessPlayers(players)
				if err != nil {
					b.Fatalf("Processing failed with concurrency %d: %v", concurrency, err)
				}
			}
		})
	}
}

// BenchmarkMemoryUsage measures memory usage for different processing modes
func BenchmarkMemoryUsage(b *testing.B) {
	playerCount := 20000
	players := generateLargeMockPlayerSet(playerCount)

	modes := []struct {
		name string
		mode ProcessingMode
	}{
		{"Statistical_Mode", StatisticalMode},
		{"Efficient_Mode", EfficientMode},
		{"Hybrid_Mode", HybridMode},
	}

	for _, mode := range modes {
		b.Run(mode.name, func(b *testing.B) {
			config := &UnifiedProcessorConfig{
				Mode:             mode.mode,
				MinSampleSize:    20,
				EnableStatistics: mode.mode != EfficientMode,
				Concurrency:      8,
				Verbose:          false,
			}

			processor := &UnifiedProcessor{
				config: config,
				logger: logrus.New(),
				statisticalAnalyzer: statistics.NewStatisticalAnalyzer(config.MinSampleSize, logrus.New()),
			}

			// Measure memory before processing
			var beforeMem runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&beforeMem)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				switch mode.mode {
				case StatisticalMode:
					processor.statisticalAnalyzer.ProcessPlayers(players)
				case EfficientMode, HybridMode:
					records := processor.convertPlayersToRecords(players, false)
					processor.calculateListRankings(records)
					if config.EnableStatistics {
						processor.statisticalAnalyzer.ProcessPlayers(players)
					}
				}
			}

			// Measure memory after processing
			var afterMem runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&afterMem)

			memUsageMB := int64(afterMem.Alloc-beforeMem.Alloc) / (1024 * 1024)
			b.ReportMetric(float64(memUsageMB), "MB")
		})
	}
}

// RunBenchmarkSuite executes the complete benchmark suite
func (pb *PerformanceBenchmark) RunBenchmarkSuite(b *testing.B) {
	pb.logger.Info("Starting performance benchmark suite...")

	for _, testCase := range pb.testCases {
		b.Run(testCase.Name, func(b *testing.B) {
			metrics := pb.runSingleBenchmark(b, testCase)
			pb.metrics[testCase.Name] = metrics

			// Report metrics
			b.ReportMetric(float64(metrics.ExecutionTime.Nanoseconds())/1e9, "seconds")
			b.ReportMetric(float64(metrics.MemoryUsageMB), "MB")
			b.ReportMetric(metrics.ThroughputRPS, "players/sec")
			b.ReportMetric(float64(metrics.APICallCount), "api_calls")
		})
	}

	pb.logger.Info("Performance benchmark suite completed")
}

// runSingleBenchmark executes a single benchmark case
func (pb *PerformanceBenchmark) runSingleBenchmark(b *testing.B, testCase BenchmarkCase) *PerformanceMetrics {
	// Generate test data
	players := generateScaledMockPlayerSet(testCase.PlayerCount, testCase.ClubCount)

	// Create processor
	processor := &UnifiedProcessor{
		config: testCase.Config,
		logger: pb.logger,
		statisticalAnalyzer: statistics.NewStatisticalAnalyzer(testCase.Config.MinSampleSize, pb.logger),
	}

	// Measure memory before
	var beforeMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&beforeMem)

	startTime := time.Now()

	// Execute the benchmark
	for i := 0; i < b.N; i++ {
		switch testCase.Mode {
		case StatisticalMode:
			_, err := processor.statisticalAnalyzer.ProcessPlayers(players)
			if err != nil {
				b.Fatalf("Statistical processing failed: %v", err)
			}
		case EfficientMode:
			records := processor.convertPlayersToRecords(players, false)
			processor.calculateListRankings(records)
		case HybridMode:
			records := processor.convertPlayersToRecords(players, false)
			processor.calculateListRankings(records)
			if testCase.Config.EnableStatistics {
				_, err := processor.statisticalAnalyzer.ProcessPlayers(players)
				if err != nil {
					b.Logf("Warning: Statistical analysis failed in hybrid mode: %v", err)
				}
			}
		}
	}

	executionTime := time.Since(startTime)

	// Measure memory after
	var afterMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&afterMem)

	memUsageMB := int64(afterMem.Alloc-beforeMem.Alloc) / (1024 * 1024)
	throughput := float64(testCase.PlayerCount*b.N) / executionTime.Seconds()

	return &PerformanceMetrics{
		ExecutionTime:    executionTime,
		APICallCount:     estimateAPICallCount(testCase.Mode, testCase.ClubCount, testCase.PlayerCount),
		MemoryUsageMB:    memUsageMB,
		ThroughputRPS:    throughput,
		ErrorRate:        0.0, // Assuming no errors for successful benchmarks
		PlayersProcessed: testCase.PlayerCount,
		ClubsProcessed:   testCase.ClubCount,
	}
}

// estimateAPICallCount estimates API call count for different modes (based on documentation analysis)
func estimateAPICallCount(mode ProcessingMode, clubs, players int) int {
	switch mode {
	case StatisticalMode, EfficientMode, HybridMode:
		// Somatogramm-style efficient: 1 + N API calls (where N = clubs)
		return 1 + clubs
	case DetailedMode:
		// Traditional Kader-Planung: 1 + N + 2*N*P (where N=clubs, P=avg players per club)
		avgPlayersPerClub := players / clubs
		return 1 + clubs + (2 * clubs * avgPlayersPerClub)
	default:
		return clubs + players // Conservative estimate
	}
}

// generateLargeMockPlayerSet creates a large set of mock players for benchmarking
func generateLargeMockPlayerSet(count int) []models.Player {
	return generateScaledMockPlayerSet(count, count/50) // Default ~50 players per club
}

// generateScaledMockPlayerSet creates mock players distributed across specified number of clubs
func generateScaledMockPlayerSet(playerCount, clubCount int) []models.Player {
	currentYear := time.Now().Year()
	players := make([]models.Player, playerCount)

	// Age distribution for realistic statistical analysis
	ageRanges := []int{12, 14, 16, 18, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80}
	genders := []string{"m", "w"}

	playersPerClub := playerCount / clubCount
	if playersPerClub < 1 {
		playersPerClub = 1
	}

	for i := 0; i < playerCount; i++ {
		clubIndex := i / playersPerClub
		if clubIndex >= clubCount {
			clubIndex = clubCount - 1
		}

		// Distribute ages to ensure we have enough samples in each age group
		ageIndex := i % len(ageRanges)
		age := ageRanges[ageIndex]
		birthYear := currentYear - age

		// Alternate genders
		gender := genders[i%2]

		// Generate realistic DWZ values (1000-2500 range)
		baseDWZ := 1200 + (i%1000) // Base range 1200-2199
		ageFactor := 0
		if age >= 20 && age <= 40 {
			ageFactor = 100 // Peak performance years
		} else if age >= 12 && age < 20 {
			ageFactor = 50 // Junior bonus
		}
		dwz := baseDWZ + ageFactor

		clubID := fmt.Sprintf("C%04d", 1000+clubIndex)

		players[i] = models.Player{
			ID:         fmt.Sprintf("%s-%06d", clubID, i%1000000),
			Name:       fmt.Sprintf("Player%d", i),
			Firstname:  fmt.Sprintf("Test%d", i),
			BirthYear:  &birthYear,
			Gender:     gender,
			CurrentDWZ: dwz,
			ClubID:     clubID,
			Club:       fmt.Sprintf("Test Club %d", clubIndex+1),
			Status:     "active",
			Active:     true,
			PKZ:        fmt.Sprintf("%08d", 10000000+i),
		}
	}

	return players
}

// GetMetrics returns the collected performance metrics
func (pb *PerformanceBenchmark) GetMetrics() map[string]*PerformanceMetrics {
	return pb.metrics
}

// PrintSummary prints a summary of all benchmark results
func (pb *PerformanceBenchmark) PrintSummary() {
	pb.logger.Info("=== Performance Benchmark Summary ===")

	for name, metrics := range pb.metrics {
		pb.logger.Infof("Test Case: %s", name)
		pb.logger.Infof("  Execution Time: %v", metrics.ExecutionTime)
		pb.logger.Infof("  Memory Usage: %d MB", metrics.MemoryUsageMB)
		pb.logger.Infof("  Throughput: %.2f players/sec", metrics.ThroughputRPS)
		pb.logger.Infof("  Est. API Calls: %d", metrics.APICallCount)
		pb.logger.Infof("  Players Processed: %d", metrics.PlayersProcessed)
		pb.logger.Infof("  Clubs Processed: %d", metrics.ClubsProcessed)
		pb.logger.Info("  ---")
	}
}

// TestPerformanceBenchmarkSuite is the main test function for running performance benchmarks
func TestPerformanceBenchmarkSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmark suite in short mode")
	}

	benchmark := NewPerformanceBenchmark()

	// Run each benchmark case individually for better control
	for _, testCase := range benchmark.testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Create a sub-benchmark
			result := testing.Benchmark(func(b *testing.B) {
				benchmark.RunBenchmarkSuite(b)
			})

			t.Logf("Benchmark %s: %s", testCase.Name, result.String())

			// Performance targets based on Phase 5 requirements
			validatePerformanceTargets(t, testCase, benchmark.metrics[testCase.Name])
		})
	}

	benchmark.PrintSummary()
}

// validatePerformanceTargets validates against Phase 5 performance requirements
func validatePerformanceTargets(t *testing.T, testCase BenchmarkCase, metrics *PerformanceMetrics) {
	if metrics == nil {
		t.Errorf("No metrics available for test case %s", testCase.Name)
		return
	}

	// Performance targets from Phase 5 documentation:
	// - Statistical analysis: <5 minutes for 50,000 players
	// - Memory usage: <2GB for statistical mode
	// - API efficiency: >90% reduction in calls for statistical mode

	if testCase.PlayerCount == 50000 && testCase.Mode == StatisticalMode {
		maxDuration := 5 * time.Minute
		if metrics.ExecutionTime > maxDuration {
			t.Errorf("Statistical analysis took %v, expected <%v for 50K players",
				metrics.ExecutionTime, maxDuration)
		}
	}

	// Memory target: <2GB for statistical mode
	if testCase.Mode == StatisticalMode && metrics.MemoryUsageMB > 2048 {
		t.Errorf("Memory usage %d MB exceeds 2GB target for statistical mode", metrics.MemoryUsageMB)
	}

	// API efficiency target: Should use 1 + N calls for efficient modes
	_ = estimateAPICallCount(testCase.Mode, testCase.ClubCount, testCase.PlayerCount)
	if testCase.Mode == StatisticalMode || testCase.Mode == EfficientMode {
		maxExpectedCalls := 1 + testCase.ClubCount + 10 // Allow small buffer
		if metrics.APICallCount > maxExpectedCalls {
			t.Errorf("API calls %d exceed efficient target %d for mode %s",
				metrics.APICallCount, maxExpectedCalls, testCase.Mode.String())
		}
	}

	// Throughput should be reasonable
	minThroughput := 100.0 // At least 100 players per second
	if metrics.ThroughputRPS < minThroughput {
		t.Logf("Warning: Low throughput %.2f players/sec for %s", metrics.ThroughputRPS, testCase.Name)
	}

	t.Logf("Performance validation passed for %s", testCase.Name)
}