package services

import (
	"context"
	"time"
	
	"portal64api/internal/cache"
	"portal64api/internal/models"
	"portal64api/internal/repositories"
	"portal64api/pkg/errors"
)

// AddressService handles address-related business logic
type AddressService struct {
	addressRepo  *repositories.AddressRepository
	cacheService cache.CacheService
	keyGen       *cache.KeyGenerator
}

// NewAddressService creates a new address service
func NewAddressService(addressRepo *repositories.AddressRepository, cacheService cache.CacheService) *AddressService {
	return &AddressService{
		addressRepo:  addressRepo,
		cacheService: cacheService,
		keyGen:       cache.NewKeyGenerator(),
	}
}

// GetRegionAddresses retrieves addresses for officials/functionaries in a specific region
func (s *AddressService) GetRegionAddresses(region string, addressType string) ([]models.RegionAddressResponse, error) {
	// Validate region parameter
	if region == "" {
		return nil, errors.NewBadRequestError("Region parameter is required")
	}

	ctx := context.Background()
	cacheKey := s.keyGen.AddressRegionKey(region)
	if addressType != "" {
		cacheKey = s.keyGen.AddressTypesKey(region) // Use types key for filtered requests
	}
	
	// Try cache first with background refresh
	var cachedAddresses []models.RegionAddressResponse
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedAddresses,
		func() (interface{}, error) {
			return s.addressRepo.GetRegionAddresses(region, addressType)
		}, 24*time.Hour) // Cache addresses for 24 hours (static reference data)
		
	if err == nil {
		return cachedAddresses, nil
	}
	
	// Cache miss or error - load directly from database
	addresses, err := s.addressRepo.GetRegionAddresses(region, addressType)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// GetAvailableRegions retrieves all regions that have addresses
func (s *AddressService) GetAvailableRegions() ([]models.RegionInfo, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.AddressRegionsKey()
	
	// Try cache first with background refresh
	var cachedRegions []models.RegionInfo
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedRegions,
		func() (interface{}, error) {
			return s.addressRepo.GetAvailableRegions()
		}, 24*time.Hour) // Cache regions for 24 hours (static reference data)
		
	if err == nil {
		return cachedRegions, nil
	}
	
	// Cache miss or error - load directly from database
	regions, err := s.addressRepo.GetAvailableRegions()
	if err != nil {
		return nil, err
	}

	return regions, nil
}

// GetAddressTypes retrieves available address types for a region
func (s *AddressService) GetAddressTypes(region string) ([]models.AddressTypeInfo, error) {
	ctx := context.Background()
	cacheKey := s.keyGen.AddressTypesKey(region)
	
	// Try cache first with background refresh
	var cachedTypes []models.AddressTypeInfo
	err := s.cacheService.GetWithRefresh(ctx, cacheKey, &cachedTypes,
		func() (interface{}, error) {
			return s.addressRepo.GetAddressTypes(region)
		}, 24*time.Hour) // Cache address types for 24 hours (static reference data)
		
	if err == nil {
		return cachedTypes, nil
	}
	
	// Cache miss or error - load directly from database
	types, err := s.addressRepo.GetAddressTypes(region)
	if err != nil {
		return nil, err
	}

	return types, nil
}
