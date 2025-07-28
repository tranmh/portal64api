package services

import (
	"fmt"
	"time"

	"portal64api/internal/models"
	"portal64api/internal/repositories"
	"portal64api/pkg/errors"
)

// TournamentService handles tournament business logic
type TournamentService struct {
	tournamentRepo *repositories.TournamentRepository
}

// NewTournamentService creates a new tournament service
func NewTournamentService(tournamentRepo *repositories.TournamentRepository) *TournamentService {
	return &TournamentService{tournamentRepo: tournamentRepo}
}

// GetTournamentByID gets a tournament by its code/ID
func (s *TournamentService) GetTournamentByID(tournamentID string) (*models.TournamentResponse, error) {
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

// SearchTournaments searches tournaments
func (s *TournamentService) SearchTournaments(req models.SearchRequest) ([]models.TournamentResponse, *models.Meta, error) {
	tournaments, total, err := s.tournamentRepo.SearchTournaments(req)
	if err != nil {
		return nil, nil, errors.NewInternalServerError("Failed to search tournaments")
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

// GetUpcomingTournaments gets upcoming tournaments
func (s *TournamentService) GetUpcomingTournaments(limit int) ([]models.TournamentResponse, error) {
	if limit == 0 {
		limit = 20
	}

	tournaments, err := s.tournamentRepo.GetUpcomingTournaments(limit)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get upcoming tournaments")
	}

	responses := make([]models.TournamentResponse, len(tournaments))
	for i, tournament := range tournaments {
		responses[i] = models.TournamentResponse{
			ID:               fmt.Sprintf("SVW-%d", tournament.TID),
			Name:             tournament.TName,
			Type:             "League", // Assuming league tournaments
			StartDate:        tournament.Teilnahmeschluss,
			EndDate:          tournament.Meldeschluss,
			Status:           getSVWTournamentStatus(&tournament),
			ParticipantCount: tournament.AnzStammspieler + tournament.AnzErsatzspieler,
		}
	}

	return responses, nil
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
