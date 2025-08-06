package cache

import (
	"context"
	"fmt"
	"time"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	// Basic cache operations
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	FlushAll(ctx context.Context) error
	
	// Background refresh operations
	GetWithRefresh(ctx context.Context, key string, dest interface{}, 
		refreshFunc func() (interface{}, error), ttl time.Duration) error
	
	// Batch operations
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
	MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	
	// Health and metrics
	Ping(ctx context.Context) error
	GetStats() CacheStats
	
	// Lifecycle
	Close() error
}

// CacheStats holds cache performance metrics
type CacheStats struct {
	// Hit/Miss Statistics
	TotalRequests    int64   `json:"total_requests"`
	CacheHits        int64   `json:"cache_hits"`
	CacheMisses      int64   `json:"cache_misses"`
	HitRatio         float64 `json:"hit_ratio"`
	
	// Performance Metrics
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	CacheOperations  int64         `json:"cache_operations"`
	BackgroundRefreshes int64      `json:"background_refreshes"`
	
	// Error Tracking
	CacheErrors      int64 `json:"cache_errors"`
	RefreshErrors    int64 `json:"refresh_errors"`
	
	// Memory Usage
	MemoryUsed       int64 `json:"memory_used"`
	KeyCount         int64 `json:"key_count"`
	
	// Connection Stats
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
}

// CacheError represents cache-specific errors
type CacheError struct {
	Operation string
	Key       string
	Err       error
}

func (e CacheError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("cache %s failed for key '%s': %v", e.Operation, e.Key, e.Err)
	}
	return fmt.Sprintf("cache %s failed: %v", e.Operation, e.Err)
}

// Error types
var (
	ErrCacheNotEnabled = &CacheError{Operation: "access", Err: fmt.Errorf("cache is not enabled")}
	ErrKeyNotFound     = &CacheError{Operation: "get", Err: fmt.Errorf("key not found")}
	ErrSerialization   = &CacheError{Operation: "serialize", Err: fmt.Errorf("serialization failed")}
	ErrDeserialization = &CacheError{Operation: "deserialize", Err: fmt.Errorf("deserialization failed")}
)
