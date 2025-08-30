package services

import (
	"context"
	"fmt"
	"time"

	"portal64api/internal/cache"
	"portal64api/internal/interfaces"
	"portal64api/internal/models"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"
)

// searchResult holds cached search results
type searchResult struct {
	Responses []models.PlayerResponse `json:"responses"`
	Meta      *models.Meta            `json:"meta"`
}

// clubPlayersResult holds cached club players results
type clubPlayersResult struct {
	Responses []models.PlayerResponse `json:"responses"`
	Meta      *models.Meta            `json:"meta"`
}

// PlayerService handles player business logic
type PlayerService struct {
	playerRepo     interfaces.PlayerRepositoryInterface
	clubRepo       interfaces.ClubRepositoryInterface
	tournamentRepo interfaces.TournamentRepositoryInterface
	cacheService   cache.CacheService
	keyGen         *cache.KeyGenerator
}

// NewPlayerService creates a new player service
func NewPlayerService(playerRepo interfaces.PlayerRepositoryInterface, clubRepo interfaces.ClubRepositoryInterface, tournamentRepo interfaces.TournamentRepositoryInterface, cacheService cache.CacheService) *PlayerService {
	return &PlayerService{
		playerRepo:     playerRepo,
		clubRepo:       clubRepo,
		tournamentRepo: tournamentRepo,
		cacheService:   cacheService,
		keyGen:         cache.NewKeyGenerator(),
	}
}

// GetPlayerByID gets a player by their ID
func (s *PlayerService) GetPlayerByID(playerID string) (*models.PlayerResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.PlayerKey(playerID)

	// Try cache first with background refresh
	var cachedPlayer models.PlayerResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedPlayer,
		func() (interface{}, error) {
			return s.loadPlayerFromDB(playerID)
		}, 1*time.Hour)

	if err == nil {
		return &cachedPlayer, nil
	}

	// Cache miss or error - load directly from database
	return s.loadPlayerFromDB(playerID)
}

// loadPlayerFromDB loads player data from database (used by cache refresh)
func (s *PlayerService) loadPlayerFromDB(playerID string) (*models.PlayerResponse, error) {
	// Parse player ID
	vkz, spielernummer, err := utils.ParsePlayerID(playerID)
	if err != nil {
		return nil, errors.NewBadRequestError("Invalid player ID format")
	}

	// Get player data
	person, org, evaluation, err := s.playerRepo.GetPlayerByID(vkz, spielernummer)
	if err != nil {
		return nil, errors.NewNotFoundError("Player")
	}

	// Convert to response format
	response := &models.PlayerResponse{
		ID:        playerID,
		Name:      person.Name,
		Firstname: person.Vorname,
		BirthYear: utils.ExtractBirthYear(person.Geburtsdatum), // GDPR compliant: only birth year
		Nation:    person.Nation,
		FideID:    person.IDFide,
		Status:    getPlayerStatus(person.Status),
	}

	// Add club information if available
	if org != nil {
		response.Club = org.Name
		response.ClubID = org.VKZ
	}

	// Add DWZ information if available
	if evaluation != nil {
		response.CurrentDWZ = evaluation.DWZNew
		response.DWZIndex = evaluation.DWZNewIndex
	}

	// Set gender
	response.Gender = getGenderString(person.Geschlecht)

	return response, nil
}

// SearchPlayers searches players by name
func (s *PlayerService) SearchPlayers(req models.SearchRequest, showActive bool) ([]models.PlayerResponse, *models.Meta, error) {
	ctx := context.Background()

	// Generate cache key for this search
	searchHash := s.keyGen.GenerateSearchHash(req, showActive)
	cacheKey := s.keyGen.SearchKey("player", searchHash)

	// Try cache first with background refresh
	var cachedResult searchResult
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedResult,
		func() (interface{}, error) {
			return s.executePlayerSearch(req, showActive)
		}, 15*time.Minute) // Cache search results for 15 minutes

	if err == nil {
		return cachedResult.Responses, cachedResult.Meta, nil
	}

	// Cache miss or error - execute search directly
	result, err := s.executePlayerSearch(req, showActive)
	if err != nil {
		return nil, nil, err
	}

	searchResult := result.(*searchResult)
	return searchResult.Responses, searchResult.Meta, nil
}

// executePlayerSearch performs the actual player search (used by cache refresh)
func (s *PlayerService) executePlayerSearch(req models.SearchRequest, showActive bool) (interface{}, error) {
	players, _, err := s.playerRepo.SearchPlayers(req, showActive)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to search players")
	}

	// Convert to response format, but only include players with valid club memberships when showActive is true
	responses := make([]models.PlayerResponse, 0, len(players))
	for _, player := range players {
		// Try to get club information
		club, clubErr := s.getPlayerCurrentClub(player.ID)
		membership, membershipErr := s.getPlayerCurrentMembership(player.ID)

		// If showActive is true, skip players without valid memberships
		if showActive && (clubErr != nil || membershipErr != nil || club == nil || membership == nil) {
			continue
		}

		response := models.PlayerResponse{
			Name:      player.Name,
			Firstname: player.Vorname,
			BirthYear: utils.ExtractBirthYear(player.Geburtsdatum), // GDPR compliant: only birth year
			Nation:    player.Nation,
			FideID:    player.IDFide,
			Gender:    getGenderString(player.Geschlecht),
			Status:    getPlayerStatus(player.Status),
		}

		if club != nil && membership != nil {
			response.Club = club.Name
			response.ClubID = club.VKZ
			response.ID = utils.GeneratePlayerID(club.VKZ, membership.Spielernummer)
		} else {
			// Only show UNKNOWN- IDs when showActive is false
			if !showActive {
				response.ID = fmt.Sprintf("UNKNOWN-%d", player.ID)
			} else {
				// Skip this player as we already handled this case above
				continue
			}
		}

		// Try to get DWZ information
		if evaluation, err := s.getPlayerLatestEvaluation(player.ID); err == nil && evaluation != nil {
			response.CurrentDWZ = evaluation.DWZNew
			response.DWZIndex = evaluation.DWZNewIndex
		}

		responses = append(responses, response)
	}

	meta := &models.Meta{
		Total:  len(responses), // Update total to reflect actual returned count
		Limit:  req.Limit,
		Offset: req.Offset,
		Count:  len(responses),
	}

	// Return as search result structure for caching
	return &searchResult{
		Responses: responses,
		Meta:      meta,
	}, nil
}

// GetPlayersByClub gets all players in a specific club
func (s *PlayerService) GetPlayersByClub(clubID string, req models.SearchRequest, showActive bool) ([]models.PlayerResponse, *models.Meta, error) {
	ctx := context.Background()

	// Generate cache key for club players (include sort order and showActive flag)
	sortKey := fmt.Sprintf("%s:%s:%t", req.SortBy, req.SortOrder, showActive)
	cacheKey := s.keyGen.ClubPlayersKey(clubID, sortKey)

	// Try cache first with background refresh
	var cachedResult clubPlayersResult
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedResult,
		func() (interface{}, error) {
			return s.executeClubPlayersSearch(clubID, req, showActive)
		}, 30*time.Minute) // Cache club players for 30 minutes

	if err == nil {
		return cachedResult.Responses, cachedResult.Meta, nil
	}

	// Cache miss or error - execute search directly
	result, err := s.executeClubPlayersSearch(clubID, req, showActive)
	if err != nil {
		return nil, nil, err
	}

	clubResult := result.(*clubPlayersResult)
	return clubResult.Responses, clubResult.Meta, nil
}

// executeClubPlayersSearch performs the actual club players search (used by cache refresh)
func (s *PlayerService) executeClubPlayersSearch(clubID string, req models.SearchRequest, showActive bool) (interface{}, error) {
	players, _, err := s.playerRepo.GetPlayersByClub(clubID, req, showActive)
	if err != nil {
		return nil, errors.NewNotFoundError("Club or players")
	}

	// Get club info
	club, err := s.clubRepo.GetClubByVKZ(clubID)
	if err != nil {
		return nil, errors.NewNotFoundError("Club")
	}

	// Convert to response format, but only include players with valid memberships when showActive is true
	responses := make([]models.PlayerResponse, 0, len(players))
	for _, player := range players {
		// Get membership to retrieve spielernummer
		membership, membershipErr := s.getPlayerCurrentMembership(player.ID)

		// If showActive is true, skip players without valid memberships
		if showActive && (membershipErr != nil || membership == nil) {
			continue
		}

		response := models.PlayerResponse{
			Name:      player.Name,
			Firstname: player.Vorname,
			Club:      club.Name,
			ClubID:    clubID,
			BirthYear: utils.ExtractBirthYear(player.Geburtsdatum), // GDPR compliant: only birth year
			Nation:    player.Nation,
			FideID:    player.IDFide,
			Gender:    getGenderString(player.Geschlecht),
			Status:    getPlayerStatus(player.Status),
		}

		if membership != nil {
			response.ID = utils.GeneratePlayerID(clubID, membership.Spielernummer)
		} else {
			// Only show UNKNOWN- IDs when showActive is false
			if !showActive {
				response.ID = fmt.Sprintf("UNKNOWN-%d", player.ID)
			} else {
				// Skip this player as we already handled this case above
				continue
			}
		}

		// Get DWZ information
		if evaluation, err := s.getPlayerLatestEvaluation(player.ID); err == nil && evaluation != nil {
			response.CurrentDWZ = evaluation.DWZNew
			response.DWZIndex = evaluation.DWZNewIndex
		}

		responses = append(responses, response)
	}

	meta := &models.Meta{
		Total:  len(responses), // Update total to reflect actual returned count
		Limit:  req.Limit,
		Offset: req.Offset,
		Count:  len(responses),
	}

	// Return as club players result structure for caching
	return &clubPlayersResult{
		Responses: responses,
		Meta:      meta,
	}, nil
}

// GetPlayerRatingHistory gets rating history for a player
func (s *PlayerService) GetPlayerRatingHistory(playerID string) ([]models.RatingHistoryResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.PlayerRatingHistoryKey(playerID)

	// Try cache first with background refresh
	var cachedHistory []models.RatingHistoryResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedHistory,
		func() (interface{}, error) {
			return s.loadPlayerRatingHistoryFromDB(playerID)
		}, 7*24*time.Hour) // Cache rating history for 7 days (historical data rarely changes)

	if err == nil {
		return cachedHistory, nil
	}

	// Cache miss or error - load directly from database
	return s.loadPlayerRatingHistoryFromDB(playerID)
}

// loadPlayerRatingHistoryFromDB loads player rating history from database (used by cache refresh)
// Uses optimized single query with JOIN to eliminate N+1 query problem
func (s *PlayerService) loadPlayerRatingHistoryFromDB(playerID string) ([]models.RatingHistoryResponse, error) {
	// Parse player ID to get VKZ and spielernummer
	vkz, spielernummer, err := utils.ParsePlayerID(playerID)
	if err != nil {
		return nil, errors.NewBadRequestError("Invalid player ID format")
	}

	// Get player data to find the actual person ID
	person, _, _, err := s.playerRepo.GetPlayerByID(vkz, spielernummer)
	if err != nil {
		return nil, errors.NewNotFoundError("Player")
	}

	// Get rating history with tournament details in single optimized query
	results, err := s.playerRepo.GetPlayerRatingHistory(person.ID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get rating history")
	}

	// Convert results to response format - no more N+1 queries needed!
	validEvaluations := []models.RatingHistoryResponse{}
	for _, result := range results {
		// Tournament code and name are already available from the JOIN
		if result.TournamentCode == "" {
			continue // Skip evaluations without valid tournament codes
		}

		// Select tournament date: prefer finishedOn over computedOn as requested
		var tournamentDate *time.Time
		if result.TournamentFinishedOn != nil {
			tournamentDate = result.TournamentFinishedOn
		} else if result.TournamentComputedOn != nil {
			tournamentDate = result.TournamentComputedOn
		}

		validEvaluation := models.RatingHistoryResponse{
			ID:             result.ID,
			TournamentID:   result.TournamentCode, // From JOIN - no separate query needed
			TournamentName: result.TournamentName, // From JOIN - new field for demo
			TournamentDate: tournamentDate,        // From JOIN - new field for kader-planung
			ECoefficient:   result.ECoefficient,
			We:             result.We,
			Achievement:    result.Achievement,
			Level:          result.Level,
			Games:          result.Games,
			UnratedGames:   result.UnratedGames,
			Points:         result.Points,
			DWZOld:         result.DWZOld,
			DWZOldIndex:    result.DWZOldIndex,
			DWZNew:         result.DWZNew,
			DWZNewIndex:    result.DWZNewIndex,
		}

		validEvaluations = append(validEvaluations, validEvaluation)
	}

	return validEvaluations, nil
}

// Helper methods

// getTournamentCodeByID gets tournament code by tournament ID
func (s *PlayerService) getTournamentCodeByID(tournamentID uint) (string, error) {
	return s.tournamentRepo.GetTournamentCodeByID(tournamentID)
}

// getPlayerCurrentClub gets the current club for a player
func (s *PlayerService) getPlayerCurrentClub(personID uint) (*models.Organisation, error) {
	return s.playerRepo.GetPlayerCurrentClub(personID)
}

// getPlayerCurrentMembership gets the current membership for a player
func (s *PlayerService) getPlayerCurrentMembership(personID uint) (*models.Mitgliedschaft, error) {
	return s.playerRepo.GetPlayerCurrentMembership(personID)
}

// getPlayerLatestEvaluation gets the latest DWZ evaluation for a player
func (s *PlayerService) getPlayerLatestEvaluation(personID uint) (*models.Evaluation, error) {
	results, err := s.playerRepo.GetPlayerRatingHistory(personID)
	if err != nil || len(results) == 0 {
		return nil, err
	}
	// Extract the Evaluation part from EvaluationWithTournament
	return &results[0].Evaluation, nil
}

// getGenderString converts gender code to string
func getGenderString(gender int) string {
	switch gender {
	case 1:
		return "male"
	case 2:
		return "female"
	default:
		return "unknown"
	}
}

// getPlayerStatus converts status code to string
func getPlayerStatus(status uint) string {
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
