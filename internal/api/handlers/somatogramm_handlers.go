package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"portal64api/internal/services"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// DEPRECATED: SomatogrammHandler is deprecated and will be removed in a future version.
// Use SomatogrammCompatibilityHandler or migrate to unified Kader-Planung endpoints.
// See MIGRATION_PHASE4_DEPRECATION_NOTICE.md for migration instructions.
//
// SomatogrammHandler handles Somatogramm related requests
type SomatogrammHandler struct {
	service *services.SomatogrammService
}

// NewSomatogrammHandler creates a new Somatogramm handler
func NewSomatogrammHandler(service *services.SomatogrammService) *SomatogrammHandler {
	return &SomatogrammHandler{
		service: service,
	}
}

// GetSomatogrammStatus returns the current execution status
// @Summary Get Somatogramm execution status
// @Description Returns current execution status including running state, last execution time, and available files
// @Tags somatogramm
// @Accept json
// @Produce json
// @Success 200 {object} services.ExecutionStatus
// @Failure 500 {object} errors.APIError
// @Router /api/v1/somatogramm/status [get]
func (h *SomatogrammHandler) GetSomatogrammStatus(c *gin.Context) {
	status := h.service.GetStatus()

	// Add available files to the status
	files, err := h.service.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to list available files"))
		return
	}

	response := map[string]interface{}{
		"status":          status,
		"available_files": files,
	}

	utils.SendJSONResponse(c, http.StatusOK, response)
}

// StartSomatogrammExecution starts manual Somatogramm execution
// @Summary Start manual Somatogramm execution
// @Description Starts manual execution of Somatogramm with optional parameters
// @Tags somatogramm
// @Accept json
// @Produce json
// @Param request body SomatogrammRequest false "Execution parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.APIError
// @Failure 409 {object} errors.APIError "Already running"
// @Failure 500 {object} errors.APIError
// @Router /api/v1/somatogramm/start [post]
func (h *SomatogrammHandler) StartSomatogrammExecution(c *gin.Context) {
	var request SomatogrammRequest
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

	if err := h.service.ExecuteManually(params); err != nil {
		if strings.Contains(err.Error(), "already running") {
			utils.SendJSONResponse(c, http.StatusConflict,
				errors.NewAPIError(409, "EXECUTION_ALREADY_RUNNING", "Execution already in progress"))
			return
		}

		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to start execution"))
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, map[string]string{
		"message": "Somatogramm execution started",
	})
}

// ListSomatogrammFiles returns list of available CSV/JSON files
// @Summary List available Somatogramm files
// @Description Returns a list of all available Somatogramm CSV/JSON files with metadata
// @Tags somatogramm
// @Accept json
// @Produce json
// @Success 200 {array} services.FileInfo
// @Failure 500 {object} errors.APIError
// @Router /api/v1/somatogramm/files [get]
func (h *SomatogrammHandler) ListSomatogrammFiles(c *gin.Context) {
	files, err := h.service.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to list files"))
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, files)
}

// DownloadSomatogrammFile serves a specific CSV/JSON file for download
// @Summary Download Somatogramm file
// @Description Downloads a specific Somatogramm CSV or JSON file
// @Tags somatogramm
// @Accept json
// @Produce application/octet-stream
// @Param filename path string true "Filename to download"
// @Success 200 {file} file "CSV or JSON file content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/somatogramm/download/{filename} [get]
func (h *SomatogrammHandler) DownloadSomatogrammFile(c *gin.Context) {
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
	files, err := h.service.ListAvailableFiles()
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
	outputDir := h.service.GetOutputDir()
	fullPath := filepath.Join(outputDir, targetFile.Path)

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(filename, ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(filename, ".csv") {
		contentType = "text/csv"
	}

	// Serve the file
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)
	c.File(fullPath)
}

// SomatogrammRequest represents the request body for manual execution
type SomatogrammRequest struct {
	OutputFormat  string `json:"output_format,omitempty" example:"csv"`
	Timeout       int    `json:"timeout,omitempty" example:"30"`
	Concurrency   int    `json:"concurrency,omitempty" example:"4"`
	MinSampleSize int    `json:"min_sample_size,omitempty" example:"100"`
	Verbose       *bool  `json:"verbose,omitempty" example:"false"`
}