package interfaces

import "portal64api/internal/models"

// PlayerRepositoryInterface defines the interface for player repository operations
type PlayerRepositoryInterface interface {
	GetPlayerByID(vkz string, spielernummer uint) (*models.Person, *models.Organisation, *models.Evaluation, error)
	SearchPlayers(req models.SearchRequest) ([]models.Person, int64, error)
	GetPlayersByClub(vkz string, req models.SearchRequest) ([]models.Person, int64, error)
	GetPlayerRatingHistory(personID uint) ([]models.Evaluation, error)
	GetPlayerCurrentClub(personID uint) (*models.Organisation, error)
	GetPlayerCurrentMembership(personID uint) (*models.Mitgliedschaft, error)
}

// ClubRepositoryInterface defines the interface for club repository operations
type ClubRepositoryInterface interface {
	GetClubByVKZ(vkz string) (*models.Organisation, error)
	SearchClubs(req models.SearchRequest) ([]models.Organisation, int64, error)
	GetClubMemberCount(organizationID uint) (int64, error)
	GetClubAverageDWZ(organizationID uint) (float64, error)
	GetAllClubs() ([]models.Organisation, error)
}
