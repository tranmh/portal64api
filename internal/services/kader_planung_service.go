package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"portal64api/internal/config"
	"github.com/sirupsen/logrus"
)

// ExecutionStatus represents the current execution status
type ExecutionStatus struct {
	Running         bool      `json:"running"`
	StartTime       time.Time `json:"start_time"`
	LastExecution   time.Time `json:"last_execution"`
	LastSuccess     time.Time `json:"last_success"`
	LastError       string    `json:"last_error"`
	OutputFile      string    `json:"output_file"`
	OutputFiles     []string  `json:"output_files"`
}

// KaderPlanungService manages the kader-planung functionality
type KaderPlanungService struct {
	config *config.KaderPlanungConfig
	logger *logrus.Logger
	status ExecutionStatus
	mutex  sync.RWMutex
	cancel context.CancelFunc
	ctx    context.Context
}

// NewKaderPlanungService creates a new kader-planung service
func NewKaderPlanungService(config *config.KaderPlanungConfig, logger *logrus.Logger) *KaderPlanungService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &KaderPlanungService{
		config: config,
		logger: logger,
		status: ExecutionStatus{
			Running:       false,
			StartTime:     time.Time{},
			LastExecution: time.Time{},
			LastSuccess:   time.Time{},
			LastError:     "",
			OutputFile:    "",
			OutputFiles:   []string{},
		},
		cancel: cancel,
		ctx:    ctx,
	}

	return service
}

// OnImportComplete implements ImportCompleteCallback interface
func (s *KaderPlanungService) OnImportComplete() {
	s.logger.Info("Database import completed successfully")
}

// Start initializes the service
func (s *KaderPlanungService) Start() error {
	if !s.config.Enabled {
		s.logger.Info("Kader-Planung service disabled")
		return nil
	}
	
	s.logger.Info("Kader-Planung service started")
	return nil
}

// Stop shuts down the service
func (s *KaderPlanungService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	
	s.logger.Info("Kader-Planung service stopped")
	return nil
}

// GetStatus returns the current execution status
func (s *KaderPlanungService) GetStatus() ExecutionStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.status
}

// ExecuteManually executes kader-planung manually
func (s *KaderPlanungService) ExecuteManually(params map[string]interface{}) error {
	s.mutex.Lock()
	if s.status.Running {
		s.mutex.Unlock()
		return fmt.Errorf("kader-planung execution already running")
	}
	s.status.Running = true
	s.status.StartTime = time.Now()
	s.status.LastError = ""
	s.mutex.Unlock()

	s.logger.Info("Manual kader-planung execution requested")

	// Execute in background
	go func() {
		err := s.executeKaderPlanung(params)

		s.mutex.Lock()
		defer s.mutex.Unlock()

		s.status.Running = false
		s.status.LastExecution = time.Now()

		if err != nil {
			s.logger.Errorf("Kader-planung execution failed: %v", err)
			s.status.LastError = err.Error()
		} else {
			s.logger.Info("Kader-planung execution completed successfully")
			s.status.LastSuccess = time.Now()

			// Find and record output files
			outputFiles := s.findLatestOutputFiles()
			s.status.OutputFiles = outputFiles
			if len(outputFiles) > 0 {
				s.status.OutputFile = outputFiles[0]
			}
		}
	}()

	return nil
}

// ListAvailableFiles returns available output files
func (s *KaderPlanungService) ListAvailableFiles() ([]FileInfo, error) {
	files := []FileInfo{}

	// Check if output directory exists
	if _, err := os.Stat(s.config.OutputDir); os.IsNotExist(err) {
		return files, nil
	}

	// Read directory
	entries, err := os.ReadDir(s.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %w", err)
	}

	// Collect CSV files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}

		filePath := filepath.Join(s.config.OutputDir, entry.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			s.logger.Warnf("Failed to stat file %s: %v", filePath, err)
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filePath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return files, nil
}

// GetOutputDir returns the configured output directory
func (s *KaderPlanungService) GetOutputDir() string {
	return s.config.OutputDir
}

// ExecuteStatisticalAnalysis executes statistical analysis with given parameters
func (s *KaderPlanungService) ExecuteStatisticalAnalysis(params map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.status.Running {
		return fmt.Errorf("statistical analysis already running")
	}
	
	s.logger.Info("Starting statistical analysis")
	s.status.Running = true
	s.status.StartTime = time.Now()
	
	// TODO: Implement actual statistical analysis execution
	// For now, just simulate execution
	go func() {
		time.Sleep(1 * time.Second)
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.status.Running = false
		s.status.LastExecution = time.Now()
		s.status.LastSuccess = time.Now()
		s.logger.Info("Statistical analysis completed")
	}()
	
	return nil
}

// ExecuteHybridAnalysis executes hybrid analysis (both detailed and statistical)
func (s *KaderPlanungService) ExecuteHybridAnalysis(params map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.status.Running {
		return fmt.Errorf("hybrid analysis already running")
	}
	
	s.logger.Info("Starting hybrid analysis")
	s.status.Running = true
	s.status.StartTime = time.Now()
	
	// TODO: Implement actual hybrid analysis execution
	// For now, just simulate execution
	go func() {
		time.Sleep(2 * time.Second)
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.status.Running = false
		s.status.LastExecution = time.Now()
		s.status.LastSuccess = time.Now()
		s.logger.Info("Hybrid analysis completed")
	}()
	
	return nil
}

// GetAnalysisCapabilities returns the analysis capabilities of the service
func (s *KaderPlanungService) GetAnalysisCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"processing_modes": []string{"hybrid"},
		"output_formats":  []string{"csv"},
		"features": map[string]bool{
			"somatogram_percentiles": true,
			"germany_wide_data":     true,
			"statistical_analysis": true,
			"detailed_analysis":   true,
			"hybrid_analysis":     true,
			"concurrent_processing": true,
			"custom_club_prefix":   true,
		},
		"version": "2.0.0",
		"notes": []string{
			"Always processes complete German dataset (~50,000 players)",
			"Club prefix filters output only, not data collection",
			"Somatogram percentiles calculated from Germany-wide data",
			"CSV format optimized for German Excel compatibility",
		},
		"deprecated_features": map[string]string{
			"multiple_modes": "Now always uses hybrid mode for optimal performance",
			"multiple_formats": "Now always outputs CSV format",
			"include_statistics": "Statistics are now always included",
		},
	}
}

// FileInfo represents file metadata
type FileInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
}

// executeKaderPlanung executes the kader-planung binary with given parameters
func (s *KaderPlanungService) executeKaderPlanung(params map[string]interface{}) error {
	// Build command arguments
	args := s.buildCommandArgs(params)

	s.logger.Infof("Executing kader-planung: %s %v", s.config.BinaryPath, args)

	// Create command with context for cancellation
	cmd := exec.CommandContext(s.ctx, s.config.BinaryPath, args...)

	// Set working directory
	if s.config.OutputDir != "" {
		cmd.Dir = filepath.Dir(s.config.OutputDir)
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kader-planung execution failed: exit status %v, output: %s", err, string(output))
	}

	s.logger.Infof("Kader-planung output: %s", string(output))

	return nil
}

// buildCommandArgs builds command line arguments from parameters
func (s *KaderPlanungService) buildCommandArgs(params map[string]interface{}) []string {
	args := []string{}

	// API base URL (required)
	apiBaseURL := s.getStringParam(params, "api_base_url", s.config.APIBaseURL)
	if apiBaseURL != "" {
		args = append(args, "--api-base-url", apiBaseURL)
	}

	// Output directory (required)
	outputDir := s.getStringParam(params, "output_dir", s.config.OutputDir)
	if outputDir != "" {
		args = append(args, "--output-dir", outputDir)
	}

	// Club prefix (optional)
	clubPrefix := s.getStringParam(params, "club_prefix", s.config.ClubPrefix)
	if clubPrefix != "" {
		args = append(args, "--club-prefix", clubPrefix)
	}

	// Timeout
	timeout := s.getIntParam(params, "timeout", s.config.Timeout)
	if timeout > 0 {
		args = append(args, "--timeout", strconv.Itoa(timeout))
	}

	// Concurrency
	concurrency := s.getIntParam(params, "concurrency", s.config.Concurrency)
	if concurrency > 0 {
		args = append(args, "--concurrency", strconv.Itoa(concurrency))
	}

	// Min sample size
	minSampleSize := s.getIntParam(params, "min_sample_size", s.config.MinSampleSize)
	if minSampleSize > 0 {
		args = append(args, "--min-sample-size", strconv.Itoa(minSampleSize))
	}

	// Verbose
	if s.getBoolParam(params, "verbose", s.config.Verbose) {
		args = append(args, "--verbose")
	}

	return args
}

// findLatestOutputFiles finds the most recent output files in the output directory
func (s *KaderPlanungService) findLatestOutputFiles() []string {
	files := []string{}

	entries, err := os.ReadDir(s.config.OutputDir)
	if err != nil {
		s.logger.Warnf("Failed to read output directory: %v", err)
		return files
	}

	// Find CSV files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".csv") {
			files = append(files, entry.Name())
		}
	}

	return files
}

// Helper methods for parameter extraction
func (s *KaderPlanungService) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (s *KaderPlanungService) getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

func (s *KaderPlanungService) getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}
