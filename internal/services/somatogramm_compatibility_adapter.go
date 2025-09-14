package services

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// SomatogrammCompatibilityAdapter provides backward compatibility for existing Somatogramm service calls
// It adapts Somatogramm calls to use the unified Kader-Planung service in statistical mode
type SomatogrammCompatibilityAdapter struct {
	kaderPlanungService *KaderPlanungService
	logger              *logrus.Logger
}

// NewSomatogrammCompatibilityAdapter creates a new compatibility adapter
func NewSomatogrammCompatibilityAdapter(kaderPlanungService *KaderPlanungService, logger *logrus.Logger) *SomatogrammCompatibilityAdapter {
	return &SomatogrammCompatibilityAdapter{
		kaderPlanungService: kaderPlanungService,
		logger:              logger,
	}
}

// OnImportComplete implements ImportCompleteCallback interface for backward compatibility
func (s *SomatogrammCompatibilityAdapter) OnImportComplete() {
	s.logger.Info("Database import completed, triggering Somatogramm-compatible statistical analysis")
	s.ExecuteAfterImport()
}

// Start provides backward compatibility for the Start method
func (s *SomatogrammCompatibilityAdapter) Start() error {
	s.logger.Info("Starting Somatogramm compatibility adapter (using unified Kader-Planung service)")
	return s.kaderPlanungService.Start()
}

// Stop provides backward compatibility for the Stop method
func (s *SomatogrammCompatibilityAdapter) Stop() error {
	s.logger.Info("Stopping Somatogramm compatibility adapter")
	return s.kaderPlanungService.Stop()
}
// ExecuteAfterImport executes statistical analysis after successful database import
// This provides backward compatibility with the original Somatogramm service behavior
func (s *SomatogrammCompatibilityAdapter) ExecuteAfterImport() {
	s.logger.Info("Executing Somatogramm-compatible analysis after import")

	// Use default Somatogramm-style parameters for statistical analysis
	params := s.getDefaultSomatogrammParams()

	if err := s.kaderPlanungService.ExecuteStatisticalAnalysis(params); err != nil {
		s.logger.Errorf("Somatogramm-compatible analysis failed: %v", err)
	}
}

// ExecuteManually executes somatogramm manually via API call with backward compatibility
func (s *SomatogrammCompatibilityAdapter) ExecuteManually(params map[string]interface{}) error {
	s.logger.Info("Executing Somatogramm-compatible manual analysis")

	// Convert Somatogramm params to unified format
	unifiedParams := s.convertSomatogrammParams(params)

	return s.kaderPlanungService.ExecuteStatisticalAnalysis(unifiedParams)
}

// GetStatus returns the current execution status in Somatogramm-compatible format
func (s *SomatogrammCompatibilityAdapter) GetStatus() ExecutionStatus {
	status := s.kaderPlanungService.GetStatus()

	// Ensure backward compatibility by adapting the status for Somatogramm
	s.adaptStatusForSomatogrammCompatibility(&status)

	return status
}

// ListAvailableFiles returns available files in Somatogramm-compatible format
func (s *SomatogrammCompatibilityAdapter) ListAvailableFiles() ([]FileInfo, error) {
	allFiles, err := s.kaderPlanungService.ListAvailableFiles()
	if err != nil {
		return nil, err
	}

	// Filter for statistical output files that match Somatogramm patterns
	var somatogrammFiles []FileInfo
	for _, file := range allFiles {
		if s.isSomatogrammCompatibleFile(file) {
			somatogrammFiles = append(somatogrammFiles, file)
		}
	}

	return somatogrammFiles, nil
}

// GetOutputDir returns the configured output directory
func (s *SomatogrammCompatibilityAdapter) GetOutputDir() string {
	return s.kaderPlanungService.GetOutputDir()
}
// convertSomatogrammParams converts legacy Somatogramm parameters to unified format
func (s *SomatogrammCompatibilityAdapter) convertSomatogrammParams(params map[string]interface{}) map[string]interface{} {
	unifiedParams := make(map[string]interface{})

	// Copy all original params
	for k, v := range params {
		unifiedParams[k] = v
	}

	// Ensure statistical processing mode is set
	unifiedParams["processing_mode"] = "statistical"

	// Map Somatogramm-specific parameters
	if val, ok := params["output_format"]; ok {
		unifiedParams["output_format"] = val
	} else {
		unifiedParams["output_format"] = "csv" // Somatogramm default
	}

	if val, ok := params["min_sample_size"]; ok {
		unifiedParams["min_sample_size"] = val
	} else {
		unifiedParams["min_sample_size"] = 10 // Somatogramm default
	}

	if val, ok := params["concurrency"]; ok {
		unifiedParams["concurrency"] = val
	}

	if val, ok := params["timeout"]; ok {
		unifiedParams["timeout"] = val
	}

	if val, ok := params["verbose"]; ok {
		unifiedParams["verbose"] = val
	}

	s.logger.Debugf("Converted Somatogramm params: %+v -> %+v", params, unifiedParams)

	return unifiedParams
}

// getDefaultSomatogrammParams returns default parameters matching original Somatogramm behavior
func (s *SomatogrammCompatibilityAdapter) getDefaultSomatogrammParams() map[string]interface{} {
	return map[string]interface{}{
		"processing_mode":  "statistical",
		"output_format":    "csv",
		"min_sample_size":  10,    // Somatogramm default
		"concurrency":      0,     // Use system CPU count
		"verbose":          false,
	}
}
// adaptStatusForSomatogrammCompatibility modifies status to match Somatogramm expectations
func (s *SomatogrammCompatibilityAdapter) adaptStatusForSomatogrammCompatibility(status *ExecutionStatus) {
	// If we have multiple output files from statistical analysis,
	// format them as comma-separated string for backward compatibility
	if len(status.OutputFiles) > 0 {
		// Join all statistical output files
		var fileNames []string
		for _, fileName := range status.OutputFiles {
			if s.isSomatogrammStyleFileName(fileName) {
				fileNames = append(fileNames, fileName)
			}
		}
		if len(fileNames) > 0 {
			// Use the most recent statistical file as primary output
			status.OutputFile = fileNames[0]
		}
	}
}

// isSomatogrammCompatibleFile checks if a file matches Somatogramm output patterns
func (s *SomatogrammCompatibilityAdapter) isSomatogrammCompatibleFile(file FileInfo) bool {
	return s.isSomatogrammStyleFileName(file.Name)
}

// isSomatogrammStyleFileName checks if filename matches Somatogramm patterns
func (s *SomatogrammCompatibilityAdapter) isSomatogrammStyleFileName(filename string) bool {
	// Match patterns like:
	// - kader-planung-statistical-male-YYYYMMDD-HHMMSS.csv
	// - kader-planung-statistical-female-YYYYMMDD-HHMMSS.csv
	// - somatogramm-male-YYYYMMDD-HHMMSS.csv (legacy)
	// - somatogramm-female-YYYYMMDD-HHMMSS.csv (legacy)

	return (strings.Contains(filename, "statistical-male-") ||
			strings.Contains(filename, "statistical-female-") ||
			strings.Contains(filename, "somatogramm-male-") ||
			strings.Contains(filename, "somatogramm-female-")) &&
		   (strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".json"))
}

// GetAnalysisCapabilities returns capabilities in Somatogramm-compatible format
func (s *SomatogrammCompatibilityAdapter) GetAnalysisCapabilities() map[string]interface{} {
	capabilities := s.kaderPlanungService.GetAnalysisCapabilities()

	// Add Somatogramm-specific compatibility info
	capabilities["somatogramm_compatible"] = true
	capabilities["migration_status"] = "compatibility_mode"
	capabilities["recommended_migration"] = "Use unified Kader-Planung service for new integrations"

	return capabilities
}
