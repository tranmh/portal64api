package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MockCacheService provides an in-memory cache implementation for testing
type MockCacheService struct {
	mu           sync.RWMutex
	data         map[string]cacheItem
	metrics      *MetricsCollector
	enabled      bool
	keyGenerator *KeyGenerator
	
	// Test configuration
	simulateErrors   bool
	errorRate       float64 // 0.0 to 1.0
	simulateLatency bool
	latency         time.Duration
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

// NewMockCacheService creates a new mock cache service
func NewMockCacheService(enabled bool) *MockCacheService {
	return &MockCacheService{
		data:         make(map[string]cacheItem),
		metrics:      NewMetricsCollector(),
		enabled:      enabled,
		keyGenerator: NewKeyGenerator(),
	}
}

// SetTestOptions configures the mock for testing scenarios
func (mcs *MockCacheService) SetTestOptions(simulateErrors bool, errorRate float64, simulateLatency bool, latency time.Duration) {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()
	
	mcs.simulateErrors = simulateErrors
	mcs.errorRate = errorRate
	mcs.simulateLatency = simulateLatency
	mcs.latency = latency
}

// Get retrieves a value from the mock cache
func (mcs *MockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	defer func() {
		mcs.metrics.RecordResponseTime(time.Since(start))
	}()
	
	mcs.metrics.RecordRequest()
	mcs.metrics.RecordOperation()
	
	if mcs.simulateLatency {
		time.Sleep(mcs.latency)
	}
	
	if mcs.simulateErrors && mcs.shouldSimulateError() {
		mcs.metrics.RecordError()
		return &CacheError{Operation: "get", Key: key, Err: fmt.Errorf("simulated error")}
	}
	
	if !mcs.enabled {
		mcs.metrics.RecordMiss()
		return ErrCacheNotEnabled
	}
	
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()
	
	item, exists := mcs.data[key]
	if !exists {
		mcs.metrics.RecordMiss()
		return &CacheError{Operation: "get", Key: key, Err: ErrKeyNotFound.Err}
	}
	
	// Check expiration
	if time.Now().After(item.expiresAt) {
		mcs.metrics.RecordMiss()
		delete(mcs.data, key) // Clean up expired item
		return &CacheError{Operation: "get", Key: key, Err: ErrKeyNotFound.Err}
	}
	
	// Deserialize
	if err := json.Unmarshal(item.value, dest); err != nil {
		mcs.metrics.RecordError()
		return &CacheError{Operation: "get", Key: key, Err: ErrDeserialization.Err}
	}
	
	mcs.metrics.RecordHit()
	return nil
}

// Set stores a value in the mock cache
func (mcs *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		mcs.metrics.RecordResponseTime(time.Since(start))
	}()
	
	mcs.metrics.RecordOperation()
	
	if mcs.simulateLatency {
		time.Sleep(mcs.latency)
	}
	
	if mcs.simulateErrors && mcs.shouldSimulateError() {
		mcs.metrics.RecordError()
		return &CacheError{Operation: "set", Key: key, Err: fmt.Errorf("simulated error")}
	}
	
	if !mcs.enabled {
		return ErrCacheNotEnabled
	}
	
	// Serialize
	data, err := json.Marshal(value)
	if err != nil {
		mcs.metrics.RecordError()
		return &CacheError{Operation: "set", Key: key, Err: ErrSerialization.Err}
	}
	
	mcs.mu.Lock()
	defer mcs.mu.Unlock()
	
	mcs.data[key] = cacheItem{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}
	
	return nil
}

// Delete removes a key from the mock cache
func (mcs *MockCacheService) Delete(ctx context.Context, key string) error {
	mcs.metrics.RecordOperation()
	
	if !mcs.enabled {
		return ErrCacheNotEnabled
	}
	
	mcs.mu.Lock()
	defer mcs.mu.Unlock()
	
	delete(mcs.data, key)
	return nil
}

// Exists checks if a key exists in the mock cache
func (mcs *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	mcs.metrics.RecordOperation()
	
	if !mcs.enabled {
		return false, ErrCacheNotEnabled
	}
	
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()
	
	item, exists := mcs.data[key]
	if !exists {
		return false, nil
	}
	
	// Check expiration
	if time.Now().After(item.expiresAt) {
		delete(mcs.data, key) // Clean up expired item
		return false, nil
	}
	
	return true, nil
}

// GetWithRefresh gets a value and optionally calls refresh function
func (mcs *MockCacheService) GetWithRefresh(ctx context.Context, key string, dest interface{}, 
	refreshFunc func() (interface{}, error), ttl time.Duration) error {
	
	// Try to get from cache first
	err := mcs.Get(ctx, key, dest)
	if err == nil {
		return nil
	}
	
	// Cache miss or error - call refresh function
	value, refreshErr := refreshFunc()
	if refreshErr != nil {
		mcs.metrics.RecordRefreshError()
		return refreshErr
	}
	
	// Store in cache
	if setErr := mcs.Set(ctx, key, value, ttl); setErr != nil {
		// Log error but don't fail the request
	}
	
	// Return the freshly fetched data
	if jsonData, jsonErr := json.Marshal(value); jsonErr == nil {
		if unmarshalErr := json.Unmarshal(jsonData, dest); unmarshalErr != nil {
			return &CacheError{Operation: "deserialize", Key: key, Err: unmarshalErr}
		}
	}
	
	return nil
}

// MGet retrieves multiple keys at once
func (mcs *MockCacheService) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	mcs.metrics.RecordOperation()
	
	if !mcs.enabled {
		return nil, ErrCacheNotEnabled
	}
	
	result := make(map[string]interface{})
	
	for _, key := range keys {
		var value interface{}
		if err := mcs.Get(ctx, key, &value); err == nil {
			result[key] = value
		}
	}
	
	return result, nil
}

// MSet stores multiple key-value pairs
func (mcs *MockCacheService) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	mcs.metrics.RecordOperation()
	
	if !mcs.enabled {
		return ErrCacheNotEnabled
	}
	
	for key, value := range items {
		if err := mcs.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	
	return nil
}

// Ping always succeeds for mock service
func (mcs *MockCacheService) Ping(ctx context.Context) error {
	if !mcs.enabled {
		return ErrCacheNotEnabled
	}
	return nil
}

// GetStats returns current mock cache statistics
func (mcs *MockCacheService) GetStats() CacheStats {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()
	
	// Count non-expired keys
	var keyCount int64
	now := time.Now()
	for _, item := range mcs.data {
		if now.Before(item.expiresAt) {
			keyCount++
		}
	}
	
	mcs.metrics.UpdateMemoryStats(int64(len(mcs.data)*100), keyCount) // Rough memory estimate
	mcs.metrics.UpdateConnectionStats(1, 0) // Mock always has 1 active connection
	
	return mcs.metrics.GetStats()
}

// Close is a no-op for mock service
func (mcs *MockCacheService) Close() error {
	return nil
}

// ClearAll removes all items from the mock cache (useful for testing)
func (mcs *MockCacheService) ClearAll() {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()
	
	mcs.data = make(map[string]cacheItem)
	mcs.metrics.Reset()
}

// shouldSimulateError returns true based on configured error rate
func (mcs *MockCacheService) shouldSimulateError() bool {
	// Simple random error simulation
	// In a real implementation, you might use math/rand with proper seeding
	return false // Simplified for now
}
