# Redis Caching Design for Portal64API

## Overview

This document outlines the Redis caching strategy for Portal64API to improve performance and reduce database load. The design implements a **Cache-Aside (Lazy Loading)** pattern with background refresh capabilities and comprehensive monitoring.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Handler   │───▶│  Cache Service  │───▶│ Repository/DB   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Redis Instance │
                       └─────────────────┘
```

### Key Components

1. **Cache Service**: Central caching layer with Redis client
2. **Background Refresh**: Proactive cache warming before expiration
3. **Fallback Mechanism**: Direct DB access when Redis unavailable
4. **Monitoring**: Metrics collection and health checks

## Cache Key Strategy

### Hierarchical Key Structure

```
Format: {domain}:{entity}:{identifier}[:{operation}]
```

**Examples:**
- `player:C0101-1014` - Individual player data
- `club:C0327` - Club details
- `club:C0327:players` - Club player list
- `tournament:T123456` - Tournament details
- `tournament:upcoming` - Upcoming tournaments list
- `address:region:C` - Region C addresses
- `search:player:hash:abc123` - Hashed search results

### Key Categories

| Category | Pattern | Example |
|----------|---------|---------|
| **Player Data** | `player:{id}` | `player:C0101-1014` |
| **Player Rating History** | `player:{id}:rating-history` | `player:C0101-1014:rating-history` |
| **Club Data** | `club:{id}` | `club:C0327` |
| **Club Players** | `club:{id}:players:{sort?}` | `club:C0327:players:dwz` |
| **Club Profile** | `club:{id}:profile` | `club:C0327:profile` |
| **Tournament Data** | `tournament:{id}` | `tournament:T123456` |
| **Tournament Lists** | `tournament:{type}` | `tournament:upcoming` |
| **Tournament Search** | `tournament:search:hash:{hash}` | `tournament:search:hash:abc123` |
| **Address Data** | `address:region:{region}` | `address:region:C` |
| **Search Results** | `search:{type}:hash:{hash}` | `search:player:hash:def456` |

### Hash Generation for Complex Searches

```go
func GenerateSearchHash(req SearchRequest) string {
    // Include all search parameters: query, limit, offset, sort, filters
    data := fmt.Sprintf("%s:%d:%d:%s:%v", 
        req.Query, req.Limit, req.Offset, req.Sort, req.ShowActive)
    return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}
```

## TTL Strategy

### Data Classification & TTL Values

| Data Type | TTL | Rationale | Examples |
|-----------|-----|-----------|----------|
| **Static Reference** | 24 hours | Rarely changes | Address regions, types |
| **Semi-Static** | 1 hour | Changes infrequently | Player details, club info |
| **Dynamic Lists** | 15 minutes | Regular updates | Tournament lists, search results |
| **Historical Data** | 7 days | Immutable past data | Rating history, past tournaments |
| **Active Memberships** | 30 minutes | Membership changes | Club player lists |

### Background Refresh Timing

- Refresh trigger: 80% of TTL elapsed
- Examples:
  - 1 hour TTL → Refresh after 48 minutes
  - 15 minutes TTL → Refresh after 12 minutes
  - 24 hours TTL → Refresh after 19.2 hours

## Implementation Details

### Cache Service Interface

```go
type CacheService interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    
    // Background refresh
    GetWithRefresh(ctx context.Context, key string, dest interface{}, 
                   refreshFunc func() (interface{}, error), ttl time.Duration) error
    
    // Batch operations
    MGet(ctx context.Context, keys []string) (map[string]interface{}, error)
    MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    
    // Health check
    Ping(ctx context.Context) error
    
    // Metrics
    GetStats() CacheStats
}
```

### Service Integration Pattern

```go
func (s *PlayerService) GetPlayerByID(playerID string) (*models.PlayerResponse, error) {
    ctx := context.Background()
    cacheKey := fmt.Sprintf("player:%s", playerID)
    
    var cachedPlayer models.PlayerResponse
    
    // Try cache with background refresh
    err := s.cache.GetWithRefresh(ctx, cacheKey, &cachedPlayer, 
        func() (interface{}, error) {
            // Fallback to database
            return s.loadPlayerFromDB(playerID)
        }, 1*time.Hour)
    
    if err != nil {
        // Cache miss or error - load directly from DB
        return s.loadPlayerFromDB(playerID)
    }
    
    return &cachedPlayer, nil
}
```

### Redis Configuration

```go
type CacheConfig struct {
    Enabled          bool          `yaml:"enabled"`
    Address          string        `yaml:"address"`           // "localhost:6379"
    Password         string        `yaml:"password"`          // ""
    Database         int           `yaml:"database"`          // 0
    MaxRetries       int           `yaml:"max_retries"`       // 3
    PoolSize         int           `yaml:"pool_size"`         // 10
    MinIdleConns     int           `yaml:"min_idle_conns"`    // 5
    DialTimeout      time.Duration `yaml:"dial_timeout"`      // 5s
    ReadTimeout      time.Duration `yaml:"read_timeout"`      // 3s
    WriteTimeout     time.Duration `yaml:"write_timeout"`     // 3s
    PoolTimeout      time.Duration `yaml:"pool_timeout"`      // 4s
    IdleTimeout      time.Duration `yaml:"idle_timeout"`      // 5m
    MaxConnAge       time.Duration `yaml:"max_conn_age"`      // 0 (no limit)
    
    // Background refresh settings
    RefreshThreshold float64       `yaml:"refresh_threshold"` // 0.8 (80%)
    RefreshWorkers   int           `yaml:"refresh_workers"`   // 5
}
```

## Deployment Options

### Option 1: External Redis Instance (Recommended)

**Advantages:**
- Production-ready
- Better resource management
- Persistent storage options
- Monitoring tools available

**Setup:**
```bash
# Using Docker
docker run -d --name redis-portal64 -p 6379:6379 redis:7-alpine

# Or install locally
# Windows: Use Redis for Windows or WSL
# Add to docker-compose.yml for easy management
```

### Option 2: Embedded Redis (Development Only)

**Use Cases:**
- Development environment
- Testing
- Simplified deployment

**Implementation:**
```go
// Using miniredis for testing/development
import "github.com/alicebob/miniredis/v2"

func NewEmbeddedRedis() (*miniredis.Miniredis, error) {
    s, err := miniredis.Run()
    if err != nil {
        return nil, err
    }
    return s, nil
}
```

## Monitoring & Metrics

### Key Metrics to Track

```go
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
```

### Health Check Endpoint

```go
// GET /health/cache
func (h *HealthHandler) CacheHealth(c *gin.Context) {
    ctx := context.Background()
    
    // Test Redis connectivity
    err := h.cache.Ping(ctx)
    if err != nil {
        c.JSON(503, gin.H{
            "status": "unhealthy",
            "redis": "disconnected",
            "error": err.Error(),
        })
        return
    }
    
    stats := h.cache.GetStats()
    c.JSON(200, gin.H{
        "status": "healthy",
        "redis": "connected",
        "stats": stats,
    })
}
```

### Monitoring Dashboard Data

```go
// GET /api/v1/admin/cache/stats
func (h *AdminHandler) CacheStats(c *gin.Context) {
    stats := h.cache.GetStats()
    
    response := gin.H{
        "performance": gin.H{
            "hit_ratio": stats.HitRatio,
            "avg_response_time": stats.AvgResponseTime.Milliseconds(),
            "total_requests": stats.TotalRequests,
        },
        "usage": gin.H{
            "memory_used_mb": stats.MemoryUsed / 1024 / 1024,
            "key_count": stats.KeyCount,
            "active_connections": stats.ActiveConnections,
        },
        "errors": gin.H{
            "cache_errors": stats.CacheErrors,
            "refresh_errors": stats.RefreshErrors,
        },
    }
    
    c.JSON(200, response)
}
```

## Route-Specific Caching Strategy

### Player Routes

| Route | Cache Key | TTL | Notes |
|-------|-----------|-----|-------|
| `GET /players/:id` | `player:{id}` | 1h | Semi-static player data |
| `GET /players/:id/rating-history` | `player:{id}:rating-history` | 7d | Historical data rarely changes |
| `GET /players?search=...` | `search:player:hash:{hash}` | 15m | Dynamic search results |

### Club Routes

| Route | Cache Key | TTL | Notes |
|-------|-----------|-----|-------|
| `GET /clubs/:id` | `club:{id}` | 1h | Club details |
| `GET /clubs/:id/players` | `club:{id}:players:{sort}` | 30m | Active memberships |
| `GET /clubs/:id/profile` | `club:{id}:profile` | 1h | Club profile with stats |
| `GET /clubs/all` | `clubs:all` | 2h | Full club list |

### Tournament Routes

| Route | Cache Key | TTL | Notes |
|-------|-----------|-----|-------|
| `GET /tournaments/:id` | `tournament:{id}` | 7d | Tournament details (historical) |
| `GET /tournaments/recent` | `tournament:recent` | 15m | Dynamic list |
| `GET /tournaments?search=...` | `search:tournament:hash:{hash}` | 15m | Search results |

### Address Routes

| Route | Cache Key | TTL | Notes |
|-------|-----------|-----|-------|
| `GET /addresses/regions` | `address:regions` | 24h | Static reference data |
| `GET /addresses/:region` | `address:region:{region}` | 24h | Regional addresses |
| `GET /addresses/:region/types` | `address:region:{region}:types` | 24h | Address types |

## Configuration Integration

### Environment Variables

```bash
# Redis Configuration
CACHE_ENABLED=true
CACHE_ADDRESS=localhost:6379
CACHE_PASSWORD=
CACHE_DATABASE=0
CACHE_POOL_SIZE=10
CACHE_DIAL_TIMEOUT=5s
CACHE_READ_TIMEOUT=3s
CACHE_WRITE_TIMEOUT=3s

# Background Refresh
CACHE_REFRESH_THRESHOLD=0.8
CACHE_REFRESH_WORKERS=5
```

### Config File Integration

```yaml
# config.yaml
cache:
  enabled: true
  address: "localhost:6379"
  password: ""
  database: 0
  pool_size: 10
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
  refresh_threshold: 0.8
  refresh_workers: 5
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. Add Redis client dependency (`go-redis/redis/v9`)
2. Implement `CacheService` interface
3. Add cache configuration to config system
4. Implement basic cache-aside pattern

### Phase 2: Route Integration
1. Integrate caching in PlayerService
2. Integrate caching in ClubService  
3. Integrate caching in TournamentService
4. Integrate caching in AddressService

### Phase 3: Advanced Features
1. Implement background refresh mechanism
2. Add search result caching with hashing
3. Implement batch operations for efficiency
4. Add comprehensive error handling

### Phase 4: Monitoring & Operations
1. Implement metrics collection
2. Add health check endpoints
3. Create admin cache management endpoints
4. Add logging and alerting

## Testing Strategy

### Unit Tests
- Test cache service methods
- Test key generation functions
- Test TTL and refresh logic
- Test fallback mechanisms

### Integration Tests
- Test full request flow with cache
- Test Redis connectivity failures
- Test cache warming scenarios
- Test metrics collection

### Performance Tests
- Measure response time improvements
- Test cache hit ratios under load
- Validate memory usage patterns
- Test background refresh performance

## Dependencies

```go
// Add to go.mod
require (
    github.com/redis/go-redis/v9 v9.0.5
    github.com/prometheus/client_golang v1.16.0 // For metrics (optional)
)
```

This design provides a robust, scalable caching solution that will significantly improve Portal64API performance while maintaining data consistency and providing comprehensive monitoring capabilities.
