package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"portal64api/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisService implements CacheService using Redis
type RedisService struct {
	client        *redis.Client
	config        config.CacheConfig
	keyGenerator  *KeyGenerator
	metrics       *MetricsCollector
	enabled       bool
	
	// Background refresh
	refreshChan   chan refreshRequest
	stopChan      chan struct{}
}

type refreshRequest struct {
	key         string
	refreshFunc func() (interface{}, error)
	ttl         time.Duration
}

// NewRedisService creates a new Redis cache service
func NewRedisService(config config.CacheConfig) (*RedisService, error) {
	if !config.Enabled {
		return &RedisService{
			enabled: false,
			metrics: NewMetricsCollector(),
		}, nil
	}
	
	// Create Redis client with configuration
	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.Database,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})
	
	service := &RedisService{
		client:       client,
		config:       config,
		keyGenerator: NewKeyGenerator(),
		metrics:      NewMetricsCollector(),
		enabled:      true,
		refreshChan:  make(chan refreshRequest, 100),
		stopChan:     make(chan struct{}),
	}
	
	// Start background refresh workers
	service.startRefreshWorkers()
	
	return service, nil
}

// Get retrieves a value from cache and deserializes it
func (rs *RedisService) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	defer func() {
		rs.metrics.RecordResponseTime(time.Since(start))
	}()
	
	rs.metrics.RecordRequest()
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		rs.metrics.RecordMiss()
		return ErrCacheNotEnabled
	}
	
	// Validate key
	if !rs.keyGenerator.ValidateKey(key) {
		rs.metrics.RecordError()
		return &CacheError{Operation: "get", Key: key, Err: fmt.Errorf("invalid key format")}
	}
	
	// Get from Redis
	val, err := rs.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			rs.metrics.RecordMiss()
			return &CacheError{Operation: "get", Key: key, Err: ErrKeyNotFound.Err}
		}
		rs.metrics.RecordError()
		return &CacheError{Operation: "get", Key: key, Err: err}
	}
	
	// Deserialize
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		rs.metrics.RecordError()
		return &CacheError{Operation: "get", Key: key, Err: ErrDeserialization.Err}
	}
	
	rs.metrics.RecordHit()
	return nil
}

// Set stores a value in cache with TTL
func (rs *RedisService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		rs.metrics.RecordResponseTime(time.Since(start))
	}()
	
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		return ErrCacheNotEnabled
	}
	
	// Validate key
	if !rs.keyGenerator.ValidateKey(key) {
		rs.metrics.RecordError()
		return &CacheError{Operation: "set", Key: key, Err: fmt.Errorf("invalid key format")}
	}
	
	// Serialize
	data, err := json.Marshal(value)
	if err != nil {
		rs.metrics.RecordError()
		return &CacheError{Operation: "set", Key: key, Err: ErrSerialization.Err}
	}
	
	// Set in Redis
	if err := rs.client.Set(ctx, key, data, ttl).Err(); err != nil {
		rs.metrics.RecordError()
		return &CacheError{Operation: "set", Key: key, Err: err}
	}
	
	return nil
}

// Delete removes a key from cache
func (rs *RedisService) Delete(ctx context.Context, key string) error {
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		return ErrCacheNotEnabled
	}
	
	if err := rs.client.Del(ctx, key).Err(); err != nil {
		rs.metrics.RecordError()
		return &CacheError{Operation: "delete", Key: key, Err: err}
	}
	
	return nil
}

// Exists checks if a key exists in cache
func (rs *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		return false, ErrCacheNotEnabled
	}
	
	count, err := rs.client.Exists(ctx, key).Result()
	if err != nil {
		rs.metrics.RecordError()
		return false, &CacheError{Operation: "exists", Key: key, Err: err}
	}
	
	return count > 0, nil
}

// FlushAll clears all keys from the cache
func (rs *RedisService) FlushAll(ctx context.Context) error {
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		return ErrCacheNotEnabled
	}
	
	if err := rs.client.FlushDB(ctx).Err(); err != nil {
		rs.metrics.RecordError()
		return &CacheError{Operation: "flush_all", Err: err}
	}
	
	return nil
}

// GetWithRefresh gets a value and schedules background refresh if needed
func (rs *RedisService) GetWithRefresh(ctx context.Context, key string, dest interface{}, 
	refreshFunc func() (interface{}, error), ttl time.Duration) error {
	
	// Try to get from cache first
	err := rs.Get(ctx, key, dest)
	if err == nil {
		// Cache hit - check if we need to schedule refresh
		rs.scheduleRefreshIfNeeded(key, refreshFunc, ttl)
		return nil
	}
	
	// Cache miss or error - call refresh function directly
	value, refreshErr := refreshFunc()
	if refreshErr != nil {
		rs.metrics.RecordRefreshError()
		return refreshErr
	}
	
	// Store in cache for next time
	if setErr := rs.Set(ctx, key, value, ttl); setErr != nil {
		// Log error but don't fail the request
		// The caller still gets their data
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
func (rs *RedisService) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		return nil, ErrCacheNotEnabled
	}
	
	if len(keys) == 0 {
		return map[string]interface{}{}, nil
	}
	
	// Validate all keys
	for _, key := range keys {
		if !rs.keyGenerator.ValidateKey(key) {
			rs.metrics.RecordError()
			return nil, &CacheError{Operation: "mget", Key: key, Err: fmt.Errorf("invalid key format")}
		}
	}
	
	// Get from Redis
	values, err := rs.client.MGet(ctx, keys...).Result()
	if err != nil {
		rs.metrics.RecordError()
		return nil, &CacheError{Operation: "mget", Err: err}
	}
	
	result := make(map[string]interface{})
	for i, val := range values {
		if val != nil {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(val.(string)), &jsonData); err == nil {
				result[keys[i]] = jsonData
				rs.metrics.RecordHit()
			} else {
				rs.metrics.RecordError()
			}
		} else {
			rs.metrics.RecordMiss()
		}
	}
	
	return result, nil
}

// MSet stores multiple key-value pairs with the same TTL
func (rs *RedisService) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	rs.metrics.RecordOperation()
	
	if !rs.enabled {
		return ErrCacheNotEnabled
	}
	
	if len(items) == 0 {
		return nil
	}
	
	// Use pipeline for efficiency
	pipe := rs.client.Pipeline()
	
	for key, value := range items {
		// Validate key
		if !rs.keyGenerator.ValidateKey(key) {
			rs.metrics.RecordError()
			return &CacheError{Operation: "mset", Key: key, Err: fmt.Errorf("invalid key format")}
		}
		
		// Serialize value
		data, err := json.Marshal(value)
		if err != nil {
			rs.metrics.RecordError()
			return &CacheError{Operation: "mset", Key: key, Err: ErrSerialization.Err}
		}
		
		pipe.Set(ctx, key, data, ttl)
	}
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		rs.metrics.RecordError()
		return &CacheError{Operation: "mset", Err: err}
	}
	
	return nil
}

// Ping checks Redis connectivity
func (rs *RedisService) Ping(ctx context.Context) error {
	if !rs.enabled {
		return ErrCacheNotEnabled
	}
	
	return rs.client.Ping(ctx).Err()
}

// GetStats returns current cache statistics
func (rs *RedisService) GetStats() CacheStats {
	if rs.enabled && rs.client != nil {
		// Update memory and connection stats from Redis
		ctx := context.Background()
		if info, err := rs.client.Info(ctx, "memory", "stats").Result(); err == nil {
			rs.updateStatsFromRedisInfo(info)
		}
		
		// Update connection stats
		poolStats := rs.client.PoolStats()
		rs.metrics.UpdateConnectionStats(int(poolStats.TotalConns-poolStats.IdleConns), int(poolStats.IdleConns))
	}
	
	return rs.metrics.GetStats()
}

// Close closes the Redis connection
func (rs *RedisService) Close() error {
	if !rs.enabled || rs.client == nil {
		return nil
	}
	
	// Stop background workers
	close(rs.stopChan)
	
	// Close Redis client
	return rs.client.Close()
}

// scheduleRefreshIfNeeded checks TTL and schedules refresh if needed
func (rs *RedisService) scheduleRefreshIfNeeded(key string, refreshFunc func() (interface{}, error), ttl time.Duration) {
	if !rs.enabled {
		return
	}
	
	// Check remaining TTL
	remaining, err := rs.client.TTL(context.Background(), key).Result()
	if err != nil {
		return
	}
	
	// Calculate refresh threshold (default 80% of TTL)
	refreshThreshold := time.Duration(float64(ttl) * rs.config.RefreshThreshold)
	
	// If remaining time is less than threshold, schedule refresh
	if remaining > 0 && remaining < refreshThreshold {
		select {
		case rs.refreshChan <- refreshRequest{
			key:         key,
			refreshFunc: refreshFunc,
			ttl:         ttl,
		}:
			// Refresh scheduled
		default:
			// Channel full, skip refresh
		}
	}
}

// startRefreshWorkers starts background refresh worker goroutines
func (rs *RedisService) startRefreshWorkers() {
	if !rs.enabled {
		return
	}
	
	for i := 0; i < rs.config.RefreshWorkers; i++ {
		go rs.refreshWorker()
	}
}

// refreshWorker processes background refresh requests
func (rs *RedisService) refreshWorker() {
	for {
		select {
		case req := <-rs.refreshChan:
			rs.processRefreshRequest(req)
		case <-rs.stopChan:
			return
		}
	}
}

// processRefreshRequest handles a single refresh request
func (rs *RedisService) processRefreshRequest(req refreshRequest) {
	ctx := context.Background()
	
	// Call refresh function
	value, err := req.refreshFunc()
	if err != nil {
		rs.metrics.RecordRefreshError()
		return
	}
	
	// Update cache
	if err := rs.Set(ctx, req.key, value, req.ttl); err != nil {
		rs.metrics.RecordRefreshError()
		return
	}
	
	rs.metrics.RecordBackgroundRefresh()
}

// updateStatsFromRedisInfo parses Redis INFO output and updates metrics
func (rs *RedisService) updateStatsFromRedisInfo(info string) {
	// Parse Redis INFO output for memory usage
	// This is a simplified version - in production you might want more detailed parsing
	var memoryUsed int64
	var keyCount int64
	
	// In a real implementation, you would parse the INFO string
	// For now, we'll use Redis commands to get the info we need
	ctx := context.Background()
	
	// Get memory usage
	if memInfo, err := rs.client.Info(ctx, "memory").Result(); err == nil {
		// Parse used_memory from the info string
		// Simplified parsing - you might want to use a proper parser
		_ = memInfo // Use this to extract actual memory usage
	}
	
	// Get key count using DBSIZE
	if size, err := rs.client.DBSize(ctx).Result(); err == nil {
		keyCount = size
	}
	
	rs.metrics.UpdateMemoryStats(memoryUsed, keyCount)
}
