package repositories

import (
	"fmt"
	"time"

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

// GetPlayerByID gets a player by their ID (VKZ-Spielernummer format)
func (r *PlayerRepository) GetPlayerByID(vkz string, spielernummer uint) (*models.Person, *models.Organisation, *models.Evaluation, error) {
	// First, get the organization by VKZ
	var org models.Organisation
	if err := r.dbs.MVDSB.Where("vkz = ?", vkz).First(&org).Error; err != nil {
		return nil, nil, nil, err
	}

	// Get the membership by organization and spielernummer - PHP-style: include future-ending memberships
	var membership models.Mitgliedschaft
	err := r.dbs.MVDSB.Where("organisation = ? AND spielernummer = ? AND (bis IS NULL OR bis > CURDATE())", org.ID, spielernummer).
		First(&membership).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Get person from MVDSB database using the person ID from membership
	var person models.Person
	if err := r.dbs.MVDSB.Where("id = ?", membership.Person).First(&person).Error; err != nil {
		return nil, nil, nil, err
	}

	// Get latest DWZ evaluation from Portal64_BDW
	var evaluation models.Evaluation
	r.dbs.Portal64BDW.Where("idPerson = ?", membership.Person).
		Order("id DESC").First(&evaluation)

	return &person, &org, &evaluation, nil
}

// SearchPlayers searches for players by name
func (r *PlayerRepository) SearchPlayers(req models.SearchRequest, showActive bool) ([]models.Person, int64, error) {
	players := make([]models.Person, 0)
	var total int64

	// Start with basic person query
	query := r.dbs.MVDSB.Model(&models.Person{}).Where("person.status = 0")

	// If showActive is true, optimize by using a subquery instead of large IN clause
	if showActive {
		// Use EXISTS subquery to avoid MySQL prepared statement placeholder limit
		// This is more efficient than a large IN clause and avoids the 65K placeholder limit
		// PHP-style: include future-ending memberships
		query = query.Where("EXISTS (SELECT 1 FROM mitgliedschaft WHERE mitgliedschaft.person = person.id AND (mitgliedschaft.bis IS NULL OR mitgliedschaft.bis > CURDATE()))")
	}

	// Add search filter with efficient prefix matching (like the old PHP code)
	if req.Query != "" {
		// Use range-based prefix matching for better performance and index usage
		// This matches the original PHP implementation: name >= 'query' AND name < 'queryzz'
		upperBound := req.Query + "zz"
		query = query.Where(
			"(person.name >= ? AND person.name < ?) OR (person.vorname >= ? AND person.vorname < ?)",
			req.Query, upperBound, req.Query, upperBound)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting
	orderBy := "person.name ASC"
	if req.SortBy != "" {
		direction := "ASC"
		if req.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("person.%s %s", req.SortBy, direction)
	}

	// Apply pagination and execute
	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&players).Error

	return players, total, err
}

// GetPlayersByClub gets all players in a club
func (r *PlayerRepository) GetPlayersByClub(vkz string, req models.SearchRequest, showActive bool) ([]models.Person, int64, error) {
	// First get the organization ID by VKZ
	var org models.Organisation
	if err := r.dbs.MVDSB.Where("vkz = ?", vkz).First(&org).Error; err != nil {
		return nil, 0, err
	}

	players := make([]models.Person, 0)
	var total int64

	// Handle special sorting cases that require JOINs or special field handling
	if req.SortBy == "current_dwz" {
		// For DWZ sorting, we need to join with evaluations from Portal64_BDW database
		// Use EXISTS subquery instead of IN clause to avoid MySQL parameter limit
		// PHP-style: include future-ending memberships
		query := r.dbs.MVDSB.Model(&models.Person{}).
			Where("status = 0 AND EXISTS (SELECT 1 FROM mitgliedschaft WHERE mitgliedschaft.person = person.id AND mitgliedschaft.organisation = ? AND (mitgliedschaft.bis IS NULL OR mitgliedschaft.bis > CURDATE()))", org.ID)

		// Add search filter
		if req.Query != "" {
			upperBound := req.Query + "zz"
			query = query.Where(
				"(name >= ? AND name < ?) OR (vorname >= ? AND vorname < ?)",
				req.Query, upperBound, req.Query, upperBound)
		}

		// Get total count
		query.Count(&total)

		// Get all matching players (without pagination for sorting)
		allPlayers := make([]models.Person, 0)
		err := query.Find(&allPlayers).Error
		if err != nil {
			return nil, 0, err
		}

		// FIXED: Batch fetch evaluations to avoid N+1 queries
		// Extract all player IDs for batch query
		allPersonIDs := make([]uint, len(allPlayers))
		for i, player := range allPlayers {
			allPersonIDs[i] = player.ID
		}

		// Fetch latest evaluations in batches to avoid MySQL parameter limit
		const batchSize = 1000
		allEvaluations := make([]models.Evaluation, 0)

		for i := 0; i < len(allPersonIDs); i += batchSize {
			end := i + batchSize
			if end > len(allPersonIDs) {
				end = len(allPersonIDs)
			}

			var batchEvaluations []models.Evaluation
			err = r.dbs.Portal64BDW.Where("idPerson IN ?", allPersonIDs[i:end]).
				Order("idPerson, id DESC").Find(&batchEvaluations).Error
			if err != nil {
				return nil, 0, err
			}

			allEvaluations = append(allEvaluations, batchEvaluations...)
		}

		// Create a map of latest evaluations by person ID for fast lookup
		latestEvaluations := make(map[uint]models.Evaluation)
		for _, eval := range allEvaluations {
			if _, exists := latestEvaluations[eval.IDPerson]; !exists {
				latestEvaluations[eval.IDPerson] = eval
			}
		}

		// Get latest evaluations for all players and sort
		type PlayerWithDWZ struct {
			Player models.Person
			DWZ    int
		}

		playersWithDWZ := make([]PlayerWithDWZ, len(allPlayers))
		for i, player := range allPlayers {
			dwz := 0
			if eval, exists := latestEvaluations[player.ID]; exists {
				dwz = eval.DWZNew
			}

			playersWithDWZ[i] = PlayerWithDWZ{
				Player: player,
				DWZ:    dwz,
			}
		}

		// Sort by DWZ
		for i := 0; i < len(playersWithDWZ)-1; i++ {
			for j := i + 1; j < len(playersWithDWZ); j++ {
				shouldSwap := false
				if req.SortOrder == "desc" {
					shouldSwap = playersWithDWZ[i].DWZ < playersWithDWZ[j].DWZ
				} else {
					shouldSwap = playersWithDWZ[i].DWZ > playersWithDWZ[j].DWZ
				}
				if shouldSwap {
					playersWithDWZ[i], playersWithDWZ[j] = playersWithDWZ[j], playersWithDWZ[i]
				}
			}
		}

		// Apply pagination
		start := req.Offset
		end := req.Offset + req.Limit
		if start > len(playersWithDWZ) {
			start = len(playersWithDWZ)
		}
		if end > len(playersWithDWZ) {
			end = len(playersWithDWZ)
		}

		// Extract paginated players
		players = make([]models.Person, end-start)
		for i := start; i < end; i++ {
			players[i-start] = playersWithDWZ[i].Player
		}

		return players, total, nil

	} else if req.SortBy == "birth_year" {
		// For birth year sorting, use EXISTS subquery instead of IN clause
		// PHP-style: include future-ending memberships
		query := r.dbs.MVDSB.Model(&models.Person{}).
			Where("status = 0 AND EXISTS (SELECT 1 FROM mitgliedschaft WHERE mitgliedschaft.person = person.id AND mitgliedschaft.organisation = ? AND (mitgliedschaft.bis IS NULL OR mitgliedschaft.bis > CURDATE()))", org.ID)

		// Add search filter
		if req.Query != "" {
			upperBound := req.Query + "zz"
			query = query.Where(
				"(name >= ? AND name < ?) OR (vorname >= ? AND vorname < ?)",
				req.Query, upperBound, req.Query, upperBound)
		}

		// Get total count
		query.Count(&total)

		// Use SQL YEAR function for sorting by birth year
		direction := "ASC"
		if req.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy := fmt.Sprintf("YEAR(geburtsdatum) %s", direction)

		err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&players).Error
		return players, total, err

	} else {
		// Standard sorting for other fields - use EXISTS subquery instead of IN clause
		// PHP-style: include future-ending memberships
		query := r.dbs.MVDSB.Model(&models.Person{}).
			Where("status = 0 AND EXISTS (SELECT 1 FROM mitgliedschaft WHERE mitgliedschaft.person = person.id AND mitgliedschaft.organisation = ? AND (mitgliedschaft.bis IS NULL OR mitgliedschaft.bis > CURDATE()))", org.ID)

		// Add search filter
		if req.Query != "" {
			// Use range-based prefix matching for better performance and index usage
			upperBound := req.Query + "zz"
			query = query.Where(
				"(name >= ? AND name < ?) OR (vorname >= ? AND vorname < ?)",
				req.Query, upperBound, req.Query, upperBound)
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
}

// EvaluationWithTournament represents an evaluation with joined tournament data
// This eliminates N+1 queries by getting tournament name and date in a single query
type EvaluationWithTournament struct {
	models.Evaluation
	TournamentName       string     `gorm:"column:tname"`
	TournamentCode       string     `gorm:"column:tcode"`
	TournamentFinishedOn *time.Time `gorm:"column:finishedOn"`
	TournamentComputedOn *time.Time `gorm:"column:computedOn"`
}

// GetPlayerRatingHistory gets rating history for a player with tournament details in single query
// This eliminates the previous N+1 query problem where we had to fetch tournament details separately
func (r *PlayerRepository) GetPlayerRatingHistory(personID uint) ([]EvaluationWithTournament, error) {
	var results []EvaluationWithTournament
	err := r.dbs.Portal64BDW.Table("evaluation e").
		Select("e.*, tm.tname, tm.tcode, tm.finishedOn, tm.computedOn").
		Joins("INNER JOIN tournamentmaster tm ON e.idMaster = tm.id").
		Where("e.idPerson = ? AND tm.computedOn IS NOT NULL", personID).
		Order("e.id DESC").Find(&results).Error
	return results, err
}

// GetPlayerCurrentClub gets the current club for a player
func (r *PlayerRepository) GetPlayerCurrentClub(personID uint) (*models.Organisation, error) {
	// Get current club membership - PHP-style: include future-ending memberships
	var membership models.Mitgliedschaft
	err := r.dbs.MVDSB.Where("person = ? AND (bis IS NULL OR bis > CURDATE())", personID).
		Order("CASE WHEN bis IS NULL THEN '9999-12-31' ELSE bis END DESC").First(&membership).Error
	if err != nil {
		return nil, err
	}

	// Get organization details
	var org models.Organisation
	err = r.dbs.MVDSB.Where("id = ?", membership.Organisation).First(&org).Error
	if err != nil {
		return nil, err
	}

	return &org, nil
}

// GetPlayerCurrentMembership gets the current membership for a player including spielernummer
func (r *PlayerRepository) GetPlayerCurrentMembership(personID uint) (*models.Mitgliedschaft, error) {
	var membership models.Mitgliedschaft
	err := r.dbs.MVDSB.Where("person = ? AND (bis IS NULL OR bis >= CURDATE())", personID).
		Order("IFNULL(bis, '9999-12-31') DESC, status ASC").First(&membership).Error
	return &membership, err
}

// FormatPlayerID formats player ID in VKZ-XXX format
func (r *PlayerRepository) FormatPlayerID(pkz, vkz string) string {
	if pkz != "" {
		return pkz
	}
	return "UNKNOWN"
}

// FormatGender formats gender code to string
func (r *PlayerRepository) FormatGender(gender int) string {
	switch gender {
	case 1:
		return "male"
	case 2:
		return "female"
	default:
		return "unknown"
	}
}

// GetPlayerCurrentDWZ gets the current DWZ rating for a player
func (r *PlayerRepository) GetPlayerCurrentDWZ(personID uint) int {
	var evaluation models.Evaluation
	err := r.dbs.Portal64BDW.Where("idPerson = ?", personID).
		Order("id DESC").First(&evaluation).Error
	if err != nil {
		return 0
	}
	return evaluation.DWZNew
}

// GetPlayerDWZIndex gets the current DWZ index for a player
func (r *PlayerRepository) GetPlayerDWZIndex(personID uint) int {
	var evaluation models.Evaluation
	err := r.dbs.Portal64BDW.Where("idPerson = ?", personID).
		Order("id DESC").First(&evaluation).Error
	if err != nil {
		return 0
	}
	return evaluation.DWZNewIndex
}

// GetPlayerStatus formats player status
func (r *PlayerRepository) GetPlayerStatus(status uint) string {
	switch status {
	case 0:
		return "active" // Primary membership
	case 1:
		return "active" // Approved membership (was incorrectly "inactive")
	case 2:
		return "inactive" // Passive membership
	default:
		return "unknown"
	}
}
