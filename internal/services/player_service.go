package services

import (
	"fmt"

	"portal64api/internal/interfaces"
	"portal64api/internal/models"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"
)

// PlayerService handles player business logic
type PlayerService struct {
	playerRepo interfaces.PlayerRepositoryInterface
	clubRepo   interfaces.ClubRepositoryInterface
}

// NewPlayerService creates a new player service
func NewPlayerService(playerRepo interfaces.PlayerRepositoryInterface, clubRepo interfaces.ClubRepositoryInterface) *PlayerService {
	return &PlayerService{
		playerRepo: playerRepo,
		clubRepo:   clubRepo,
	}
}

// GetPlayerByID gets a player by their ID
func (s *PlayerService) GetPlayerByID(playerID string) (*models.PlayerResponse, error) {
	// Parse player ID
	vkz, personID, err := utils.ParsePlayerID(playerID)
	if err != nil {
		return nil, errors.NewBadRequestError("Invalid player ID format")
	}

	// Get player data
	person, org, evaluation, err := s.playerRepo.GetPlayerByID(vkz, personID)
	if err != nil {
		return nil, errors.NewNotFoundError("Player")
	}

	// Convert to response format
	response := &models.PlayerResponse{
		ID:        playerID,
		Name:      person.Name,
		Firstname: person.Vorname,
		Birth:     person.Geburtsdatum,
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
func (s *PlayerService) SearchPlayers(req models.SearchRequest) ([]models.PlayerResponse, *models.Meta, error) {
	players, total, err := s.playerRepo.SearchPlayers(req)
	if err != nil {
		return nil, nil, errors.NewInternalServerError("Failed to search players")
	}

	// Convert to response format
	responses := make([]models.PlayerResponse, len(players))
	for i, player := range players {
		responses[i] = models.PlayerResponse{
			ID:        fmt.Sprintf("UNKNOWN-%d", player.ID), // Will be updated with club info
			Name:      player.Name,
			Firstname: player.Vorname,
			Birth:     player.Geburtsdatum,
			Nation:    player.Nation,
			FideID:    player.IDFide,
			Gender:    getGenderString(player.Geschlecht),
			Status:    getPlayerStatus(player.Status),
		}

		// Try to get club information
		if club, err := s.getPlayerCurrentClub(player.ID); err == nil && club != nil {
			responses[i].Club = club.Name
			responses[i].ClubID = club.VKZ
			responses[i].ID = utils.GeneratePlayerID(club.VKZ, player.ID)
		}

		// Try to get DWZ information
		if evaluation, err := s.getPlayerLatestEvaluation(player.ID); err == nil && evaluation != nil {
			responses[i].CurrentDWZ = evaluation.DWZNew
			responses[i].DWZIndex = evaluation.DWZNewIndex
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

// GetPlayersByClub gets all players in a specific club
func (s *PlayerService) GetPlayersByClub(clubID string, req models.SearchRequest) ([]models.PlayerResponse, *models.Meta, error) {
	players, total, err := s.playerRepo.GetPlayersByClub(clubID, req)
	if err != nil {
		return nil, nil, errors.NewNotFoundError("Club or players")
	}

	// Get club info
	club, err := s.clubRepo.GetClubByVKZ(clubID)
	if err != nil {
		return nil, nil, errors.NewNotFoundError("Club")
	}

	// Convert to response format
	responses := make([]models.PlayerResponse, len(players))
	for i, player := range players {
		responses[i] = models.PlayerResponse{
			ID:        utils.GeneratePlayerID(clubID, player.ID),
			Name:      player.Name,
			Firstname: player.Vorname,
			Club:      club.Name,
			ClubID:    clubID,
			Birth:     player.Geburtsdatum,
			Nation:    player.Nation,
			FideID:    player.IDFide,
			Gender:    getGenderString(player.Geschlecht),
			Status:    getPlayerStatus(player.Status),
		}

		// Get DWZ information
		if evaluation, err := s.getPlayerLatestEvaluation(player.ID); err == nil && evaluation != nil {
			responses[i].CurrentDWZ = evaluation.DWZNew
			responses[i].DWZIndex = evaluation.DWZNewIndex
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

// GetPlayerRatingHistory gets rating history for a player
func (s *PlayerService) GetPlayerRatingHistory(playerID string) ([]models.Evaluation, error) {
	// Parse player ID
	_, personID, err := utils.ParsePlayerID(playerID)
	if err != nil {
		return nil, errors.NewBadRequestError("Invalid player ID format")
	}

	evaluations, err := s.playerRepo.GetPlayerRatingHistory(personID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get rating history")
	}

	return evaluations, nil
}

// Helper methods

// getPlayerCurrentClub gets the current club for a player
func (s *PlayerService) getPlayerCurrentClub(personID uint) (*models.Organisation, error) {
	// This would require a membership query - simplified for now
	return nil, fmt.Errorf("not implemented")
}

// getPlayerLatestEvaluation gets the latest DWZ evaluation for a player
func (s *PlayerService) getPlayerLatestEvaluation(personID uint) (*models.Evaluation, error) {
	evaluations, err := s.playerRepo.GetPlayerRatingHistory(personID)
	if err != nil || len(evaluations) == 0 {
		return nil, err
	}
	return &evaluations[0], nil
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
		return "active"
	case 1:
		return "inactive"
	default:
		return "unknown"
	}
}
