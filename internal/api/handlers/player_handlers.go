package handlers

import (
	"net/http"
	"strconv"

	"portal64api/internal/models"
	"portal64api/internal/services"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// PlayerHandler handles player-related HTTP requests
type PlayerHandler struct {
	playerService *services.PlayerService
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(playerService *services.PlayerService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

// GetPlayer godoc
// @Summary Get player by ID
// @Description Get a player by their ID (format: C0101-1014)
// @Tags players
// @Accept json
// @Produce json,text/csv
// @Param id path string true "Player ID (format: C0101-1014)"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.PlayerResponse
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/players/{id} [get]
func (h *PlayerHandler) GetPlayer(c *gin.Context) {
	playerID := c.Param("id")
	
	// Validate player ID format
	if err := utils.ValidatePlayerID(playerID); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}
	
	player, err := h.playerService.GetPlayerByID(playerID)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get player"))
		return
	}

	utils.HandleResponse(c, player, "player.csv")
}

// SearchPlayers godoc
// @Summary Search players
// @Description Search players by name with pagination
// @Tags players
// @Accept json
// @Produce json,text/csv
// @Param query query string false "Search query"
// @Param limit query int false "Limit (max 100)" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(name)
// @Param sort_order query string false "Sort order (asc/desc)" default(asc)
// @Param active query bool false "Show only active players with valid club memberships" default(true)
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.PlayerResponse,meta=models.Meta}
// @Failure 400 {object} models.Response
// @Router /api/v1/players [get]
func (h *PlayerHandler) SearchPlayers(c *gin.Context) {
	req, err := utils.ParseSearchParams(c)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}

	// Parse active parameter, default to true
	activeStr := c.DefaultQuery("active", "true")
	showActive, err := strconv.ParseBool(activeStr)
	if err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, 
			errors.NewBadRequestError("Invalid active parameter"))
		return
	}

	players, meta, err := h.playerService.SearchPlayers(req, showActive)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to search players"))
		return
	}

	response := struct {
		Data []models.PlayerResponse `json:"data"`
		Meta interface{}             `json:"meta"`
	}{
		Data: players,
		Meta: meta,
	}

	utils.HandleResponse(c, response, "players.csv")
}

// GetPlayerRatingHistory godoc
// @Summary Get player rating history
// @Description Get DWZ rating history for a player
// @Tags players
// @Accept json
// @Produce json,text/csv
// @Param id path string true "Player ID (format: C0101-1014)"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.RatingHistoryResponse}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/players/{id}/rating-history [get]
func (h *PlayerHandler) GetPlayerRatingHistory(c *gin.Context) {
	playerID := c.Param("id")
	
	// Add cache-control headers to prevent browser caching
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	
	// Validate player ID format
	if err := utils.ValidatePlayerID(playerID); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}
	
	history, err := h.playerService.GetPlayerRatingHistory(playerID)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get rating history"))
		return
	}

	utils.HandleResponse(c, history, "rating_history.csv")
}

// GetPlayersByClub godoc
// @Summary Get players by club
// @Description Get all players in a specific club
// @Tags players
// @Accept json
// @Produce json,text/csv
// @Param id path string true "Club ID (format: C0101)"
// @Param query query string false "Search query"
// @Param limit query int false "Limit (max 100)" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(current_dwz)
// @Param sort_order query string false "Sort order (asc/desc)" default(desc)
// @Param active query bool false "Show only active players with valid club memberships" default(true)
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.PlayerResponse,meta=models.Meta}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/clubs/{id}/players [get]
func (h *PlayerHandler) GetPlayersByClub(c *gin.Context) {
	clubID := c.Param("id")
	
	// Validate club ID format
	if err := utils.ValidateClubID(clubID); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}
	
	req, err := utils.ParseSearchParamsWithDefaults(c, "current_dwz", "desc")
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}

	// Parse active parameter, default to true
	activeStr := c.DefaultQuery("active", "true")
	showActive, err := strconv.ParseBool(activeStr)
	if err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, 
			errors.NewBadRequestError("Invalid active parameter"))
		return
	}

	players, meta, err := h.playerService.GetPlayersByClub(clubID, req, showActive)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get club players"))
		return
	}

	response := struct {
		Data []models.PlayerResponse `json:"data"`
		Meta interface{}             `json:"meta"`
	}{
		Data: players,
		Meta: meta,
	}

	utils.HandleResponse(c, response, "club_players.csv")
}
