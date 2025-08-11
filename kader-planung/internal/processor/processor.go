package processor

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/portal64/kader-planung/internal/api"
	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/resume"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

// Processor handles the main data processing workflow
type Processor struct {
	apiClient     *api.Client
	resumeManager *resume.Manager
	concurrency   int
	logger        *logrus.Logger
}

// New creates a new processor instance
func New(apiClient *api.Client, resumeManager *resume.Manager, concurrency int) *Processor {
	return &Processor{
		apiClient:     apiClient,
		resumeManager: resumeManager,
		concurrency:   concurrency,
		logger:        logrus.StandardLogger(),
	}
}
// ProcessKaderPlanung runs the complete Kader-Planung data collection process
func (p *Processor) ProcessKaderPlanung(checkpoint *resume.Checkpoint, clubPrefix string) ([]models.KaderPlanungRecord, error) {
	var allRecords []models.KaderPlanungRecord

	// Phase 1: Club Discovery
	p.logger.Info("Phase 1: Discovering clubs...")
	clubs, err := p.discoverClubs(checkpoint, clubPrefix)
	if err != nil {
		return nil, fmt.Errorf("club discovery failed: %w", err)
	}

	if len(clubs) == 0 {
		p.logger.Warn("No clubs found matching criteria")
		return allRecords, nil
	}

	// Update progress
	p.resumeManager.UpdateProgress(checkpoint, len(clubs), 0, "player_processing")
	if err := p.resumeManager.SaveCheckpoint(checkpoint); err != nil {
		p.logger.Warnf("Failed to save checkpoint: %v", err)
	}

	// Phase 2: Player Data Collection and Processing
	p.logger.Infof("Phase 2: Processing %d clubs with concurrency %d...", len(clubs), p.concurrency)
	records, err := p.processClubsConcurrently(checkpoint, clubs)
	if err != nil {
		return nil, fmt.Errorf("player processing failed: %w", err)
	}

	allRecords = append(allRecords, records...)

	p.logger.Infof("Processing complete: %d total records generated", len(allRecords))
	return allRecords, nil
}
// discoverClubs fetches all clubs matching the criteria
func (p *Processor) discoverClubs(checkpoint *resume.Checkpoint, clubPrefix string) ([]models.Club, error) {
	clubs, err := p.apiClient.GetAllClubs(clubPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clubs: %w", err)
	}

	prefixStr := clubPrefix
	if prefixStr == "" {
		prefixStr = "all"
	}
	p.logger.Infof("Found %d clubs matching prefix '%s'", len(clubs), prefixStr)

	return clubs, nil
}

// processClubsConcurrently processes all clubs using a worker pool
func (p *Processor) processClubsConcurrently(checkpoint *resume.Checkpoint, clubs []models.Club) ([]models.KaderPlanungRecord, error) {
	// Create channels for job distribution
	clubJobs := make(chan models.Club, len(clubs))
	results := make(chan []models.KaderPlanungRecord, len(clubs))
	errors := make(chan error, len(clubs))

	// Create progress bar
	bar := progressbar.NewOptions(len(clubs),
		progressbar.OptionSetDescription("Processing clubs"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(65*time.Millisecond),
	)
	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < p.concurrency; i++ {
		wg.Add(1)
		go p.clubWorker(&wg, clubJobs, results, errors, checkpoint, bar)
	}

	// Queue jobs, skipping already processed clubs
	processedClubs := p.resumeManager.GetProcessedClubs(checkpoint)
	processedMap := make(map[string]bool)
	for _, clubID := range processedClubs {
		processedMap[clubID] = true
	}

	queuedJobs := 0
	for _, club := range clubs {
		if !processedMap[club.ID] {
			clubJobs <- club
			queuedJobs++
		} else {
			bar.Add(1) // Count already processed clubs
			p.logger.Debugf("Skipping already processed club: %s", club.ID)
		}
	}
	close(clubJobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)
	close(errors)

	// Collect results and errors
	var allRecords []models.KaderPlanungRecord
	for records := range results {
		allRecords = append(allRecords, records...)
	}
	// Log any errors
	errorCount := 0
	for err := range errors {
		p.logger.Errorf("Processing error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		p.logger.Warnf("Completed with %d errors", errorCount)
	}

	bar.Finish()
	return allRecords, nil
}

// clubWorker processes individual clubs
func (p *Processor) clubWorker(wg *sync.WaitGroup, jobs <-chan models.Club, results chan<- []models.KaderPlanungRecord, errors chan<- error, checkpoint *resume.Checkpoint, bar *progressbar.ProgressBar) {
	defer wg.Done()

	for club := range jobs {
		records, err := p.processClub(checkpoint, club)
		if err != nil {
			errors <- fmt.Errorf("failed to process club %s (%s): %w", club.ID, club.Name, err)
			// Mark club as failed but continue processing
			p.resumeManager.MarkProcessed(checkpoint, "club", club.ID, "failed")
		} else {
			results <- records
			// Mark club as completed
			p.resumeManager.MarkProcessed(checkpoint, "club", club.ID, "completed")
		}

		// Update progress
		checkpoint.Progress.ProcessedClubs++
		if checkpoint.Progress.ProcessedClubs%10 == 0 {
			if err := p.resumeManager.SaveCheckpoint(checkpoint); err != nil {
				p.logger.Warnf("Failed to save checkpoint: %v", err)
			}
		}

		bar.Add(1)
	}
}
// processClub processes all players for a single club
func (p *Processor) processClub(checkpoint *resume.Checkpoint, club models.Club) ([]models.KaderPlanungRecord, error) {
	// Fetch all players for the club
	players, err := p.apiClient.GetAllClubPlayers(club.ID, false) // Include inactive players
	if err != nil {
		return nil, fmt.Errorf("failed to fetch players for club %s: %w", club.ID, err)
	}

	p.logger.Debugf("Processing %d players for club %s (%s)", len(players), club.ID, club.Name)

	var records []models.KaderPlanungRecord
	processedPlayers := p.resumeManager.GetProcessedPlayers(checkpoint)
	processedMap := make(map[string]bool)
	for _, playerID := range processedPlayers {
		processedMap[playerID] = true
	}

	// Process each player
	for _, player := range players {
		if processedMap[player.ID] {
			p.logger.Debugf("Skipping already processed player: %s", player.ID)
			
			// Try to get the record from partial data
			if partialData := p.resumeManager.GetPartialData(checkpoint, player.ID); partialData != nil {
				if record := p.createKaderPlanungRecord(club, partialData.Player, partialData.Analysis); record != nil {
					records = append(records, *record)
				}
			}
			continue
		}

		record, err := p.processPlayer(checkpoint, club, player)
		if err != nil {
			p.logger.Warnf("Failed to process player %s: %v", player.ID, err)
			p.resumeManager.MarkProcessed(checkpoint, "player", player.ID, "failed")
			continue
		}
		if record != nil {
			records = append(records, *record)
		}

		p.resumeManager.MarkProcessed(checkpoint, "player", player.ID, "completed")
		p.resumeManager.RemovePartialData(checkpoint, player.ID)
	}

	return records, nil
}

// processPlayer processes a single player and generates their record
func (p *Processor) processPlayer(checkpoint *resume.Checkpoint, club models.Club, player models.Player) (*models.KaderPlanungRecord, error) {
	// Check if we have partial data
	partialData := p.resumeManager.GetPartialData(checkpoint, player.ID)
	
	var history *models.RatingHistory
	var err error

	if partialData != nil && partialData.History != nil {
		// Use cached history
		history = partialData.History
	} else {
		// Fetch rating history
		history, err = p.apiClient.GetPlayerRatingHistory(player.ID)
		if err != nil {
			p.logger.Debugf("Could not fetch rating history for player %s: %v", player.ID, err)
			// Continue without history - we'll mark missing data appropriately
		}

		// Save partial data
		p.resumeManager.AddPartialData(checkpoint, resume.PartialPlayerData{
			ClubID:   club.ID,
			ClubName: club.Name,
			Player:   player,
			History:  history,
		})
	}
	// Analyze historical data
	var analysis *models.HistoricalAnalysis
	if partialData != nil && partialData.Analysis != nil {
		analysis = partialData.Analysis
	} else {
		analysis = p.analyzeHistoricalData(history)
		
		// Update partial data with analysis
		if partialData == nil {
			partialData = &resume.PartialPlayerData{
				ClubID:   club.ID,
				ClubName: club.Name,
				Player:   player,
				History:  history,
			}
		}
		partialData.Analysis = analysis
		p.resumeManager.AddPartialData(checkpoint, *partialData)
	}

	return p.createKaderPlanungRecord(club, player, analysis), nil
}

// analyzeHistoricalData performs historical analysis on a player's rating history
func (p *Processor) analyzeHistoricalData(history *models.RatingHistory) *models.HistoricalAnalysis {
	analysis := &models.HistoricalAnalysis{
		DWZ12MonthsAgo:          models.DataNotAvailable,
		GamesLast12Months:       0,
		SuccessRateLast12Months: 0,
		HasHistoricalData:       false,
	}

	if history == nil || len(history.Points) == 0 {
		return analysis
	}

	analysis.HasHistoricalData = true
	target12MonthsAgo := time.Now().AddDate(0, -12, 0)
	cutoff12MonthsAgo := time.Now().AddDate(0, -12, 0)
	// Find DWZ 12 months ago (closest point before target date)
	var closestPoint *models.RatingPoint
	var minDiff time.Duration = time.Duration(math.MaxInt64)

	for i := range history.Points {
		point := &history.Points[i]
		if point.Date.Before(target12MonthsAgo) {
			diff := target12MonthsAgo.Sub(point.Date)
			if diff < minDiff {
				closestPoint = point
				minDiff = diff
			}
		}
	}

	if closestPoint != nil {
		analysis.DWZ12MonthsAgo = fmt.Sprintf("%d", closestPoint.DWZ)
	}

	// Calculate games and success rate in last 12 months
	var totalGames int
	var totalPoints float64

	for _, point := range history.Points {
		if point.Date.After(cutoff12MonthsAgo) {
			totalGames += point.Games
			totalPoints += point.Points
		}
	}

	analysis.GamesLast12Months = totalGames

	if totalGames > 0 {
		// Success rate = points / games * 100
		successRate := totalPoints / float64(totalGames) * 100
		analysis.SuccessRateLast12Months = successRate
	}

	return analysis
}
// createKaderPlanungRecord creates a final record for export
func (p *Processor) createKaderPlanungRecord(club models.Club, player models.Player, analysis *models.HistoricalAnalysis) *models.KaderPlanungRecord {
	// Handle birthyear safely (API returns pointer)
	birthyear := 0
	if player.BirthYear != nil {
		birthyear = *player.BirthYear
	}
	
	record := &models.KaderPlanungRecord{
		ClubName:                club.Name,
		ClubID:                  club.ID,
		PlayerID:                player.ID,
		Lastname:                player.Name,        // API returns last name in "name" field
		Firstname:               player.Firstname,
		Birthyear:               birthyear,
		CurrentDWZ:              player.CurrentDWZ,
		DWZ12MonthsAgo:          models.DataNotAvailable,
		GamesLast12Months:       models.DataNotAvailable,
		SuccessRateLast12Months: models.DataNotAvailable,
	}

	if analysis != nil && analysis.HasHistoricalData {
		record.DWZ12MonthsAgo = analysis.DWZ12MonthsAgo
		
		if analysis.GamesLast12Months > 0 {
			record.GamesLast12Months = fmt.Sprintf("%d", analysis.GamesLast12Months)
			record.SuccessRateLast12Months = fmt.Sprintf("%.1f", analysis.SuccessRateLast12Months)
		}
	}

	return record
}

// SetLogger sets a custom logger for the processor
func (p *Processor) SetLogger(logger *logrus.Logger) {
	p.logger = logger
}