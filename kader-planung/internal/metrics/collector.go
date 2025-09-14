package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsCollector handles collection and reporting of performance metrics
type MetricsCollector struct {
	logger       *logrus.Logger
	startTime    time.Time
	endTime      time.Time
	metrics      *DetailedMetrics
	memoryStats  []MemorySnapshot
	mutex        sync.RWMutex
	isCollecting bool
	stopChan     chan struct{}
}

// DetailedMetrics provides comprehensive performance metrics
type DetailedMetrics struct {
	// Execution metrics
	TotalDuration       time.Duration `json:"total_duration"`
	PreProcessingTime   time.Duration `json:"pre_processing_time"`
	ProcessingTime      time.Duration `json:"processing_time"`
	PostProcessingTime  time.Duration `json:"post_processing_time"`

	// Resource usage
	PeakMemoryUsageMB   int64   `json:"peak_memory_usage_mb"`
	AvgMemoryUsageMB    float64 `json:"avg_memory_usage_mb"`
	CPUUsagePercent     float64 `json:"cpu_usage_percent"`
	GoroutineCount      int     `json:"goroutine_count"`

	// Processing metrics
	TotalPlayersProcessed    int     `json:"total_players_processed"`
	TotalClubsProcessed      int     `json:"total_clubs_processed"`
	PlayersPerSecond         float64 `json:"players_per_second"`
	RecordsGenerated         int     `json:"records_generated"`
	StatisticalGroupsGenerated int   `json:"statistical_groups_generated"`

	// API and network metrics
	EstimatedAPICallCount    int     `json:"estimated_api_call_count"`
	APICalls95thPercentile   float64 `json:"api_calls_95th_percentile_ms"`
	NetworkLatencyAvgMs      float64 `json:"network_latency_avg_ms"`

	// Quality metrics
	ErrorCount               int     `json:"error_count"`
	WarningCount            int     `json:"warning_count"`
	DataQualityScore        float64 `json:"data_quality_score"`
	ProcessingSuccessRate   float64 `json:"processing_success_rate"`

	// Mode-specific metrics
	ProcessingMode          string  `json:"processing_mode"`
	ConcurrencyLevel        int     `json:"concurrency_level"`
	MinSampleSize           int     `json:"min_sample_size"`
	StatisticsEnabled       bool    `json:"statistics_enabled"`

	// System information
	SystemInfo              SystemInfo      `json:"system_info"`
	MemorySnapshots         []MemorySnapshot `json:"memory_snapshots"`
	Timestamp               time.Time       `json:"timestamp"`
}

// SystemInfo captures system configuration information
type SystemInfo struct {
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	CPUCount        int    `json:"cpu_count"`
	GoVersion       string `json:"go_version"`
	MaxMemoryMB     int64  `json:"max_memory_mb"`
}

// MemorySnapshot captures memory usage at a specific point in time
type MemorySnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	AllocMB      int64     `json:"alloc_mb"`
	TotalAllocMB int64     `json:"total_alloc_mb"`
	SysMB        int64     `json:"sys_mb"`
	NumGC        uint32    `json:"num_gc"`
	GoroutineCount int     `json:"goroutine_count"`
}

// NewMetricsCollector creates a new performance metrics collector
func NewMetricsCollector(logger *logrus.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:       logger,
		metrics:      &DetailedMetrics{},
		memoryStats:  make([]MemorySnapshot, 0),
		stopChan:     make(chan struct{}),
		isCollecting: false,
	}
}

// StartCollection begins performance metrics collection
func (mc *MetricsCollector) StartCollection(ctx context.Context) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.isCollecting {
		mc.logger.Warn("Metrics collection already in progress")
		return
	}

	mc.startTime = time.Now()
	mc.isCollecting = true
	mc.metrics.Timestamp = mc.startTime
	mc.metrics.SystemInfo = mc.getSystemInfo()

	mc.logger.Info("Starting performance metrics collection")

	// Start memory monitoring goroutine
	go mc.monitorMemoryUsage(ctx)
}

// StopCollection ends performance metrics collection and finalizes metrics
func (mc *MetricsCollector) StopCollection() *DetailedMetrics {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.isCollecting {
		mc.logger.Warn("Metrics collection was not started")
		return mc.metrics
	}

	mc.endTime = time.Now()
	mc.isCollecting = false

	// Signal monitoring goroutine to stop
	close(mc.stopChan)

	// Finalize metrics
	mc.finalizeMetrics()

	mc.logger.Info("Performance metrics collection completed")
	return mc.metrics
}

// RecordProcessingStart marks the start of main processing
func (mc *MetricsCollector) RecordProcessingStart() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics.PreProcessingTime = time.Since(mc.startTime)
}

// RecordProcessingEnd marks the end of main processing
func (mc *MetricsCollector) RecordProcessingEnd() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics.ProcessingTime = time.Since(mc.startTime) - mc.metrics.PreProcessingTime
}

// RecordProcessingResult records the results of processing
func (mc *MetricsCollector) RecordProcessingResult(processingMode string, totalPlayers, totalClubs, recordsGenerated, statisticalGroups int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics.ProcessingMode = processingMode
	mc.metrics.TotalPlayersProcessed = totalPlayers
	mc.metrics.TotalClubsProcessed = totalClubs
	mc.metrics.RecordsGenerated = recordsGenerated
	mc.metrics.StatisticalGroupsGenerated = statisticalGroups
}

// RecordError increments error count
func (mc *MetricsCollector) RecordError() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.metrics.ErrorCount++
}

// RecordWarning increments warning count
func (mc *MetricsCollector) RecordWarning() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.metrics.WarningCount++
}

// SetConcurrencyLevel records the concurrency level used
func (mc *MetricsCollector) SetConcurrencyLevel(level int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.metrics.ConcurrencyLevel = level
}

// SetMinSampleSize records the minimum sample size used
func (mc *MetricsCollector) SetMinSampleSize(size int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.metrics.MinSampleSize = size
}

// SetStatisticsEnabled records whether statistics were enabled
func (mc *MetricsCollector) SetStatisticsEnabled(enabled bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.metrics.StatisticsEnabled = enabled
}

// EstimateAPICallCount estimates the number of API calls for the given mode
func (mc *MetricsCollector) EstimateAPICallCount(mode string, clubs, players int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	switch mode {
	case "statistical", "efficient", "hybrid":
		// Somatogramm-style efficient: 1 + N API calls (where N = clubs)
		mc.metrics.EstimatedAPICallCount = 1 + clubs
	case "detailed":
		// Traditional Kader-Planung: 1 + N + 2*N*P (where N=clubs, P=avg players per club)
		avgPlayersPerClub := 1
		if clubs > 0 {
			avgPlayersPerClub = players / clubs
		}
		mc.metrics.EstimatedAPICallCount = 1 + clubs + (2 * clubs * avgPlayersPerClub)
	default:
		mc.metrics.EstimatedAPICallCount = clubs + players
	}
}

// monitorMemoryUsage continuously monitors memory usage during processing
func (mc *MetricsCollector) monitorMemoryUsage(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mc.stopChan:
			return
		case <-ticker.C:
			mc.captureMemorySnapshot()
		}
	}
}

// captureMemorySnapshot takes a snapshot of current memory usage
func (mc *MetricsCollector) captureMemorySnapshot() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := MemorySnapshot{
		Timestamp:      time.Now(),
		AllocMB:        int64(m.Alloc) / (1024 * 1024),
		TotalAllocMB:   int64(m.TotalAlloc) / (1024 * 1024),
		SysMB:          int64(m.Sys) / (1024 * 1024),
		NumGC:          m.NumGC,
		GoroutineCount: runtime.NumGoroutine(),
	}

	mc.mutex.Lock()
	mc.memoryStats = append(mc.memoryStats, snapshot)
	mc.mutex.Unlock()
}

// finalizeMetrics calculates final derived metrics
func (mc *MetricsCollector) finalizeMetrics() {
	mc.metrics.TotalDuration = mc.endTime.Sub(mc.startTime)
	mc.metrics.PostProcessingTime = mc.metrics.TotalDuration - mc.metrics.PreProcessingTime - mc.metrics.ProcessingTime

	// Calculate throughput
	if mc.metrics.TotalDuration.Seconds() > 0 {
		mc.metrics.PlayersPerSecond = float64(mc.metrics.TotalPlayersProcessed) / mc.metrics.TotalDuration.Seconds()
	}

	// Calculate memory statistics
	mc.calculateMemoryStatistics()

	// Calculate processing success rate
	totalOperations := mc.metrics.TotalPlayersProcessed
	if totalOperations > 0 {
		mc.metrics.ProcessingSuccessRate = float64(totalOperations-mc.metrics.ErrorCount) / float64(totalOperations) * 100
	} else {
		mc.metrics.ProcessingSuccessRate = 100.0
	}

	// Calculate data quality score (simplified)
	mc.metrics.DataQualityScore = mc.calculateDataQualityScore()

	// Copy memory snapshots
	mc.metrics.MemorySnapshots = make([]MemorySnapshot, len(mc.memoryStats))
	copy(mc.metrics.MemorySnapshots, mc.memoryStats)
}

// calculateMemoryStatistics computes memory usage statistics
func (mc *MetricsCollector) calculateMemoryStatistics() {
	if len(mc.memoryStats) == 0 {
		return
	}

	var peak int64
	var total int64
	var count int64

	for _, snapshot := range mc.memoryStats {
		if snapshot.AllocMB > peak {
			peak = snapshot.AllocMB
		}
		total += snapshot.AllocMB
		count++
	}

	mc.metrics.PeakMemoryUsageMB = peak
	if count > 0 {
		mc.metrics.AvgMemoryUsageMB = float64(total) / float64(count)
	}

	// Get final goroutine count
	if len(mc.memoryStats) > 0 {
		mc.metrics.GoroutineCount = mc.memoryStats[len(mc.memoryStats)-1].GoroutineCount
	}
}

// calculateDataQualityScore computes a simple data quality score
func (mc *MetricsCollector) calculateDataQualityScore() float64 {
	// Simple scoring based on error/warning rates
	totalIssues := mc.metrics.ErrorCount + mc.metrics.WarningCount
	totalProcessed := mc.metrics.TotalPlayersProcessed

	if totalProcessed == 0 {
		return 100.0
	}

	issueRate := float64(totalIssues) / float64(totalProcessed)
	qualityScore := (1.0 - issueRate) * 100.0

	if qualityScore < 0 {
		qualityScore = 0
	}
	if qualityScore > 100 {
		qualityScore = 100
	}

	return qualityScore
}

// getSystemInfo collects system information
func (mc *MetricsCollector) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUCount:     runtime.NumCPU(),
		GoVersion:    runtime.Version(),
		MaxMemoryMB:  int64(m.Sys) / (1024 * 1024),
	}
}

// ExportToJSON exports metrics to JSON format
func (mc *MetricsCollector) ExportToJSON(filename string) error {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	data, err := json.MarshalIndent(mc.metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics to JSON: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metrics to file: %w", err)
	}

	mc.logger.Infof("Performance metrics exported to %s", filename)
	return nil
}

// PrintSummary prints a human-readable summary of the metrics
func (mc *MetricsCollector) PrintSummary() {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	m := mc.metrics

	mc.logger.Info("=== Performance Metrics Summary ===")
	mc.logger.Infof("Processing Mode: %s", m.ProcessingMode)
	mc.logger.Infof("Total Duration: %v", m.TotalDuration)
	mc.logger.Infof("Processing Time: %v", m.ProcessingTime)
	mc.logger.Infof("Players Processed: %d", m.TotalPlayersProcessed)
	mc.logger.Infof("Clubs Processed: %d", m.TotalClubsProcessed)
	mc.logger.Infof("Throughput: %.2f players/sec", m.PlayersPerSecond)
	mc.logger.Infof("Peak Memory: %d MB", m.PeakMemoryUsageMB)
	mc.logger.Infof("Avg Memory: %.2f MB", m.AvgMemoryUsageMB)
	mc.logger.Infof("Est. API Calls: %d", m.EstimatedAPICallCount)
	mc.logger.Infof("Success Rate: %.2f%%", m.ProcessingSuccessRate)
	mc.logger.Infof("Data Quality Score: %.2f", m.DataQualityScore)
	mc.logger.Infof("Errors: %d, Warnings: %d", m.ErrorCount, m.WarningCount)

	if m.StatisticsEnabled {
		mc.logger.Infof("Statistical Groups: %d", m.StatisticalGroupsGenerated)
	}

	mc.logger.Infof("System: %s/%s, %d CPUs", m.SystemInfo.OS, m.SystemInfo.Architecture, m.SystemInfo.CPUCount)
}

// GetMetrics returns the current metrics (thread-safe)
func (mc *MetricsCollector) GetMetrics() *DetailedMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	// Return a copy to prevent external modifications
	metricsCopy := *mc.metrics
	return &metricsCopy
}

// CompareMetrics compares two sets of metrics and returns a comparison report
func CompareMetrics(baseline, current *DetailedMetrics) *MetricsComparison {
	return &MetricsComparison{
		BaselineTimestamp: baseline.Timestamp,
		CurrentTimestamp:  current.Timestamp,

		DurationChange:    current.TotalDuration - baseline.TotalDuration,
		ThroughputChange:  current.PlayersPerSecond - baseline.PlayersPerSecond,
		MemoryChange:      current.PeakMemoryUsageMB - baseline.PeakMemoryUsageMB,

		DurationChangePercent:  calculatePercentChange(float64(baseline.TotalDuration), float64(current.TotalDuration)),
		ThroughputChangePercent: calculatePercentChange(baseline.PlayersPerSecond, current.PlayersPerSecond),
		MemoryChangePercent:    calculatePercentChange(float64(baseline.PeakMemoryUsageMB), float64(current.PeakMemoryUsageMB)),

		QualityScoreChange: current.DataQualityScore - baseline.DataQualityScore,
		ErrorCountChange:   current.ErrorCount - baseline.ErrorCount,
	}
}

// MetricsComparison provides a comparison between two metric sets
type MetricsComparison struct {
	BaselineTimestamp time.Time `json:"baseline_timestamp"`
	CurrentTimestamp  time.Time `json:"current_timestamp"`

	DurationChange    time.Duration `json:"duration_change"`
	ThroughputChange  float64       `json:"throughput_change"`
	MemoryChange      int64         `json:"memory_change_mb"`

	DurationChangePercent  float64 `json:"duration_change_percent"`
	ThroughputChangePercent float64 `json:"throughput_change_percent"`
	MemoryChangePercent    float64 `json:"memory_change_percent"`

	QualityScoreChange float64 `json:"quality_score_change"`
	ErrorCountChange   int     `json:"error_count_change"`
}

// calculatePercentChange calculates percentage change between two values
func calculatePercentChange(baseline, current float64) float64 {
	if baseline == 0 {
		if current == 0 {
			return 0
		}
		return 100 // Treat as 100% change if baseline was zero
	}
	return ((current - baseline) / baseline) * 100
}

// PrintComparison prints a human-readable metrics comparison
func (mc *MetricsComparison) PrintComparison(logger *logrus.Logger) {
	logger.Info("=== Metrics Comparison ===")
	logger.Infof("Baseline: %v", mc.BaselineTimestamp.Format("2006-01-02 15:04:05"))
	logger.Infof("Current:  %v", mc.CurrentTimestamp.Format("2006-01-02 15:04:05"))
	logger.Infof("Duration Change: %v (%.2f%%)", mc.DurationChange, mc.DurationChangePercent)
	logger.Infof("Throughput Change: %.2f players/sec (%.2f%%)", mc.ThroughputChange, mc.ThroughputChangePercent)
	logger.Infof("Memory Change: %d MB (%.2f%%)", mc.MemoryChange, mc.MemoryChangePercent)
	logger.Infof("Quality Score Change: %.2f", mc.QualityScoreChange)
	logger.Infof("Error Count Change: %d", mc.ErrorCountChange)
}