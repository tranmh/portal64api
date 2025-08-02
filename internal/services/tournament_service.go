package services

import (
	"context"
	"fmt"
	"time"

	"portal64api/internal/cache"
	"portal64api/internal/models"
	"portal64api/internal/repositories"
	"portal64api/pkg/errors"
)

// TournamentService handles tournament business logic
type TournamentService struct {
	tournamentRepo *repositories.TournamentRepository
	cacheService   cache.CacheService
	keyGen         *cache.KeyGenerator
}

// NewTournamentService creates a new tournament service
func NewTournamentService(tournamentRepo *repositories.TournamentRepository, cacheService cache.CacheService) *TournamentService {
	return &TournamentService{
		tournamentRepo: tournamentRepo,
		cacheService:   cacheService,
		keyGen:         cache.NewKeyGenerator(),
	}
}

// GetTournamentByID gets a tournament by its code/ID
func (s *TournamentService) GetTournamentByID(tournamentID string) (*models.EnhancedTournamentResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.TournamentKey(tournamentID)

	// Try cache first with background refresh
	var cachedTournament models.EnhancedTournamentResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedTournament,
		func() (interface{}, error) {
			return s.loadTournamentFromDB(tournamentID)
		}, 1*time.Hour) // Cache tournament details for 1 hour

	if err == nil {
		return &cachedTournament, nil
	}

	// Fallback to direct DB access if cache fails
	return s.loadTournamentFromDB(tournamentID)
}

// loadTournamentFromDB loads tournament data from database
func (s *TournamentService) loadTournamentFromDB(tournamentID string) (*models.EnhancedTournamentResponse, error) {
	// Get comprehensive tournament data
	tournament, err := s.tournamentRepo.GetEnhancedTournamentData(tournamentID)
	if err != nil {
		return nil, errors.NewNotFoundError("Tournament")
	}

	return tournament, nil
}

// GetBasicTournamentByID gets basic tournament info (for backward compatibility)
func (s *TournamentService) GetBasicTournamentByID(tournamentID string) (*models.TournamentResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.TournamentKey(fmt.Sprintf("basic_%s", tournamentID))

	// Try cache first with background refresh
	var cachedTournament models.TournamentResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedTournament,
		func() (interface{}, error) {
			return s.loadBasicTournamentFromDB(tournamentID)
		}, 1*time.Hour) // Cache basic tournament details for 1 hour

	if err == nil {
		return &cachedTournament, nil
	}

	// Fallback to direct DB access if cache fails
	return s.loadBasicTournamentFromDB(tournamentID)
}

// loadBasicTournamentFromDB loads basic tournament data from database
func (s *TournamentService) loadBasicTournamentFromDB(tournamentID string) (*models.TournamentResponse, error) {
	tournament, err := s.tournamentRepo.GetTournamentByCode(tournamentID)
	if err != nil {
		return nil, errors.NewNotFoundError("Tournament")
	}

	// Get participant count
	participantCount, err := s.tournamentRepo.GetParticipantCount(tournament.ID)
	if err != nil {
		// Log error but don't fail the entire request
		participantCount = 0
	}

	response := &models.TournamentResponse{
		ID:               tournament.TCode,
		Name:             tournament.TName,
		Code:             tournament.TCode,
		Type:             tournament.Type,
		Rounds:           tournament.Rounds,
		StartDate:        tournament.FinishedOn, // Adjust based on actual schema
		EndDate:          tournament.FinishedOn,
		Status:           getTournamentStatus(tournament),
		ParticipantCount: participantCount,
	}

	return response, nil
}

// tournamentSearchResult wraps search results for caching
type tournamentSearchResult struct {
	Responses []models.TournamentResponse
	Meta      *models.Meta
}

// SearchTournaments searches tournaments
func (s *TournamentService) SearchTournaments(req models.SearchRequest) ([]models.TournamentResponse, *models.Meta, error) {
	ctx := context.Background()
	searchHash := s.keyGen.GenerateSearchHash(req, false) // tournaments don't have active flag
	cacheKey := s.keyGen.SearchKey("tournament", searchHash)

	// Try cache first with background refresh
	var cachedResult tournamentSearchResult
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedResult,
		func() (interface{}, error) {
			return s.executeTournamentSearch(req)
		}, 15*time.Minute) // Cache search results for 15 minutes

	if err == nil {
		return cachedResult.Responses, cachedResult.Meta, nil
	}

	// Fallback to direct execution if cache fails
	result, err := s.executeTournamentSearch(req)
	if err != nil {
		return nil, nil, err
	}
	return result.Responses, result.Meta, nil
}

// executeTournamentSearch performs the actual tournament search
func (s *TournamentService) executeTournamentSearch(req models.SearchRequest) (*tournamentSearchResult, error) {
	tournaments, total, err := s.tournamentRepo.SearchTournaments(req)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to search tournaments")
	}

	responses := make([]models.TournamentResponse, len(tournaments))
	for i, tournament := range tournaments {
		// Get participant count
		participantCount, err := s.tournamentRepo.GetParticipantCount(tournament.ID)
		if err != nil {
			// Log error but don't fail the entire request
			participantCount = 0
		}

		responses[i] = models.TournamentResponse{
			ID:               tournament.TCode,
			Name:             tournament.TName,
			Code:             tournament.TCode,
			Type:             tournament.Type,
			Rounds:           tournament.Rounds,
			StartDate:        tournament.FinishedOn,
			EndDate:          tournament.FinishedOn,
			Status:           getTournamentStatus(&tournament),
			ParticipantCount: participantCount,
		}
	}

	meta := &models.Meta{
		Total:  int(total),
		Limit:  req.Limit,
		Offset: req.Offset,
		Count:  len(responses),
	}

	return &tournamentSearchResult{
		Responses: responses,
		Meta:      meta,
	}, nil
}

// GetTournamentsByDateRange gets tournaments within a date range
func (s *TournamentService) GetTournamentsByDateRange(startDate, endDate time.Time, req models.SearchRequest) ([]models.TournamentResponse, *models.Meta, error) {
	tournaments, total, err := s.tournamentRepo.GetTournamentsByDateRange(startDate, endDate, req)
	if err != nil {
		return nil, nil, errors.NewInternalServerError("Failed to get tournaments by date range")
	}

	responses := make([]models.TournamentResponse, len(tournaments))
	for i, tournament := range tournaments {
		// Get participant count
		participantCount, err := s.tournamentRepo.GetParticipantCount(tournament.ID)
		if err != nil {
			// Log error but don't fail the entire request
			participantCount = 0
		}

		responses[i] = models.TournamentResponse{
			ID:               tournament.TCode,
			Name:             tournament.TName,
			Code:             tournament.TCode,
			Type:             tournament.Type,
			Rounds:           tournament.Rounds,
			StartDate:        tournament.FinishedOn,
			EndDate:          tournament.FinishedOn,
			Status:           getTournamentStatus(&tournament),
			ParticipantCount: participantCount,
		}
	}

	meta := &models.Meta{
		Total:  int(total),
		Limit:  req.Limit,
		Offset: req.Offset,
		Count:  len(responses),
	}

	return responses, meta, nil
}

// GetRecentTournaments gets recently finished tournaments
func (s *TournamentService) GetRecentTournaments(days, limit int) ([]models.TournamentResponse, error) {
	if days == 0 {
		days = 30 // Default to last 30 days
	}
	if limit == 0 {
		limit = 20
	}

	tournaments, err := s.tournamentRepo.GetRecentTournaments(days, limit)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get recent tournaments")
	}

	responses := make([]models.TournamentResponse, len(tournaments))
	for i, tournament := range tournaments {
		// Get participant count
		participantCount, err := s.tournamentRepo.GetParticipantCount(tournament.ID)
		if err != nil {
			// Log error but don't fail the entire request
			participantCount = 0
		}

		responses[i] = models.TournamentResponse{
			ID:               tournament.TCode,
			Name:             tournament.TName,
			Code:             tournament.TCode,
			Type:             tournament.Type,
			Rounds:           tournament.Rounds,
			StartDate:        tournament.FinishedOn,
			EndDate:          tournament.FinishedOn,
			Status:           getTournamentStatus(&tournament),
			ParticipantCount: participantCount,
		}
	}

	return responses, nil
}

// Helper methods

// getTournamentStatus determines tournament status
func getTournamentStatus(tournament *models.Tournament) string {
	if tournament.FinishedOn != nil {
		return "finished"
	}
	if tournament.LockedOn != nil {
		return "locked"
	}
	return "active"
}

// getSVWTournamentStatus determines SVW tournament status
func getSVWTournamentStatus(tournament *models.Turnier) string {
	now := time.Now()
	
	if tournament.IsFreigegeben {
		if tournament.Teilnahmeschluss != nil && tournament.Teilnahmeschluss.After(now) {
			return "open"
		}
		if tournament.Meldeschluss != nil && tournament.Meldeschluss.After(now) {
			return "registration_open"
		}
		return "running"
	}
	
	return "pending"
}
