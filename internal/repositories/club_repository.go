package repositories

import (
	"fmt"

	"portal64api/internal/database"
	"portal64api/internal/models"
)

// ClubRepository handles club data operations
type ClubRepository struct {
	dbs *database.Databases
}

// NewClubRepository creates a new club repository
func NewClubRepository(dbs *database.Databases) *ClubRepository {
	return &ClubRepository{dbs: dbs}
}

// GetClubByVKZ gets a club by its VKZ (Club ID)
func (r *ClubRepository) GetClubByVKZ(vkz string) (*models.Organisation, error) {
	var org models.Organisation
	err := r.dbs.MVDSB.Where("vkz = ? AND status = 1", vkz).First(&org).Error
	return &org, err
}

// SearchClubs searches for clubs by name or VKZ
func (r *ClubRepository) SearchClubs(req models.SearchRequest) ([]models.Organisation, int64, error) {
	var clubs []models.Organisation
	var total int64

	query := r.dbs.MVDSB.Model(&models.Organisation{}).Where("status = 1 AND organisationsart = 20")

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("name LIKE ? OR kurzname LIKE ? OR vkz LIKE ?", 
			searchPattern, searchPattern, searchPattern)
	}

	// Apply additional filters
	if req.FilterBy != "" && req.FilterValue != "" {
		switch req.FilterBy {
		case "region":
			query = query.Where("verband = ?", req.FilterValue)
		case "district":
			query = query.Where("bezirk = ?", req.FilterValue)
		}
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
	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&clubs).Error
	
	return clubs, total, err
}

// GetClubMemberCount gets the number of active members in a club
func (r *ClubRepository) GetClubMemberCount(organizationID uint) (int64, error) {
	var count int64
	err := r.dbs.MVDSB.Model(&models.Mitgliedschaft{}).
		Where("organisation = ? AND bis IS NULL AND status = 1", organizationID).
		Count(&count).Error
	return count, err
}

// GetClubAverageDWZ calculates the average DWZ rating for a club's active members (optimized)
func (r *ClubRepository) GetClubAverageDWZ(organizationID uint) (float64, error) {
	type Result struct {
		AvgDWZ float64
	}

	var result Result
	
	// Optimized query using window function - eliminates correlated subquery
	// This should reduce query time from ~260ms to ~20-50ms for typical clubs
	// Note: mitgliedschaft table is in MVDSB database, evaluation table is in Portal64BDW database
	err := r.dbs.Portal64BDW.Raw(`
		SELECT AVG(latest_dwz) as avg_dwz
		FROM (
			SELECT DISTINCT 
				e.idPerson,
				FIRST_VALUE(e.dwzNew) OVER (
					PARTITION BY e.idPerson 
					ORDER BY e.id DESC
				) as latest_dwz
			FROM evaluation e
			INNER JOIN mvdsb.mitgliedschaft m ON e.idPerson = m.person
			WHERE m.organisation = ? 
				AND m.bis IS NULL 
				AND m.status = 1
				AND e.dwzNew > 0
		) latest_evaluations
		WHERE latest_dwz > 0
	`, organizationID).Scan(&result).Error

	return result.AvgDWZ, err
}

// GetAllClubs gets all clubs for listing
func (r *ClubRepository) GetAllClubs() ([]models.Organisation, error) {
	var clubs []models.Organisation
	err := r.dbs.MVDSB.Where("status = 1 AND organisationsart = 20 AND vkz != ''").
		Order("name ASC").Find(&clubs).Error
	return clubs, err
}
