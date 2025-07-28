package repositories

import (
	"fmt"
	"time"

	"portal64api/internal/database"
	"portal64api/internal/models"
)

// TournamentRepository handles tournament data operations
type TournamentRepository struct {
	dbs *database.Databases
}

// NewTournamentRepository creates a new tournament repository
func NewTournamentRepository(dbs *database.Databases) *TournamentRepository {
	return &TournamentRepository{dbs: dbs}
}

// GetTournamentByCode gets a tournament by its code
func (r *TournamentRepository) GetTournamentByCode(code string) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.dbs.Portal64BDW.Where("tcode = ?", code).First(&tournament).Error
	return &tournament, err
}

// SearchTournaments searches for tournaments
func (r *TournamentRepository) SearchTournaments(req models.SearchRequest) ([]models.Tournament, int64, error) {
	var tournaments []models.Tournament
	var total int64

	query := r.dbs.Portal64BDW.Model(&models.Tournament{})

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("tname LIKE ? OR tcode LIKE ?", searchPattern, searchPattern)
	}

	// Apply date filters if provided
	if req.FilterBy == "year" && req.FilterValue != "" {
		query = query.Where("YEAR(finishedOn) = ?", req.FilterValue)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting
	orderBy := "finishedOn DESC"
	if req.SortBy != "" {
		direction := "DESC"
		if req.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, direction)
	}

	// Apply pagination and execute
	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&tournaments).Error
	
	return tournaments, total, err
}

// GetUpcomingTournaments gets upcoming tournaments from SVW database
func (r *TournamentRepository) GetUpcomingTournaments(limit int) ([]models.Turnier, error) {
	var tournaments []models.Turnier
	now := time.Now()
	
	err := r.dbs.Portal64SVW.Where("isFreigegeben = 1 AND (Teilnahmeschluss > ? OR Meldeschluss > ?)", 
		now, now).Order("Teilnahmeschluss ASC").Limit(limit).Find(&tournaments).Error
	
	return tournaments, err
}

// GetTournamentsByDateRange gets tournaments within a date range
func (r *TournamentRepository) GetTournamentsByDateRange(startDate, endDate time.Time, req models.SearchRequest) ([]models.Tournament, int64, error) {
	var tournaments []models.Tournament
	var total int64

	query := r.dbs.Portal64BDW.Model(&models.Tournament{}).
		Where("finishedOn BETWEEN ? AND ?", startDate, endDate)

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("tname LIKE ? OR tcode LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting and pagination
	orderBy := "finishedOn DESC"
	if req.SortBy != "" {
		direction := "DESC"
		if req.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, direction)
	}

	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&tournaments).Error
	
	return tournaments, total, err
}

// GetTournamentResults gets results for a specific tournament
func (r *TournamentRepository) GetTournamentResults(tournamentID uint) ([]models.Evaluation, error) {
	var evaluations []models.Evaluation
	err := r.dbs.Portal64BDW.Where("idMaster = ?", tournamentID).
		Order("dwzNew DESC").Find(&evaluations).Error
	return evaluations, err
}

// GetRecentTournaments gets recently finished tournaments
func (r *TournamentRepository) GetRecentTournaments(days int, limit int) ([]models.Tournament, error) {
	var tournaments []models.Tournament
	cutoff := time.Now().AddDate(0, 0, -days)
	
	err := r.dbs.Portal64BDW.Where("finishedOn >= ?", cutoff).
		Order("finishedOn DESC").Limit(limit).Find(&tournaments).Error
	
	return tournaments, err
}

// GetParticipantCount gets the number of participants for a tournament
func (r *TournamentRepository) GetParticipantCount(tournamentID uint) (int, error) {
	var count int64
	err := r.dbs.Portal64BDW.Model(&models.Participant{}).
		Where("idTournament = ?", tournamentID).Count(&count).Error
	return int(count), err
}
