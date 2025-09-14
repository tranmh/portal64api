package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"portal64api/internal/services"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SomatogrammCompatibilityHandler provides backward compatibility for Somatogramm API endpoints
// by using the unified Kader-Planung service with the compatibility adapter
type SomatogrammCompatibilityHandler struct {
	adapter *services.SomatogrammCompatibilityAdapter
	logger  *logrus.Logger
}

// NewSomatogrammCompatibilityHandler creates a new compatibility handler
func NewSomatogrammCompatibilityHandler(kaderPlanungService *services.KaderPlanungService) *SomatogrammCompatibilityHandler {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	adapter := services.NewSomatogrammCompatibilityAdapter(kaderPlanungService, logger)

	return &SomatogrammCompatibilityHandler{
		adapter: adapter,
		logger:  logger,
	}
}

// GetSomatogrammStatus returns the current execution status (compatibility wrapper)
// @Summary Get Somatogramm execution status (DEPRECATED)
// @Description Returns current execution status including running state, last execution time, and available files. This endpoint is deprecated - use /api/v1/kader-planung/status instead.
// @Tags somatogramm
// @Accept json
// @Produce json
// @Success 200 {object} services.ExecutionStatus
// @Failure 500 {object} errors.APIError
// @Router /api/v1/somatogramm/status [get]
// @Deprecated
func (h *SomatogrammCompatibilityHandler) GetSomatogrammStatus(c *gin.Context) {
	h.logger.Warn("DEPRECATED: /api/v1/somatogramm/status endpoint used. Please migrate to /api/v1/kader-planung/status")

	status := h.adapter.GetStatus()

	// Add available files to the status for backward compatibility
	files, err := h.adapter.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to list available files"))
		return
	}

	response := map[string]interface{}{
		"status":          status,
		"available_files": files,
		"migration_notice": "This endpoint is deprecated. Please use /api/v1/kader-planung/status for new integrations.",
	}

	utils.SendJSONResponse(c, http.StatusOK, response)
}

// StartSomatogrammExecution starts manual Somatogramm execution (compatibility wrapper)
// @Summary Start manual Somatogramm execution (DEPRECATED)
// @Description Starts manual execution of Somatogramm using unified Kader-Planung service in statistical mode. This endpoint is deprecated - use /api/v1/kader-planung/statistical instead.
// @Tags somatogramm
// @Accept json
// @Produce json
// @Param request body SomatogrammCompatibilityRequest false "Execution parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.APIError
// @Failure 409 {object} errors.APIError "Already running"
// @Failure 500 {object} errors.APIError
// @Router /api/v1/somatogramm/start [post]
// @Deprecated
func (h *SomatogrammCompatibilityHandler) StartSomatogrammExecution(c *gin.Context) {
	h.logger.Warn("DEPRECATED: /api/v1/somatogramm/start endpoint used. Please migrate to /api/v1/kader-planung/statistical")

	var request SomatogrammCompatibilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid request format"))
		return
	}

	// Convert request to parameters map
	params := map[string]interface{}{}
	if request.OutputFormat != "" {
		params["output_format"] = request.OutputFormat
	}
	if request.Timeout > 0 {
		params["timeout"] = request.Timeout
	}
	if request.Concurrency > 0 {
		params["concurrency"] = request.Concurrency
	}
	if request.MinSampleSize > 0 {
		params["min_sample_size"] = request.MinSampleSize
	}
	if request.Verbose != nil {
		params["verbose"] = *request.Verbose
	}

	if err := h.adapter.ExecuteManually(params); err != nil {
		if strings.Contains(err.Error(), "already running") {
			utils.SendJSONResponse(c, http.StatusConflict,
				errors.NewAPIError(409, "EXECUTION_ALREADY_RUNNING", "Somatogramm-compatible execution already in progress"))
			return
		}

		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to start Somatogramm-compatible execution"))
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, map[string]interface{}{
		"message": "Somatogramm-compatible statistical analysis started",
		"migration_notice": "This endpoint is deprecated. Please use /api/v1/kader-planung/statistical for new integrations.",
	})
}

// ListSomatogrammFiles returns list of available CSV/JSON files (compatibility wrapper)
// @Summary List available Somatogramm files (DEPRECATED)
// @Description Returns a list of all available Somatogramm-compatible statistical analysis files with metadata. This endpoint is deprecated - use /api/v1/kader-planung/statistical/files instead.
// @Tags somatogramm
// @Accept json
// @Produce json
// @Success 200 {array} services.FileInfo
// @Failure 500 {object} errors.APIError
// @Router /api/v1/somatogramm/files [get]
// @Deprecated
func (h *SomatogrammCompatibilityHandler) ListSomatogrammFiles(c *gin.Context) {
	h.logger.Warn("DEPRECATED: /api/v1/somatogramm/files endpoint used. Please migrate to /api/v1/kader-planung/statistical/files")

	files, err := h.adapter.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to list Somatogramm-compatible files"))
		return
	}

	// Add migration notice
	response := map[string]interface{}{
		"files": files,
		"migration_notice": "This endpoint is deprecated. Please use /api/v1/kader-planung/statistical/files for new integrations.",
	}

	utils.SendJSONResponse(c, http.StatusOK, response)
}

// DownloadSomatogrammFile serves a specific CSV/JSON file for download (compatibility wrapper)
// @Summary Download Somatogramm file (DEPRECATED)
// @Description Downloads a specific Somatogramm-compatible statistical analysis CSV or JSON file. This endpoint is deprecated - use /api/v1/kader-planung/download/{filename} instead.
// @Tags somatogramm
// @Accept json
// @Produce application/octet-stream
// @Param filename path string true "Filename to download"
// @Success 200 {file} file "CSV/JSON file content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/somatogramm/download/{filename} [get]
// @Deprecated
func (h *SomatogrammCompatibilityHandler) DownloadSomatogrammFile(c *gin.Context) {
	h.logger.Warn("DEPRECATED: /api/v1/somatogramm/download endpoint used. Please migrate to /api/v1/kader-planung/download")

	filename := c.Param("filename")
	if filename == "" {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Filename parameter is required"))
		return
	}

	// Validate filename for security
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Filename contains invalid characters"))
		return
	}

	// Check if file exists in available files
	files, err := h.adapter.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to list files"))
		return
	}

	var targetFile *services.FileInfo
	for _, file := range files {
		if file.Name == filename {
			targetFile = &file
			break
		}
	}

	if targetFile == nil {
		utils.SendJSONResponse(c, http.StatusNotFound,
			errors.NewNotFoundError("The requested file does not exist"))
		return
	}

	// Build full file path
	outputDir := h.adapter.GetOutputDir()
	fullPath := filepath.Join(outputDir, targetFile.Path)

	// Add deprecation header
	c.Header("X-Deprecation-Warning", "This endpoint is deprecated. Please use /api/v1/kader-planung/download/{filename} for new integrations.")

	// Serve the file
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(fullPath)
}

// SomatogrammCompatibilityRequest represents the request body for manual execution (backward compatibility)
type SomatogrammCompatibilityRequest struct {
	OutputFormat  string `json:"output_format,omitempty" example:"csv"`
	Timeout       int    `json:"timeout,omitempty" example:"30"`
	Concurrency   int    `json:"concurrency,omitempty" example:"4"`
	MinSampleSize int    `json:"min_sample_size,omitempty" example:"100"`
	Verbose       *bool  `json:"verbose,omitempty" example:"false"`
}