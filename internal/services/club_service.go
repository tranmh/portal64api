package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	
	"portal64api/internal/cache"
	"portal64api/internal/models"
	"portal64api/internal/repositories"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"
)

// ClubService handles club business logic
type ClubService struct {
	clubRepo     *repositories.ClubRepository
	playerRepo   *repositories.PlayerRepository
	cacheService cache.CacheService
	keyGen       *cache.KeyGenerator
}

// NewClubService creates a new club service
func NewClubService(clubRepo *repositories.ClubRepository, cacheService cache.CacheService) *ClubService {
	return &ClubService{
		clubRepo:     clubRepo,
		playerRepo:   nil, // Will be set by dependency injection if needed
		cacheService: cacheService,
		keyGen:       cache.NewKeyGenerator(),
	}
}

// SetPlayerRepository sets the player repository for club-player operations
func (s *ClubService) SetPlayerRepository(playerRepo *repositories.PlayerRepository) {
	s.playerRepo = playerRepo
}

// GetClubByID gets a club by its VKZ (ID)
func (s *ClubService) GetClubByID(clubID string) (*models.ClubResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.ClubKey(clubID)
	
	// Try cache first with background refresh
	var cachedClub models.ClubResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedClub,
		func() (interface{}, error) {
			return s.loadClubFromDB(clubID)
		}, 1*time.Hour) // Cache club details for 1 hour
	
	if err == nil {
		return &cachedClub, nil
	}
	
	// Cache miss or error - load directly from database
	return s.loadClubFromDB(clubID)
}

// loadClubFromDB loads club data from database (used by cache refresh)
func (s *ClubService) loadClubFromDB(clubID string) (*models.ClubResponse, error) {
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

// clubSearchResult wraps search results for caching
type clubSearchResult struct {
	Responses []models.ClubResponse `json:"responses"`
	Meta      *models.Meta         `json:"meta"`
}

// SearchClubs searches clubs by name or other criteria
func (s *ClubService) SearchClubs(req models.SearchRequest) ([]models.ClubResponse, *models.Meta, error) {
	ctx := context.Background()
	searchHash := s.keyGen.GenerateSearchHash(req, false) // clubs don't have active flag
	cacheKey := s.keyGen.SearchKey("club", searchHash)

	// Try cache first with background refresh
	var cachedResult clubSearchResult
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedResult,
		func() (interface{}, error) {
			return s.executeClubSearch(req)
		}, 15*time.Minute) // Cache search results for 15 minutes

	if err == nil {
		return cachedResult.Responses, cachedResult.Meta, nil
	}

	// Fallback to direct execution if cache fails
	result, err := s.executeClubSearch(req)
	if err != nil {
		return nil, nil, err
	}
	return result.Responses, result.Meta, nil
}

// executeClubSearch performs the actual club search
func (s *ClubService) executeClubSearch(req models.SearchRequest) (*clubSearchResult, error) {
	clubs, total, err := s.clubRepo.SearchClubs(req)
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to search clubs")
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

	return &clubSearchResult{
		Responses: responses,
		Meta:      meta,
	}, nil
}

// GetAllClubs gets all clubs
func (s *ClubService) GetAllClubs() ([]models.ClubResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.ClubListKey("all_clubs")

	// Try cache first with background refresh
	var cachedClubs []models.ClubResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedClubs,
		func() (interface{}, error) {
			return s.loadAllClubsFromDB()
		}, 1*time.Hour) // Cache all clubs for 1 hour

	if err == nil {
		return cachedClubs, nil
	}

	// Fallback to direct DB access if cache fails
	return s.loadAllClubsFromDB()
}

// loadAllClubsFromDB loads all clubs from database
func (s *ClubService) loadAllClubsFromDB() ([]models.ClubResponse, error) {
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

// GetClubProfile gets comprehensive club profile with players and statistics
func (s *ClubService) GetClubProfile(clubID string) (*models.ClubProfileResponse, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.ClubProfileKey(clubID)

	// Try cache first with background refresh
	var cachedProfile models.ClubProfileResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedProfile,
		func() (interface{}, error) {
			return s.loadClubProfileFromDB(clubID)
		}, 30*time.Minute) // Cache club profile for 30 minutes

	if err == nil {
		return &cachedProfile, nil
	}

	// Fallback to direct DB access if cache fails
	return s.loadClubProfileFromDB(clubID)
}

// loadClubProfileFromDB loads club profile from database
func (s *ClubService) loadClubProfileFromDB(clubID string) (*models.ClubProfileResponse, error) {
	// Get basic club information
	club, err := s.GetClubByID(clubID)
	if err != nil {
		return nil, err
	}

	// Initialize the profile response
	profile := &models.ClubProfileResponse{
		Club:            *club,
		Players:         []models.PlayerResponse{},
		PlayerCount:     club.MemberCount,
		ActivePlayerCount: 0,
		RatingStats: models.ClubRatingStats{
			AverageDWZ:         club.AverageDWZ,
			RatingDistribution: make(map[string]int),
		},
		RecentTournaments: []models.TournamentResponse{},
		TournamentCount:   0,
		Teams:            []models.ClubTeam{},
		Contact:          s.generateClubContact(clubID, club.Name),
	}

	// Get club players if player repository is available
	if s.playerRepo != nil {
		players, err := s.getClubPlayersForProfile(clubID)
		if err == nil {
			profile.Players = players
			profile.PlayerCount = len(players)
			
			// Calculate rating statistics
			profile.RatingStats = s.calculateRatingStats(players)
			
			// Count active players
			activeCount := 0
			for _, player := range players {
				if player.Status == "active" {
					activeCount++
				}
			}
			profile.ActivePlayerCount = activeCount
		}
	}

	return profile, nil
}

// getClubPlayersForProfile gets club players for the profile display
func (s *ClubService) getClubPlayersForProfile(clubID string) ([]models.PlayerResponse, error) {
	// Get players for this club (active only)
	req := models.SearchRequest{
		Limit:  100, // Limit to first 100 players for profile display
		Offset: 0,
		SortBy: "current_dwz",
		SortOrder: "desc", // Highest rated first
	}
	
	players, _, err := s.playerRepo.GetPlayersByClub(clubID, req, true) // true = active only
	if err != nil {
		return nil, err
	}

	// Get club info for responses
	club, err := s.clubRepo.GetClubByVKZ(clubID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	responses := make([]models.PlayerResponse, len(players))
	for i, player := range players {
		responses[i] = models.PlayerResponse{
			ID:         s.playerRepo.FormatPlayerID(player.PKZ, club.VKZ),
			Name:       player.Name,
			Firstname:  player.Vorname,
			Club:       club.Name,
			ClubID:     club.VKZ,
			BirthYear:  utils.ExtractBirthYear(player.Geburtsdatum), // GDPR compliant: only birth year
			Gender:     s.playerRepo.FormatGender(player.Geschlecht),
			Nation:     player.Nation,
			FideID:     player.IDFide,
			CurrentDWZ: s.playerRepo.GetPlayerCurrentDWZ(player.ID),
			DWZIndex:   s.playerRepo.GetPlayerDWZIndex(player.ID),
			Status:     s.playerRepo.GetPlayerStatus(player.Status),
		}
	}

	return responses, nil
}

// calculateRatingStats calculates rating statistics for club players
func (s *ClubService) calculateRatingStats(players []models.PlayerResponse) models.ClubRatingStats {
	stats := models.ClubRatingStats{
		RatingDistribution: make(map[string]int),
	}

	if len(players) == 0 {
		return stats
	}

	// Collect DWZ ratings
	var dwzRatings []int
	playersWithDWZ := 0
	totalDWZ := 0

	for _, player := range players {
		if player.CurrentDWZ > 0 {
			dwzRatings = append(dwzRatings, player.CurrentDWZ)
			totalDWZ += player.CurrentDWZ
			playersWithDWZ++

			// Rating distribution
			category := s.getRatingCategory(player.CurrentDWZ)
			stats.RatingDistribution[category]++
		}
	}

	stats.PlayersWithDWZ = playersWithDWZ

	if playersWithDWZ > 0 {
		// Average DWZ
		stats.AverageDWZ = float64(totalDWZ) / float64(playersWithDWZ)

		// Sort ratings for median and min/max
		sort.Ints(dwzRatings)
		
		// Median DWZ
		if len(dwzRatings)%2 == 0 {
			stats.MedianDWZ = float64(dwzRatings[len(dwzRatings)/2-1]+dwzRatings[len(dwzRatings)/2]) / 2
		} else {
			stats.MedianDWZ = float64(dwzRatings[len(dwzRatings)/2])
		}

		// Min/Max DWZ
		stats.HighestDWZ = dwzRatings[len(dwzRatings)-1]
		stats.LowestDWZ = dwzRatings[0]
	}

	return stats
}

// getRatingCategory categorizes a DWZ rating
func (s *ClubService) getRatingCategory(dwz int) string {
	switch {
	case dwz >= 2200:
		return "Expert (2200+)"
	case dwz >= 2000:
		return "Advanced (2000-2199)"
	case dwz >= 1800:
		return "Intermediate (1800-1999)"
	case dwz >= 1600:
		return "Beginner+ (1600-1799)"
	case dwz >= 1400:
		return "Beginner (1400-1599)"
	case dwz >= 1200:
		return "Novice (1200-1399)"
	default:
		return "Learning (<1200)"
	}
}

// generateClubContact generates contact information for a club
func (s *ClubService) generateClubContact(clubID, clubName string) models.ClubContact {
	contact := models.ClubContact{}
	
	// Get the organization record first to get the ID
	club, err := s.clubRepo.GetClubByVKZ(clubID)
	if err != nil {
		return contact // Return empty contact if club not found
	}
	
	// Get contact information from the address tables
	contactInfo, err := s.clubRepo.GetClubContactInfo(club.ID)
	if err != nil {
		return contact // Return empty contact if query fails
	}
	
	// Map the contact information to the response model
	if website, exists := contactInfo["website"]; exists {
		// Ensure website has proper protocol
		if website != "" && !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
			if strings.HasPrefix(website, "www.") {
				contact.Website = "http://" + website
			} else {
				contact.Website = "http://www." + website
			}
		} else {
			contact.Website = website
		}
	}
	
	if email, exists := contactInfo["email"]; exists {
		contact.Email = email
	}
	
	if phone, exists := contactInfo["phone"]; exists {
		contact.Phone = phone
	}
	
	if address, exists := contactInfo["address"]; exists {
		contact.Address = address
	}
	
	if meetingTime, exists := contactInfo["meeting_time"]; exists {
		contact.MeetingTime = meetingTime
	}
	
	// If we have additional address components, use them for meeting place
	if city, exists := contactInfo["city"]; exists {
		if additional, hasAdditional := contactInfo["additional"]; hasAdditional && additional != "" {
			contact.MeetingPlace = fmt.Sprintf("%s, %s", additional, city)
		} else if remarks, hasRemarks := contactInfo["remarks"]; hasRemarks && remarks != "" {
			contact.MeetingPlace = fmt.Sprintf("%s, %s", remarks, city)
		} else {
			contact.MeetingPlace = city
		}
	}
	
	return contact
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
