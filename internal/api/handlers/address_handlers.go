package handlers

import (
	"net/http"

	"portal64api/internal/models"
	"portal64api/internal/services"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AddressHandler handles address-related HTTP requests
type AddressHandler struct {
	addressService *services.AddressService
}

// NewAddressHandler creates a new address handler
func NewAddressHandler(addressService *services.AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

// GetRegionAddresses godoc
// @Summary Get addresses for a specific region
// @Description Get addresses of officials/functionaries for a specific region
// @Tags addresses
// @Accept json
// @Produce json,text/csv
// @Param region path string true "Region code (e.g., C, B, W)"
// @Param type query string false "Address type (e.g., praesidium, vorstand)"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.RegionAddressResponse}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/addresses/{region} [get]
func (h *AddressHandler) GetRegionAddresses(c *gin.Context) {
	region := c.Param("region")
	addressType := c.Query("type")
	
	// Validate region parameter
	if region == "" {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Region parameter is required"))
		return
	}

	// Get addresses from service
	addresses, err := h.addressService.GetRegionAddresses(region, addressType)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to get region addresses"))
		return
	}

	// Check for CSV format
	format := c.Query("format")
	if format == "csv" {
		h.sendAddressesCSV(c, addresses)
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, addresses)
}

// GetRegionAddressesByType godoc
// @Summary Get addresses for a specific region and type
// @Description Get addresses of officials/functionaries for a specific region and address type
// @Tags addresses
// @Accept json
// @Produce json,text/csv
// @Param region path string true "Region code (e.g., C, B, W)"
// @Param type path string true "Address type (e.g., praesidium, vorstand)"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.RegionAddressResponse}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/addresses/{region}/{type} [get]
func (h *AddressHandler) GetRegionAddressesByType(c *gin.Context) {
	region := c.Param("region")
	addressType := c.Param("type")
	
	// Validate parameters
	if region == "" {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Region parameter is required"))
		return
	}
	if addressType == "" {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Address type parameter is required"))
		return
	}

	// Get addresses from service
	addresses, err := h.addressService.GetRegionAddresses(region, addressType)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError,
			errors.NewInternalServerError("Failed to get region addresses"))
		return
	}

	// Check for CSV format
	format := c.Query("format")
	if format == "csv" {
		h.sendAddressesCSV(c, addresses)
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, addresses)
}

// GetAvailableRegions godoc
// @Summary Get available regions
// @Description Get all regions that have address information
// @Tags addresses
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.RegionInfo}
// @Failure 500 {object} models.Response
// @Router /api/v1/addresses/regions [get]
func (h *AddressHandler) GetAvailableRegions(c *gin.Context) {
	regions, err := h.addressService.GetAvailableRegions()
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError, err)
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, regions)
}

// GetAddressTypes godoc
// @Summary Get available address types for a region
// @Description Get all available address/function types for a specific region
// @Tags addresses
// @Accept json
// @Produce json
// @Param region path string true "Region code (e.g., C, B, W)"
// @Success 200 {object} models.Response{data=[]models.AddressTypeInfo}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/addresses/{region}/types [get]
func (h *AddressHandler) GetAddressTypes(c *gin.Context) {
	region := c.Param("region")
	
	// Validate region parameter
	if region == "" {
		utils.SendJSONResponse(c, http.StatusBadRequest,
			errors.NewBadRequestError("Region parameter is required"))
		return
	}

	types, err := h.addressService.GetAddressTypes(region)
	if err != nil {
		utils.SendJSONResponse(c, http.StatusInternalServerError, err)
		return
	}

	utils.SendJSONResponse(c, http.StatusOK, types)
}

// sendAddressesCSV sends addresses in CSV format
func (h *AddressHandler) sendAddressesCSV(c *gin.Context, addresses []models.RegionAddressResponse) {
	// Transform addresses to a flat structure for CSV
	type FlatAddress struct {
		ID               uint   `json:"id" csv:"ID"`
		Name             string `json:"name" csv:"Name"`
		Function         string `json:"function" csv:"Function"`
		Organisation     string `json:"organisation" csv:"Organisation"`
		Region           string `json:"region" csv:"Region"`
		ContactType      string `json:"contact_type" csv:"Contact Type"`
		ContactValue     string `json:"contact_value" csv:"Contact Value"`
	}

	var flatAddresses []FlatAddress

	for _, addr := range addresses {
		if len(addr.ContactDetails) == 0 {
			// Include address even if no contact details
			flatAddr := FlatAddress{
				ID:           addr.ID,
				Name:         addr.Name,
				Function:     addr.FunctionName,
				Organisation: addr.OrganisationName,
				Region:       addr.Region,
				ContactType:  "",
				ContactValue: "",
			}
			flatAddresses = append(flatAddresses, flatAddr)
		} else {
			// Include one row per contact detail
			for _, contact := range addr.ContactDetails {
				flatAddr := FlatAddress{
					ID:           addr.ID,
					Name:         addr.Name,
					Function:     addr.FunctionName,
					Organisation: addr.OrganisationName,
					Region:       addr.Region,
					ContactType:  contact.Type,
					ContactValue: contact.Value,
				}
				flatAddresses = append(flatAddresses, flatAddr)
			}
		}
	}

	utils.SendCSVResponse(c, "addresses.csv", flatAddresses)
}
