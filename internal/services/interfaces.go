package services

import "portal64api/internal/models"

// ImportServiceInterface defines the contract for import services
type ImportServiceInterface interface {
	Start() error
	Stop() error
	TriggerManualImport() error
	GetStatus() *models.ImportStatus
	GetLogs(limit int) []models.ImportLogEntry
	TestConnection() error
}
