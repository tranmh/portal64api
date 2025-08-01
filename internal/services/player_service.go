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
func (s *PlayerService) SearchPlayers(req models.SearchRequest, showActive bool) ([]models.PlayerResponse, *models.Meta, error) {
	players, _, err := s.playerRepo.SearchPlayers(req, showActive)
	if err != nil {
		return nil, nil, errors.NewInternalServerError("Failed to search players")
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
			Birth:     player.Geburtsdatum,
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

	return responses, meta, nil
}

// GetPlayersByClub gets all players in a specific club
func (s *PlayerService) GetPlayersByClub(clubID string, req models.SearchRequest, showActive bool) ([]models.PlayerResponse, *models.Meta, error) {
	players, _, err := s.playerRepo.GetPlayersByClub(clubID, req, showActive)
	if err != nil {
		return nil, nil, errors.NewNotFoundError("Club or players")
	}

	// Get club info
	club, err := s.clubRepo.GetClubByVKZ(clubID)
	if err != nil {
		return nil, nil, errors.NewNotFoundError("Club")
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
			Birth:     player.Geburtsdatum,
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

	return responses, meta, nil
}

// GetPlayerRatingHistory gets rating history for a player
func (s *PlayerService) GetPlayerRatingHistory(playerID string) ([]models.Evaluation, error) {
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

	evaluations, err := s.playerRepo.GetPlayerRatingHistory(person.ID)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get rating history")
	}

	return evaluations, nil
}

// Helper methods

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
