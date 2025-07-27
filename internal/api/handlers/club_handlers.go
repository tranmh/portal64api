package handlers

import (
	"net/http"
	"strings"

	"portal64api/internal/services"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ClubHandler handles club-related HTTP requests
type ClubHandler struct {
	clubService *services.ClubService
}

// NewClubHandler creates a new club handler
func NewClubHandler(clubService *services.ClubService) *ClubHandler {
	return &ClubHandler{clubService: clubService}
}

// GetClub godoc
// @Summary Get club by ID
// @Description Get a club by its VKZ/ID (format: C0101)
// @Tags clubs
// @Accept json
// @Produce json,text/csv
// @Param id path string true "Club ID (format: C0101)"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.ClubResponse
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/clubs/{id} [get]
func (h *ClubHandler) GetClub(c *gin.Context) {
	clubID := c.Param("id")
	
	// Validate club ID format
	if err := utils.ValidateClubID(clubID); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}
	
	club, err := h.clubService.GetClubByID(clubID)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get club"))
		return
	}

	utils.HandleResponse(c, club, "club.csv")
}

// SearchClubs godoc
// @Summary Search clubs
// @Description Search clubs by name, VKZ, or other criteria
// @Tags clubs
// @Accept json
// @Produce json,text/csv
// @Param query query string false "Search query"
// @Param limit query int false "Limit (max 100)" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(name)
// @Param sort_order query string false "Sort order (asc/desc)" default(asc)
// @Param filter_by query string false "Filter by field (region, district)"
// @Param filter_value query string false "Filter value"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.ClubResponse,meta=models.Meta}
// @Failure 400 {object} models.Response
// @Router /api/v1/clubs [get]
func (h *ClubHandler) SearchClubs(c *gin.Context) {
	req, err := utils.ParseSearchParams(c)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}

	clubs, meta, err := h.clubService.SearchClubs(req)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to search clubs"))
		return
	}

	// Check if CSV format is requested
	format := c.Query("format")
	accept := c.GetHeader("Accept")
	
	if format == "csv" || strings.Contains(accept, "text/csv") {
		// For CSV, send just the clubs data directly
		utils.SendCSVResponse(c, "clubs.csv", clubs)
	} else {
		// For JSON, wrap in response structure with metadata
		response := struct {
			Data []interface{} `json:"data"`
			Meta interface{}   `json:"meta"`
		}{
			Data: make([]interface{}, len(clubs)),
			Meta: meta,
		}

		for i, club := range clubs {
			response.Data[i] = club
		}

		utils.SendJSONResponse(c, http.StatusOK, response)
	}
}

// GetAllClubs godoc
// @Summary Get all clubs
// @Description Get all active clubs
// @Tags clubs
// @Accept json
// @Produce json,text/csv
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.ClubResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/clubs/all [get]
func (h *ClubHandler) GetAllClubs(c *gin.Context) {
	clubs, err := h.clubService.GetAllClubs()
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get all clubs"))
		return
	}

	utils.HandleResponse(c, clubs, "all_clubs.csv")
}
