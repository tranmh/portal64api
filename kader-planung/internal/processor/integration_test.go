package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/portal64/kader-planung/internal/api"
	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/metrics"
	"github.com/portal64/kader-planung/internal/resume"
	"github.com/sirupsen/logrus"
)

// IntegrationTestSuite manages comprehensive integration tests for the unified processor
type IntegrationTestSuite struct {
	logger           *logrus.Logger
	testDataDir      string
	metricsCollector *metrics.MetricsCollector
	testResults      map[string]*TestResult
}

// TestResult captures the results of an integration test
type TestResult struct {
	TestName         string                          `json:"test_name"`
	Success          bool                            `json:"success"`
	Duration         time.Duration                   `json:"duration"`
	ErrorMessage     string                          `json:"error_message,omitempty"`
	ProcessingResult *ProcessingResult               `json:"processing_result,omitempty"`
	Metrics          *metrics.DetailedMetrics    `json:"metrics,omitempty"`
	DataQuality      *DataQualityReport              `json:"data_quality,omitempty"`
}

// DataQualityReport provides data quality validation results
type DataQualityReport struct {
	TotalRecords           int     `json:"total_records"`
	ValidRecords           int     `json:"valid_records"`
	InvalidRecords         int     `json:"invalid_records"`
	MissingDataRecords     int     `json:"missing_data_records"`
	DataCompletenessRate   float64 `json:"data_completeness_rate"`
	DWZRangeValid          bool    `json:"dwz_range_valid"`
	AgeDistributionValid   bool    `json:"age_distribution_valid"`
	GenderDistributionValid bool   `json:"gender_distribution_valid"`
}

// NewIntegrationTestSuite creates a new integration test suite
func NewIntegrationTestSuite() *IntegrationTestSuite {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create test data directory
	testDataDir := filepath.Join("testdata", "integration")
	os.MkdirAll(testDataDir, 0755)

	return &IntegrationTestSuite{
		logger:           logger,
		testDataDir:      testDataDir,
		metricsCollector: metrics.NewMetricsCollector(logger),
		testResults:      make(map[string]*TestResult),
	}
}

// RunFullIntegrationSuite executes the complete integration test suite
func TestFullIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test suite in short mode")
	}

	suite := NewIntegrationTestSuite()

	// Test cases covering all processing modes and data sizes
	testCases := []struct {
		name        string
		dataSize    string
		mode        ProcessingMode
		playerCount int
		clubCount   int
		config      *UnifiedProcessorConfig
	}{
		{
			name:        "Small_Dataset_Statistical_Mode",
			dataSize:    "small",
			mode:        StatisticalMode,
			playerCount: 500,
			clubCount:   10,
			config: &UnifiedProcessorConfig{
				Mode:             StatisticalMode,
				MinSampleSize:    10,
				EnableStatistics: true,
				Concurrency:      4,
				Verbose:          false,
			},
		},
		{
			name:        "Small_Dataset_Efficient_Mode",
			dataSize:    "small",
			mode:        EfficientMode,
			playerCount: 500,
			clubCount:   10,
			config: &UnifiedProcessorConfig{
				Mode:             EfficientMode,
				MinSampleSize:    10,
				EnableStatistics: false,
				Concurrency:      4,
				Verbose:          false,
			},
		},
		{
			name:        "Small_Dataset_Hybrid_Mode",
			dataSize:    "small",
			mode:        HybridMode,
			playerCount: 500,
			clubCount:   10,
			config: &UnifiedProcessorConfig{
				Mode:             HybridMode,
				MinSampleSize:    10,
				EnableStatistics: true,
				Concurrency:      4,
				Verbose:          false,
			},
		},
		{
			name:        "Medium_Dataset_Statistical_Mode",
			dataSize:    "medium",
			mode:        StatisticalMode,
			playerCount: 5000,
			clubCount:   50,
			config: &UnifiedProcessorConfig{
				Mode:             StatisticalMode,
				MinSampleSize:    20,
				EnableStatistics: true,
				Concurrency:      8,
				Verbose:          false,
			},
		},
		{
			name:        "Medium_Dataset_Efficient_Mode",
			dataSize:    "medium",
			mode:        EfficientMode,
			playerCount: 5000,
			clubCount:   50,
			config: &UnifiedProcessorConfig{
				Mode:             EfficientMode,
				MinSampleSize:    20,
				EnableStatistics: false,
				Concurrency:      8,
				Verbose:          false,
			},
		},
		{
			name:        "Large_Dataset_Statistical_Mode",
			dataSize:    "large",
			mode:        StatisticalMode,
			playerCount: 20000,
			clubCount:   200,
			config: &UnifiedProcessorConfig{
				Mode:             StatisticalMode,
				MinSampleSize:    50,
				EnableStatistics: true,
				Concurrency:      16,
				Verbose:          false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := suite.runIntegrationTest(t, tc.name, tc.config, tc.playerCount, tc.clubCount)

			if !result.Success {
				t.Errorf("Integration test %s failed: %s", tc.name, result.ErrorMessage)
			}

			// Validate performance targets based on dataset size
			suite.validatePerformanceTargets(t, tc.dataSize, result)

			// Validate data quality
			suite.validateDataQuality(t, result)
		})
	}

	// Generate comprehensive test report
	suite.generateTestReport()
}

// runIntegrationTest executes a single integration test with full metrics collection
func (suite *IntegrationTestSuite) runIntegrationTest(t *testing.T, testName string, config *UnifiedProcessorConfig, playerCount, clubCount int) *TestResult {
	startTime := time.Now()
	ctx := context.Background()

	result := &TestResult{
		TestName: testName,
		Success:  false,
	}

	suite.logger.Infof("Starting integration test: %s", testName)

	// Start metrics collection
	suite.metricsCollector = metrics.NewMetricsCollector(suite.logger)
	suite.metricsCollector.StartCollection(ctx)
	suite.metricsCollector.SetConcurrencyLevel(config.Concurrency)
	suite.metricsCollector.SetMinSampleSize(config.MinSampleSize)
	suite.metricsCollector.SetStatisticsEnabled(config.EnableStatistics)

	// Generate test data
	testPlayers := generateScaledMockPlayerSet(playerCount, clubCount)
	suite.logger.Infof("Generated %d test players across %d clubs", len(testPlayers), clubCount)

	// Create mock API client
	mockClient := createMockAPIClientForIntegration(testPlayers)

	// Create resume manager (not used in tests but required)
	resumeManager := &resume.Manager{}

	// Record processing start
	suite.metricsCollector.RecordProcessingStart()

	// Create unified processor
	processor := NewUnifiedProcessor(mockClient, resumeManager, config)
	processor.SetLogger(suite.logger)

	// Estimate API call count
	suite.metricsCollector.EstimateAPICallCount(config.Mode.String(), clubCount, playerCount)

	defer func() {
		// Always stop metrics collection
		result.Metrics = suite.metricsCollector.StopCollection()
		result.Duration = time.Since(startTime)
	}()

	// Execute processing based on mode
	var processingResult *ProcessingResult
	var err error

	switch config.Mode {
	case StatisticalMode:
		processingResult, err = suite.runStatisticalModeTest(processor, testPlayers)
	case EfficientMode:
		processingResult, err = suite.runEfficientModeTest(processor, testPlayers)
	case HybridMode:
		processingResult, err = suite.runHybridModeTest(processor, testPlayers)
	case DetailedMode:
		processingResult, err = suite.runDetailedModeTest(processor, nil) // Uses resume.Checkpoint
	default:
		err = fmt.Errorf("unsupported processing mode: %s", config.Mode.String())
	}

	// Record processing end
	suite.metricsCollector.RecordProcessingEnd()

	if err != nil {
		result.ErrorMessage = err.Error()
		suite.metricsCollector.RecordError()
		suite.logger.Errorf("Integration test %s failed: %v", testName, err)
		return result
	}

	// Record processing result
	statisticalGroups := 0
	if processingResult.StatisticalData != nil {
		for _, data := range processingResult.StatisticalData {
			statisticalGroups += len(data.Percentiles)
		}
	}
	suite.metricsCollector.RecordProcessingResult(
		processingResult.Mode.String(),
		processingResult.TotalPlayers,
		processingResult.ProcessedClubs,
		len(processingResult.Records),
		statisticalGroups,
	)

	// Validate results
	dataQuality := suite.validateProcessingResults(processingResult, testPlayers)
	result.DataQuality = dataQuality

	result.ProcessingResult = processingResult
	result.Success = true

	suite.testResults[testName] = result
	suite.logger.Infof("Integration test %s completed successfully", testName)

	return result
}

// runStatisticalModeTest executes statistical mode processing
func (suite *IntegrationTestSuite) runStatisticalModeTest(processor *UnifiedProcessor, testPlayers []models.Player) (*ProcessingResult, error) {
	// Process players using statistical analyzer directly
	validPlayers := filterValidPlayersForStatistics(testPlayers)
	statisticalData, err := processor.statisticalAnalyzer.ProcessPlayers(validPlayers)
	if err != nil {
		return nil, fmt.Errorf("statistical processing failed: %w", err)
	}

	return &ProcessingResult{
		Mode:            StatisticalMode,
		Records:         nil, // No individual records in statistical mode
		StatisticalData: statisticalData,
		TotalPlayers:    len(validPlayers),
		ProcessedClubs:  countUniqueClubs(validPlayers),
		ProcessingTime:  0, // Will be set by metrics collector
	}, nil
}

// runEfficientModeTest executes efficient mode processing
func (suite *IntegrationTestSuite) runEfficientModeTest(processor *UnifiedProcessor, testPlayers []models.Player) (*ProcessingResult, error) {
	// Convert players to records without historical analysis
	records := processor.convertPlayersToRecords(testPlayers, false)
	processor.calculateListRankings(records)

	return &ProcessingResult{
		Mode:           EfficientMode,
		Records:        records,
		StatisticalData: nil,
		TotalPlayers:   len(testPlayers),
		ProcessedClubs: countUniqueClubs(testPlayers),
		ProcessingTime: 0, // Will be set by metrics collector
	}, nil
}

// runHybridModeTest executes hybrid mode processing
func (suite *IntegrationTestSuite) runHybridModeTest(processor *UnifiedProcessor, testPlayers []models.Player) (*ProcessingResult, error) {
	// Convert players to records
	records := processor.convertPlayersToRecords(testPlayers, false)
	processor.calculateListRankings(records)

	// Perform statistical analysis
	validPlayers := filterValidPlayersForStatistics(testPlayers)
	statisticalData, err := processor.statisticalAnalyzer.ProcessPlayers(validPlayers)
	if err != nil {
		suite.metricsCollector.RecordWarning()
		suite.logger.Warnf("Statistical analysis failed in hybrid mode: %v", err)
		statisticalData = nil
	}

	return &ProcessingResult{
		Mode:            HybridMode,
		Records:         records,
		StatisticalData: statisticalData,
		TotalPlayers:    len(testPlayers),
		ProcessedClubs:  countUniqueClubs(testPlayers),
		ProcessingTime:  0, // Will be set by metrics collector
	}, nil
}

// runDetailedModeTest executes detailed mode processing (legacy)
func (suite *IntegrationTestSuite) runDetailedModeTest(processor *UnifiedProcessor, checkpoint *resume.Checkpoint) (*ProcessingResult, error) {
	// For integration tests, we simulate detailed mode with efficient mode + historical data
	// In real usage, this would use the legacy processor with API calls

	// This is a simplified version for testing
	return &ProcessingResult{
		Mode:           DetailedMode,
		Records:        []models.KaderPlanungRecord{}, // Would be populated by legacy processor
		TotalPlayers:   0,
		ProcessedClubs: 0,
		ProcessingTime: time.Millisecond, // Minimal processing time for test
	}, nil
}

// validateProcessingResults validates the quality of processing results
func (suite *IntegrationTestSuite) validateProcessingResults(result *ProcessingResult, originalPlayers []models.Player) *DataQualityReport {
	report := &DataQualityReport{
		TotalRecords: len(originalPlayers),
	}

	// Validate records if present
	if result.Records != nil {
		report.ValidRecords = len(result.Records)

		// Check for missing or invalid data
		for _, record := range result.Records {
			if record.CurrentDWZ <= 0 || record.PlayerID == "" {
				report.InvalidRecords++
			}
			if record.Birthyear == 0 && record.DWZ12MonthsAgo == models.DataNotAvailable {
				report.MissingDataRecords++
			}
		}
	}

	// Validate statistical data if present
	if result.StatisticalData != nil {
		for gender, data := range result.StatisticalData {
			suite.logger.Infof("Statistical data for gender %s: %d players, %d age groups",
				gender, data.Metadata.TotalPlayers, data.Metadata.ValidAgeGroups)
		}
	}

	// Calculate completeness rate
	if report.TotalRecords > 0 {
		report.DataCompletenessRate = float64(report.ValidRecords) / float64(report.TotalRecords) * 100
	}

	// Validate DWZ range (realistic chess ratings: 500-3000)
	report.DWZRangeValid = true
	if result.Records != nil {
		for _, record := range result.Records {
			if record.CurrentDWZ < 500 || record.CurrentDWZ > 3000 {
				report.DWZRangeValid = false
				break
			}
		}
	}

	// Validate age distribution (reasonable age range: 5-100)
	report.AgeDistributionValid = true
	currentYear := time.Now().Year()
	for _, player := range originalPlayers {
		if player.BirthYear != nil {
			age := currentYear - *player.BirthYear
			if age < 5 || age > 100 {
				report.AgeDistributionValid = false
				break
			}
		}
	}

	// Validate gender distribution (should have both male and female players)
	maleCount := 0
	femaleCount := 0
	for _, player := range originalPlayers {
		switch player.Gender {
		case "m":
			maleCount++
		case "w":
			femaleCount++
		}
	}
	report.GenderDistributionValid = maleCount > 0 && femaleCount > 0

	return report
}

// validatePerformanceTargets validates performance against Phase 5 targets
func (suite *IntegrationTestSuite) validatePerformanceTargets(t *testing.T, dataSize string, result *TestResult) {
	if result.Metrics == nil {
		t.Errorf("No metrics available for performance validation")
		return
	}

	metrics := result.Metrics

	// Performance targets based on Phase 5 requirements
	switch dataSize {
	case "small":
		// Small dataset should complete very quickly
		maxDuration := 30 * time.Second
		if metrics.TotalDuration > maxDuration {
			t.Errorf("Small dataset took %v, expected <%v", metrics.TotalDuration, maxDuration)
		}

		// Memory should be minimal
		if metrics.PeakMemoryUsageMB > 100 {
			t.Errorf("Small dataset used %d MB memory, expected <100 MB", metrics.PeakMemoryUsageMB)
		}

	case "medium":
		// Medium dataset targets
		maxDuration := 2 * time.Minute
		if metrics.TotalDuration > maxDuration {
			t.Errorf("Medium dataset took %v, expected <%v", metrics.TotalDuration, maxDuration)
		}

		if metrics.PeakMemoryUsageMB > 500 {
			t.Errorf("Medium dataset used %d MB memory, expected <500 MB", metrics.PeakMemoryUsageMB)
		}

	case "large":
		// Large dataset targets (but not 50K target from Phase 5)
		maxDuration := 5 * time.Minute
		if metrics.TotalDuration > maxDuration {
			t.Errorf("Large dataset took %v, expected <%v", metrics.TotalDuration, maxDuration)
		}

		if metrics.PeakMemoryUsageMB > 1024 {
			t.Errorf("Large dataset used %d MB memory, expected <1024 MB", metrics.PeakMemoryUsageMB)
		}
	}

	// General throughput target
	minThroughput := 50.0 // At least 50 players per second
	if metrics.PlayersPerSecond < minThroughput {
		t.Logf("Warning: Low throughput %.2f players/sec for %s dataset", metrics.PlayersPerSecond, dataSize)
	}

	// Success rate should be high
	if metrics.ProcessingSuccessRate < 95.0 {
		t.Errorf("Low success rate %.2f%%, expected >95%%", metrics.ProcessingSuccessRate)
	}
}

// validateDataQuality validates data quality metrics
func (suite *IntegrationTestSuite) validateDataQuality(t *testing.T, result *TestResult) {
	if result.DataQuality == nil {
		t.Errorf("No data quality report available")
		return
	}

	quality := result.DataQuality

	// Data completeness should be high
	if quality.DataCompletenessRate < 90.0 {
		t.Errorf("Low data completeness rate %.2f%%, expected >90%%", quality.DataCompletenessRate)
	}

	// DWZ values should be in valid range
	if !quality.DWZRangeValid {
		t.Errorf("DWZ values outside valid range detected")
	}

	// Age distribution should be valid
	if !quality.AgeDistributionValid {
		t.Errorf("Invalid age distribution detected")
	}

	// Gender distribution should be balanced
	if !quality.GenderDistributionValid {
		t.Errorf("Unbalanced gender distribution detected")
	}
}

// generateTestReport creates a comprehensive test report
func (suite *IntegrationTestSuite) generateTestReport() {
	suite.logger.Info("=== Integration Test Suite Report ===")

	totalTests := len(suite.testResults)
	successfulTests := 0

	for testName, result := range suite.testResults {
		if result.Success {
			successfulTests++
		}

		suite.logger.Infof("Test: %s", testName)
		suite.logger.Infof("  Success: %t", result.Success)
		suite.logger.Infof("  Duration: %v", result.Duration)

		if result.Metrics != nil {
			suite.logger.Infof("  Players Processed: %d", result.Metrics.TotalPlayersProcessed)
			suite.logger.Infof("  Peak Memory: %d MB", result.Metrics.PeakMemoryUsageMB)
			suite.logger.Infof("  Throughput: %.2f players/sec", result.Metrics.PlayersPerSecond)
		}

		if result.DataQuality != nil {
			suite.logger.Infof("  Data Quality: %.2f%%", result.DataQuality.DataCompletenessRate)
		}

		if !result.Success {
			suite.logger.Errorf("  Error: %s", result.ErrorMessage)
		}

		suite.logger.Info("  ---")
	}

	successRate := float64(successfulTests) / float64(totalTests) * 100
	suite.logger.Infof("Overall Success Rate: %.2f%% (%d/%d tests passed)", successRate, successfulTests, totalTests)

	// Export detailed report to JSON
	reportFile := filepath.Join(suite.testDataDir, fmt.Sprintf("integration_test_report_%s.json", time.Now().Format("20060102_150405")))
	suite.exportTestReport(reportFile)
}

// exportTestReport exports the test results to JSON
func (suite *IntegrationTestSuite) exportTestReport(filename string) {
	// Implementation would marshal suite.testResults to JSON and write to file
	suite.logger.Infof("Test report would be exported to: %s", filename)
}

// Helper functions

// createMockAPIClientForIntegration creates a mock API client for integration testing
func createMockAPIClientForIntegration(players []models.Player) *api.Client {
	// This would be a more sophisticated mock that implements the Client interface
	// For now, return a basic client (in real implementation, this would be properly mocked)
	return &api.Client{}
}

// filterValidPlayersForStatistics filters players suitable for statistical analysis
func filterValidPlayersForStatistics(players []models.Player) []models.Player {
	var validPlayers []models.Player
	for _, player := range players {
		if player.BirthYear != nil && player.CurrentDWZ > 0 && player.Active && player.Status == "active" {
			validPlayers = append(validPlayers, player)
		}
	}
	return validPlayers
}

// countUniqueClubs counts unique clubs in player list
func countUniqueClubs(players []models.Player) int {
	clubs := make(map[string]bool)
	for _, player := range players {
		clubs[player.ClubID] = true
	}
	return len(clubs)
}