package cache

import (
	"crypto/md5"
	"fmt"
	"portal64api/internal/models"
	"strings"
)

// Key generation constants
const (
	PlayerKeyPrefix      = "player"
	ClubKeyPrefix        = "club"
	TournamentKeyPrefix  = "tournament"
	AddressKeyPrefix     = "address"
	SearchKeyPrefix      = "search"
)

// KeyGenerator provides cache key generation utilities
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// Player-related keys
func (kg *KeyGenerator) PlayerKey(playerID string) string {
	return fmt.Sprintf("%s:%s", PlayerKeyPrefix, playerID)
}

func (kg *KeyGenerator) PlayerRatingHistoryKey(playerID string) string {
	return fmt.Sprintf("%s:%s:rating-history", PlayerKeyPrefix, playerID)
}

// Club-related keys
func (kg *KeyGenerator) ClubKey(clubID string) string {
	return fmt.Sprintf("%s:%s", ClubKeyPrefix, clubID)
}

func (kg *KeyGenerator) ClubPlayersKey(clubID string, sort string) string {
	if sort == "" {
		return fmt.Sprintf("%s:%s:players", ClubKeyPrefix, clubID)
	}
	return fmt.Sprintf("%s:%s:players:%s", ClubKeyPrefix, clubID, sort)
}

func (kg *KeyGenerator) ClubProfileKey(clubID string) string {
	return fmt.Sprintf("%s:%s:profile", ClubKeyPrefix, clubID)
}

func (kg *KeyGenerator) ClubsAllKey() string {
	return fmt.Sprintf("%s:all", ClubKeyPrefix)
}

func (kg *KeyGenerator) ClubListKey(listType string) string {
	return fmt.Sprintf("%s:%s", ClubKeyPrefix, listType)
}

// Tournament-related keys
func (kg *KeyGenerator) TournamentKey(tournamentID string) string {
	return fmt.Sprintf("%s:%s", TournamentKeyPrefix, tournamentID)
}

func (kg *KeyGenerator) TournamentListKey(listType string) string {
	return fmt.Sprintf("%s:%s", TournamentKeyPrefix, listType)
}

// Address-related keys
func (kg *KeyGenerator) AddressRegionKey(region string) string {
	return fmt.Sprintf("%s:region:%s", AddressKeyPrefix, region)
}

func (kg *KeyGenerator) AddressRegionsKey() string {
	return fmt.Sprintf("%s:regions", AddressKeyPrefix)
}

func (kg *KeyGenerator) AddressTypesKey(region string) string {
	return fmt.Sprintf("%s:region:%s:types", AddressKeyPrefix, region)
}

// Search-related keys
func (kg *KeyGenerator) SearchKey(entityType string, hash string) string {
	return fmt.Sprintf("%s:%s:hash:%s", SearchKeyPrefix, entityType, hash)
}

// Hash generation for search requests
func (kg *KeyGenerator) GenerateSearchHash(req models.SearchRequest, showActive bool) string {
	// Include all search parameters that affect results
	sortKey := fmt.Sprintf("%s:%s", req.SortBy, req.SortOrder)
	data := fmt.Sprintf("%s:%d:%d:%s:%t", 
		strings.ToLower(req.Query), req.Limit, req.Offset, sortKey, showActive)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// Club search hash with additional parameters
func (kg *KeyGenerator) GenerateClubSearchHash(req models.SearchRequest) string {
	sortKey := fmt.Sprintf("%s:%s", req.SortBy, req.SortOrder)
	data := fmt.Sprintf("%s:%d:%d:%s", 
		strings.ToLower(req.Query), req.Limit, req.Offset, sortKey)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// Tournament search hash
func (kg *KeyGenerator) GenerateTournamentSearchHash(req models.SearchRequest) string {
	sortKey := fmt.Sprintf("%s:%s", req.SortBy, req.SortOrder)
	data := fmt.Sprintf("%s:%d:%d:%s", 
		strings.ToLower(req.Query), req.Limit, req.Offset, sortKey)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// Key validation
func (kg *KeyGenerator) ValidateKey(key string) bool {
	if key == "" {
		return false
	}
	
	// Check for valid prefixes
	validPrefixes := []string{
		PlayerKeyPrefix, ClubKeyPrefix, TournamentKeyPrefix, 
		AddressKeyPrefix, SearchKeyPrefix,
	}
	
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(key, prefix+":") {
			return true
		}
	}
	
	return false
}

// Extract entity type from key
func (kg *KeyGenerator) GetEntityType(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// Extract entity ID from key
func (kg *KeyGenerator) GetEntityID(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
