package services

import (
	"portal64api/internal/models"
	"portal64api/internal/repositories"
	"portal64api/pkg/errors"
)

// AddressService handles address-related business logic
type AddressService struct {
	addressRepo *repositories.AddressRepository
}

// NewAddressService creates a new address service
func NewAddressService(addressRepo *repositories.AddressRepository) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
	}
}

// GetRegionAddresses retrieves addresses for officials/functionaries in a specific region
func (s *AddressService) GetRegionAddresses(region string, addressType string) ([]models.RegionAddressResponse, error) {
	// Validate region parameter
	if region == "" {
		return nil, errors.NewBadRequestError("Region parameter is required")
	}

	// Get addresses from repository
	addresses, err := s.addressRepo.GetRegionAddresses(region, addressType)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

// GetAvailableRegions retrieves all regions that have addresses
func (s *AddressService) GetAvailableRegions() ([]models.RegionInfo, error) {
	regions, err := s.addressRepo.GetAvailableRegions()
	if err != nil {
		return nil, err
	}

	return regions, nil
}

// GetAddressTypes retrieves available address types for a region
func (s *AddressService) GetAddressTypes(region string) ([]models.AddressTypeInfo, error) {
	types, err := s.addressRepo.GetAddressTypes(region)
	if err != nil {
		return nil, err
	}

	return types, nil
}
