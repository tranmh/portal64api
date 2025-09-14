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

// KaderPlanungHandler handles Kader-Planung related requests
type KaderPlanungHandler struct {
	service *services.KaderPlanungService
}

// NewKaderPlanungHandler creates a new Kader-Planung handler
func NewKaderPlanungHandler(service *services.KaderPlanungService) *KaderPlanungHandler {
	return &KaderPlanungHandler{
		service: service,
	}
}

// GetKaderPlanungStatus returns the current execution status
// @Summary Get Kader-Planung execution status
// @Description Returns current execution status including running state, last execution time, and available files
// @Tags kader-planung
// @Accept json
// @Produce json
// @Success 200 {object} services.ExecutionStatus
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/status [get]
func (h *KaderPlanungHandler) GetKaderPlanungStatus(c *gin.Context) {
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

// StartKaderPlanungExecution starts manual Kader-Planung execution
// @Summary Start manual Kader-Planung execution  
// @Description Starts manual execution of Kader-Planung with optional parameters
// @Tags kader-planung
// @Accept json
// @Produce json
// @Param request body KaderPlanungRequest false "Execution parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.APIError
// @Failure 409 {object} errors.APIError "Already running"
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/start [post]
func (h *KaderPlanungHandler) StartKaderPlanungExecution(c *gin.Context) {
	var request KaderPlanungRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, 
			errors.NewBadRequestError("Invalid request format"))
		return
	}
	
	// Convert request to parameters map
	params := map[string]interface{}{}
	if request.ClubPrefix != "" {
		params["club_prefix"] = request.ClubPrefix
	}
	if request.OutputFormat != "" {
		params["output_format"] = request.OutputFormat
	}
	if request.Timeout > 0 {
		params["timeout"] = request.Timeout
	}
	if request.Concurrency > 0 {
		params["concurrency"] = request.Concurrency
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
		"message": "Kader-Planung execution started",
	})
}

// ListKaderPlanungFiles returns list of available CSV files
// @Summary List available Kader-Planung files
// @Description Returns a list of all available Kader-Planung CSV files with metadata
// @Tags kader-planung
// @Accept json
// @Produce json
// @Success 200 {array} services.FileInfo
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/files [get]
func (h *KaderPlanungHandler) ListKaderPlanungFiles(c *gin.Context) {
	files, err := h.service.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to list files"))
		return
	}
	
	utils.SendJSONResponse(c, http.StatusOK, files)
}

// DownloadKaderPlanungFile serves a specific CSV file for download
// @Summary Download Kader-Planung file
// @Description Downloads a specific Kader-Planung CSV file
// @Tags kader-planung
// @Accept json
// @Produce application/octet-stream
// @Param filename path string true "Filename to download"
// @Success 200 {file} file "CSV file content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/kader-planung/download/{filename} [get]
func (h *KaderPlanungHandler) DownloadKaderPlanungFile(c *gin.Context) {
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
	
	// Serve the file
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(fullPath)
}

// KaderPlanungRequest represents the request body for manual execution
type KaderPlanungRequest struct {
	ClubPrefix   string `json:"club_prefix,omitempty" example:"C0"`
	OutputFormat string `json:"output_format,omitempty" example:"csv"`
	Timeout      int    `json:"timeout,omitempty" example:"30"`
	Concurrency  int    `json:"concurrency,omitempty" example:"4"`
	Verbose      *bool  `json:"verbose,omitempty" example:"false"`
}

// StatisticalAnalysisRequest represents the request body for statistical analysis
type StatisticalAnalysisRequest struct {
	MinSampleSize int      `json:"min_sample_size,omitempty" example:"100"`
	Timeout       int      `json:"timeout,omitempty" example:"30"`
	Concurrency   int      `json:"concurrency,omitempty" example:"4"`
	Verbose       *bool    `json:"verbose,omitempty" example:"false"`
	Formats       []string `json:"formats,omitempty" example:"csv,json"`
}

// HybridAnalysisRequest represents the request body for hybrid analysis
type HybridAnalysisRequest struct {
	ClubPrefix    string   `json:"club_prefix,omitempty" example:"C0"`
	OutputFormat  string   `json:"output_format,omitempty" example:"csv"`
	MinSampleSize int      `json:"min_sample_size,omitempty" example:"100"`
	Timeout       int      `json:"timeout,omitempty" example:"30"`
	Concurrency   int      `json:"concurrency,omitempty" example:"4"`
	Verbose       *bool    `json:"verbose,omitempty" example:"false"`
	Formats       []string `json:"formats,omitempty" example:"csv,json"`
}

// ExecuteStatisticalAnalysis starts statistical analysis execution
// @Summary Start statistical analysis execution
// @Description Starts statistical analysis execution (Somatogramm-style) with optimized API calls
// @Tags kader-planung
// @Accept json
// @Produce json
// @Param request body StatisticalAnalysisRequest false "Statistical analysis parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.APIError
// @Failure 409 {object} errors.APIError "Already running"
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/statistical [post]
func (h *KaderPlanungHandler) ExecuteStatisticalAnalysis(c *gin.Context) {
	var request StatisticalAnalysisRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid request format"))
		return
	}

	// Convert request to parameters map
	params := map[string]interface{}{}
	if request.MinSampleSize > 0 {
		params["min_sample_size"] = request.MinSampleSize
	}
	if request.Timeout > 0 {
		params["timeout"] = request.Timeout
	}
	if request.Concurrency > 0 {
		params["concurrency"] = request.Concurrency
	}
	if request.Verbose != nil {
		params["verbose"] = *request.Verbose
	}
	if len(request.Formats) > 0 {
		params["formats"] = request.Formats
	}

	if err := h.service.ExecuteStatisticalAnalysis(params); err != nil {
		if strings.Contains(err.Error(), "already running") {
			utils.SendJSONResponse(c, http.StatusConflict,
				errors.NewAPIError(409, "EXECUTION_ALREADY_RUNNING", "Statistical analysis already in progress"))
			return
		}

		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to start statistical analysis"))
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, map[string]string{
		"message": "Statistical analysis started",
	})
}

// ExecuteHybridAnalysis starts hybrid analysis execution
// @Summary Start hybrid analysis execution
// @Description Starts hybrid analysis execution (both detailed and statistical analysis)
// @Tags kader-planung
// @Accept json
// @Produce json
// @Param request body HybridAnalysisRequest false "Hybrid analysis parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.APIError
// @Failure 409 {object} errors.APIError "Already running"
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/hybrid [post]
func (h *KaderPlanungHandler) ExecuteHybridAnalysis(c *gin.Context) {
	var request HybridAnalysisRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid request format"))
		return
	}

	// Convert request to parameters map
	params := map[string]interface{}{}
	if request.ClubPrefix != "" {
		params["club_prefix"] = request.ClubPrefix
	}
	if request.OutputFormat != "" {
		params["output_format"] = request.OutputFormat
	}
	if request.MinSampleSize > 0 {
		params["min_sample_size"] = request.MinSampleSize
	}
	if request.Timeout > 0 {
		params["timeout"] = request.Timeout
	}
	if request.Concurrency > 0 {
		params["concurrency"] = request.Concurrency
	}
	if request.Verbose != nil {
		params["verbose"] = *request.Verbose
	}
	if len(request.Formats) > 0 {
		params["formats"] = request.Formats
	}

	if err := h.service.ExecuteHybridAnalysis(params); err != nil {
		if strings.Contains(err.Error(), "already running") {
			utils.SendJSONResponse(c, http.StatusConflict,
				errors.NewAPIError(409, "EXECUTION_ALREADY_RUNNING", "Hybrid analysis already in progress"))
			return
		}

		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to start hybrid analysis"))
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, map[string]string{
		"message": "Hybrid analysis started",
	})
}

// GetStatisticalResults returns results from statistical analysis
// @Summary Get statistical analysis results
// @Description Returns available statistical analysis result files
// @Tags kader-planung
// @Accept json
// @Produce json
// @Success 200 {array} services.FileInfo
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/statistical/files [get]
func (h *KaderPlanungHandler) GetStatisticalResults(c *gin.Context) {
	files, err := h.service.ListAvailableFiles()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to list statistical results"))
		return
	}

	// Filter for statistical files
	statisticalFiles := []services.FileInfo{}
	for _, file := range files {
		if strings.Contains(file.Name, "statistical") || strings.Contains(file.Name, "male") || strings.Contains(file.Name, "female") {
			statisticalFiles = append(statisticalFiles, file)
		}
	}

	utils.SendJSONResponse(c, http.StatusOK, statisticalFiles)
}

// GetAnalysisCapabilities returns supported analysis modes and formats
// @Summary Get analysis capabilities
// @Description Returns supported processing modes, output formats, and features
// @Tags kader-planung
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} errors.APIError
// @Router /api/v1/kader-planung/capabilities [get]
func (h *KaderPlanungHandler) GetAnalysisCapabilities(c *gin.Context) {
	capabilities := h.service.GetAnalysisCapabilities()
	utils.SendJSONResponse(c, http.StatusOK, capabilities)
}
