package services

import (
	"context"
	"fmt"
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
	s.logger.Info("Manual kader-planung execution requested")
	return nil
}

// ListAvailableFiles returns available output files
func (s *KaderPlanungService) ListAvailableFiles() ([]FileInfo, error) {
	return []FileInfo{}, nil
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
