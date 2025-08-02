# Redis Caching Implementation File Plan

## Overview

This document outlines which files will be created or modified to implement the Redis caching system as designed in `docs/RedisCaching.md`. The implementation follows the existing project structure and patterns.

## File Structure Plan

```
portal64api/
├── internal/
│   ├── cache/                          # NEW - Cache service package
│   │   ├── interface.go               # Cache service interface
│   │   ├── redis_service.go           # Redis implementation
│   │   ├── mock_service.go            # Mock for testing
│   │   ├── key_generator.go           # Cache key generation utils
│   │   ├── background_refresh.go      # Background refresh workers
│   │   └── metrics.go                 # Cache metrics collection
│   ├── config/
│   │   └── config.go                  # MODIFIED - Add cache config
│   ├── services/
│   │   ├── player_service.go          # MODIFIED - Add caching
│   │   ├── club_service.go            # MODIFIED - Add caching  
│   │   ├── tournament_service.go      # MODIFIED - Add caching
│   │   └── address_service.go         # MODIFIED - Add caching
│   ├── api/
│   │   └── handlers/
│   │       ├── health_handler.go      # MODIFIED - Add cache health
│   │       └── admin_handler.go       # NEW - Cache admin endpoints
│   └── middleware/
│       └── cache_middleware.go        # NEW - Optional cache middleware
├── pkg/
│   └── cache/
│       └── utils.go                   # NEW - Cache utility functions
├── cmd/server/
│   └── main.go                        # MODIFIED - Initialize cache service
├── configs/
│   ├── config.yaml                    # MODIFIED - Add cache config
│   └── config.example.yaml            # MODIFIED - Cache config example
├── tests/
│   ├── cache/                         # NEW - Cache-specific tests
│   │   ├── cache_service_test.go
│   │   ├── integration_test.go
│   │   └── performance_test.go
│   └── services/                      # MODIFIED - Add cache tests
│       ├── player_service_test.go     # Add cache integration tests
│       ├── club_service_test.go       # Add cache integration tests
│       └── tournament_service_test.go # Add cache integration tests
├── docker-compose.yml                 # MODIFIED - Add Redis service
└── .env.example                       # MODIFIED - Add cache env vars
```

## Implementation Plan by File

### Phase 1: Core Infrastructure

#### 1. `internal/cache/interface.go`
**Purpose**: Define the cache service interface
**Contents**:
- `CacheService` interface from the design document
- `CacheStats` struct for metrics
- `CacheConfig` struct for configuration
- Error types for cache operations

```go
type CacheService interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    // ... other methods from design
}
```

#### 2. `internal/cache/redis_service.go`
**Purpose**: Redis implementation of CacheService
**Contents**:
- Redis client wrapper
- Connection management
- Cache-aside pattern implementation
- Error handling and fallback logic
- Connection pooling configuration

**Dependencies**: `github.com/redis/go-redis/v9`

#### 3. `internal/cache/key_generator.go`
**Purpose**: Centralized cache key generation
**Contents**:
- Key generation functions for each entity type
- Hash generation for complex searches
- Key pattern constants
- Validation functions

**Key Functions**:
```go
func PlayerKey(playerID string) string
func ClubKey(clubID string) string
func ClubPlayersKey(clubID string, sort string) string
func SearchHashKey(entityType string, hash string) string
func GenerateSearchHash(req SearchRequest) string
```

#### 4. `internal/cache/background_refresh.go`
**Purpose**: Background cache refresh mechanism
**Contents**:
- Worker pool for background refresh
- TTL monitoring logic
- Refresh trigger mechanisms
- Error handling for refresh failures

#### 5. `internal/cache/metrics.go`
**Purpose**: Cache performance metrics
**Contents**:
- Metrics collection (hit/miss ratios, response times)
- Statistics aggregation
- Memory usage tracking
- Performance counters

#### 6. `internal/cache/mock_service.go`
**Purpose**: Mock implementation for testing
**Contents**:
- In-memory mock cache
- Test utilities
- Configurable behaviors for testing

### Phase 2: Configuration Integration

#### 7. `internal/config/config.go` (MODIFIED)
**Purpose**: Add cache configuration
**Changes**:
- Add `CacheConfig` struct to main `Config`
- Add cache-related environment variable parsing
- Add cache configuration validation

**New Config Section**:
```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Cache    CacheConfig    `yaml:"cache"`     // NEW
}

type CacheConfig struct {
    Enabled          bool          `yaml:"enabled"`
    Address          string        `yaml:"address"`
    Password         string        `yaml:"password"`
    Database         int           `yaml:"database"`
    PoolSize         int           `yaml:"pool_size"`
    // ... other Redis config fields
}
```

#### 8. `.env.example` (MODIFIED)
**Purpose**: Add cache environment variables
**New Variables**:
```bash
# Redis Cache Configuration
CACHE_ENABLED=true
CACHE_ADDRESS=localhost:6379
CACHE_PASSWORD=
CACHE_DATABASE=0
CACHE_POOL_SIZE=10
CACHE_DIAL_TIMEOUT=5s
CACHE_READ_TIMEOUT=3s
CACHE_WRITE_TIMEOUT=3s
CACHE_REFRESH_THRESHOLD=0.8
CACHE_REFRESH_WORKERS=5
```

### Phase 3: Service Integration

#### 9. `internal/services/player_service.go` (MODIFIED)
**Purpose**: Add caching to player operations
**Changes**:
- Inject `CacheService` into `PlayerService`
- Add caching to `GetPlayerByID`
- Add caching to `SearchPlayers`
- Add caching to `GetPlayerRatingHistory`
- Implement cache-aside pattern with fallback

**Example Pattern**:
```go
func (s *PlayerService) GetPlayerByID(playerID string) (*models.PlayerResponse, error) {
    // Try cache first
    cacheKey := s.cache.PlayerKey(playerID)
    var cachedPlayer models.PlayerResponse
    
    if err := s.cache.Get(ctx, cacheKey, &cachedPlayer); err == nil {
        return &cachedPlayer, nil
    }
    
    // Fallback to database
    player, err := s.loadPlayerFromDB(playerID)
    if err != nil {
        return nil, err
    }
    
    // Update cache
    s.cache.Set(ctx, cacheKey, player, 1*time.Hour)
    return player, nil
}
```

#### 10. `internal/services/club_service.go` (MODIFIED)
**Purpose**: Add caching to club operations
**Changes**:
- Add caching to `GetClubByID`
- Add caching to `GetClubPlayers` (with sort-specific keys)
- Add caching to `SearchClubs`
- Add caching to club statistics

#### 11. `internal/services/tournament_service.go` (MODIFIED)
**Purpose**: Add caching to tournament operations
**Changes**:
- Add caching to tournament details
- Add caching to tournament lists (upcoming, recent)
- Add caching to tournament search results

#### 12. `internal/services/address_service.go` (MODIFIED)
**Purpose**: Add caching to address operations
**Changes**:
- Add caching to regional addresses
- Add caching to address types
- Long TTL for mostly static data

### Phase 4: API Integration

#### 13. `cmd/server/main.go` (MODIFIED)
**Purpose**: Initialize cache service
**Changes**:
- Initialize Redis client
- Create cache service instance
- Inject cache service into all services
- Add graceful shutdown for cache connections

**New Initialization**:
```go
// Initialize cache service
cacheService, err := cache.NewCacheService(cfg.Cache)
if err != nil {
    log.Fatalf("Failed to initialize cache: %v", err)
}
defer cacheService.Close()

// Inject cache into services
playerService := services.NewPlayerService(playerRepo, clubRepo, cacheService)
clubService := services.NewClubService(clubRepo, cacheService)
// ... other services
```

#### 14. `internal/api/handlers/health_handler.go` (MODIFIED)
**Purpose**: Add cache health checks
**Changes**:
- Add cache connectivity check
- Add cache statistics to health endpoint
- Add detailed cache metrics

**New Endpoint**: `GET /health/cache`

#### 15. `internal/api/handlers/admin_handler.go` (NEW)
**Purpose**: Cache administration endpoints
**Contents**:
- Cache statistics endpoint
- Cache flush operations
- Cache key inspection
- Performance metrics

**New Endpoints**:
- `GET /api/v1/admin/cache/stats`
- `POST /api/v1/admin/cache/flush`
- `GET /api/v1/admin/cache/keys`

#### 16. `internal/middleware/cache_middleware.go` (NEW)
**Purpose**: Optional HTTP caching middleware
**Contents**:
- HTTP cache headers
- ETag generation
- Conditional requests handling
- Cache-Control headers

### Phase 5: Utilities and Support

#### 17. `pkg/cache/utils.go` (NEW)
**Purpose**: Cache utility functions
**Contents**:
- Serialization helpers
- Cache key validation
- TTL calculation utilities
- Error handling helpers

#### 18. `docker-compose.yml` (MODIFIED)
**Purpose**: Add Redis service for development
**Changes**:
```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  redis_data:
```

### Phase 6: Testing

#### 19. `tests/cache/cache_service_test.go` (NEW)
**Purpose**: Unit tests for cache service
**Contents**:
- Redis service method tests
- Key generation tests
- Error handling tests
- Mock service tests

#### 20. `tests/cache/integration_test.go` (NEW)
**Purpose**: Integration tests with real Redis
**Contents**:
- Full cache flow tests
- Fallback mechanism tests
- Background refresh tests
- Connection failure scenarios

#### 21. `tests/cache/performance_test.go` (NEW)
**Purpose**: Performance benchmarks
**Contents**:
- Cache hit/miss performance
- Memory usage tests
- Concurrent access tests
- Load testing scenarios

#### 22. `tests/services/*_test.go` (MODIFIED)
**Purpose**: Add cache integration to service tests
**Changes**:
- Test caching behavior in service methods
- Test cache invalidation scenarios
- Test fallback mechanisms
- Performance improvement validation

### Phase 7: Documentation and Operations

#### 23. `docs/CacheDeployment.md` (NEW)
**Purpose**: Deployment guide for Redis caching
**Contents**:
- Redis installation instructions
- Configuration examples
- Monitoring setup
- Troubleshooting guide

#### 24. `docs/CacheMonitoring.md` (NEW)
**Purpose**: Cache monitoring and alerting
**Contents**:
- Key metrics to monitor
- Alert thresholds
- Dashboard examples
- Performance tuning

## Dependencies to Add

Add to `go.mod`:
```go
require (
    github.com/redis/go-redis/v9 v9.0.5
    github.com/alicebob/miniredis/v2 v2.30.4 // For testing
)
```

## Configuration Files

### Environment Variables
All cache configuration will be loaded from environment variables or `.env` file:
- `CACHE_ENABLED`
- `CACHE_ADDRESS`
- `CACHE_PASSWORD`
- `CACHE_DATABASE`
- etc.

### YAML Configuration
Cache configuration will be integrated into the existing `config.yaml` structure.

## Testing Strategy

### Unit Tests
- All cache service methods
- Key generation functions
- Configuration parsing
- Error scenarios

### Integration Tests
- Full request-response cycle with caching
- Database fallback scenarios
- Redis connectivity issues
- Background refresh functionality

### Performance Tests
- Cache hit ratio improvements
- Response time improvements
- Memory usage validation
- Concurrent access performance

## Implementation Order

1. **Core Infrastructure** (Files 1-6): Cache service foundation
2. **Configuration** (Files 7-8): Configuration integration
3. **Service Integration** (Files 9-12): Add caching to business logic
4. **API Integration** (Files 13-15): HTTP layer and admin endpoints
5. **Support & Utils** (Files 16-18): Utilities and development support
6. **Testing** (Files 19-22): Comprehensive test coverage
7. **Documentation** (Files 23-24): Operational documentation

## Validation Criteria

Each phase should be validated with:
- ✅ All tests pass
- ✅ Performance improvements measured
- ✅ Fallback mechanisms work
- ✅ Configuration is properly loaded
- ✅ Cache keys follow the design patterns
- ✅ TTL values match the design specifications
- ✅ Background refresh works correctly
- ✅ Metrics are properly collected

This implementation plan ensures a systematic, well-tested rollout of Redis caching that follows the existing codebase patterns and the comprehensive design outlined in `docs/RedisCaching.md`.
