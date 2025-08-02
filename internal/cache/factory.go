package cache

import (
	"fmt"
	"portal64api/internal/config"
)

// NewCacheService creates a cache service based on configuration
func NewCacheService(cfg config.CacheConfig) (CacheService, error) {
	if !cfg.Enabled {
		// Return a disabled mock service for development/testing
		return NewMockCacheService(false), nil
	}
	
	// Create Redis service
	redisService, err := NewRedisService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis cache service: %w", err)
	}
	
	return redisService, nil
}

// NewTestCacheService creates a cache service for testing
func NewTestCacheService(enabled bool) CacheService {
	return NewMockCacheService(enabled)
}
