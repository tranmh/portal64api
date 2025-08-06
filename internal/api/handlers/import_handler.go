package handlers

import (
	"fmt"
	"net/http"
	"portal64api/internal/models"
	"portal64api/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	importService services.ImportServiceInterface
}

// NewImportHandler creates a new import handler instance
func NewImportHandler(importService services.ImportServiceInterface) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// GetImportStatus returns the current import status
// @Summary Get import status
// @Description Get the current status of database import operations
// @Tags import
// @Accept json
// @Produce json
// @Success 200 {object} models.ImportStatus
// @Router /api/v1/import/status [get]
func (ih *ImportHandler) GetImportStatus(c *gin.Context) {
	if ih.importService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Import service is not available",
		})
		return
	}

	status := ih.importService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// StartManualImport triggers a manual import
// @Summary Start manual import
// @Description Trigger a manual database import operation
// @Tags import
// @Accept json
// @Produce json
// @Success 200 {object} models.ImportStartResponse
// @Failure 400 {object} gin.H
// @Failure 409 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/import/start [post]
func (ih *ImportHandler) StartManualImport(c *gin.Context) {
	if ih.importService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Import service is not available",
		})
		return
	}

	err := ih.importService.TriggerManualImport()
	if err != nil {
		if err.Error() == "import is already running" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Import is already running",
			})
			return
		}
		if err.Error() == "import service is disabled" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Import service is disabled",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := models.ImportStartResponse{
		Message:   "Manual import started",
		StartedAt: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetImportLogs returns recent import log entries
// @Summary Get import logs
// @Description Get recent import log entries with optional limit
// @Tags import
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of log entries to return (default: 100)"
// @Success 200 {object} models.ImportLogsResponse
// @Router /api/v1/import/logs [get]
func (ih *ImportHandler) GetImportLogs(c *gin.Context) {
	if ih.importService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Import service is not available",
		})
		return
	}

	// Parse limit parameter
	limit := 100 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			// Cap at 1000 to prevent excessive memory usage
			if limit > 1000 {
				limit = 1000
			}
		}
	}

	logs := ih.importService.GetLogs(limit)
	response := models.ImportLogsResponse{
		Logs: logs,
	}

	c.JSON(http.StatusOK, response)
}

// TestImportConnection tests the SCP connection
// @Summary Test import connection
// @Description Test the SCP connection for import operations
// @Tags import
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/import/test-connection [post]
func (ih *ImportHandler) TestImportConnection(c *gin.Context) {
	if ih.importService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Import service is not available",
		})
		return
	}

	err := ih.importService.TestConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Connection test failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection test successful",
		"status":  "ok",
	})
}

// GetImportHealth returns import service health information
// @Summary Get import service health
// @Description Get health and configuration information for import service
// @Tags import
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/import/health [get]
func (ih *ImportHandler) GetImportHealth(c *gin.Context) {
	if ih.importService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unavailable",
			"message": "Import service is not available",
		})
		return
	}

	status := ih.importService.GetStatus()
	
	health := gin.H{
		"status":       "available",
		"service":      "import",
		"current_status": status.Status,
		"last_success": status.LastSuccess,
		"next_scheduled": status.NextScheduled,
	}

	// Add additional health indicators
	if status.Status == models.StatusRunning {
		health["current_step"] = status.CurrentStep
		health["progress"] = status.Progress
	}

	if status.Error != "" {
		health["last_error"] = status.Error
	}

	c.JSON(http.StatusOK, health)
}

// GetImportConfig returns import configuration (without sensitive data)
// @Summary Get import configuration
// @Description Get import service configuration (excluding sensitive information)
// @Tags import
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/import/config [get]
func (ih *ImportHandler) GetImportConfig(c *gin.Context) {
	// This would typically be restricted to admin users
	// For now, return basic configuration without sensitive data
	
	config := gin.H{
		"enabled":    true, // This could be read from actual config
		"schedule":   "0 2 * * *", // Daily at 2 AM
		"components": gin.H{
			"freshness_check": true,
			"scp_download":    true,
			"zip_extraction":  true,
			"database_import": true,
			"cache_cleanup":   true,
		},
		"target_databases": []string{"mvdsb", "portal64_bdw"},
	}

	c.JSON(http.StatusOK, config)
}
