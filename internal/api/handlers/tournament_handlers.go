package handlers

import (
	"net/http"
	"strconv"
	"time"

	"portal64api/internal/models"
	"portal64api/internal/services"
	"portal64api/pkg/errors"
	"portal64api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// TournamentHandler handles tournament-related HTTP requests
type TournamentHandler struct {
	tournamentService *services.TournamentService
}

// NewTournamentHandler creates a new tournament handler
func NewTournamentHandler(tournamentService *services.TournamentService) *TournamentHandler {
	return &TournamentHandler{tournamentService: tournamentService}
}

// GetTournament godoc
// @Summary Get tournament by ID
// @Description Get a tournament by its ID/code (format: C529-K00-HT1) with comprehensive details
// @Tags tournaments
// @Accept json
// @Produce json,text/csv
// @Param id path string true "Tournament ID (format: C529-K00-HT1)"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.EnhancedTournamentResponse
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/v1/tournaments/{id} [get]
func (h *TournamentHandler) GetTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	
	// Validate tournament ID format
	if err := utils.ValidateTournamentID(tournamentID); err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}
	
	tournament, err := h.tournamentService.GetTournamentByID(tournamentID)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get tournament"))
		return
	}

	utils.HandleResponse(c, tournament, "tournament.csv")
}

// SearchTournaments godoc
// @Summary Search tournaments
// @Description Search tournaments by name, code, or other criteria
// @Tags tournaments
// @Accept json
// @Produce json,text/csv
// @Param query query string false "Search query"
// @Param limit query int false "Limit (max 100)" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(finishedOn)
// @Param sort_order query string false "Sort order (asc/desc)" default(desc)
// @Param filter_by query string false "Filter by field (year)"
// @Param filter_value query string false "Filter value"
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.TournamentResponse,meta=models.Meta}
// @Failure 400 {object} models.Response
// @Router /api/v1/tournaments [get]
func (h *TournamentHandler) SearchTournaments(c *gin.Context) {
	req, err := utils.ParseSearchParamsWithDefaults(c, "finishedOn", "desc")
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}

	tournaments, meta, err := h.tournamentService.SearchTournaments(req)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to search tournaments"))
		return
	}

	response := struct {
		Data []models.TournamentResponse `json:"data"`
		Meta interface{}                 `json:"meta"`
	}{
		Data: tournaments,
		Meta: meta,
	}

	utils.HandleResponse(c, response, "tournaments.csv")
}

// GetRecentTournaments godoc
// @Summary Get recent tournaments
// @Description Get recently finished tournaments
// @Tags tournaments
// @Accept json
// @Produce json,text/csv
// @Param days query int false "Number of days to look back" default(30)
// @Param limit query int false "Maximum number of tournaments to return" default(20)
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.TournamentResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/tournaments/recent [get]
func (h *TournamentHandler) GetRecentTournaments(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	limitStr := c.DefaultQuery("limit", "20")

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 20
	}

	tournaments, err := h.tournamentService.GetRecentTournaments(days, limit)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get recent tournaments"))
		return
	}

	utils.HandleResponse(c, tournaments, "recent_tournaments.csv")
}

// GetTournamentsByDateRange godoc
// @Summary Get tournaments by date range
// @Description Get tournaments within a specific date range
// @Tags tournaments
// @Accept json
// @Produce json,text/csv
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param query query string false "Search query"
// @Param limit query int false "Limit (max 100)" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(finishedOn)
// @Param sort_order query string false "Sort order (asc/desc)" default(desc)
// @Param format query string false "Response format (json or csv)" Enums(json,csv)
// @Success 200 {object} models.Response{data=[]models.TournamentResponse,meta=models.Meta}
// @Failure 400 {object} models.Response
// @Router /api/v1/tournaments/date-range [get]
func (h *TournamentHandler) GetTournamentsByDateRange(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.SendJSONResponse(c, http.StatusBadRequest, 
			errors.NewBadRequestError("start_date and end_date are required"))
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, 
			errors.NewBadRequestError("Invalid start_date format (use YYYY-MM-DD)"))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.SendJSONResponse(c, http.StatusBadRequest, 
			errors.NewBadRequestError("Invalid end_date format (use YYYY-MM-DD)"))
		return
	}

	req, err := utils.ParseSearchParamsWithDefaults(c, "finishedOn", "desc")
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusBadRequest, err)
		return
	}

	tournaments, meta, err := h.tournamentService.GetTournamentsByDateRange(startDate, endDate, req)
	if err != nil {
		if apiErr, ok := err.(errors.APIError); ok {
			utils.SendJSONResponse(c, apiErr.Code, apiErr)
			return
		}
		utils.SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to get tournaments by date range"))
		return
	}

	response := struct {
		Data []models.TournamentResponse `json:"data"`
		Meta interface{}                 `json:"meta"`
	}{
		Data: tournaments,
		Meta: meta,
	}

	utils.HandleResponse(c, response, "tournaments_by_date.csv")
}
