package processor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/portal64/kader-planung/internal/api"
	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/resume"
	"github.com/portal64/kader-planung/internal/statistics"
	"github.com/sirupsen/logrus"
)

// ProcessingMode defines the different processing modes available
type ProcessingMode int

const (
	// EfficientMode uses Somatogramm-style processing: minimal API calls, statistical focus
	EfficientMode ProcessingMode = iota
	// DetailedMode uses traditional Kader-Planung processing: comprehensive individual analysis
	DetailedMode
	// StatisticalMode focuses purely on statistical analysis (like original Somatogramm)
	StatisticalMode
	// HybridMode combines efficient bulk fetching with selective detailed analysis
	HybridMode
)

// String returns the string representation of a ProcessingMode
func (p ProcessingMode) String() string {
	switch p {
	case EfficientMode:
		return "efficient"
	case DetailedMode:
		return "detailed"
	case StatisticalMode:
		return "statistical"
	case HybridMode:
		return "hybrid"
	default:
		return "unknown"
	}
}

// ParseProcessingMode parses a string into a ProcessingMode
func ParseProcessingMode(mode string) (ProcessingMode, error) {
	switch mode {
	case "efficient":
		return EfficientMode, nil
	case "detailed":
		return DetailedMode, nil
	case "statistical":
		return StatisticalMode, nil
	case "hybrid":
		return HybridMode, nil
	default:
		return DetailedMode, fmt.Errorf("unknown processing mode: %s", mode)
	}
}

// UnifiedProcessorConfig holds configuration for the unified processor
type UnifiedProcessorConfig struct {
	Mode            ProcessingMode
	MinSampleSize   int  // For statistical analysis
	EnableStatistics bool // Whether to include statistics in output
	ClubPrefix      string
	Concurrency     int
	Verbose         bool
}

// UnifiedProcessor handles multiple processing modes for Kader-Planung
type UnifiedProcessor struct {
	apiClient         *api.Client
	resumeManager     *resume.Manager
	statisticalAnalyzer *statistics.StatisticalAnalyzer
	legacyProcessor   *Processor // For backward compatibility
	config            *UnifiedProcessorConfig
	logger            *logrus.Logger
}

// NewUnifiedProcessor creates a new unified processor
func NewUnifiedProcessor(
	apiClient *api.Client,
	resumeManager *resume.Manager,
	config *UnifiedProcessorConfig,
) *UnifiedProcessor {
	logger := logrus.StandardLogger()

	// Create statistical analyzer
	statisticalAnalyzer := statistics.NewStatisticalAnalyzer(config.MinSampleSize, logger)

	// Create legacy processor for backward compatibility
	legacyProcessor := New(apiClient, resumeManager, config.Concurrency)
	legacyProcessor.SetLogger(logger)

	return &UnifiedProcessor{
		apiClient:           apiClient,
		resumeManager:       resumeManager,
		statisticalAnalyzer: statisticalAnalyzer,
		legacyProcessor:     legacyProcessor,
		config:              config,
		logger:              logger,
	}
}

// ProcessData executes the appropriate processing mode
func (p *UnifiedProcessor) ProcessData(checkpoint *resume.Checkpoint) (*ProcessingResult, error) {
	startTime := time.Now()

	p.logger.Infof("Starting unified processing in %s mode", p.config.Mode.String())

	var result *ProcessingResult
	var err error

	switch p.config.Mode {
	case EfficientMode:
		result, err = p.processEfficientMode(checkpoint)
	case DetailedMode:
		result, err = p.processDetailedMode(checkpoint)
	case StatisticalMode:
		result, err = p.processStatisticalMode()
	case HybridMode:
		result, err = p.processHybridMode(checkpoint)
	default:
		return nil, fmt.Errorf("unsupported processing mode: %s", p.config.Mode.String())
	}

	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	result.ProcessingTime = duration

	p.logger.Infof("Processing completed in %v using %s mode", duration, p.config.Mode.String())
	return result, nil
}

// processEfficientMode uses efficient bulk processing for basic analysis
func (p *UnifiedProcessor) processEfficientMode(checkpoint *resume.Checkpoint) (*ProcessingResult, error) {
	p.logger.Info("Using efficient mode: bulk player fetch with minimal API calls")

	// Use the efficient bulk fetch method
	players, err := p.apiClient.FetchAllPlayersEfficient(p.config.ClubPrefix)
	if err != nil {
		return nil, fmt.Errorf("efficient bulk fetch failed: %w", err)
	}

	p.logger.Infof("Efficiently fetched %d players", len(players))

	// Convert to KaderPlanungRecords (without historical analysis to save time)
	records := p.convertPlayersToRecords(players, false)

	// Calculate rankings
	p.calculateListRankings(records)

	result := &ProcessingResult{
		Mode:           p.config.Mode,
		Records:        records,
		TotalPlayers:   len(players),
		ProcessedClubs: p.countUniqueClubs(players),
	}

	// Add statistics if enabled
	if p.config.EnableStatistics {
		validPlayers := p.apiClient.FilterValidPlayersForStatistics(players)
		statisticalData, err := p.statisticalAnalyzer.ProcessPlayers(validPlayers)
		if err != nil {
			p.logger.Warnf("Statistical analysis failed: %v", err)
		} else {
			result.StatisticalData = statisticalData
		}
	}

	return result, nil
}

// processDetailedMode uses the traditional Kader-Planung processing
func (p *UnifiedProcessor) processDetailedMode(checkpoint *resume.Checkpoint) (*ProcessingResult, error) {
	p.logger.Info("Using detailed mode: traditional Kader-Planung processing with full historical analysis")

	// Use the legacy processor for detailed analysis
	records, err := p.legacyProcessor.ProcessKaderPlanung(checkpoint, p.config.ClubPrefix)
	if err != nil {
		return nil, fmt.Errorf("detailed processing failed: %w", err)
	}

	result := &ProcessingResult{
		Mode:         p.config.Mode,
		Records:      records,
		TotalPlayers: len(records),
		// ProcessedClubs will be set by the legacy processor
	}

	return result, nil
}

// processStatisticalMode focuses purely on statistical analysis (like original Somatogramm)
func (p *UnifiedProcessor) processStatisticalMode() (*ProcessingResult, error) {
	p.logger.Info("Using statistical mode: pure statistical analysis without individual records")

	// Use efficient bulk fetch
	players, err := p.apiClient.FetchAllPlayersEfficient(p.config.ClubPrefix)
	if err != nil {
		return nil, fmt.Errorf("bulk fetch for statistics failed: %w", err)
	}

	// Filter valid players for statistics
	validPlayers := p.apiClient.FilterValidPlayersForStatistics(players)
	p.logger.Infof("Using %d valid players from %d total for statistical analysis", len(validPlayers), len(players))

	// Perform statistical analysis
	statisticalData, err := p.statisticalAnalyzer.ProcessPlayers(validPlayers)
	if err != nil {
		return nil, fmt.Errorf("statistical analysis failed: %w", err)
	}

	result := &ProcessingResult{
		Mode:            p.config.Mode,
		Records:         nil, // No individual records in statistical mode
		TotalPlayers:    len(validPlayers),
		ProcessedClubs:  p.countUniqueClubs(validPlayers),
		StatisticalData: statisticalData,
	}

	return result, nil
}

// processHybridMode combines efficient bulk fetching with selective detailed analysis
// PHASE 3 IMPLEMENTATION: Always fetch ALL German players for accurate percentile calculation
func (p *UnifiedProcessor) processHybridMode(checkpoint *resume.Checkpoint) (*ProcessingResult, error) {
	p.logger.Info("Using hybrid mode: efficient bulk fetch + selective detailed analysis")

	// Phase 1: Complete German Player Dataset Collection
	// ALWAYS fetch ALL German players regardless of club prefix for accurate percentile calculation
	p.logger.Info("Phase 1: Collecting complete German player dataset for percentile calculation...")
	allGermanPlayers, err := p.apiClient.FetchAllPlayersEfficient("") // Empty prefix = ALL German players
	if err != nil {
		return nil, fmt.Errorf("failed to collect complete German player dataset: %w", err)
	}
	p.logger.Infof("German player dataset collection completed (%d total players)", len(allGermanPlayers))

	// Phase 2: Germany-wide Percentile Calculation
	p.logger.Info("Phase 2: Calculating Germany-wide somatogram percentiles...")
	var statisticalData map[string]statistics.SomatogrammData
	if p.config.EnableStatistics {
		validGermanPlayers := p.apiClient.FilterValidPlayersForStatistics(allGermanPlayers)
		p.logger.Infof("Using %d valid players from %d total for Germany-wide percentile calculation", 
			len(validGermanPlayers), len(allGermanPlayers))
		
		statisticalData, err = p.statisticalAnalyzer.ProcessPlayers(validGermanPlayers)
		if err != nil {
			p.logger.Warnf("Statistical analysis in hybrid mode failed: %v", err)
		} else {
			p.logger.Info("Germany-wide percentile calculation completed")
		}
	}

	// Phase 3: Output Filtering and Record Generation
	p.logger.Info("Phase 3: Generating records with Germany-wide percentiles...")
	
	// Filter players by club prefix ONLY for final output (percentiles remain Germany-wide)
	var playersToProcess []models.Player
	if p.config.ClubPrefix == "" {
		playersToProcess = allGermanPlayers
		p.logger.Infof("No club prefix specified - processing all %d German players", len(allGermanPlayers))
	} else {
		for _, player := range allGermanPlayers {
			if len(player.ClubID) >= len(p.config.ClubPrefix) && 
			   player.ClubID[:len(p.config.ClubPrefix)] == p.config.ClubPrefix {
				playersToProcess = append(playersToProcess, player)
			}
		}
		p.logger.Infof("Club prefix '%s' specified - processing %d players from %d total", 
			p.config.ClubPrefix, len(playersToProcess), len(allGermanPlayers))
	}

	if len(playersToProcess) == 0 {
		p.logger.Warn("No players found matching criteria after filtering")
		return &ProcessingResult{
			Mode:            p.config.Mode,
			Records:         []models.KaderPlanungRecord{},
			StatisticalData: statisticalData,
			TotalPlayers:    0,
			ProcessedClubs:  0,
		}, nil
	}

	// Convert filtered players to records with historical analysis
	records := p.convertPlayersToRecords(playersToProcess, true)

	// Populate somatogram percentiles from statistical analysis
	if statisticalData != nil {
		p.populateSomatogramPercentiles(records, statisticalData)
	}

	// Calculate rankings for final output players only
	p.calculateListRankings(records)

	result := &ProcessingResult{
		Mode:            p.config.Mode,
		Records:         records,
		StatisticalData: statisticalData, // Germany-wide percentiles
		TotalPlayers:    len(playersToProcess), // Final filtered count
		ProcessedClubs:  p.countUniqueClubs(playersToProcess),
	}

	p.logger.Infof("Phase 3 completed: %d records generated with Germany-wide percentiles", len(records))
	return result, nil
}

// convertPlayersToRecords converts Player objects to KaderPlanungRecord objects
func (p *UnifiedProcessor) convertPlayersToRecords(players []models.Player, includeHistoricalAnalysis bool) []models.KaderPlanungRecord {
	var records []models.KaderPlanungRecord
	currentYear := time.Now().Year()
	targetDates := models.GenerateTargetDates()

	for _, player := range players {
		// Handle birthyear safely
		birthyear := 0
		if player.BirthYear != nil {
			birthyear = *player.BirthYear
		}

		// Create a minimal club object for prefix calculation
		club := models.Club{
			ID:   player.ClubID,
			Name: player.Club,
		}

		// Calculate club_id prefixes
		prefix1, prefix2, prefix3 := models.CalculateClubIDPrefixes(club.ID)

		record := models.KaderPlanungRecord{
			ClubIDPrefix1:           prefix1,
			ClubIDPrefix2:           prefix2,
			ClubIDPrefix3:           prefix3,
			ClubName:                club.Name,
			ClubID:                  club.ID,
			PlayerID:                player.ID,
			PKZ:                     player.PKZ,
			Lastname:                player.Name,
			Firstname:               player.Firstname,
			Birthyear:               birthyear,
			Gender:                  player.Gender,
			CurrentDWZ:              player.CurrentDWZ,
			ListRanking:             0, // Will be calculated later
			DWZ12MonthsAgo:          models.DataNotAvailable,
			GamesLast12Months:       models.DataNotAvailable,
			SuccessRateLast12Months: models.DataNotAvailable,
			HistoricalDWZ:           make(map[string]string), // Initialize the map
			DWZMinCurrentYear:       models.DataNotAvailable,
			DWZMaxCurrentYear:       models.DataNotAvailable,
			SomatogramPercentile:    models.DataNotAvailable, // Will be populated by Germany-wide percentile data
			DWZAgeRelation:          models.CalculateDWZAgeRelation(player.CurrentDWZ, birthyear, time.Now().Year()),
		}

		// Add historical analysis if requested
		if includeHistoricalAnalysis {
			history, err := p.apiClient.GetPlayerRatingHistory(player.ID)
			if err != nil {
				p.logger.Warnf("Failed to fetch rating history for player %s: %v", player.ID, err)
				// Still set games and success rate to 0 when API fails
				record.GamesLast12Months = "0"
				record.SuccessRateLast12Months = "0,0"
				// DWZ12MonthsAgo remains DATA_NOT_AVAILABLE
				// Initialize historical DWZ with DATA_NOT_AVAILABLE for all dates
				for _, date := range targetDates {
					key := fmt.Sprintf("%d_%02d_%02d", date.Year(), date.Month(), date.Day())
					record.HistoricalDWZ[key] = models.DataNotAvailable
				}
			} else if history == nil {
				p.logger.Warnf("Rating history is nil for player %s", player.ID)
				// Still set games and success rate to 0 when history is nil
				record.GamesLast12Months = "0"
				record.SuccessRateLast12Months = "0,0"
				// DWZ12MonthsAgo remains DATA_NOT_AVAILABLE
				// Initialize historical DWZ with DATA_NOT_AVAILABLE for all dates
				for _, date := range targetDates {
					key := fmt.Sprintf("%d_%02d_%02d", date.Year(), date.Month(), date.Day())
					record.HistoricalDWZ[key] = models.DataNotAvailable
				}
			} else {
				// Use the legacy processor's analyzeHistoricalData method
				analysis := p.legacyProcessor.AnalyzeHistoricalData(history)
				if analysis != nil {
					if analysis.HasHistoricalData {
						p.logger.Debugf("Player %s: applying historical analysis - DWZ12MonthsAgo=%s, Games=%d, Success=%.1f",
							player.ID, analysis.DWZ12MonthsAgo, analysis.GamesLast12Months, analysis.SuccessRateLast12Months)
						record.DWZ12MonthsAgo = analysis.DWZ12MonthsAgo
					} else {
						p.logger.Debugf("Player %s: no historical data found, but still recording games/success rate", player.ID)
						// Keep DWZ12MonthsAgo as DATA_NOT_AVAILABLE since we can't find it
					}

					// Always record games count and success rate, even if no historical data
					record.GamesLast12Months = fmt.Sprintf("%d", analysis.GamesLast12Months)

					if analysis.GamesLast12Months > 0 {
						record.SuccessRateLast12Months = strings.Replace(fmt.Sprintf("%.1f", analysis.SuccessRateLast12Months), ".", ",", 1)
					} else {
						// If no games, success rate is 0
						record.SuccessRateLast12Months = "0,0"
					}
				} else {
					p.logger.Debugf("Player %s: analysis is nil", player.ID)
					// Even if analysis fails, set games and success rate to 0
					record.GamesLast12Months = "0"
					record.SuccessRateLast12Months = "0,0"
					// DWZ12MonthsAgo remains DATA_NOT_AVAILABLE
				}

				// NEW: Extended historical analysis - semi-annual DWZ values
				record.HistoricalDWZ = models.AnalyzeExtendedHistoricalData(history.Points, targetDates)

				// NEW: Calculate yearly min/max for current year
				record.DWZMinCurrentYear, record.DWZMaxCurrentYear = models.CalculateYearlyMinMax(
					history.Points, player.CurrentDWZ, currentYear)
			}
		} else {
			// Initialize historical DWZ with DATA_NOT_AVAILABLE for all dates when not doing analysis
			for _, date := range targetDates {
				key := fmt.Sprintf("%d_%02d_%02d", date.Year(), date.Month(), date.Day())
				record.HistoricalDWZ[key] = models.DataNotAvailable
			}
		}

		records = append(records, record)
	}

	return records
}

// calculateListRankings calculates list rankings for all records
func (p *UnifiedProcessor) calculateListRankings(records []models.KaderPlanungRecord) {
	// Extract players from records for ranking calculation
	players := make([]models.Player, len(records))
	for i, record := range records {
		players[i] = models.Player{
			ID:         record.PlayerID,
			CurrentDWZ: record.CurrentDWZ,
			Status:     "active",
			Active:     true,
		}
	}

	// Calculate rankings
	rankings := models.CalculateListRanking(players)

	// Update records with rankings
	for i := range records {
		records[i].ListRanking = rankings[i]
	}

	p.logger.Debugf("Calculated list rankings for %d records", len(records))
}

// countUniqueClubs counts the number of unique clubs in a player list
func (p *UnifiedProcessor) countUniqueClubs(players []models.Player) int {
	clubs := make(map[string]bool)
	for _, player := range players {
		clubs[player.ClubID] = true
	}
	return len(clubs)
}

// populateSomatogramPercentiles applies statistical analysis results to populate somatogram percentiles
func (p *UnifiedProcessor) populateSomatogramPercentiles(records []models.KaderPlanungRecord, statisticalData map[string]statistics.SomatogrammData) {
	p.logger.Debug("Populating somatogram percentiles from statistical analysis...")
	
	populatedCount := 0
	
	for i := range records {
		record := &records[i]
		
		// Skip players without birth year or gender
		if record.Birthyear == 0 || record.Gender == "" {
			continue
		}
		
		// Calculate age
		currentYear := time.Now().Year()
		age := currentYear - record.Birthyear
		
		// Look up statistical data for this gender
		genderData, exists := statisticalData[record.Gender]
		if !exists {
			continue
		}
		
		// Look up percentile data for this age
		ageKey := fmt.Sprintf("%d", age)
		ageData, exists := genderData.Percentiles[ageKey]
		if !exists {
			continue
		}
		
		// Find the percentile for this player's DWZ
		percentile := p.findPercentileForDWZ(record.CurrentDWZ, ageData.Percentiles)
		if percentile >= 0 {
			record.SomatogramPercentile = strings.Replace(fmt.Sprintf("%.1f", percentile), ".", ",", 1)
			populatedCount++
		}
	}
	
	p.logger.Infof("Populated somatogram percentiles for %d/%d records", populatedCount, len(records))
}

// findPercentileForDWZ finds the percentile rank for a specific DWZ value
func (p *UnifiedProcessor) findPercentileForDWZ(dwz int, percentileThresholds map[int]int) float64 {
	// percentileThresholds maps percentile -> DWZ threshold
	// We need to find which percentile this DWZ falls into
	
	// Handle edge cases
	if len(percentileThresholds) == 0 {
		return -1
	}
	
	// Get sorted percentiles (0, 1, 2, ..., 100)
	var percentiles []int
	for p := range percentileThresholds {
		percentiles = append(percentiles, p)
	}
	sort.Ints(percentiles)
	
	// Find the percentile range this DWZ falls into
	for i, percentile := range percentiles {
		threshold := percentileThresholds[percentile]
		
		if dwz <= threshold {
			// This DWZ is at or below this percentile threshold
			if i == 0 {
				// At or below the lowest threshold (0th percentile)
				return float64(percentile)
			}
			
			// Interpolate between this percentile and the previous one
			prevPercentile := percentiles[i-1]
			prevThreshold := percentileThresholds[prevPercentile]
			
			if threshold == prevThreshold {
				// No interpolation needed if thresholds are the same
				return float64(percentile)
			}
			
			// Linear interpolation
			ratio := float64(dwz-prevThreshold) / float64(threshold-prevThreshold)
			interpolatedPercentile := float64(prevPercentile) + ratio*float64(percentile-prevPercentile)
			return interpolatedPercentile
		}
	}
	
	// DWZ is above all thresholds (above 100th percentile)
	return 100.0
}

// SetLogger sets a custom logger for the processor
func (p *UnifiedProcessor) SetLogger(logger *logrus.Logger) {
	p.logger = logger
	p.statisticalAnalyzer = statistics.NewStatisticalAnalyzer(p.config.MinSampleSize, logger)
	p.legacyProcessor.SetLogger(logger)
}

// ProcessingResult contains the results of processing
type ProcessingResult struct {
	Mode            ProcessingMode                           `json:"mode"`
	Records         []models.KaderPlanungRecord             `json:"records,omitempty"`
	StatisticalData map[string]statistics.SomatogrammData   `json:"statistical_data,omitempty"`
	TotalPlayers    int                                      `json:"total_players"`
	ProcessedClubs  int                                      `json:"processed_clubs"`
	ProcessingTime  time.Duration                            `json:"processing_time"`
}