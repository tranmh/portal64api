package services

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"portal64api/internal/config"

	"github.com/sirupsen/logrus"
)

// ExecutionStatus represents the current execution status
type ExecutionStatus struct {
	Running       bool      `json:"running"`
	StartTime     time.Time `json:"start_time"`
	LastExecution time.Time `json:"last_execution"`
	LastSuccess   time.Time `json:"last_success"`
	LastError     string    `json:"last_error"`
	OutputFile    string    `json:"output_file"`
}

// KaderPlanungService manages the integrated kader-planung functionality
// It implements ImportCompleteCallback to execute after successful imports
type KaderPlanungService struct {
	config *config.KaderPlanungConfig
	logger *logrus.Logger
	status ExecutionStatus
	mutex  sync.RWMutex
	cancel context.CancelFunc
	ctx    context.Context
}

// OnImportComplete implements ImportCompleteCallback interface
// This method is called when a database import completes successfully
func (s *KaderPlanungService) OnImportComplete() {
	s.logger.Info("Database import completed successfully, triggering Kader-Planung execution")
	s.ExecuteAfterImport()
}

// NewKaderPlanungService creates a new kader-planung service
func NewKaderPlanungService(config *config.KaderPlanungConfig, logger *logrus.Logger) *KaderPlanungService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &KaderPlanungService{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
	
	return service
}

// Start initializes the service
func (s *KaderPlanungService) Start() error {
	if !s.config.Enabled {
		s.logger.Info("Kader-Planung service disabled")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
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

// ExecuteAfterImport executes kader-planung after successful database import
func (s *KaderPlanungService) ExecuteAfterImport() {
	if !s.config.Enabled {
		return
	}

	s.mutex.Lock()
	if s.status.Running {
		s.logger.Warn("Kader-Planung already running, skipping execution")
		s.mutex.Unlock()
		return
	}
	s.status.Running = true
	s.status.StartTime = time.Now()
	s.mutex.Unlock()

	go func() {
		defer func() {
			s.mutex.Lock()
			s.status.Running = false
			s.mutex.Unlock()
		}()

		s.logger.Info("Starting Kader-Planung execution after successful import")
		
		if err := s.executeKaderPlanung(); err != nil {
			s.mutex.Lock()
			s.status.LastError = err.Error()
			s.mutex.Unlock()
			s.logger.Errorf("Kader-Planung execution failed: %v", err)
		} else {
			s.mutex.Lock()
			s.status.LastSuccess = time.Now()
			s.status.LastError = ""
			s.mutex.Unlock()
			s.logger.Info("Kader-Planung execution completed successfully")
			
			// Clean up old files
			if err := s.cleanupOldFiles(); err != nil {
				s.logger.Errorf("Failed to cleanup old files: %v", err)
			}
		}
		
		s.mutex.Lock()
		s.status.LastExecution = time.Now()
		s.mutex.Unlock()
	}()
}

// ExecuteManually executes kader-planung manually via API call
func (s *KaderPlanungService) ExecuteManually(params map[string]interface{}) error {
	if !s.config.Enabled {
		return fmt.Errorf("Kader-Planung service is disabled")
	}

	s.mutex.Lock()
	if s.status.Running {
		s.mutex.Unlock()
		return fmt.Errorf("Kader-Planung is already running")
	}
	s.status.Running = true
	s.status.StartTime = time.Now()
	s.mutex.Unlock()

	go func() {
		defer func() {
			s.mutex.Lock()
			s.status.Running = false
			s.status.LastExecution = time.Now()
			s.mutex.Unlock()
		}()

		s.logger.Info("Starting manual Kader-Planung execution")
		
		if err := s.executeKaderPlanungWithParams(params); err != nil {
			s.mutex.Lock()
			s.status.LastError = err.Error()
			s.mutex.Unlock()
			s.logger.Errorf("Manual Kader-Planung execution failed: %v", err)
		} else {
			s.mutex.Lock()
			s.status.LastSuccess = time.Now()
			s.status.LastError = ""
			s.mutex.Unlock()
			s.logger.Info("Manual Kader-Planung execution completed successfully")
			
			// Clean up old files
			if err := s.cleanupOldFiles(); err != nil {
				s.logger.Errorf("Failed to cleanup old files: %v", err)
			}
		}
	}()

	return nil
}

// GetStatus returns the current execution status
func (s *KaderPlanungService) GetStatus() ExecutionStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.status
}

// ListAvailableFiles returns a list of available CSV files
func (s *KaderPlanungService) ListAvailableFiles() ([]FileInfo, error) {
	files := []FileInfo{}
	
	err := filepath.Walk(s.config.OutputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if strings.HasSuffix(info.Name(), ".csv") && strings.HasPrefix(info.Name(), "kader-planung-") {
			relPath, _ := filepath.Rel(s.config.OutputDir, path)
			files = append(files, FileInfo{
				Name:     info.Name(),
				Path:     relPath,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
			})
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort by modification time, newest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})
	
	return files, nil
}

// FileInfo represents information about a generated file
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`  
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// executeKaderPlanung runs the kader-planung binary with default settings
func (s *KaderPlanungService) executeKaderPlanung() error {
	return s.executeKaderPlanungWithParams(map[string]interface{}{})
}

// executeKaderPlanungWithParams runs the kader-planung binary with custom parameters
func (s *KaderPlanungService) executeKaderPlanungWithParams(params map[string]interface{}) error {
	// Build command arguments
	args := s.buildCommandArgs(params)
	
	// Create command with context for cancellation
	cmd := exec.CommandContext(s.ctx, s.config.BinaryPath, args...)
	
	// Set working directory to the output directory
	cmd.Dir = s.config.OutputDir
	
	// Note: Process priority setting removed for cross-platform compatibility
	// To run with lower priority, consider using nice on Unix systems
	// or Process.SetPriorityClass on Windows after starting the process
	
	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kader-planung execution failed: %v, output: %s", err, string(output))
	}
	
	s.logger.Debugf("Kader-Planung output: %s", string(output))
	
	// Find the generated output file
	if outputFile := s.findLatestOutputFile(); outputFile != "" {
		s.mutex.Lock()
		s.status.OutputFile = outputFile
		s.mutex.Unlock()
	}
	
	return nil
}

// buildCommandArgs builds command line arguments from parameters
func (s *KaderPlanungService) buildCommandArgs(params map[string]interface{}) []string {
	args := []string{}
	
	// Use config defaults, override with params
	clubPrefix := s.getStringParam(params, "club_prefix", s.config.ClubPrefix)
	if clubPrefix != "" {
		args = append(args, "--club-prefix", clubPrefix)
	}
	
	outputFormat := s.getStringParam(params, "output_format", s.config.OutputFormat)
	args = append(args, "--output-format", outputFormat)
	
	args = append(args, "--output-dir", ".")
	args = append(args, "--api-base-url", s.config.APIBaseURL)
	
	timeout := s.getIntParam(params, "timeout", s.config.Timeout)
	args = append(args, "--timeout", strconv.Itoa(timeout))
	
	concurrency := s.getIntParam(params, "concurrency", s.config.Concurrency)
	if concurrency == 0 {
		concurrency = runtime.NumCPU()
	}
	args = append(args, "--concurrency", strconv.Itoa(concurrency))
	
	if s.getBoolParam(params, "verbose", s.config.Verbose) {
		args = append(args, "--verbose")
	}
	
	return args
}

// Helper functions for parameter extraction
func (s *KaderPlanungService) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return defaultValue
}

func (s *KaderPlanungService) getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key].(int); ok {
		return val
	}
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func (s *KaderPlanungService) getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultValue
}

// findLatestOutputFile finds the most recently created CSV file
func (s *KaderPlanungService) findLatestOutputFile() string {
	files, err := s.ListAvailableFiles()
	if err != nil || len(files) == 0 {
		return ""
	}
	return files[0].Name
}

// cleanupOldFiles removes files older than MaxVersions
func (s *KaderPlanungService) cleanupOldFiles() error {
	files, err := s.ListAvailableFiles()
	if err != nil {
		return err
	}
	
	// Keep only the newest MaxVersions files
	if len(files) <= s.config.MaxVersions {
		return nil
	}
	
	filesToDelete := files[s.config.MaxVersions:]
	for _, file := range filesToDelete {
		fullPath := filepath.Join(s.config.OutputDir, file.Path)
		if err := os.Remove(fullPath); err != nil {
			s.logger.Warnf("Failed to delete old file %s: %v", fullPath, err)
		} else {
			s.logger.Infof("Deleted old file: %s", file.Name)
		}
	}
	
	return nil
}

// GetOutputDir returns the configured output directory
func (s *KaderPlanungService) GetOutputDir() string {
	return s.config.OutputDir
}
