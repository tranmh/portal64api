package handlers

import (
	"context"
	"net/http"
	"time"

	"portal64api/internal/cache"

	"github.com/gin-gonic/gin"
)

// AdminHandler handles administrative operations
type AdminHandler struct {
	cacheService cache.CacheService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(cacheService cache.CacheService) *AdminHandler {
	return &AdminHandler{
		cacheService: cacheService,
	}
}

// GetCacheStats godoc
// @Summary Get cache statistics
// @Description Get detailed cache performance statistics
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} models.CacheStatsResponse
// @Failure 503 {object} map[string]string
// @Router /api/v1/admin/cache/stats [get]
func (h *AdminHandler) GetCacheStats(c *gin.Context) {
	stats := h.cacheService.GetStats()
	
	c.JSON(http.StatusOK, gin.H{
		"performance": gin.H{
			"hit_ratio":           stats.HitRatio,
			"avg_response_time_ms": stats.AvgResponseTime.Milliseconds(),
			"total_requests":      stats.TotalRequests,
			"cache_hits":          stats.CacheHits,
			"cache_misses":        stats.CacheMisses,
		},
		"usage": gin.H{
			"memory_used_mb":      stats.MemoryUsed / 1024 / 1024,
			"key_count":          stats.KeyCount,
			"active_connections": stats.ActiveConnections,
			"idle_connections":   stats.IdleConnections,
		},
		"operations": gin.H{
			"cache_operations":     stats.CacheOperations,
			"background_refreshes": stats.BackgroundRefreshes,
		},
		"errors": gin.H{
			"cache_errors":   stats.CacheErrors,
			"refresh_errors": stats.RefreshErrors,
		},
	})
}

// GetCacheHealth godoc
// @Summary Check cache health
// @Description Check Redis cache connectivity and status
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} models.CacheHealthResponse
// @Failure 503 {object} models.CacheHealthResponse
// @Router /api/v1/admin/cache/health [get]
func (h *AdminHandler) GetCacheHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test Redis connectivity
	err := h.cacheService.Ping(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"redis":  "disconnected",
			"error":  err.Error(),
		})
		return
	}
	
	stats := h.cacheService.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"redis":  "connected",
		"stats": gin.H{
			"hit_ratio":      stats.HitRatio,
			"total_requests": stats.TotalRequests,
			"key_count":      stats.KeyCount,
		},
	})
}
