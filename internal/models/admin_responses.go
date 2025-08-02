package models

// CacheStatsResponse represents the response structure for cache statistics
type CacheStatsResponse struct {
	Performance CachePerformanceStats `json:"performance"`
	Usage       CacheUsageStats       `json:"usage"`
	Operations  CacheOperationsStats  `json:"operations"`
	Errors      CacheErrorStats       `json:"errors"`
}

// CachePerformanceStats represents cache performance metrics
type CachePerformanceStats struct {
	HitRatio          float64 `json:"hit_ratio"`
	AvgResponseTimeMs int64   `json:"avg_response_time_ms"`
	TotalRequests     int64   `json:"total_requests"`
	CacheHits         int64   `json:"cache_hits"`
	CacheMisses       int64   `json:"cache_misses"`
}

// CacheUsageStats represents cache usage metrics
type CacheUsageStats struct {
	MemoryUsedMB        int64 `json:"memory_used_mb"`
	KeyCount           int64 `json:"key_count"`
	ActiveConnections  int   `json:"active_connections"`
	IdleConnections    int   `json:"idle_connections"`
}

// CacheOperationsStats represents cache operation metrics
type CacheOperationsStats struct {
	CacheOperations     int64 `json:"cache_operations"`
	BackgroundRefreshes int64 `json:"background_refreshes"`
}

// CacheErrorStats represents cache error metrics
type CacheErrorStats struct {
	CacheErrors   int64 `json:"cache_errors"`
	RefreshErrors int64 `json:"refresh_errors"`
}

// CacheHealthResponse represents the response structure for cache health check
type CacheHealthResponse struct {
	Status string                 `json:"status"`
	Redis  string                 `json:"redis"`
	Stats  CacheHealthStats       `json:"stats,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// CacheHealthStats represents simplified stats for health check
type CacheHealthStats struct {
	HitRatio      float64 `json:"hit_ratio"`
	TotalRequests int64   `json:"total_requests"`
	KeyCount      int64   `json:"key_count"`
}
