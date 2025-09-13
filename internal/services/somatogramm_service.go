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

// SomatogrammService manages the integrated somatogramm functionality
// It implements ImportCompleteCallback to execute after successful imports
type SomatogrammService struct {
	config *config.SomatogrammConfig
	logger *logrus.Logger
	status ExecutionStatus
	mutex  sync.RWMutex
	cancel context.CancelFunc
	ctx    context.Context
}

// OnImportComplete implements ImportCompleteCallback interface
// This method is called when a database import completes successfully
func (s *SomatogrammService) OnImportComplete() {
	s.logger.Info("Database import completed successfully, triggering Somatogramm execution")
	s.ExecuteAfterImport()
}

// NewSomatogrammService creates a new somatogramm service
func NewSomatogrammService(config *config.SomatogrammConfig, logger *logrus.Logger) *SomatogrammService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &SomatogrammService{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	return service
}

// Start initializes the service
func (s *SomatogrammService) Start() error {
	if !s.config.Enabled {
		s.logger.Info("Somatogramm service disabled")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	s.logger.Info("Somatogramm service started")
	return nil
}

// Stop shuts down the service
func (s *SomatogrammService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Info("Somatogramm service stopped")
	return nil
}

// ExecuteAfterImport executes somatogramm after successful database import
func (s *SomatogrammService) ExecuteAfterImport() {
	if !s.config.Enabled {
		return
	}

	s.mutex.Lock()
	if s.status.Running {
		s.logger.Warn("Somatogramm already running, skipping execution")
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

		s.logger.Info("Starting Somatogramm execution after successful import")

		if err := s.executeSomatogramm(); err != nil {
			s.mutex.Lock()
			s.status.LastError = err.Error()
			s.mutex.Unlock()
			s.logger.Errorf("Somatogramm execution failed: %v", err)
		} else {
			s.mutex.Lock()
			s.status.LastSuccess = time.Now()
			s.status.LastError = ""
			s.mutex.Unlock()
			s.logger.Info("Somatogramm execution completed successfully")

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

// ExecuteManually executes somatogramm manually via API call
func (s *SomatogrammService) ExecuteManually(params map[string]interface{}) error {
	if !s.config.Enabled {
		return fmt.Errorf("Somatogramm service is disabled")
	}

	s.mutex.Lock()
	if s.status.Running {
		s.mutex.Unlock()
		return fmt.Errorf("Somatogramm is already running")
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

		s.logger.Info("Starting manual Somatogramm execution")

		if err := s.executeSomatogrammWithParams(params); err != nil {
			s.mutex.Lock()
			s.status.LastError = err.Error()
			s.mutex.Unlock()
			s.logger.Errorf("Manual Somatogramm execution failed: %v", err)
		} else {
			s.mutex.Lock()
			s.status.LastSuccess = time.Now()
			s.status.LastError = ""
			s.mutex.Unlock()
			s.logger.Info("Manual Somatogramm execution completed successfully")

			// Clean up old files
			if err := s.cleanupOldFiles(); err != nil {
				s.logger.Errorf("Failed to cleanup old files: %v", err)
			}
		}
	}()

	return nil
}

// GetStatus returns the current execution status
func (s *SomatogrammService) GetStatus() ExecutionStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.status
}

// ListAvailableFiles returns a list of available CSV/JSON files
func (s *SomatogrammService) ListAvailableFiles() ([]FileInfo, error) {
	files := []FileInfo{}

	err := filepath.Walk(s.config.OutputDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if (strings.HasSuffix(info.Name(), ".csv") || strings.HasSuffix(info.Name(), ".json")) &&
		   strings.HasPrefix(info.Name(), "somatogramm-") {
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

// executeSomatogramm runs the somatogramm binary with default settings
func (s *SomatogrammService) executeSomatogramm() error {
	return s.executeSomatogrammWithParams(map[string]interface{}{})
}

// executeSomatogrammWithParams runs the somatogramm binary with custom parameters
func (s *SomatogrammService) executeSomatogrammWithParams(params map[string]interface{}) error {
	// Build command arguments
	args := s.buildCommandArgs(params)

	// Create command with context for cancellation
	cmd := exec.CommandContext(s.ctx, s.config.BinaryPath, args...)

	// Set working directory to the output directory
	cmd.Dir = s.config.OutputDir

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("somatogramm execution failed: %v, output: %s", err, string(output))
	}

	s.logger.Debugf("Somatogramm output: %s", string(output))

	// Find the generated output files
	if outputFiles := s.findLatestOutputFiles(); len(outputFiles) > 0 {
		s.mutex.Lock()
		s.status.OutputFile = strings.Join(outputFiles, ", ")
		s.mutex.Unlock()
	}

	return nil
}

// buildCommandArgs builds command line arguments from parameters
func (s *SomatogrammService) buildCommandArgs(params map[string]interface{}) []string {
	args := []string{}

	// Use config defaults, override with params
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

	minSampleSize := s.getIntParam(params, "min_sample_size", s.config.MinSampleSize)
	args = append(args, "--min-sample-size", strconv.Itoa(minSampleSize))

	if s.getBoolParam(params, "verbose", s.config.Verbose) {
		args = append(args, "--verbose")
	}

	return args
}

// Helper functions for parameter extraction
func (s *SomatogrammService) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return defaultValue
}

func (s *SomatogrammService) getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key].(int); ok {
		return val
	}
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func (s *SomatogrammService) getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultValue
}

// findLatestOutputFiles finds the most recently created CSV/JSON files
func (s *SomatogrammService) findLatestOutputFiles() []string {
	files, err := s.ListAvailableFiles()
	if err != nil || len(files) == 0 {
		return []string{}
	}

	// Return the most recent files (up to 3 for male, female, divers)
	var result []string
	for i := 0; i < len(files) && i < 3; i++ {
		result = append(result, files[i].Name)
	}
	return result
}

// cleanupOldFiles removes files older than MaxVersions
func (s *SomatogrammService) cleanupOldFiles() error {
	files, err := s.ListAvailableFiles()
	if err != nil {
		return err
	}

	// Group files by gender to keep MaxVersions per gender
	genderFiles := make(map[string][]FileInfo)
	for _, file := range files {
		gender := "unknown"
		if strings.Contains(file.Name, "-male-") {
			gender = "male"
		} else if strings.Contains(file.Name, "-female-") {
			gender = "female"
		} else if strings.Contains(file.Name, "-divers-") {
			gender = "divers"
		}
		genderFiles[gender] = append(genderFiles[gender], file)
	}

	// Clean up each gender group
	for gender, genderFileList := range genderFiles {
		if len(genderFileList) <= s.config.MaxVersions {
			continue
		}

		// Sort by modification time, newest first
		sort.Slice(genderFileList, func(i, j int) bool {
			return genderFileList[i].ModTime.After(genderFileList[j].ModTime)
		})

		filesToDelete := genderFileList[s.config.MaxVersions:]
		for _, file := range filesToDelete {
			fullPath := filepath.Join(s.config.OutputDir, file.Path)
			if err := os.Remove(fullPath); err != nil {
				s.logger.Warnf("Failed to delete old file %s: %v", fullPath, err)
			} else {
				s.logger.Infof("Deleted old file: %s", file.Name)
			}
		}

		s.logger.Infof("Cleaned up %d old files for gender %s", len(filesToDelete), gender)
	}

	return nil
}

// GetOutputDir returns the configured output directory
func (s *SomatogrammService) GetOutputDir() string {
	return s.config.OutputDir
}