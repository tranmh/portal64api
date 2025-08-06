package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"portal64api/internal/cache"
	"portal64api/internal/interfaces"
	"portal64api/internal/models"
	"portal64api/internal/services"
	"portal64api/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPlayerRepository is a mock implementation of PlayerRepositoryInterface
type MockPlayerRepository struct {
	mock.Mock
}

// Ensure MockPlayerRepository implements PlayerRepositoryInterface
var _ interfaces.PlayerRepositoryInterface = (*MockPlayerRepository)(nil)

func (m *MockPlayerRepository) GetPlayerByID(vkz string, personID uint) (*models.Person, *models.Organisation, *models.Evaluation, error) {
	args := m.Called(vkz, personID)
	
	// Handle potential nil returns properly
	var person *models.Person
	var org *models.Organisation 
	var eval *models.Evaluation
	
	if args.Get(0) != nil {
		person = args.Get(0).(*models.Person)
	}
	if args.Get(1) != nil {
		org = args.Get(1).(*models.Organisation)
	}
	if args.Get(2) != nil {
		eval = args.Get(2).(*models.Evaluation)
	}
	
	return person, org, eval, args.Error(3)
}

func (m *MockPlayerRepository) SearchPlayers(req models.SearchRequest, showActive bool) ([]models.Person, int64, error) {
	args := m.Called(req, showActive)
	return args.Get(0).([]models.Person), args.Get(1).(int64), args.Error(2)
}

func (m *MockPlayerRepository) GetPlayersByClub(vkz string, req models.SearchRequest, showActive bool) ([]models.Person, int64, error) {
	args := m.Called(vkz, req, showActive)
	return args.Get(0).([]models.Person), args.Get(1).(int64), args.Error(2)
}

func (m *MockPlayerRepository) GetPlayerRatingHistory(personID uint) ([]models.Evaluation, error) {
	args := m.Called(personID)
	return args.Get(0).([]models.Evaluation), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerCurrentClub(personID uint) (*models.Organisation, error) {
	args := m.Called(personID)
	return args.Get(0).(*models.Organisation), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerCurrentMembership(personID uint) (*models.Mitgliedschaft, error) {
	args := m.Called(personID)
	return args.Get(0).(*models.Mitgliedschaft), args.Error(1)
}

// MockClubRepository is a mock implementation of ClubRepository
// MockClubRepository is a mock implementation of ClubRepositoryInterface
type MockClubRepository struct {
	mock.Mock
}

// Ensure MockClubRepository implements ClubRepositoryInterface
var _ interfaces.ClubRepositoryInterface = (*MockClubRepository)(nil)

func (m *MockClubRepository) GetClubByVKZ(vkz string) (*models.Organisation, error) {
	args := m.Called(vkz)
	return args.Get(0).(*models.Organisation), args.Error(1)
}

func (m *MockClubRepository) SearchClubs(req models.SearchRequest) ([]models.Organisation, int64, error) {
	args := m.Called(req)
	return args.Get(0).([]models.Organisation), args.Get(1).(int64), args.Error(2)
}

func (m *MockClubRepository) GetClubMemberCount(organizationID uint) (int64, error) {
	args := m.Called(organizationID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClubRepository) GetClubAverageDWZ(organizationID uint) (float64, error) {
	args := m.Called(organizationID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockClubRepository) GetAllClubs() ([]models.Organisation, error) {
	args := m.Called()
	return args.Get(0).([]models.Organisation), args.Error(1)
}

// MockTournamentRepository is a mock implementation of TournamentRepositoryInterface
type MockTournamentRepository struct {
	mock.Mock
}

// Ensure MockTournamentRepository implements TournamentRepositoryInterface
var _ interfaces.TournamentRepositoryInterface = (*MockTournamentRepository)(nil)

func (m *MockTournamentRepository) GetTournamentByID(tournamentID string) (*models.Tournament, error) {
	args := m.Called(tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) SearchTournaments(req models.SearchRequest) ([]models.Tournament, int64, error) {
	args := m.Called(req)
	return args.Get(0).([]models.Tournament), args.Get(1).(int64), args.Error(2)
}

func (m *MockTournamentRepository) GetTournamentsByDateRange(startDate, endDate string) ([]models.Tournament, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]models.Tournament), args.Error(1)
}

func (m *MockTournamentRepository) GetTournamentCodeByID(tournamentID uint) (string, error) {
	args := m.Called(tournamentID)
	return args.String(0), args.Error(1)
}

// MockCacheServiceForPlayer is a simple mock cache service for player tests
type MockCacheServiceForPlayer struct{}

var _ cache.CacheService = (*MockCacheServiceForPlayer)(nil)

func (m *MockCacheServiceForPlayer) Get(ctx context.Context, key string, dest interface{}) error { return nil }
func (m *MockCacheServiceForPlayer) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error { return nil }
func (m *MockCacheServiceForPlayer) Delete(ctx context.Context, key string) error { return nil }
func (m *MockCacheServiceForPlayer) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (m *MockCacheServiceForPlayer) FlushAll(ctx context.Context) error { return nil }
func (m *MockCacheServiceForPlayer) GetWithRefresh(ctx context.Context, key string, dest interface{}, refreshFunc func() (interface{}, error), ttl time.Duration) error { 
	// Simulate cache miss to trigger database lookup
	return fmt.Errorf("cache miss") 
}
func (m *MockCacheServiceForPlayer) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) { return nil, nil }
func (m *MockCacheServiceForPlayer) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error { return nil }
func (m *MockCacheServiceForPlayer) Ping(ctx context.Context) error { return nil }
func (m *MockCacheServiceForPlayer) GetStats() cache.CacheStats { return cache.CacheStats{} }
func (m *MockCacheServiceForPlayer) Close() error { return nil }

func TestPlayerService_GetPlayerByID(t *testing.T) {
	// Setup
	mockPlayerRepo := new(MockPlayerRepository)
	mockClubRepo := new(MockClubRepository)
	mockTournamentRepo := new(MockTournamentRepository)
	mockCacheService := &MockCacheServiceForPlayer{}
	service := services.NewPlayerService(mockPlayerRepo, mockClubRepo, mockTournamentRepo, mockCacheService)

	// Test data
	birth := time.Date(1980, 5, 15, 0, 0, 0, 0, time.UTC)
	person := &models.Person{
		ID:           1014,
		Name:         "Sick",
		Vorname:      "Oliver",
		Geburtsdatum: &birth,
		Nation:       "GER",
		IDFide:       12345,
		Geschlecht:   1,
		Status:       1,
	}

	org := &models.Organisation{
		ID:       1,
		Name:     "Post-SV Ulm",
		VKZ:      "C0101",
		Status:   1,
	}

	evaluation := &models.Evaluation{
		ID:          1,
		IDPerson:    1014,
		DWZNew:      2156,
		DWZNewIndex: 45,
	}

	// Test cases
	tests := []struct {
		name        string
		playerID    string
		setupMocks  func()
		expectError bool
		errorType   error
	}{
		{
			name:     "Valid player ID",
			playerID: "C0101-1014",
			setupMocks: func() {
				mockPlayerRepo.On("GetPlayerByID", "C0101", uint(1014)).Return(person, org, evaluation, nil)
			},
			expectError: false,
		},
		{
			name:     "Invalid player ID format",
			playerID: "invalid-id",
			setupMocks: func() {
				// No mock setup needed as validation fails before repository call
			},
			expectError: true,
			errorType:   errors.NewBadRequestError("Invalid player ID format"),
		},
		{
			name:     "Player not found",
			playerID: "C0101-9999",
			setupMocks: func() {
				// Return nil values to simulate not found
				mockPlayerRepo.On("GetPlayerByID", "C0101", uint(9999)).Return((*models.Person)(nil), (*models.Organisation)(nil), (*models.Evaluation)(nil), errors.NewNotFoundError("Player"))
			},
			expectError: true,
			errorType:   errors.NewNotFoundError("Player"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Execute
			result, err := service.GetPlayerByID(tt.playerID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.playerID, result.ID)
				assert.Equal(t, "Sick", result.Name)
				assert.Equal(t, "Oliver", result.Firstname)
				assert.Equal(t, "Post-SV Ulm", result.Club)
				assert.Equal(t, "C0101", result.ClubID)
				assert.Equal(t, 2156, result.CurrentDWZ)
				assert.Equal(t, 45, result.DWZIndex)
				assert.Equal(t, "male", result.Gender)
				assert.Equal(t, "inactive", result.Status)
			}

			// Clear mock expectations
			mockPlayerRepo.ExpectedCalls = nil
			mockClubRepo.ExpectedCalls = nil
		})
	}
}

func TestPlayerService_SearchPlayers(t *testing.T) {
	// Setup
	mockPlayerRepo := new(MockPlayerRepository)
	mockClubRepo := new(MockClubRepository)
	mockTournamentRepo := new(MockTournamentRepository)
	mockCacheService := &MockCacheServiceForPlayer{}
	service := services.NewPlayerService(mockPlayerRepo, mockClubRepo, mockTournamentRepo, mockCacheService)

	// Test data
	players := []models.Person{
		{
			ID:         1,
			Name:       "Müller",
			Vorname:    "Hans",
			Nation:     "GER",
			Geschlecht: 1,
			Status:     1,
		},
		{
			ID:         2,
			Name:       "Schmidt",
			Vorname:    "Anna",
			Nation:     "GER",
			Geschlecht: 2,
			Status:     1,
		},
	}

	req := models.SearchRequest{
		Query:  "Müller",
		Limit:  20,
		Offset: 0,
	}

	// Setup mock
	mockPlayerRepo.On("SearchPlayers", req, true).Return(players, int64(2), nil)
	// Mock GetPlayerRatingHistory for both players
	mockPlayerRepo.On("GetPlayerRatingHistory", uint(1)).Return([]models.Evaluation{}, nil)
	mockPlayerRepo.On("GetPlayerRatingHistory", uint(2)).Return([]models.Evaluation{}, nil)
	// Mock GetPlayerCurrentClub for both players
	club := &models.Organisation{ID: 1, Name: "Test Club", VKZ: "C0101"}
	mockPlayerRepo.On("GetPlayerCurrentClub", uint(1)).Return(club, nil)
	mockPlayerRepo.On("GetPlayerCurrentClub", uint(2)).Return(club, nil)
	// Mock GetPlayerCurrentMembership for both players
	membership := &models.Mitgliedschaft{ID: 1, Person: 1, Organisation: 1, Spielernummer: 1}
	mockPlayerRepo.On("GetPlayerCurrentMembership", uint(1)).Return(membership, nil)
	mockPlayerRepo.On("GetPlayerCurrentMembership", uint(2)).Return(membership, nil)

	// Execute
	results, meta, err := service.SearchPlayers(req, true)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.NotNil(t, meta)
	assert.Len(t, results, 2)
	assert.Equal(t, int64(2), int64(meta.Total))
	assert.Equal(t, 20, meta.Limit)
	assert.Equal(t, 0, meta.Offset)
	assert.Equal(t, 2, meta.Count)

	// Verify mock was called
	mockPlayerRepo.AssertExpectations(t)
}
