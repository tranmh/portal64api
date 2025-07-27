package repositories

import (
	"fmt"

	"portal64api/internal/database"
	"portal64api/internal/models"
)

// PlayerRepository handles player data operations
type PlayerRepository struct {
	dbs *database.Databases
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository(dbs *database.Databases) *PlayerRepository {
	return &PlayerRepository{dbs: dbs}
}

// GetPlayerByID gets a player by their ID (VKZ-PersonID format)
func (r *PlayerRepository) GetPlayerByID(vkz string, personID uint) (*models.Person, *models.Organisation, *models.Evaluation, error) {
	// Get person from MVDSB database
	var person models.Person
	if err := r.dbs.MVDSB.Where("id = ?", personID).First(&person).Error; err != nil {
		return nil, nil, nil, err
	}

	// Get current club membership
	var membership models.Mitgliedschaft
	var org models.Organisation
	err := r.dbs.MVDSB.Where("person = ? AND bis IS NULL AND status = 1", personID).
		Order("von DESC").First(&membership).Error
	if err == nil {
		r.dbs.MVDSB.Where("id = ?", membership.Organisation).First(&org)
	}

	// Get latest DWZ evaluation from Portal64_BDW
	var evaluation models.Evaluation
	r.dbs.Portal64BDW.Where("idPerson = ?", personID).
		Order("id DESC").First(&evaluation)

	return &person, &org, &evaluation, nil
}

// SearchPlayers searches for players by name
func (r *PlayerRepository) SearchPlayers(req models.SearchRequest) ([]models.Person, int64, error) {
	var players []models.Person
	var total int64

	query := r.dbs.MVDSB.Model(&models.Person{}).Where("status = 1")

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("name LIKE ? OR vorname LIKE ? OR CONCAT(vorname, ' ', name) LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting
	orderBy := "name ASC"
	if req.SortBy != "" {
		direction := "ASC"
		if req.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, direction)
	}

	// Apply pagination and execute
	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&players).Error
	
	return players, total, err
}

// GetPlayersByClub gets all players in a club
func (r *PlayerRepository) GetPlayersByClub(vkz string, req models.SearchRequest) ([]models.Person, int64, error) {
	// First get the organization ID by VKZ
	var org models.Organisation
	if err := r.dbs.MVDSB.Where("vkz = ? AND status = 1", vkz).First(&org).Error; err != nil {
		return nil, 0, err
	}

	// Get current memberships for this organization
	var memberships []models.Mitgliedschaft
	memberQuery := r.dbs.MVDSB.Where("organisation = ? AND bis IS NULL AND status = 1", org.ID)
	if err := memberQuery.Find(&memberships).Error; err != nil {
		return nil, 0, err
	}

	// Extract person IDs
	personIDs := make([]uint, len(memberships))
	for i, m := range memberships {
		personIDs[i] = m.Person
	}

	if len(personIDs) == 0 {
		return []models.Person{}, 0, nil
	}

	var players []models.Person
	var total int64

	query := r.dbs.MVDSB.Model(&models.Person{}).Where("id IN ? AND status = 1", personIDs)

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("name LIKE ? OR vorname LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting and pagination
	orderBy := "name ASC"
	if req.SortBy != "" {
		direction := "ASC"
		if req.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, direction)
	}

	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&players).Error
	return players, total, err
}

// GetPlayerRatingHistory gets rating history for a player
func (r *PlayerRepository) GetPlayerRatingHistory(personID uint) ([]models.Evaluation, error) {
	var evaluations []models.Evaluation
	err := r.dbs.Portal64BDW.Where("idPerson = ?", personID).
		Order("id DESC").Find(&evaluations).Error
	return evaluations, err
}
