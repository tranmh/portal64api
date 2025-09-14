package processor

import (
	"fmt"
	"math"
	"sort"
	"strings"
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
// Phase 3 Implementation: Always fetch ALL German players for accurate percentile calculation
func (p *Processor) ProcessKaderPlanung(checkpoint *resume.Checkpoint, clubPrefix string) ([]models.KaderPlanungRecord, error) {
	overallStartTime := time.Now()

	// Phase 1: Complete German Player Dataset Collection
	// ALWAYS fetch ALL German players regardless of club prefix for accurate percentile calculation
	p.logger.Info("Phase 1: Collecting complete German player dataset for percentile calculation...")
	playerCollectionStart := time.Now()
	
	allGermanPlayers, err := p.fetchAllGermanPlayers()
	if err != nil {
		return nil, fmt.Errorf("failed to collect complete German player dataset: %w", err)
	}
	playerCollectionDuration := time.Since(playerCollectionStart)
	p.logger.Infof("German player dataset collection completed in %v (%d total players)", 
		playerCollectionDuration, len(allGermanPlayers))

	// Phase 2: Germany-wide Percentile Calculation
	p.logger.Info("Phase 2: Calculating Germany-wide somatogram percentiles...")
	percentileCalcStart := time.Now()
	
	percentileMap, err := p.calculateSomatogramPercentiles(allGermanPlayers)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate somatogram percentiles: %w", err)
	}
	percentileCalcDuration := time.Since(percentileCalcStart)
	p.logger.Infof("Percentile calculation completed in %v", percentileCalcDuration)

	// Phase 3: Output Filtering and Record Generation
	p.logger.Info("Phase 3: Generating records with historical analysis and percentiles...")
	recordGenStart := time.Now()
	
	// Filter players by club prefix ONLY for final output (percentiles remain Germany-wide)
	var playersToProcess []models.Player
	if clubPrefix == "" {
		playersToProcess = allGermanPlayers
		p.logger.Infof("No club prefix specified - processing all %d German players", len(allGermanPlayers))
	} else {
		for _, player := range allGermanPlayers {
			if strings.HasPrefix(player.ClubID, clubPrefix) {
				playersToProcess = append(playersToProcess, player)
			}
		}
		p.logger.Infof("Club prefix '%s' specified - processing %d players from %d total", 
			clubPrefix, len(playersToProcess), len(allGermanPlayers))
	}

	if len(playersToProcess) == 0 {
		p.logger.Warn("No players found matching criteria after filtering")
		return []models.KaderPlanungRecord{}, nil
	}

	// Generate records with historical analysis and Germany-wide percentiles
	allRecords, err := p.generateRecordsWithPercentiles(checkpoint, playersToProcess, percentileMap)
	if err != nil {
		return nil, fmt.Errorf("record generation failed: %w", err)
	}
	recordGenDuration := time.Since(recordGenStart)
	p.logger.Infof("Record generation completed in %v (%d records generated)", recordGenDuration, len(allRecords))

	// Phase 4: Calculate rankings for final output (active players only)
	p.logger.Info("Phase 4: Calculating list rankings...")
	rankingStartTime := time.Now()
	
	// Extract players from records for ranking calculation
	players := make([]models.Player, len(allRecords))
	for i, record := range allRecords {
		players[i] = models.Player{
			ID:         record.PlayerID,
			CurrentDWZ: record.CurrentDWZ,
			Status:     "active", // Only include active players in ranking as requested
			Active:     true,     // All players in export are considered active for ranking
		}
	}
	
	// Calculate rankings
	rankings := models.CalculateListRanking(players)
	
	// Update records with rankings
	for i := range allRecords {
		allRecords[i].ListRanking = rankings[i]
	}
	
	rankingDuration := time.Since(rankingStartTime)
	p.logger.Infof("Ranking calculation completed in %v", rankingDuration)

	overallDuration := time.Since(overallStartTime)
	p.logger.Infof("Processing complete: %d total records generated in %v", len(allRecords), overallDuration)
	
	// Performance summary
	p.logger.Infof("Performance Summary:")
	p.logger.Infof("  German Player Collection: %v", playerCollectionDuration)
	p.logger.Infof("  Percentile Calculation: %v", percentileCalcDuration)
	p.logger.Infof("  Record Generation: %v", recordGenDuration)
	p.logger.Infof("  Ranking Calculation: %v", rankingDuration)
	p.logger.Infof("  Total Time: %v", overallDuration)
	
	return allRecords, nil
}
// ========================================
// PHASE 3: SOMATOGRAM INTEGRATION METHODS
// ========================================

// fetchAllGermanPlayers collects ALL German players for accurate percentile calculation
func (p *Processor) fetchAllGermanPlayers() ([]models.Player, error) {
	p.logger.Debug("Fetching all German players using efficient bulk operation...")

	// Use the efficient bulk fetch method to get ALL players (no club prefix filter)
	allPlayers, err := p.apiClient.FetchAllPlayersEfficient("") // Empty prefix = all German clubs
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all German players: %w", err)
	}

	// Filter for valid players for statistical analysis (like somatogram does)
	validPlayers := p.apiClient.FilterValidPlayersForStatistics(allPlayers)

	p.logger.Infof("Collected %d valid German players from %d total for somatogram analysis", 
		len(validPlayers), len(allPlayers))

	return validPlayers, nil
}

// calculateSomatogramPercentiles calculates Germany-wide percentiles using somatogram logic
func (p *Processor) calculateSomatogramPercentiles(allPlayers []models.Player) (map[string]float64, error) {
	p.logger.Debug("Calculating Germany-wide somatogram percentiles...")

	// Group players by age and gender (adapted from somatogram processor)
	ageGenderGroups := p.groupPlayersByAgeAndGender(allPlayers)
	p.logger.Debugf("Created %d age-gender groups for percentile calculation", len(ageGenderGroups))

	// Filter groups by minimum sample size (default from config or 50)
	minSampleSize := 50 // Reasonable default for percentile accuracy
	validGroups := p.filterGroupsBySampleSize(ageGenderGroups, minSampleSize)
	p.logger.Debugf("%d groups meet minimum sample size requirement (%d)", len(validGroups), minSampleSize)

	// Create percentile lookup map: "playerID" -> percentile value
	percentileMap := make(map[string]float64)

	// Process each valid group
	for _, group := range validGroups {
		groupPercentiles := p.calculatePercentilesForGroup(group.Players)
		
		// Assign percentile to each player in the group
		for _, player := range group.Players {
			playerPercentile := p.findPercentileForPlayer(player, groupPercentiles)
			percentileMap[player.ID] = playerPercentile
		}

		p.logger.Debugf("Calculated percentiles for age %d, gender %s (%d players)", 
			group.Age, group.Gender, len(group.Players))
	}

	p.logger.Infof("Generated percentile data for %d players from %d valid age-gender groups", 
		len(percentileMap), len(validGroups))

	return percentileMap, nil
}

// generateRecordsWithPercentiles generates kader-planung records with historical analysis and percentiles
func (p *Processor) generateRecordsWithPercentiles(checkpoint *resume.Checkpoint, players []models.Player, percentileMap map[string]float64) ([]models.KaderPlanungRecord, error) {
	p.logger.Debug("Generating records with historical analysis and percentiles...")

	var allRecords []models.KaderPlanungRecord

	// Process players concurrently for historical analysis
	playerChan := make(chan models.Player, len(players))
	resultChan := make(chan *models.KaderPlanungRecord, len(players))
	errorChan := make(chan error, len(players))

	var wg sync.WaitGroup

	// Start worker pool for concurrent processing
	for i := 0; i < p.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for player := range playerChan {
				record, err := p.generateRecordWithPercentile(checkpoint, player, percentileMap)
				if err != nil {
					p.logger.Warnf("Worker %d: Failed to process player %s: %v", workerID, player.ID, err)
					errorChan <- err
					resultChan <- nil
					continue
				}
				resultChan <- record
			}
		}(i)
	}

	// Send players to workers
	for _, player := range players {
		playerChan <- player
	}
	close(playerChan)

	// Wait for workers and close channels
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	processedCount := 0
	errorCount := 0

	for result := range resultChan {
		processedCount++
		if result != nil {
			allRecords = append(allRecords, *result)
		} else {
			errorCount++
		}
	}

	// Log any errors
	for err := range errorChan {
		if err != nil {
			p.logger.Debugf("Error during record generation: %v", err)
		}
	}

	p.logger.Infof("Record generation completed: %d records generated, %d errors", 
		len(allRecords), errorCount)

	return allRecords, nil
}

// ========================================
// SOMATOGRAM CALCULATION HELPERS
// ========================================

// AgeGenderGroup represents a group of players by age and gender (adapted from somatogram)
type AgeGenderGroup struct {
	Age        int
	Gender     string
	Players    []models.Player
	SampleSize int
	AvgDWZ     float64
	MedianDWZ  int
}

// groupPlayersByAgeAndGender groups players by age and gender (adapted from somatogram)
func (p *Processor) groupPlayersByAgeAndGender(players []models.Player) map[string]AgeGenderGroup {
	groups := make(map[string]AgeGenderGroup)
	currentYear := time.Now().Year()

	for _, player := range players {
		if player.BirthYear == nil || player.Gender == "" {
			continue
		}

		age := currentYear - *player.BirthYear
		key := fmt.Sprintf("%s-%d", player.Gender, age)

		group, exists := groups[key]
		if !exists {
			group = AgeGenderGroup{
				Age:     age,
				Gender:  player.Gender,
				Players: []models.Player{},
			}
		}

		group.Players = append(group.Players, player)
		groups[key] = group
	}

	// Calculate statistics for each group
	for key, group := range groups {
		group.SampleSize = len(group.Players)
		group.AvgDWZ = p.calculateAverageDWZ(group.Players)
		group.MedianDWZ = p.calculateMedianDWZ(group.Players)
		groups[key] = group
	}

	return groups
}

// filterGroupsBySampleSize filters groups by minimum sample size (adapted from somatogram)
func (p *Processor) filterGroupsBySampleSize(groups map[string]AgeGenderGroup, minSampleSize int) []AgeGenderGroup {
	var validGroups []AgeGenderGroup

	for _, group := range groups {
		if group.SampleSize >= minSampleSize {
			validGroups = append(validGroups, group)
		} else {
			p.logger.Debugf("Skipping age %d, gender %s: only %d players (minimum: %d)",
				group.Age, group.Gender, group.SampleSize, minSampleSize)
		}
	}

	return validGroups
}

// calculatePercentilesForGroup calculates percentiles for a group of players (adapted from somatogram)
func (p *Processor) calculatePercentilesForGroup(players []models.Player) map[int]int {
	if len(players) == 0 {
		return make(map[int]int)
	}

	// Extract DWZ values and sort
	dwzValues := make([]int, len(players))
	for i, player := range players {
		dwzValues[i] = player.CurrentDWZ
	}
	sort.Ints(dwzValues)

	// Calculate percentiles (0-100)
	n := len(dwzValues)
	percentiles := make(map[int]int)

	for p := 0; p <= 100; p++ {
		if p == 0 {
			percentiles[p] = dwzValues[0]
		} else if p == 100 {
			percentiles[p] = dwzValues[n-1]
		} else {
			rank := float64(p) * float64(n-1) / 100.0
			lower := int(rank)
			upper := lower + 1

			if upper >= n {
				percentiles[p] = dwzValues[n-1]
			} else {
				weight := rank - float64(lower)
				percentiles[p] = int(float64(dwzValues[lower])*(1-weight) + float64(dwzValues[upper])*weight)
			}
		}
	}

	return percentiles
}

// findPercentileForPlayer finds the percentile rank for a specific player's DWZ
func (p *Processor) findPercentileForPlayer(player models.Player, groupPercentiles map[int]int) float64 {
	playerDWZ := player.CurrentDWZ

	// Find the percentile rank for this player's DWZ
	for percentile := 0; percentile <= 100; percentile++ {
		if dwzThreshold, exists := groupPercentiles[percentile]; exists {
			if playerDWZ <= dwzThreshold {
				return float64(percentile)
			}
		}
	}

	// If player's DWZ is above 100th percentile, return 100
	return 100.0
}

// generateRecordWithPercentile creates a record with historical analysis and somatogram percentile
func (p *Processor) generateRecordWithPercentile(checkpoint *resume.Checkpoint, player models.Player, percentileMap map[string]float64) (*models.KaderPlanungRecord, error) {
	// Get club information
	club := models.Club{
		ID:   player.ClubID,
		Name: player.Club,
	}

	// Check if we have partial data for this player
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
			p.logger.Warnf("Failed to fetch rating history for player %s: %v", player.ID, err)
			// Continue without history - we'll mark missing data appropriately
		} else if history == nil {
			p.logger.Warnf("Rating history is nil for player %s", player.ID)
		} else {
			p.logger.Debugf("Successfully fetched rating history for player %s: %d points", player.ID, len(history.Points))
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
		analysis = p.AnalyzeHistoricalData(history)
		
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

	// Create record with percentile data
	record := p.createKaderPlanungRecord(club, player, analysis)
	
	// Add somatogram percentile from Germany-wide calculation
	if percentile, exists := percentileMap[player.ID]; exists {
		record.SomatogramPercentile = strings.Replace(fmt.Sprintf("%.1f", percentile), ".", ",", 1)
	} else {
		record.SomatogramPercentile = models.DataNotAvailable
	}

	return record, nil
}

// Helper methods adapted from somatogram processor

// calculateAverageDWZ calculates average DWZ for a group of players
func (p *Processor) calculateAverageDWZ(players []models.Player) float64 {
	if len(players) == 0 {
		return 0
	}

	total := 0
	for _, player := range players {
		total += player.CurrentDWZ
	}

	return float64(total) / float64(len(players))
}

// calculateMedianDWZ calculates median DWZ for a group of players
func (p *Processor) calculateMedianDWZ(players []models.Player) int {
	if len(players) == 0 {
		return 0
	}

	dwzValues := make([]int, len(players))
	for i, player := range players {
		dwzValues[i] = player.CurrentDWZ
	}

	sort.Ints(dwzValues)
	n := len(dwzValues)

	if n%2 == 1 {
		return dwzValues[n/2]
	} else {
		return (dwzValues[n/2-1] + dwzValues[n/2]) / 2
	}
}

// ========================================
// EXISTING METHODS (LEGACY SUPPORT)
// ========================================

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
		analysis = p.AnalyzeHistoricalData(history)
		
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

// AnalyzeHistoricalData performs historical analysis on a player's rating history
func (p *Processor) AnalyzeHistoricalData(history *models.RatingHistory) *models.HistoricalAnalysis {
	startTime := time.Now()
	
	analysis := &models.HistoricalAnalysis{
		DWZ12MonthsAgo:          models.DataNotAvailable,
		GamesLast12Months:       0,
		SuccessRateLast12Months: 0,
		HasHistoricalData:       false,
	}

	if history == nil {
		p.logger.Debugf("Historical analysis: history is nil")
		return analysis
	}
	
	if len(history.Points) == 0 {
		p.logger.Debugf("Historical analysis: history.Points is empty")
		return analysis
	}
	
	p.logger.Debugf("Historical analysis: processing %d rating points", len(history.Points))

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

	duration := time.Since(startTime)
	p.logger.Debugf("Historical analysis completed in %v for player %s (HasData: %t, Games: %d)", 
		duration, "unknown", analysis.HasHistoricalData, analysis.GamesLast12Months)
		
	return analysis
}
// createKaderPlanungRecord creates a final record for export
func (p *Processor) createKaderPlanungRecord(club models.Club, player models.Player, analysis *models.HistoricalAnalysis) *models.KaderPlanungRecord {
	// Handle birthyear safely (API returns pointer)
	birthyear := 0
	if player.BirthYear != nil {
		birthyear = *player.BirthYear
	}
	
	// Calculate club_id prefixes
	prefix1, prefix2, prefix3 := models.CalculateClubIDPrefixes(club.ID)
	
	record := &models.KaderPlanungRecord{
		ClubIDPrefix1:           prefix1,
		ClubIDPrefix2:           prefix2,
		ClubIDPrefix3:           prefix3,
		ClubName:                club.Name,
		ClubID:                  club.ID,
		PlayerID:                player.ID,
		PKZ:                     player.PKZ,         // NEW: Player identification number
		Lastname:                player.Name,        // API returns last name in "name" field
		Firstname:               player.Firstname,
		Birthyear:               birthyear,
		Gender:                  player.Gender,      // NEW: 'm', 'w', 'd' for man/woman/divers
		CurrentDWZ:              player.CurrentDWZ,
		ListRanking:             0,                  // NEW: Will be calculated after all players are processed
		DWZ12MonthsAgo:          models.DataNotAvailable,
		GamesLast12Months:       models.DataNotAvailable,
		SuccessRateLast12Months: models.DataNotAvailable,
		SomatogramPercentile:    models.DataNotAvailable,  // NEW: Somatogram percentile (Phase 2 - placeholder)
		DWZAgeRelation:          models.CalculateDWZAgeRelation(player.CurrentDWZ, birthyear, time.Now().Year()), // NEW: Calculate age relation
	}

	if analysis != nil && analysis.HasHistoricalData {
		p.logger.Debugf("Player %s: applying historical analysis - DWZ12MonthsAgo=%s, Games=%d, Success=%.1f",
			player.ID, analysis.DWZ12MonthsAgo, analysis.GamesLast12Months, analysis.SuccessRateLast12Months)
		record.DWZ12MonthsAgo = analysis.DWZ12MonthsAgo

		// Always record games count, even if 0
		record.GamesLast12Months = fmt.Sprintf("%d", analysis.GamesLast12Months)

		if analysis.GamesLast12Months > 0 {
			record.SuccessRateLast12Months = strings.Replace(fmt.Sprintf("%.1f", analysis.SuccessRateLast12Months), ".", ",", 1)
		} else {
			// If no games, success rate cannot be calculated
			record.SuccessRateLast12Months = "0,0"
		}
	} else {
		if analysis == nil {
			p.logger.Debugf("Player %s: analysis is nil", player.ID)
		} else if !analysis.HasHistoricalData {
			p.logger.Debugf("Player %s: analysis.HasHistoricalData is false", player.ID)
		}
	}

	return record
}

// SetLogger sets a custom logger for the processor
func (p *Processor) SetLogger(logger *logrus.Logger) {
	p.logger = logger
}