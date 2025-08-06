package importers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/config"
	"portal64api/internal/models"
	"strings"
	"time"
)

// FreshnessChecker handles file freshness comparison
type FreshnessChecker struct {
	config       *config.FreshnessConfig
	metadataFile string
	logger       *log.Logger
}

// NewFreshnessChecker creates a new freshness checker instance
func NewFreshnessChecker(config *config.FreshnessConfig, metadataFile string, logger *log.Logger) *FreshnessChecker {
	return &FreshnessChecker{
		config:       config,
		metadataFile: metadataFile,
		logger:       logger,
	}
}

// CheckFreshness compares remote files with last imported files
func (fc *FreshnessChecker) CheckFreshness(remoteFiles []models.FileMetadata) (*models.FreshnessResult, error) {
	if !fc.config.Enabled {
		fc.logger.Println("Freshness checking is disabled, proceeding with import")
		return &models.FreshnessResult{
			ShouldImport: true,
			Reason:       "freshness_check_disabled",
			RemoteFiles:  remoteFiles,
		}, nil
	}

	// Load last import metadata
	lastImport, err := fc.loadLastImportMetadata()
	if err != nil {
		fc.logger.Printf("Failed to load last import metadata: %v", err)
		// First import - no metadata exists
		return &models.FreshnessResult{
			ShouldImport: true,
			Reason:       "first_import",
			RemoteFiles:  remoteFiles,
		}, nil
	}

	result := &models.FreshnessResult{
		ShouldImport: false,
		RemoteFiles:  remoteFiles,
		LastImported: lastImport.LastImport.Files,
		Comparisons:  make([]models.FileComparison, 0),
	}

	// Compare each remote file with last imported files
	for _, remoteFile := range remoteFiles {
		lastFile := fc.findMatchingFile(lastImport.LastImport.Files, remoteFile)
		comparison := fc.compareFiles(remoteFile, lastFile)
		result.Comparisons = append(result.Comparisons, comparison)

		if comparison.IsNewer {
			result.ShouldImport = true
			result.Reason = "newer_files_available"
		}
	}

	if !result.ShouldImport {
		result.Reason = "no_newer_files"
	}

	fc.logger.Printf("Freshness check completed: %s", result.Reason)
	return result, nil
}

// SaveImportMetadata saves successful import metadata to file
func (fc *FreshnessChecker) SaveImportMetadata(files []models.FileMetadata) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fc.metadataFile), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metadata := models.LastImportMetadata{
		LastImport: models.ImportRecord{
			Timestamp: time.Now(),
			Success:   true,
			Files:     files,
		},
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal import metadata: %w", err)
	}

	if err := os.WriteFile(fc.metadataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write import metadata: %w", err)
	}

	fc.logger.Printf("Saved import metadata to %s", fc.metadataFile)
	return nil
}

// loadLastImportMetadata loads the last import metadata from file
func (fc *FreshnessChecker) loadLastImportMetadata() (*models.LastImportMetadata, error) {
	if _, err := os.Stat(fc.metadataFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("metadata file does not exist: %s", fc.metadataFile)
	}

	data, err := os.ReadFile(fc.metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata models.LastImportMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// findMatchingFile finds a file in the last imported files that matches the remote file
func (fc *FreshnessChecker) findMatchingFile(lastFiles []models.FileMetadata, remoteFile models.FileMetadata) *models.FileMetadata {
	for _, lastFile := range lastFiles {
		// Try exact filename match first
		if lastFile.Filename == remoteFile.Filename {
			return &lastFile
		}
		
		// Try pattern matching if patterns are available
		if remoteFile.Pattern != "" && lastFile.Pattern != "" &&
		   remoteFile.Pattern == lastFile.Pattern {
			return &lastFile
		}
		
		// Try fuzzy matching based on database target
		if remoteFile.Database != "" && lastFile.Database != "" &&
		   remoteFile.Database == lastFile.Database {
			return &lastFile
		}
	}
	
	return nil
}

// compareFiles compares a remote file with the last imported file
func (fc *FreshnessChecker) compareFiles(remote models.FileMetadata, last *models.FileMetadata) models.FileComparison {
	comparison := models.FileComparison{
		RemoteFile: &remote,
		LastFile:   last,
		IsNewer:    false,
		Reasons:    make([]string, 0),
	}

	if last == nil {
		comparison.IsNewer = true
		comparison.Reasons = append(comparison.Reasons, "file_not_found_in_last_import")
		return comparison
	}

	// Compare modification time
	if fc.config.CompareTimestamp && remote.ModTime.After(last.ModTime) {
		comparison.IsNewer = true
		comparison.Reasons = append(comparison.Reasons, "newer_timestamp")
	}

	// Compare file size
	if fc.config.CompareSize && remote.Size != last.Size {
		comparison.IsNewer = true
		comparison.Reasons = append(comparison.Reasons, "different_size")
	}

	// Compare checksum (optional, slower)
	if fc.config.CompareChecksum && remote.Checksum != "" &&
		last.Checksum != "" && remote.Checksum != last.Checksum {
		comparison.IsNewer = true
		comparison.Reasons = append(comparison.Reasons, "different_checksum")
	}

	return comparison
}

// GetLastImportInfo returns information about the last successful import
func (fc *FreshnessChecker) GetLastImportInfo() (*models.ImportRecord, error) {
	metadata, err := fc.loadLastImportMetadata()
	if err != nil {
		return nil, err
	}
	return &metadata.LastImport, nil
}

// RemoveMetadataFile removes the metadata file (useful for testing or reset)
func (fc *FreshnessChecker) RemoveMetadataFile() error {
	if _, err := os.Stat(fc.metadataFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to remove
	}
	
	if err := os.Remove(fc.metadataFile); err != nil {
		return fmt.Errorf("failed to remove metadata file: %w", err)
	}
	
	fc.logger.Printf("Removed metadata file: %s", fc.metadataFile)
	return nil
}

// ValidateMetadataFile checks if the metadata file is valid
func (fc *FreshnessChecker) ValidateMetadataFile() error {
	_, err := fc.loadLastImportMetadata()
	return err
}

// matchesPattern checks if a filename matches a pattern
func matchesPattern(filename, pattern string) bool {
	if pattern == "" {
		return false
	}
	
	// Simple wildcard matching - replace * with regex equivalent
	if strings.Contains(pattern, "*") {
		prefix := strings.Split(pattern, "*")[0]
		suffix := ""
		if parts := strings.Split(pattern, "*"); len(parts) > 1 {
			suffix = parts[len(parts)-1]
		}
		
		return strings.HasPrefix(filename, prefix) && strings.HasSuffix(filename, suffix)
	}
	
	return filename == pattern
}
