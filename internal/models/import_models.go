package models

import (
	"encoding/json"
	"time"
)

// ImportStatus represents the current status of an import operation
type ImportStatus struct {
	Status        string              `json:"status"`          // idle, running, success, failed, skipped
	Progress      int                 `json:"progress"`        // 0-100
	CurrentStep   string              `json:"current_step"`    // current operation step
	StartedAt     *time.Time          `json:"started_at"`
	CompletedAt   *time.Time          `json:"completed_at"`
	LastSuccess   *time.Time          `json:"last_success"`
	NextScheduled *time.Time          `json:"next_scheduled"`
	RetryCount    int                 `json:"retry_count"`
	MaxRetries    int                 `json:"max_retries"`
	Error         string              `json:"error,omitempty"`
	SkipReason    string              `json:"skip_reason,omitempty"`
	FilesInfo     *ImportFilesInfo    `json:"files_info,omitempty"`
}

// ImportFilesInfo contains information about files involved in the import
type ImportFilesInfo struct {
	RemoteFiles   []FileMetadata `json:"remote_files"`
	LastImported  []FileMetadata `json:"last_imported"`
	Downloaded    []string       `json:"downloaded"`
	Extracted     []string       `json:"extracted"`
	Imported      []string       `json:"imported"`
}

// FileMetadata contains metadata about a file
type FileMetadata struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	Checksum string    `json:"checksum,omitempty"`
	Pattern  string    `json:"pattern,omitempty"`
	Database string    `json:"database,omitempty"`
	Imported bool      `json:"imported,omitempty"`
	IsNewer  bool      `json:"is_newer,omitempty"`
}

// ImportLogEntry represents a single log entry
type ImportLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`      // INFO, WARN, ERROR
	Message   string    `json:"message"`
	Step      string    `json:"step"`       // download, extract, import, cleanup
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	FileSize  int64     `json:"file_size,omitempty"`
}

// LastImportMetadata stores information about the last successful import
type LastImportMetadata struct {
	LastImport ImportRecord `json:"last_import"`
}

// ImportRecord contains details of an import operation
type ImportRecord struct {
	Timestamp time.Time      `json:"timestamp"`
	Success   bool           `json:"success"`
	Files     []FileMetadata `json:"files"`
}

// FreshnessResult contains the result of a file freshness check
type FreshnessResult struct {
	ShouldImport  bool             `json:"should_import"`
	Reason        string           `json:"reason"`
	RemoteFiles   []FileMetadata   `json:"remote_files"`
	LastImported  []FileMetadata   `json:"last_imported"`
	Comparisons   []FileComparison `json:"comparisons"`
}

// FileComparison contains the comparison result between remote and last imported file
type FileComparison struct {
	RemoteFile *FileMetadata `json:"remote_file"`
	LastFile   *FileMetadata `json:"last_file"`
	IsNewer    bool          `json:"is_newer"`
	Reasons    []string      `json:"reasons"`
}

// ImportStartResponse is the response for manual import start
type ImportStartResponse struct {
	Message   string    `json:"message"`
	StartedAt time.Time `json:"started_at"`
}

// ImportLogsResponse is the response for import logs
type ImportLogsResponse struct {
	Logs []ImportLogEntry `json:"logs"`
}

// LoadMetrics contains current system load information
type LoadMetrics struct {
	ConcurrentRequests  int     `json:"concurrent_requests"`
	ActiveDBConnections int     `json:"active_db_connections"`
	MemoryUsagePercent  float64 `json:"memory_usage_percent"`
	CPUUsagePercent     float64 `json:"cpu_usage_percent"`
}

// ImportSteps contains constants for import step names
const (
	StepInitialization       = "initialization"
	StepCheckingFreshness    = "checking_file_freshness"
	StepDownload             = "download"
	StepExtraction           = "extraction"
	StepDatabaseImport       = "importing_database"
	StepCacheCleanup         = "cache_cleanup"
	StepCleanup              = "cleanup"
	StepCompleted            = "completed"
)

// ImportStatus constants
const (
	StatusIdle    = "idle"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Error type constants
const (
	ErrorCritical    = "critical"
	ErrorRecoverable = "recoverable"
	ErrorWarning     = "warning"
)

// Helper method to convert ImportStatus to JSON string
func (is *ImportStatus) ToJSON() string {
	data, _ := json.Marshal(is)
	return string(data)
}

// Helper method to create a new ImportStatus
func NewImportStatus() *ImportStatus {
	return &ImportStatus{
		Status:     StatusIdle,
		Progress:   0,
		RetryCount: 0,
		MaxRetries: 2,
	}
}

// Helper method to update status with step and progress
func (is *ImportStatus) UpdateProgress(step string, progress int) {
	is.CurrentStep = step
	is.Progress = progress
	if progress > 0 && is.Status == StatusIdle {
		is.Status = StatusRunning
	}
}

// Helper method to mark import as completed successfully
func (is *ImportStatus) MarkSuccess() {
	now := time.Now()
	is.Status = StatusSuccess
	is.Progress = 100
	is.CurrentStep = StepCompleted
	is.CompletedAt = &now
	is.LastSuccess = &now
	is.Error = ""
	is.SkipReason = ""
}

// Helper method to mark import as failed
func (is *ImportStatus) MarkFailed(err error) {
	now := time.Now()
	is.Status = StatusFailed
	is.CompletedAt = &now
	if err != nil {
		is.Error = err.Error()
	}
}

// Helper method to mark import as skipped
func (is *ImportStatus) MarkSkipped(reason string) {
	now := time.Now()
	is.Status = StatusSkipped
	is.Progress = 100
	is.CurrentStep = StepCompleted
	is.CompletedAt = &now
	is.SkipReason = reason
	is.Error = ""
}
