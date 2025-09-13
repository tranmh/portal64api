package processor

import (
	"fmt"
	"sort"
	"time"

	"somatogramm/internal/models"
)

type Processor struct {
	MinSampleSize int
	Verbose       bool
}

func NewProcessor(minSampleSize int, verbose bool) *Processor {
	return &Processor{
		MinSampleSize: minSampleSize,
		Verbose:       verbose,
	}
}

func (p *Processor) log(message string) {
	if p.Verbose {
		fmt.Printf("[PROCESSOR] %s\n", message)
	}
}

func (p *Processor) ProcessPlayers(players []models.Player) (map[string]models.SomatogrammData, error) {
	p.log(fmt.Sprintf("Processing %d players...", len(players)))

	ageGenderGroups := p.groupPlayersByAgeAndGender(players)
	p.log(fmt.Sprintf("Created %d age-gender groups", len(ageGenderGroups)))

	validGroups := p.filterGroupsBySampleSize(ageGenderGroups)
	p.log(fmt.Sprintf("%d groups meet minimum sample size requirement (%d)", len(validGroups), p.MinSampleSize))

	results := make(map[string]models.SomatogrammData)

	genderGroups := map[string][]models.AgeGenderGroup{
		"m": {},
		"w": {},
		"d": {},
	}

	for _, group := range validGroups {
		genderGroups[group.Gender] = append(genderGroups[group.Gender], group)
	}

	for gender, groups := range genderGroups {
		if len(groups) == 0 {
			continue
		}

		percentilesByAge := make(map[string]models.PercentileData)
		totalPlayers := 0
		validAgeGroups := 0

		for _, group := range groups {
			percentiles := p.calculatePercentiles(group.Players)

			percentilesByAge[fmt.Sprintf("%d", group.Age)] = models.PercentileData{
				Age:         group.Age,
				SampleSize:  group.SampleSize,
				AvgDWZ:      group.AvgDWZ,
				MedianDWZ:   group.MedianDWZ,
				Percentiles: percentiles,
			}

			totalPlayers += len(group.Players)
			validAgeGroups++
		}

		results[gender] = models.SomatogrammData{
			Metadata: models.SomatogrammMetadata{
				GeneratedAt:    time.Now(),
				Gender:         gender,
				TotalPlayers:   totalPlayers,
				ValidAgeGroups: validAgeGroups,
				MinSampleSize:  p.MinSampleSize,
			},
			Percentiles: percentilesByAge,
		}

		p.log(fmt.Sprintf("Generated data for gender %s: %d players, %d age groups", gender, totalPlayers, validAgeGroups))
	}

	return results, nil
}

func (p *Processor) groupPlayersByAgeAndGender(players []models.Player) map[string]models.AgeGenderGroup {
	groups := make(map[string]models.AgeGenderGroup)
	currentYear := time.Now().Year()

	for _, player := range players {
		if player.BirthYear == nil {
			continue
		}

		age := currentYear - *player.BirthYear
		key := fmt.Sprintf("%s-%d", player.Gender, age)

		group, exists := groups[key]
		if !exists {
			group = models.AgeGenderGroup{
				Age:     age,
				Gender:  player.Gender,
				Players: []models.Player{},
			}
		}

		group.Players = append(group.Players, player)
		groups[key] = group
	}

	for key, group := range groups {
		group.SampleSize = len(group.Players)
		group.AvgDWZ = p.calculateAverageDWZ(group.Players)
		group.MedianDWZ = p.calculateMedianDWZ(group.Players)
		groups[key] = group
	}

	return groups
}

func (p *Processor) filterGroupsBySampleSize(groups map[string]models.AgeGenderGroup) []models.AgeGenderGroup {
	var validGroups []models.AgeGenderGroup

	for _, group := range groups {
		if group.SampleSize >= p.MinSampleSize {
			validGroups = append(validGroups, group)
		} else {
			p.log(fmt.Sprintf("Skipping age %d, gender %s: only %d players (minimum: %d)",
				group.Age, group.Gender, group.SampleSize, p.MinSampleSize))
		}
	}

	sort.Slice(validGroups, func(i, j int) bool {
		if validGroups[i].Gender == validGroups[j].Gender {
			return validGroups[i].Age < validGroups[j].Age
		}
		return validGroups[i].Gender < validGroups[j].Gender
	})

	return validGroups
}

func (p *Processor) calculatePercentiles(players []models.Player) map[int]int {
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