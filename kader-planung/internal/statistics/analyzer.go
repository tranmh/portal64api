package statistics

import (
	"fmt"
	"sort"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
)

// StatisticalAnalyzer handles statistical analysis and percentile calculations
type StatisticalAnalyzer struct {
	minSampleSize int
	logger        *logrus.Logger
}

// NewStatisticalAnalyzer creates a new statistical analyzer
func NewStatisticalAnalyzer(minSampleSize int, logger *logrus.Logger) *StatisticalAnalyzer {
	return &StatisticalAnalyzer{
		minSampleSize: minSampleSize,
		logger:        logger,
	}
}

// SomatogrammData represents the complete statistical analysis results
type SomatogrammData struct {
	Metadata    SomatogrammMetadata        `json:"metadata"`
	Percentiles map[string]PercentileData `json:"percentiles"` // Age -> PercentileData
}

// SomatogrammMetadata contains metadata about the statistical analysis
type SomatogrammMetadata struct {
	GeneratedAt    time.Time `json:"generated_at"`
	Gender         string    `json:"gender"`
	TotalPlayers   int       `json:"total_players"`
	ValidAgeGroups int       `json:"valid_age_groups"`
	MinSampleSize  int       `json:"min_sample_size"`
}

// PercentileData represents percentile statistics for a specific age group
type PercentileData struct {
	Age         int            `json:"age"`
	SampleSize  int            `json:"sample_size"`
	AvgDWZ      float64        `json:"avg_dwz"`
	MedianDWZ   int            `json:"median_dwz"`
	Percentiles map[int]int    `json:"percentiles"` // Percentile -> DWZ value
}

// AgeGenderGroup represents a group of players with the same age and gender
type AgeGenderGroup struct {
	Age        int             `json:"age"`
	Gender     string          `json:"gender"`
	Players    []models.Player `json:"-"` // Don't serialize player data
	SampleSize int             `json:"sample_size"`
	AvgDWZ     float64         `json:"avg_dwz"`
	MedianDWZ  int             `json:"median_dwz"`
}

// ProcessPlayers processes players for statistical analysis (ported from Somatogramm)
func (s *StatisticalAnalyzer) ProcessPlayers(players []models.Player) (map[string]SomatogrammData, error) {
	s.logger.Infof("Processing %d players for statistical analysis...", len(players))

	ageGenderGroups := s.groupPlayersByAgeAndGender(players)
	s.logger.Debugf("Created %d age-gender groups", len(ageGenderGroups))

	validGroups := s.filterGroupsBySampleSize(ageGenderGroups)
	s.logger.Infof("%d groups meet minimum sample size requirement (%d)", len(validGroups), s.minSampleSize)

	results := make(map[string]SomatogrammData)

	// Group by gender for separate processing
	genderGroups := map[string][]AgeGenderGroup{
		"m": {},
		"w": {},
		"d": {},
	}

	for _, group := range validGroups {
		genderGroups[group.Gender] = append(genderGroups[group.Gender], group)
	}

	// Process each gender separately
	for gender, groups := range genderGroups {
		if len(groups) == 0 {
			continue
		}

		percentilesByAge := make(map[string]PercentileData)
		totalPlayers := 0
		validAgeGroups := 0

		for _, group := range groups {
			percentiles := s.calculatePercentiles(group.Players)

			percentilesByAge[fmt.Sprintf("%d", group.Age)] = PercentileData{
				Age:         group.Age,
				SampleSize:  group.SampleSize,
				AvgDWZ:      group.AvgDWZ,
				MedianDWZ:   group.MedianDWZ,
				Percentiles: percentiles,
			}

			totalPlayers += len(group.Players)
			validAgeGroups++
		}

		results[gender] = SomatogrammData{
			Metadata: SomatogrammMetadata{
				GeneratedAt:    time.Now(),
				Gender:         gender,
				TotalPlayers:   totalPlayers,
				ValidAgeGroups: validAgeGroups,
				MinSampleSize:  s.minSampleSize,
			},
			Percentiles: percentilesByAge,
		}

		s.logger.Infof("Generated statistical data for gender %s: %d players, %d age groups", gender, totalPlayers, validAgeGroups)
	}

	return results, nil
}

// groupPlayersByAgeAndGender groups players by age and gender (ported from Somatogramm)
func (s *StatisticalAnalyzer) groupPlayersByAgeAndGender(players []models.Player) map[string]AgeGenderGroup {
	groups := make(map[string]AgeGenderGroup)
	currentYear := time.Now().Year()

	for _, player := range players {
		if player.BirthYear == nil {
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
		group.AvgDWZ = s.calculateAverageDWZ(group.Players)
		group.MedianDWZ = s.calculateMedianDWZ(group.Players)
		groups[key] = group
	}

	return groups
}

// filterGroupsBySampleSize filters groups by minimum sample size (ported from Somatogramm)
func (s *StatisticalAnalyzer) filterGroupsBySampleSize(groups map[string]AgeGenderGroup) []AgeGenderGroup {
	var validGroups []AgeGenderGroup

	for _, group := range groups {
		if group.SampleSize >= s.minSampleSize {
			validGroups = append(validGroups, group)
		} else {
			s.logger.Debugf("Skipping age %d, gender %s: only %d players (minimum: %d)",
				group.Age, group.Gender, group.SampleSize, s.minSampleSize)
		}
	}

	// Sort by gender first, then by age
	sort.Slice(validGroups, func(i, j int) bool {
		if validGroups[i].Gender == validGroups[j].Gender {
			return validGroups[i].Age < validGroups[j].Age
		}
		return validGroups[i].Gender < validGroups[j].Gender
	})

	return validGroups
}

// calculatePercentiles calculates percentile values for a group of players (ported from Somatogramm)
func (s *StatisticalAnalyzer) calculatePercentiles(players []models.Player) map[int]int {
	if len(players) == 0 {
		return make(map[int]int)
	}

	dwzValues := make([]int, len(players))
	for i, player := range players {
		dwzValues[i] = player.CurrentDWZ
	}

	sort.Ints(dwzValues)
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

// calculateAverageDWZ calculates the average DWZ for a group of players (ported from Somatogramm)
func (s *StatisticalAnalyzer) calculateAverageDWZ(players []models.Player) float64 {
	if len(players) == 0 {
		return 0
	}

	total := 0
	for _, player := range players {
		total += player.CurrentDWZ
	}

	return float64(total) / float64(len(players))
}

// calculateMedianDWZ calculates the median DWZ for a group of players (ported from Somatogramm)
func (s *StatisticalAnalyzer) calculateMedianDWZ(players []models.Player) int {
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

// SetMinSampleSize updates the minimum sample size
func (s *StatisticalAnalyzer) SetMinSampleSize(minSampleSize int) {
	s.minSampleSize = minSampleSize
}

// GetMinSampleSize returns the current minimum sample size
func (s *StatisticalAnalyzer) GetMinSampleSize() int {
	return s.minSampleSize
}