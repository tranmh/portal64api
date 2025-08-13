package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"portal64api/internal/cache"
	"portal64api/internal/config"
	"portal64api/internal/importers"
	"portal64api/internal/models"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ImportCompleteCallback defines the interface for import completion callbacks
type ImportCompleteCallback interface {
	OnImportComplete()
}

// ImportService handles scheduled and manual database imports
type ImportService struct {
	config            *config.ImportConfig
	dbConfig          *config.DatabaseConfig
	cacheService      cache.CacheService
	logger            *log.Logger
	
	// Components
	downloader        *importers.SCPDownloader
	extractor         *importers.ZIPExtractor
	importer          *importers.DatabaseImporter
	freshnessChecker  *importers.FreshnessChecker
	statusTracker     *importers.StatusTracker
	
	// Scheduling
	cron              *cron.Cron
	cronJobID         cron.EntryID
	
	// Control
	isRunning         bool
	stopChan          chan struct{}
	mutex             sync.RWMutex
	
	// Callbacks
	onCompleteCallbacks []ImportCompleteCallback
}

// NewImportService creates a new import service instance
func NewImportService(importConfig *config.ImportConfig, dbConfig *config.DatabaseConfig, cacheService cache.CacheService, logger *log.Logger) *ImportService {
	service := &ImportService{
		config:       importConfig,
		dbConfig:     dbConfig,
		cacheService: cacheService,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}

	// Initialize components
	service.downloader = importers.NewSCPDownloader(&importConfig.SCP, logger)
	service.extractor = importers.NewZIPExtractor(&importConfig.ZIP, logger)
	service.importer = importers.NewDatabaseImporter(&importConfig.Database, dbConfig, logger)
	service.freshnessChecker = importers.NewFreshnessChecker(&importConfig.Freshness, importConfig.Storage.MetadataFile, logger)
	service.statusTracker = importers.NewStatusTracker(1000, logger) // Keep 1000 log entries

	// Initialize cron scheduler
	service.cron = cron.New()

	return service
}

// Start starts the import service with scheduled execution
func (is *ImportService) Start() error {
	if !is.config.Enabled {
		is.logger.Println("Import service is disabled")
		return nil
	}

	is.logger.Println("Starting import service...")

	// Schedule the import job
	job := cron.FuncJob(func() {
		if err := is.executeScheduledImport(); err != nil {
			is.logger.Printf("Scheduled import failed: %v", err)
		}
	})

	entryID, err := is.cron.AddJob(is.config.Schedule, job)
	if err != nil {
		return fmt.Errorf("failed to schedule import job: %w", err)
	}

	is.cronJobID = entryID
	is.cron.Start()

	// Set next scheduled time
	entries := is.cron.Entries()
	for _, entry := range entries {
		if entry.ID == is.cronJobID {
			is.statusTracker.SetNextScheduled(entry.Next)
			break
		}
	}

	is.logger.Printf("Import service started with schedule: %s", is.config.Schedule)
	return nil
}

// Stop stops the import service
func (is *ImportService) Stop() error {
	is.logger.Println("Stopping import service...")

	// Stop cron scheduler
	if is.cron != nil {
		is.cron.Stop()
	}

	// Signal stop
	close(is.stopChan)

	is.logger.Println("Import service stopped")
	return nil
}

// TriggerManualImport triggers a manual import immediately
func (is *ImportService) TriggerManualImport() error {
	if !is.config.Enabled {
		return fmt.Errorf("import service is disabled")
	}

	is.mutex.RLock()
	running := is.isRunning
	is.mutex.RUnlock()

	if running {
		return fmt.Errorf("import is already running")
	}

	is.logger.Println("Manual import triggered")

	// Execute import in background
	go func() {
		if err := is.executeImport(); err != nil {
			is.logger.Printf("Manual import failed: %v", err)
		}
	}()

	return nil
}

// GetStatus returns the current import status
func (is *ImportService) GetStatus() *models.ImportStatus {
	return is.statusTracker.GetStatus()
}

// GetLogs returns recent import log entries
func (is *ImportService) GetLogs(limit int) []models.ImportLogEntry {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	return is.statusTracker.GetLogs(limit)
}

// TestConnection tests the SCP connection
func (is *ImportService) TestConnection() error {
	return is.downloader.TestConnection()
}

// executeScheduledImport handles scheduled import execution with load checking
func (is *ImportService) executeScheduledImport() error {
	is.logger.Println("Scheduled import starting...")

	// Check if already running
	is.mutex.RLock()
	running := is.isRunning
	is.mutex.RUnlock()

	if running {
		is.logger.Println("Import already running, skipping scheduled execution")
		return nil
	}

	// Check API load if enabled
	if is.config.LoadCheck.Enabled {
		if err := is.checkLoadAndDelay(); err != nil {
			is.statusTracker.MarkSkipped("api_under_heavy_load", models.StepInitialization)
			return fmt.Errorf("skipping import due to heavy load: %w", err)
		}
	}

	return is.executeImport()
}

// executeImport performs the complete import process
func (is *ImportService) executeImport() error {
	is.mutex.Lock()
	is.isRunning = true
	is.mutex.Unlock()

	defer func() {
		is.mutex.Lock()
		is.isRunning = false
		is.mutex.Unlock()
	}()

	is.statusTracker.UpdateStatus(models.StatusRunning, models.StepInitialization, 0)
	is.logger.Println("Starting import process...")

	start := time.Now()

	// Execute import phases
	if err := is.executeImportPhases(); err != nil {
		is.statusTracker.MarkFailed(err, is.statusTracker.GetCurrentStep())
		return err
	}

	// Mark as successful
	is.statusTracker.MarkSuccess()
	duration := time.Since(start)
	is.statusTracker.LogDuration(models.StepCompleted, "Import completed successfully", duration)

	is.logger.Printf("Import process completed successfully in %s", duration)
	
	// Notify completion callbacks
	is.notifyCompletionCallbacks()
	
	return nil
}

// executeImportPhases executes all import phases
func (is *ImportService) executeImportPhases() error {
	// Phase 1: File freshness check
	if err := is.executeFreshnessCheck(); err != nil {
		return fmt.Errorf("freshness check failed: %w", err)
	}

	// Phase 2: Download files
	downloadedFiles, err := is.executeDownloadPhase()
	if err != nil {
		return fmt.Errorf("download phase failed: %w", err)
	}

	// Phase 3: Extract ZIP files
	extractedFiles, err := is.executeExtractionPhase(downloadedFiles)
	if err != nil {
		return fmt.Errorf("extraction phase failed: %w", err)
	}

	// Phase 4: Import databases
	if err := is.executeImportPhase(extractedFiles); err != nil {
		return fmt.Errorf("database import phase failed: %w", err)
	}

	// Phase 5: Clear cache
	if err := is.executeCacheCleanup(); err != nil {
		return fmt.Errorf("cache cleanup failed: %w", err)
	}

	// Phase 6: Cleanup and save metadata
	if err := is.executeCleanupPhase(downloadedFiles); err != nil {
		return fmt.Errorf("cleanup phase failed: %w", err)
	}

	return nil
}

// executeFreshnessCheck performs file freshness checking
func (is *ImportService) executeFreshnessCheck() error {
	if !is.config.Freshness.Enabled {
		is.statusTracker.UpdateProgress(models.StepCheckingFreshness, 15)
		is.logger.Println("File freshness check disabled, proceeding with import")
		return nil
	}

	is.statusTracker.UpdateProgress(models.StepCheckingFreshness, 10)
	is.logger.Println("Checking file freshness...")

	// List remote files
	remoteFiles, err := is.downloader.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}

	// Check freshness
	freshnessResult, err := is.freshnessChecker.CheckFreshness(remoteFiles)
	if err != nil {
		is.logger.Printf("Freshness check failed, proceeding with import: %v", err)
		return nil // Continue with import even if freshness check fails
	}

	// Update files info
	filesInfo := &models.ImportFilesInfo{
		RemoteFiles: remoteFiles,
	}
	
	if lastImport, err := is.freshnessChecker.GetLastImportInfo(); err == nil {
		filesInfo.LastImported = lastImport.Files
	}

	is.statusTracker.SetFilesInfo(filesInfo)

	// Check if we should skip
	if !freshnessResult.ShouldImport && is.config.Freshness.SkipIfNotNewer {
		is.statusTracker.MarkSkipped(freshnessResult.Reason, models.StepCheckingFreshness)
		is.logger.Printf("Import skipped: %s", freshnessResult.Reason)
		return fmt.Errorf("import skipped: %s", freshnessResult.Reason)
	}

	is.statusTracker.UpdateProgress(models.StepCheckingFreshness, 15)
	is.logger.Printf("Freshness check completed: %s", freshnessResult.Reason)
	return nil
}

// executeDownloadPhase downloads files from remote server
func (is *ImportService) executeDownloadPhase() ([]models.FileMetadata, error) {
	is.statusTracker.UpdateProgress(models.StepDownload, 20)
	is.logger.Println("Starting download phase...")

	// List remote files
	remoteFiles, err := is.downloader.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote files: %w", err)
	}

	// Create temporary directory
	tempDir := is.config.Storage.TempDir
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download files
	if err := is.downloader.DownloadFiles(remoteFiles, tempDir); err != nil {
		return nil, fmt.Errorf("failed to download files: %w", err)
	}

	// Update files info
	if filesInfo := is.statusTracker.GetStatus().FilesInfo; filesInfo != nil {
		for _, file := range remoteFiles {
			filesInfo.Downloaded = append(filesInfo.Downloaded, file.Filename)
		}
		is.statusTracker.SetFilesInfo(filesInfo)
	}

	is.statusTracker.UpdateProgress(models.StepDownload, 40)
	is.logger.Printf("Download phase completed: %d files", len(remoteFiles))
	return remoteFiles, nil
}

// executeExtractionPhase extracts ZIP files
func (is *ImportService) executeExtractionPhase(downloadedFiles []models.FileMetadata) (map[string]string, error) {
	is.statusTracker.UpdateProgress(models.StepExtraction, 50)
	is.logger.Println("Starting extraction phase...")

	tempDir := is.config.Storage.TempDir
	extractDir := filepath.Join(tempDir, "extracted")

	// Extract each ZIP file
	var zipPaths []string
	for _, file := range downloadedFiles {
		zipPath := filepath.Join(tempDir, file.Filename)
		zipPaths = append(zipPaths, zipPath)
	}

	results, err := is.extractor.ExtractFiles(zipPaths, extractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract files: %w", err)
	}

	// Find database dumps
	databaseDumps, err := is.extractor.FindDatabaseDumps(extractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find database dumps: %w", err)
	}

	// Update files info
	if filesInfo := is.statusTracker.GetStatus().FilesInfo; filesInfo != nil {
		for _, extractedFiles := range results {
			for _, file := range extractedFiles {
				filesInfo.Extracted = append(filesInfo.Extracted, filepath.Base(file))
			}
		}
		is.statusTracker.SetFilesInfo(filesInfo)
	}

	is.statusTracker.UpdateProgress(models.StepExtraction, 60)
	is.logger.Printf("Extraction phase completed: %d database dumps found", len(databaseDumps))
	return databaseDumps, nil
}

// executeImportPhase imports database dumps
func (is *ImportService) executeImportPhase(databaseDumps map[string]string) error {
	is.statusTracker.UpdateProgress(models.StepDatabaseImport, 70)
	is.logger.Println("Starting database import phase...")

	if err := is.importer.ImportDatabases(databaseDumps); err != nil {
		return fmt.Errorf("failed to import databases: %w", err)
	}

	// Update files info
	if filesInfo := is.statusTracker.GetStatus().FilesInfo; filesInfo != nil {
		for database := range databaseDumps {
			filesInfo.Imported = append(filesInfo.Imported, database)
		}
		is.statusTracker.SetFilesInfo(filesInfo)
	}

	is.statusTracker.UpdateProgress(models.StepDatabaseImport, 85)
	is.logger.Printf("Database import phase completed: %d databases imported", len(databaseDumps))
	return nil
}

// executeCacheCleanup clears Redis cache
func (is *ImportService) executeCacheCleanup() error {
	is.statusTracker.UpdateProgress(models.StepCacheCleanup, 90)
	is.logger.Println("Starting cache cleanup...")

	if is.cacheService != nil {
		ctx := context.Background()
		if err := is.cacheService.FlushAll(ctx); err != nil {
			is.logger.Printf("Warning: Cache cleanup failed: %v", err)
			// Don't fail the entire import for cache issues
		} else {
			is.logger.Println("Cache cleared successfully")
		}
	} else {
		is.logger.Println("Cache service not available, skipping cleanup")
	}

	is.statusTracker.UpdateProgress(models.StepCacheCleanup, 95)
	return nil
}

// executeCleanupPhase performs cleanup and saves metadata
func (is *ImportService) executeCleanupPhase(downloadedFiles []models.FileMetadata) error {
	is.statusTracker.UpdateProgress(models.StepCleanup, 98)
	is.logger.Println("Starting cleanup phase...")

	// Save import metadata
	if err := is.freshnessChecker.SaveImportMetadata(downloadedFiles); err != nil {
		is.logger.Printf("Warning: Failed to save import metadata: %v", err)
		// Don't fail the entire import for metadata issues
	}

	// Cleanup temporary files if configured
	if is.config.Storage.CleanupOnSuccess {
		tempDir := is.config.Storage.TempDir
		if err := os.RemoveAll(tempDir); err != nil {
			is.logger.Printf("Warning: Failed to cleanup temp directory: %v", err)
		} else {
			is.logger.Printf("Temporary files cleaned up: %s", tempDir)
		}
	}

	is.statusTracker.UpdateProgress(models.StepCleanup, 100)
	is.logger.Println("Cleanup phase completed")
	return nil
}

// checkLoadAndDelay checks API load and delays if necessary
func (is *ImportService) checkLoadAndDelay() error {
	// For now, implement a simple check
	// This could be enhanced to check actual API metrics
	
	for attempt := 0; attempt < is.config.LoadCheck.MaxDelays; attempt++ {
		// Simple load check - in production this could check:
		// - Active database connections
		// - CPU usage
		// - Memory usage
		// - Request rate
		
		if !is.isAPIUnderHeavyLoad() {
			return nil // Proceed with import
		}

		is.logger.Printf("API under heavy load, delaying import for %v (attempt %d/%d)",
			is.config.LoadCheck.DelayDuration, attempt+1, is.config.LoadCheck.MaxDelays)

		select {
		case <-time.After(is.config.LoadCheck.DelayDuration):
			continue
		case <-is.stopChan:
			return fmt.Errorf("import service stopped during delay")
		}
	}

	return fmt.Errorf("max delays exceeded, API still under heavy load")
}

// isAPIUnderHeavyLoad checks if the API is under heavy load
func (is *ImportService) isAPIUnderHeavyLoad() bool {
	// Placeholder implementation
	// In production, this could check actual metrics
	return false
}

// AddCompletionCallback adds a callback to be called when import completes successfully
func (is *ImportService) AddCompletionCallback(callback ImportCompleteCallback) {
	is.mutex.Lock()
	defer is.mutex.Unlock()
	is.onCompleteCallbacks = append(is.onCompleteCallbacks, callback)
}

// notifyCompletionCallbacks calls all registered completion callbacks
func (is *ImportService) notifyCompletionCallbacks() {
	is.mutex.RLock()
	callbacks := make([]ImportCompleteCallback, len(is.onCompleteCallbacks))
	copy(callbacks, is.onCompleteCallbacks)
	is.mutex.RUnlock()
	
	for _, callback := range callbacks {
		// Call callback in a separate goroutine to avoid blocking
		go func(cb ImportCompleteCallback) {
			defer func() {
				if r := recover(); r != nil {
					is.logger.Printf("Import completion callback panicked: %v", r)
				}
			}()
			cb.OnImportComplete()
		}(callback)
	}
}
