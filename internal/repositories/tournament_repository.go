package repositories

import (
	"fmt"
	"time"

	"portal64api/internal/database"
	"portal64api/internal/models"
	"portal64api/pkg/utils"
)

// TournamentRepository handles tournament data operations
type TournamentRepository struct {
	dbs *database.Databases
}

// NewTournamentRepository creates a new tournament repository
func NewTournamentRepository(dbs *database.Databases) *TournamentRepository {
	return &TournamentRepository{dbs: dbs}
}

// GetTournamentByCode gets a tournament by its code
func (r *TournamentRepository) GetTournamentByCode(code string) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.dbs.Portal64BDW.Where("tcode = ?", code).First(&tournament).Error
	return &tournament, err
}

// SearchTournaments searches for tournaments
func (r *TournamentRepository) SearchTournaments(req models.SearchRequest) ([]models.Tournament, int64, error) {
	tournaments := make([]models.Tournament, 0)
	var total int64

	query := r.dbs.Portal64BDW.Model(&models.Tournament{})

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("tname LIKE ? OR tcode LIKE ?", searchPattern, searchPattern)
	}

	// Apply date filters if provided
	if req.FilterBy == "year" && req.FilterValue != "" {
		query = query.Where("YEAR(finishedOn) = ?", req.FilterValue)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting
	orderBy := "finishedOn DESC"
	if req.SortBy != "" {
		direction := "DESC"
		if req.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, direction)
	}

	// Apply pagination and execute
	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&tournaments).Error
	
	return tournaments, total, err
}

// GetTournamentsByDateRange gets tournaments within a date range
func (r *TournamentRepository) GetTournamentsByDateRange(startDate, endDate time.Time, req models.SearchRequest) ([]models.Tournament, int64, error) {
	var tournaments []models.Tournament
	var total int64

	query := r.dbs.Portal64BDW.Model(&models.Tournament{}).
		Where("finishedOn BETWEEN ? AND ?", startDate, endDate)

	// Add search filter
	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("tname LIKE ? OR tcode LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	query.Count(&total)

	// Apply sorting and pagination
	orderBy := "finishedOn DESC"
	if req.SortBy != "" {
		direction := "DESC"
		if req.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("%s %s", req.SortBy, direction)
	}

	err := query.Order(orderBy).Limit(req.Limit).Offset(req.Offset).Find(&tournaments).Error
	
	return tournaments, total, err
}

// GetTournamentResults gets results for a specific tournament
func (r *TournamentRepository) GetTournamentResults(tournamentID uint) ([]models.Evaluation, error) {
	var evaluations []models.Evaluation
	err := r.dbs.Portal64BDW.Where("idMaster = ?", tournamentID).
		Order("dwzNew DESC").Find(&evaluations).Error
	return evaluations, err
}

// GetRecentTournaments gets recently finished tournaments
func (r *TournamentRepository) GetRecentTournaments(days int, limit int) ([]models.Tournament, error) {
	var tournaments []models.Tournament
	cutoff := time.Now().AddDate(0, 0, -days)
	
	err := r.dbs.Portal64BDW.Where("finishedOn >= ?", cutoff).
		Order("finishedOn DESC").Limit(limit).Find(&tournaments).Error
	
	return tournaments, err
}

// GetParticipantCount gets the number of participants for a tournament
func (r *TournamentRepository) GetParticipantCount(tournamentID uint) (int, error) {
	var count int64
	err := r.dbs.Portal64BDW.Model(&models.Participant{}).
		Where("idTournament = ?", tournamentID).Count(&count).Error
	return int(count), err
}

// GetEnhancedTournamentData gets comprehensive tournament data including participants, games, and evaluations
func (r *TournamentRepository) GetEnhancedTournamentData(tournamentCode string) (*models.EnhancedTournamentResponse, error) {
	// First get the basic tournament info
	var tournament models.Tournament
	err := r.dbs.Portal64BDW.Where("tcode = ?", tournamentCode).First(&tournament).Error
	if err != nil {
		return nil, err
	}

	response := &models.EnhancedTournamentResponse{
		ID:               tournament.TCode,
		Name:             tournament.TName,
		Code:             tournament.TCode,
		Type:             tournament.Type,
		Rounds:           tournament.Rounds,
		FinishedOn:       tournament.FinishedOn,
		ComputedOn:       tournament.ComputedOn,
		RecomputedOn:     tournament.RecomputedOn,
		Status:           r.getTournamentStatus(&tournament),
		Note:             tournament.Note,
	}

	// Get organization information
	if tournament.IDOrganisation > 0 {
		orgInfo, err := r.getOrganizationInfo(tournament.IDOrganisation)
		if err == nil {
			response.Organization = orgInfo
		}
	}

	// Get assessors
	assessors, err := r.getAssessors(tournament.Assessor1, tournament.Assessor2)
	if err == nil {
		response.Assessors = assessors
	}

	// Get participants with detailed information
	participants, err := r.getParticipantsWithDetails(tournament.ID, tournament.FinishedOn)
	if err == nil {
		response.Participants = participants
		response.ParticipantCount = len(participants)
	}

	// Get games and results
	games, err := r.getGamesAndResults(tournament.ID)
	if err == nil {
		response.Games = games
	}

	// Get evaluations if tournament is computed
	if tournament.ComputedOn != nil && tournament.IDMaster > 0 {
		evaluations, err := r.getEvaluations(tournament.IDMaster)
		if err == nil {
			response.Evaluations = evaluations
		}
	}

	return response, nil
}

// getOrganizationInfo retrieves organization details
func (r *TournamentRepository) getOrganizationInfo(orgID uint) (*models.OrganizationInfo, error) {
	var org models.Organisation
	err := r.dbs.MVDSB.Where("id = ?", orgID).First(&org).Error
	if err != nil {
		return nil, err
	}

	return &models.OrganizationInfo{
		ID:        fmt.Sprintf("C%04d", orgID),
		Name:      org.Name,
		ShortName: org.Kurzname,
		VKZ:       org.VKZ,
		Region:    org.Verband,
		District:  org.Bezirk,
	}, nil
}

// getAssessors retrieves assessor information
func (r *TournamentRepository) getAssessors(assessor1ID, assessor2ID uint) ([]models.PersonInfo, error) {
	var assessors []models.PersonInfo
	
	if assessor1ID > 0 {
		var person models.Person
		err := r.dbs.MVDSB.Where("id = ?", assessor1ID).First(&person).Error
		if err == nil {
			assessors = append(assessors, models.PersonInfo{
				ID:        person.ID,
				Name:      person.Name,
				Firstname: person.Vorname,
				FullName:  person.Name + ", " + person.Vorname,
			})
		}
	}

	if assessor2ID > 0 && assessor2ID != assessor1ID {
		var person models.Person
		err := r.dbs.MVDSB.Where("id = ?", assessor2ID).First(&person).Error
		if err == nil {
			assessors = append(assessors, models.PersonInfo{
				ID:        person.ID,
				Name:      person.Name,
				Firstname: person.Vorname,
				FullName:  person.Name + ", " + person.Vorname,
			})
		}
	}

	return assessors, nil
}

// getParticipantsWithDetails retrieves all participants with detailed information
func (r *TournamentRepository) getParticipantsWithDetails(tournamentID uint, finishedOn *time.Time) ([]models.ParticipantInfo, error) {
	// First get participants from portal64_bdw database
	var participants []models.Participant
	err := r.dbs.Portal64BDW.Where("idTournament = ?", tournamentID).Order("no").Find(&participants).Error
	if err != nil {
		return nil, err
	}

	var participantInfos []models.ParticipantInfo
	for _, participant := range participants {
		participantInfo := models.ParticipantInfo{
			ID:       participant.ID,
			No:       participant.No,
			PersonID: participant.IDPerson,
			State:    r.determinePlayerState(participant.IDPerson, participant.IDMembership),
		}

		// Get rating information from the participant record
		if participant.UseRating != nil {
			participantInfo.Rating = &models.RatingInfo{
				UseRating:      participant.UseRating,
				UseRatingIndex: participant.UseRatingIndex,
			}
		}

		// Get person details from mvdsb database if person ID exists
		if participant.IDPerson > 0 {
			var person models.Person
			err := r.dbs.MVDSB.Where("id = ?", participant.IDPerson).First(&person).Error
			if err == nil {
				participantInfo.Name = person.Name
				participantInfo.Firstname = person.Vorname
				participantInfo.FullName = person.Name + ", " + person.Vorname
				participantInfo.BirthYear = utils.ExtractBirthYear(person.Geburtsdatum) // GDPR compliant: only birth year
				participantInfo.Gender = r.getGenderString(person.Geschlecht)
				participantInfo.Nation = person.Nation
				participantInfo.FideID = person.IDFide
			}

			// Get club information if membership ID exists
			if participant.IDMembership > 0 {
				var membership models.Mitgliedschaft
				err := r.dbs.MVDSB.Where("id = ?", participant.IDMembership).First(&membership).Error
				if err == nil {
					var organisation models.Organisation
					err := r.dbs.MVDSB.Where("id = ?", membership.Organisation).First(&organisation).Error
					if err == nil {
						participantInfo.Club = &models.ClubInfo{
							ID:               membership.ID,
							Name:             organisation.Name,
							VKZ:              organisation.VKZ,
							MembershipNumber: int(membership.Spielberechtigung), // Using Spielberechtigung as membership number
						}
					}
				}
			}

			// Get historical rating information
			if participantInfo.Rating == nil {
				participantInfo.Rating = &models.RatingInfo{}
			}
			historicalRating, err := r.getParticipantRating(participant.IDPerson, tournamentID, finishedOn)
			if err == nil && historicalRating != nil {
				if participantInfo.Rating.DWZOld == nil {
					participantInfo.Rating.DWZOld = historicalRating.DWZOld
				}
				if participantInfo.Rating.DWZOldIndex == nil {
					participantInfo.Rating.DWZOldIndex = historicalRating.DWZOldIndex
				}
			}
		}

		participantInfos = append(participantInfos, participantInfo)
	}

	return participantInfos, nil
}

// getParticipantRating retrieves rating information for a participant
func (r *TournamentRepository) getParticipantRating(personID uint, tournamentID uint, finishedOn *time.Time) (*models.RatingInfo, error) {
	// This is a simplified version - the full logic from PHP is quite complex
	// involving historical rating lookups and ELO ratings
	
	rating := &models.RatingInfo{}
	
	// Try to get the participant's use rating first
	var participant models.Participant
	err := r.dbs.Portal64BDW.Where("idTournament = ? AND idPerson = ?", tournamentID, personID).First(&participant).Error
	if err == nil {
		rating.UseRating = participant.UseRating
		rating.UseRatingIndex = participant.UseRatingIndex
	}

	// Get historical DWZ rating (simplified - the real logic is more complex)
	if finishedOn != nil {
		query := `
			SELECT e.dwzNew as dwz_old, e.dwzNewIndex as dwz_old_index
			FROM evaluation e
			INNER JOIN tournamentMaster tm ON e.idMaster = tm.id
			WHERE e.idPerson = ? AND tm.finishedOn < ? AND tm.computedOn IS NOT NULL
			ORDER BY tm.finishedOn DESC, tm.acron DESC
			LIMIT 1
		`
		
		type RatingRow struct {
			DWZOld      *int `db:"dwz_old"`
			DWZOldIndex *int `db:"dwz_old_index"`
		}
		
		var ratingRow RatingRow
		err := r.dbs.Portal64BDW.Raw(query, personID, finishedOn).Scan(&ratingRow).Error
		if err == nil {
			rating.DWZOld = ratingRow.DWZOld
			rating.DWZOldIndex = ratingRow.DWZOldIndex
		}
	}

	return rating, nil
}

// getGamesAndResults retrieves all games and results for a tournament
func (r *TournamentRepository) getGamesAndResults(tournamentID uint) ([]models.RoundInfo, error) {
	// Get appointments (rounds)
	var appointments []models.Appointment
	err := r.dbs.Portal64BDW.Where("idTournament = ?", tournamentID).Order("round").Find(&appointments).Error
	if err != nil {
		return nil, err
	}

	var rounds []models.RoundInfo
	for _, appointment := range appointments {
		round := models.RoundInfo{
			Round:       appointment.Round,
			Appointment: appointment.Appointment,
			Games:       []models.GameInfo{},
		}

		// Get games for this round
		var games []models.Game
		err := r.dbs.Portal64BDW.Where("idAppointment = ?", appointment.ID).Order("board").Find(&games).Error
		if err == nil {
			for _, game := range games {
				gameInfo := models.GameInfo{
					ID:    game.ID,
					Board: game.Board,
				}

				// Get results for this game
				var results []models.Result
				err := r.dbs.Portal64BDW.Where("idGame = ?", game.ID).Find(&results).Error
				if err == nil {
					// Process white and black results
					for _, result := range results {
						playerRef := models.PlayerRef{
							ID: result.IDPerson,
						}

						// Get participant number and person details
						var participant models.Participant
						err := r.dbs.Portal64BDW.Where("idTournament = ? AND idPerson = ?", tournamentID, result.IDPerson).First(&participant).Error
						if err == nil {
							playerRef.No = participant.No
						}

						// Get person name from mvdsb database
						if result.IDPerson > 0 {
							var person models.Person
							err := r.dbs.MVDSB.Where("id = ?", result.IDPerson).First(&person).Error
							if err == nil {
								playerRef.Name = person.Name + ", " + person.Vorname
								playerRef.FullName = playerRef.Name
							}
						}

						if result.Color == "W" {
							gameInfo.White = playerRef
							gameInfo.WhitePoints = result.Points
						} else if result.Color == "B" {
							gameInfo.Black = playerRef
							gameInfo.BlackPoints = result.Points
						}
					}

					// Get result display information
					if game.IDResultsDisplayRating > 0 {
						var resultDisplay models.ResultsDisplay
						err := r.dbs.Portal64BDW.Where("id = ?", game.IDResultsDisplayRating).First(&resultDisplay).Error
						if err == nil {
							gameInfo.Result = resultDisplay.Display
						}
					}
				}

				round.Games = append(round.Games, gameInfo)
			}
		}

		rounds = append(rounds, round)
	}

	return rounds, nil
}

// getEvaluations retrieves evaluation data for a computed tournament
func (r *TournamentRepository) getEvaluations(masterID uint) ([]models.EvaluationInfo, error) {
	var evaluations []models.Evaluation
	err := r.dbs.Portal64BDW.Where("idMaster = ?", masterID).Order("dwzNew DESC").Find(&evaluations).Error
	if err != nil {
		return nil, err
	}

	var evaluationInfos []models.EvaluationInfo
	for _, eval := range evaluations {
		evaluationInfo := models.EvaluationInfo{
			ID:           eval.ID,
			PersonID:     eval.IDPerson,
			ECoefficient: eval.ECoefficient,
			We:           eval.We,
			Achievement:  eval.Achievement,
			Level:        eval.Level,
			Games:        eval.Games,
			UnratedGames: eval.UnratedGames,
			Points:       eval.Points,
			DWZOld:       eval.DWZOld,
			DWZOldIndex:  eval.DWZOldIndex,
			DWZNew:       eval.DWZNew,
			DWZNewIndex:  eval.DWZNewIndex,
		}

		// Get player name from mvdsb database
		if eval.IDPerson > 0 {
			var person models.Person
			err := r.dbs.MVDSB.Where("id = ?", eval.IDPerson).First(&person).Error
			if err == nil {
				evaluationInfo.PlayerName = person.Name + ", " + person.Vorname
			}
		}

		evaluationInfos = append(evaluationInfos, evaluationInfo)
	}

	return evaluationInfos, nil
}

// Helper methods
func (r *TournamentRepository) getTournamentStatus(tournament *models.Tournament) string {
	if tournament.FinishedOn != nil {
		if tournament.ComputedOn != nil {
			return "computed"
		}
		return "finished"
	}
	if tournament.LockedOn != nil {
		return "locked"
	}
	return "active"
}

func (r *TournamentRepository) getGenderString(gender int) string {
	switch gender {
	case 1:
		return "M"
	case 2:
		return "F"
	default:
		return ""
	}
}

func (r *TournamentRepository) determinePlayerState(personID uint, membershipID uint) int {
	// Simplified logic - 0=blocked, 1=unknown, 2=ok
	if personID == 0 {
		return 1 // unknown player
	}
	if membershipID == 0 {
		return 1 // no membership found
	}
	return 2 // ok
}

// GetTournamentCodeByID gets tournament code by its ID
func (r *TournamentRepository) GetTournamentCodeByID(tournamentID uint) (string, error) {
	var tournament models.Tournament
	err := r.dbs.Portal64BDW.Where("id = ?", tournamentID).First(&tournament).Error
	if err != nil {
		return "", err
	}
	return tournament.TCode, nil
}
