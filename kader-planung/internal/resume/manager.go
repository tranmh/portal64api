package resume

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
)

// Manager handles checkpoint creation, loading, and management
type Manager struct {
	checkpointFile string
	mutex          sync.RWMutex
	logger         *logrus.Logger
}

// NewManager creates a new resume manager
func NewManager(checkpointFile string) *Manager {
	return &Manager{
		checkpointFile: checkpointFile,
		logger:         logrus.StandardLogger(),
	}
}
// Config represents the configuration used for the run
type Config struct {
	ClubPrefix   string `json:"club_prefix"`
	OutputFormat string `json:"output_format"`
	Concurrency  int    `json:"concurrency"`
}

// Progress tracks the overall progress of the operation
type Progress struct {
	TotalClubs     int    `json:"total_clubs"`
	ProcessedClubs int    `json:"processed_clubs"`
	CurrentPhase   string `json:"current_phase"`
}

// ProcessedItem represents a completed processing item
type ProcessedItem struct {
	Type   string `json:"type"`   // "club" or "player"
	ID     string `json:"id"`     // club or player ID
	Status string `json:"status"` // "completed", "failed", "partial"
}

// PartialPlayerData stores partially processed player data
type PartialPlayerData struct {
	ClubID   string                     `json:"club_id"`
	ClubName string                     `json:"club_name"`
	Player   models.Player              `json:"player"`
	History  *models.RatingHistory      `json:"history,omitempty"`
	Analysis *models.HistoricalAnalysis `json:"analysis,omitempty"`
}
// Checkpoint represents the complete state of a processing run
type Checkpoint struct {
	Timestamp      time.Time           `json:"timestamp"`
	Config         Config              `json:"config"`
	Progress       Progress            `json:"progress"`
	ProcessedItems []ProcessedItem     `json:"processed_items"`
	PartialData    []PartialPlayerData `json:"partial_data"`
}

// SaveCheckpoint saves the current state to disk
func (m *Manager) SaveCheckpoint(checkpoint *Checkpoint) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	checkpoint.Timestamp = time.Now()

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Write to temporary file first, then rename for atomic operation
	tempFile := m.checkpointFile + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary checkpoint file: %w", err)
	}

	if err := os.Rename(tempFile, m.checkpointFile); err != nil {
		os.Remove(tempFile) // Clean up temp file if rename fails
		return fmt.Errorf("failed to rename checkpoint file: %w", err)
	}
	m.logger.Debugf("Checkpoint saved: %d/%d clubs processed", 
		checkpoint.Progress.ProcessedClubs, checkpoint.Progress.TotalClubs)

	return nil
}

// LoadCheckpoint loads the checkpoint from disk
func (m *Manager) LoadCheckpoint() (*Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, err := os.Stat(m.checkpointFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("checkpoint file does not exist: %s", m.checkpointFile)
	}

	data, err := ioutil.ReadFile(m.checkpointFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	m.logger.Infof("Checkpoint loaded from %s", m.checkpointFile)
	return &checkpoint, nil
}

// IsProcessed checks if an item has already been processed
func (m *Manager) IsProcessed(checkpoint *Checkpoint, itemType, itemID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, item := range checkpoint.ProcessedItems {
		if item.Type == itemType && item.ID == itemID && item.Status == "completed" {
			return true
		}
	}
	return false
}

// MarkProcessed marks an item as processed
func (m *Manager) MarkProcessed(checkpoint *Checkpoint, itemType, itemID, status string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Remove any existing entry for this item
	for i, item := range checkpoint.ProcessedItems {
		if item.Type == itemType && item.ID == itemID {
			checkpoint.ProcessedItems = append(checkpoint.ProcessedItems[:i], checkpoint.ProcessedItems[i+1:]...)
			break
		}
	}

	// Add the new entry
	checkpoint.ProcessedItems = append(checkpoint.ProcessedItems, ProcessedItem{
		Type:   itemType,
		ID:     itemID,
		Status: status,
	})
}

// AddPartialData adds partial player data to the checkpoint
func (m *Manager) AddPartialData(checkpoint *Checkpoint, data PartialPlayerData) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Remove any existing partial data for this player
	for i, existing := range checkpoint.PartialData {
		if existing.Player.ID == data.Player.ID {
			checkpoint.PartialData = append(checkpoint.PartialData[:i], checkpoint.PartialData[i+1:]...)
			break
		}
	}

	// Add the new partial data
	checkpoint.PartialData = append(checkpoint.PartialData, data)
}
// GetPartialData retrieves partial data for a player
func (m *Manager) GetPartialData(checkpoint *Checkpoint, playerID string) *PartialPlayerData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, data := range checkpoint.PartialData {
		if data.Player.ID == playerID {
			return &data
		}
	}
	return nil
}

// RemovePartialData removes partial data for a player (when processing is complete)
func (m *Manager) RemovePartialData(checkpoint *Checkpoint, playerID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for i, data := range checkpoint.PartialData {
		if data.Player.ID == playerID {
			checkpoint.PartialData = append(checkpoint.PartialData[:i], checkpoint.PartialData[i+1:]...)
			break
		}
	}
}

// UpdateProgress updates the progress information
func (m *Manager) UpdateProgress(checkpoint *Checkpoint, totalClubs, processedClubs int, phase string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	checkpoint.Progress.TotalClubs = totalClubs
	checkpoint.Progress.ProcessedClubs = processedClubs
	checkpoint.Progress.CurrentPhase = phase
}

// GetProcessedClubs returns a list of club IDs that have been completely processed
func (m *Manager) GetProcessedClubs(checkpoint *Checkpoint) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var clubs []string
	for _, item := range checkpoint.ProcessedItems {
		if item.Type == "club" && item.Status == "completed" {
			clubs = append(clubs, item.ID)
		}
	}
	return clubs
}
// GetProcessedPlayers returns a list of player IDs that have been completely processed
func (m *Manager) GetProcessedPlayers(checkpoint *Checkpoint) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var players []string
	for _, item := range checkpoint.ProcessedItems {
		if item.Type == "player" && item.Status == "completed" {
			players = append(players, item.ID)
		}
	}
	return players
}

// ValidateConfigCompatibility checks if the checkpoint config matches current run config
func (m *Manager) ValidateConfigCompatibility(checkpoint *Checkpoint, currentConfig Config) error {
	if checkpoint.Config.ClubPrefix != currentConfig.ClubPrefix {
		return fmt.Errorf("club prefix mismatch: checkpoint has '%s', current run has '%s'",
			checkpoint.Config.ClubPrefix, currentConfig.ClubPrefix)
	}

	if checkpoint.Config.OutputFormat != currentConfig.OutputFormat {
		return fmt.Errorf("output format mismatch: checkpoint has '%s', current run has '%s'",
			checkpoint.Config.OutputFormat, currentConfig.OutputFormat)
	}

	// Concurrency can be different, just warn
	if checkpoint.Config.Concurrency != currentConfig.Concurrency {
		m.logger.Warnf("Concurrency changed from %d to %d", 
			checkpoint.Config.Concurrency, currentConfig.Concurrency)
	}

	return nil
}
// Cleanup removes the checkpoint file after successful completion
func (m *Manager) Cleanup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, err := os.Stat(m.checkpointFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clean up
	}

	if err := os.Remove(m.checkpointFile); err != nil {
		return fmt.Errorf("failed to remove checkpoint file: %w", err)
	}

	m.logger.Infof("Checkpoint file cleaned up: %s", m.checkpointFile)
	return nil
}

// SetLogger sets a custom logger for the manager
func (m *Manager) SetLogger(logger *logrus.Logger) {
	m.logger = logger
}