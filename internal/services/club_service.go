package services

import (
	"portal64api/internal/models"
	"portal64api/internal/repositories"
	"portal64api/pkg/errors"
)

// ClubService handles club business logic
type ClubService struct {
	clubRepo *repositories.ClubRepository
}

// NewClubService creates a new club service
func NewClubService(clubRepo *repositories.ClubRepository) *ClubService {
	return &ClubService{clubRepo: clubRepo}
}

// GetClubByID gets a club by its VKZ (ID)
func (s *ClubService) GetClubByID(clubID string) (*models.ClubResponse, error) {
	club, err := s.clubRepo.GetClubByVKZ(clubID)
	if err != nil {
		return nil, errors.NewNotFoundError("Club")
	}

	// Get member count
	memberCount, _ := s.clubRepo.GetClubMemberCount(club.ID)

	// Get average DWZ
	avgDWZ, _ := s.clubRepo.GetClubAverageDWZ(club.ID)

	response := &models.ClubResponse{
		ID:           club.VKZ,
		Name:         club.Name,
		ShortName:    club.Kurzname,
		Region:       getRegionName(club.Verband),
		District:     getDistrictName(club.Bezirk),
		FoundingDate: club.Grundungsdatum,
		MemberCount:  int(memberCount),
		AverageDWZ:   avgDWZ,
		Status:       getClubStatus(club.Status),
	}

	return response, nil
}

// SearchClubs searches clubs by name or other criteria
func (s *ClubService) SearchClubs(req models.SearchRequest) ([]models.ClubResponse, *models.Meta, error) {
	clubs, total, err := s.clubRepo.SearchClubs(req)
	if err != nil {
		return nil, nil, errors.NewInternalServerError("Failed to search clubs")
	}

	responses := make([]models.ClubResponse, len(clubs))
	for i, club := range clubs {
		// Get member count
		memberCount, _ := s.clubRepo.GetClubMemberCount(club.ID)

		// Get average DWZ
		avgDWZ, _ := s.clubRepo.GetClubAverageDWZ(club.ID)

		responses[i] = models.ClubResponse{
			ID:           club.VKZ,
			Name:         club.Name,
			ShortName:    club.Kurzname,
			Region:       getRegionName(club.Verband),
			District:     getDistrictName(club.Bezirk),
			FoundingDate: club.Grundungsdatum,
			MemberCount:  int(memberCount),
			AverageDWZ:   avgDWZ,
			Status:       getClubStatus(club.Status),
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

// GetAllClubs gets all clubs
func (s *ClubService) GetAllClubs() ([]models.ClubResponse, error) {
	clubs, err := s.clubRepo.GetAllClubs()
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to get clubs")
	}

	responses := make([]models.ClubResponse, len(clubs))
	for i, club := range clubs {
		responses[i] = models.ClubResponse{
			ID:           club.VKZ,
			Name:         club.Name,
			ShortName:    club.Kurzname,
			Region:       getRegionName(club.Verband),
			District:     getDistrictName(club.Bezirk),
			FoundingDate: club.Grundungsdatum,
			Status:       getClubStatus(club.Status),
		}
	}

	return responses, nil
}

// Helper methods

// getRegionName converts region code to name
func getRegionName(code string) string {
	regionMap := map[string]string{
		"W": "Württemberg",
		"B": "Baden",
		"H": "Hessen",
		// Add more regions as needed
	}
	if name, exists := regionMap[code]; exists {
		return name
	}
	return code
}

// getDistrictName converts district code to name
func getDistrictName(code string) string {
	// This would typically be a database lookup or config mapping
	// For now, return the code itself
	return code
}

// getClubStatus converts status code to string
func getClubStatus(status uint) string {
	switch status {
	case 0:
		return "active"
	case 1:
		return "inactive"
	case 2:
		return "suspended"
	default:
		return "unknown"
	}
}
