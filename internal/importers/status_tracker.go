package importers

import (
	"fmt"
	"log"
	"portal64api/internal/models"
	"sync"
	"time"
)

// StatusTracker manages import status and logging in memory
type StatusTracker struct {
	status    *models.ImportStatus
	logs      []models.ImportLogEntry
	mutex     sync.RWMutex
	maxLogs   int
	logger    *log.Logger
}

// NewStatusTracker creates a new status tracker instance
func NewStatusTracker(maxLogs int, logger *log.Logger) *StatusTracker {
	return &StatusTracker{
		status:  models.NewImportStatus(),
		logs:    make([]models.ImportLogEntry, 0),
		maxLogs: maxLogs,
		logger:  logger,
	}
}

// GetStatus returns a copy of the current import status
func (st *StatusTracker) GetStatus() *models.ImportStatus {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	// Create a deep copy to avoid race conditions
	statusCopy := *st.status
	if st.status.StartedAt != nil {
		startedAt := *st.status.StartedAt
		statusCopy.StartedAt = &startedAt
	}
	if st.status.CompletedAt != nil {
		completedAt := *st.status.CompletedAt
		statusCopy.CompletedAt = &completedAt
	}
	if st.status.LastSuccess != nil {
		lastSuccess := *st.status.LastSuccess
		statusCopy.LastSuccess = &lastSuccess
	}
	if st.status.NextScheduled != nil {
		nextScheduled := *st.status.NextScheduled
		statusCopy.NextScheduled = &nextScheduled
	}
	if st.status.FilesInfo != nil {
		filesInfoCopy := *st.status.FilesInfo
		statusCopy.FilesInfo = &filesInfoCopy
	}

	return &statusCopy
}

// UpdateStatus updates the current status
func (st *StatusTracker) UpdateStatus(status, step string, progress int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.Status = status
	st.status.CurrentStep = step
	st.status.Progress = progress

	if status == models.StatusRunning && st.status.StartedAt == nil {
		now := time.Now()
		st.status.StartedAt = &now
		st.status.CompletedAt = nil
		st.status.Error = ""
		st.status.SkipReason = ""
	}

	st.logEventUnsafe("INFO", step, fmt.Sprintf("Status updated: %s (%d%%)", status, progress), "", 0)
}

// UpdateProgress updates only the progress and current step
func (st *StatusTracker) UpdateProgress(step string, progress int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.UpdateProgress(step, progress)
	
	// Log every 25% progress or important steps
	if progress%25 == 0 || progress == 100 {
		st.logEventUnsafe("INFO", step, fmt.Sprintf("Progress: %d%%", progress), "", 0)
	}
}

// MarkSuccess marks the import as successful
func (st *StatusTracker) MarkSuccess() {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.MarkSuccess()
	st.logEventUnsafe("INFO", models.StepCompleted, "Import completed successfully", "", 0)
}

// MarkFailed marks the import as failed
func (st *StatusTracker) MarkFailed(err error, step string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.MarkFailed(err)
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	st.logEventUnsafe("ERROR", step, "Import failed", errorMsg, 0)
}

// MarkSkipped marks the import as skipped
func (st *StatusTracker) MarkSkipped(reason, step string) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.MarkSkipped(reason)
	st.logEventUnsafe("INFO", step, fmt.Sprintf("Import skipped: %s", reason), "", 0)
}

// SetRetryInfo sets retry information
func (st *StatusTracker) SetRetryInfo(retryCount, maxRetries int) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.RetryCount = retryCount
	st.status.MaxRetries = maxRetries
}

// SetNextScheduled sets the next scheduled import time
func (st *StatusTracker) SetNextScheduled(nextTime time.Time) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.NextScheduled = &nextTime
}

// SetFilesInfo sets the files information
func (st *StatusTracker) SetFilesInfo(filesInfo *models.ImportFilesInfo) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status.FilesInfo = filesInfo
}

// LogInfo logs an informational message
func (st *StatusTracker) LogInfo(step, message string) {
	st.logEvent("INFO", step, message, "", 0)
}

// LogWarning logs a warning message
func (st *StatusTracker) LogWarning(step, message string) {
	st.logEvent("WARN", step, message, "", 0)
}

// LogError logs an error message
func (st *StatusTracker) LogError(step, message, errorMsg string) {
	st.logEvent("ERROR", step, message, errorMsg, 0)
}

// LogProgress logs a progress message with file size
func (st *StatusTracker) LogProgress(step, message string, fileSize int64) {
	st.logEvent("INFO", step, message, "", fileSize)
}

// LogDuration logs a message with duration
func (st *StatusTracker) LogDuration(step, message string, duration time.Duration) {
	entry := models.ImportLogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Step:      step,
		Message:   message,
		Duration:  duration.String(),
	}

	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.addLogEntryUnsafe(entry)
}

// GetLogs returns a copy of recent log entries
func (st *StatusTracker) GetLogs(limit int) []models.ImportLogEntry {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	if limit <= 0 || limit > len(st.logs) {
		limit = len(st.logs)
	}

	// Return most recent logs
	result := make([]models.ImportLogEntry, limit)
	copy(result, st.logs[len(st.logs)-limit:])
	
	return result
}

// GetAllLogs returns all log entries
func (st *StatusTracker) GetAllLogs() []models.ImportLogEntry {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	result := make([]models.ImportLogEntry, len(st.logs))
	copy(result, st.logs)
	
	return result
}

// ClearLogs clears all log entries
func (st *StatusTracker) ClearLogs() {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.logs = st.logs[:0]
}

// Reset resets the status tracker to initial state
func (st *StatusTracker) Reset() {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.status = models.NewImportStatus()
	st.logs = st.logs[:0]
}

// IsRunning returns true if import is currently running
func (st *StatusTracker) IsRunning() bool {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	return st.status.Status == models.StatusRunning
}

// GetCurrentStep returns the current step
func (st *StatusTracker) GetCurrentStep() string {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	return st.status.CurrentStep
}

// GetProgress returns the current progress percentage
func (st *StatusTracker) GetProgress() int {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	return st.status.Progress
}

// logEvent logs an event with thread safety
func (st *StatusTracker) logEvent(level, step, message, errorMsg string, fileSize int64) {
	entry := models.ImportLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Step:      step,
		Message:   message,
		FileSize:  fileSize,
	}

	if errorMsg != "" {
		entry.Error = errorMsg
	}

	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.addLogEntryUnsafe(entry)
}

// logEventUnsafe logs an event without locking (must be called with mutex locked)
func (st *StatusTracker) logEventUnsafe(level, step, message, errorMsg string, fileSize int64) {
	entry := models.ImportLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Step:      step,
		Message:   message,
		FileSize:  fileSize,
	}

	if errorMsg != "" {
		entry.Error = errorMsg
	}

	st.addLogEntryUnsafe(entry)
}

// addLogEntryUnsafe adds a log entry without locking (must be called with mutex locked)
func (st *StatusTracker) addLogEntryUnsafe(entry models.ImportLogEntry) {
	// Also log to standard logger
	logMessage := fmt.Sprintf("[%s] %s: %s", entry.Level, entry.Step, entry.Message)
	if entry.Error != "" {
		logMessage += fmt.Sprintf(" (Error: %s)", entry.Error)
	}
	st.logger.Println(logMessage)

	// Add to in-memory logs
	st.logs = append(st.logs, entry)

	// Trim logs if exceeding max size
	if len(st.logs) > st.maxLogs {
		// Keep only the most recent logs
		excess := len(st.logs) - st.maxLogs
		st.logs = st.logs[excess:]
	}
}

// GetStatusSummary returns a brief status summary
func (st *StatusTracker) GetStatusSummary() string {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	switch st.status.Status {
	case models.StatusIdle:
		if st.status.LastSuccess != nil {
			return fmt.Sprintf("Idle (Last success: %s)", st.status.LastSuccess.Format("2006-01-02 15:04"))
		}
		return "Idle (Never run)"
	case models.StatusRunning:
		return fmt.Sprintf("Running: %s (%d%%)", st.status.CurrentStep, st.status.Progress)
	case models.StatusSuccess:
		if st.status.CompletedAt != nil {
			return fmt.Sprintf("Success (Completed: %s)", st.status.CompletedAt.Format("2006-01-02 15:04"))
		}
		return "Success"
	case models.StatusFailed:
		return fmt.Sprintf("Failed: %s", st.status.Error)
	case models.StatusSkipped:
		return fmt.Sprintf("Skipped: %s", st.status.SkipReason)
	default:
		return "Unknown status"
	}
}

// GetLogsSince returns logs since the specified time
func (st *StatusTracker) GetLogsSince(since time.Time) []models.ImportLogEntry {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	var result []models.ImportLogEntry
	for _, log := range st.logs {
		if log.Timestamp.After(since) {
			result = append(result, log)
		}
	}

	return result
}

// GetLogsByLevel returns logs of a specific level
func (st *StatusTracker) GetLogsByLevel(level string) []models.ImportLogEntry {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	var result []models.ImportLogEntry
	for _, log := range st.logs {
		if log.Level == level {
			result = append(result, log)
		}
	}

	return result
}

// GetErrorCount returns the number of error log entries
func (st *StatusTracker) GetErrorCount() int {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	count := 0
	for _, log := range st.logs {
		if log.Level == "ERROR" {
			count++
		}
	}

	return count
}

// GetWarningCount returns the number of warning log entries
func (st *StatusTracker) GetWarningCount() int {
	st.mutex.RLock()
	defer st.mutex.RUnlock()

	count := 0
	for _, log := range st.logs {
		if log.Level == "WARN" {
			count++
		}
	}

	return count
}
