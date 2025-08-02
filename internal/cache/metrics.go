package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects and tracks cache performance metrics
type MetricsCollector struct {
	mu sync.RWMutex
	
	// Atomic counters for thread-safe operations
	totalRequests       int64
	cacheHits          int64
	cacheMisses        int64
	cacheOperations    int64
	backgroundRefreshes int64
	cacheErrors        int64
	refreshErrors      int64
	
	// Response time tracking
	responseTimes      []time.Duration
	responseTimeSum    int64 // nanoseconds
	responseTimeCount  int64
	
	// Memory and connection stats (updated periodically)
	memoryUsed         int64
	keyCount          int64
	activeConnections  int
	idleConnections   int
	
	// Configuration
	maxResponseTimesSamples int
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		maxResponseTimesSamples: 1000, // Keep last 1000 response times
		responseTimes:          make([]time.Duration, 0, 1000),
	}
}

// RecordRequest records a cache request
func (mc *MetricsCollector) RecordRequest() {
	atomic.AddInt64(&mc.totalRequests, 1)
}

// RecordHit records a cache hit
func (mc *MetricsCollector) RecordHit() {
	atomic.AddInt64(&mc.cacheHits, 1)
}

// RecordMiss records a cache miss
func (mc *MetricsCollector) RecordMiss() {
	atomic.AddInt64(&mc.cacheMisses, 1)
}

// RecordOperation records a cache operation
func (mc *MetricsCollector) RecordOperation() {
	atomic.AddInt64(&mc.cacheOperations, 1)
}

// RecordBackgroundRefresh records a background refresh
func (mc *MetricsCollector) RecordBackgroundRefresh() {
	atomic.AddInt64(&mc.backgroundRefreshes, 1)
}

// RecordError records a cache error
func (mc *MetricsCollector) RecordError() {
	atomic.AddInt64(&mc.cacheErrors, 1)
}

// RecordRefreshError records a background refresh error
func (mc *MetricsCollector) RecordRefreshError() {
	atomic.AddInt64(&mc.refreshErrors, 1)
}

// RecordResponseTime records a response time
func (mc *MetricsCollector) RecordResponseTime(duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Add to response times slice (with size limit)
	if len(mc.responseTimes) >= mc.maxResponseTimesSamples {
		// Remove oldest, add newest (sliding window)
		mc.responseTimes = mc.responseTimes[1:]
	}
	mc.responseTimes = append(mc.responseTimes, duration)
	
	// Update running average using atomic operations
	atomic.AddInt64(&mc.responseTimeSum, int64(duration))
	atomic.AddInt64(&mc.responseTimeCount, 1)
}

// UpdateMemoryStats updates memory-related statistics
func (mc *MetricsCollector) UpdateMemoryStats(memoryUsed, keyCount int64) {
	atomic.StoreInt64(&mc.memoryUsed, memoryUsed)
	atomic.StoreInt64(&mc.keyCount, keyCount)
}

// UpdateConnectionStats updates connection statistics
func (mc *MetricsCollector) UpdateConnectionStats(active, idle int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.activeConnections = active
	mc.idleConnections = idle
}

// GetStats returns current cache statistics
func (mc *MetricsCollector) GetStats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	totalReq := atomic.LoadInt64(&mc.totalRequests)
	hits := atomic.LoadInt64(&mc.cacheHits)
	misses := atomic.LoadInt64(&mc.cacheMisses)
	
	// Calculate hit ratio
	var hitRatio float64
	if totalReq > 0 {
		hitRatio = float64(hits) / float64(totalReq)
	}
	
	// Calculate average response time
	var avgResponseTime time.Duration
	timeCount := atomic.LoadInt64(&mc.responseTimeCount)
	if timeCount > 0 {
		avgNanos := atomic.LoadInt64(&mc.responseTimeSum) / timeCount
		avgResponseTime = time.Duration(avgNanos)
	}
	
	return CacheStats{
		TotalRequests:       totalReq,
		CacheHits:          hits,
		CacheMisses:        misses,
		HitRatio:           hitRatio,
		AvgResponseTime:    avgResponseTime,
		CacheOperations:    atomic.LoadInt64(&mc.cacheOperations),
		BackgroundRefreshes: atomic.LoadInt64(&mc.backgroundRefreshes),
		CacheErrors:        atomic.LoadInt64(&mc.cacheErrors),
		RefreshErrors:      atomic.LoadInt64(&mc.refreshErrors),
		MemoryUsed:         atomic.LoadInt64(&mc.memoryUsed),
		KeyCount:           atomic.LoadInt64(&mc.keyCount),
		ActiveConnections:  mc.activeConnections,
		IdleConnections:    mc.idleConnections,
	}
}

// Reset resets all metrics (useful for testing)
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	atomic.StoreInt64(&mc.totalRequests, 0)
	atomic.StoreInt64(&mc.cacheHits, 0)
	atomic.StoreInt64(&mc.cacheMisses, 0)
	atomic.StoreInt64(&mc.cacheOperations, 0)
	atomic.StoreInt64(&mc.backgroundRefreshes, 0)
	atomic.StoreInt64(&mc.cacheErrors, 0)
	atomic.StoreInt64(&mc.refreshErrors, 0)
	atomic.StoreInt64(&mc.responseTimeSum, 0)
	atomic.StoreInt64(&mc.responseTimeCount, 0)
	atomic.StoreInt64(&mc.memoryUsed, 0)
	atomic.StoreInt64(&mc.keyCount, 0)
	
	mc.responseTimes = mc.responseTimes[:0]
	mc.activeConnections = 0
	mc.idleConnections = 0
}
