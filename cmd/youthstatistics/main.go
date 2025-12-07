// Youth Statistics Tool for Portal64 API
// Queries active youth player counts (U08-U18) by district (Bezirk), age class, and gender
// for a specified reference year (default: 2026)
package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// API Response structures
type ClubResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	District string `json:"district"`
}

// ClubsAPIResponse matches the actual API response: {"success":true,"data":[...]}
type ClubsAPIResponse struct {
	Success bool           `json:"success"`
	Data    []ClubResponse `json:"data"`
}

type PlayerResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Firstname string `json:"firstname"`
	BirthYear *int   `json:"birth_year"`
	Gender    string `json:"gender"`
	ClubID    string `json:"club_id"`
}

type PlayersMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type PlayersData struct {
	Data []PlayerResponse `json:"data"`
	Meta PlayersMeta      `json:"meta"`
}

// PlayersAPIResponse matches the actual API response: {"success":true,"data":{"data":[...],"meta":{...}}}
type PlayersAPIResponse struct {
	Success bool        `json:"success"`
	Data    PlayersData `json:"data"`
}

// AgeGroup represents a youth age category
type AgeGroup struct {
	Name         string // e.g., "U18", "U16"
	MinBirthYear int    // minimum birth year for this category
}

// Statistics structures
type GenderStats struct {
	Male   int `json:"male"`
	Female int `json:"female"`
	Other  int `json:"other"`
	Total  int `json:"total"`
}

type DistrictStats struct {
	District   string                  `json:"district"`
	ByAgeGroup map[string]*GenderStats `json:"by_age_group"`
	ByYear     map[int]*GenderStats    `json:"by_year"`
	Total      *GenderStats            `json:"total"`
}

type YouthStatistics struct {
	ReferenceYear int                       `json:"reference_year"`
	AgeGroups     []string                  `json:"age_groups"`
	BirthYears    []int                     `json:"birth_years"`
	Districts     map[string]*DistrictStats `json:"districts"`
	Total         *DistrictStats            `json:"total"`
	GeneratedAt   string                    `json:"generated_at"`
}

// ClubResult holds players fetched for a club
type ClubResult struct {
	Club    ClubResponse
	Players []PlayerResponse
	Error   error
}

func main() {
	// Command line flags
	serverURL := flag.String("server", "http://localhost:8080", "API server URL")
	refYear := flag.Int("year", 2026, "Reference year for age calculation")
	outputFormat := flag.String("format", "table", "Output format: table, csv, json")
	byYear := flag.Bool("by-year", false, "Show statistics by birth year instead of age groups")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	concurrency := flag.Int("concurrency", 50, "Number of concurrent requests")
	flag.Parse()

	// Create HTTP client with TLS skip verification option
	client := createHTTPClient(*serverURL)

	fmt.Printf("Youth Statistics Tool - Reference Year: %d\n", *refYear)
	fmt.Printf("Server: %s\n", *serverURL)
	fmt.Printf("Concurrency: %d\n\n", *concurrency)

	// Define age groups for 2026 (U08 to U18)
	ageGroups := calculateAgeGroups(*refYear)

	// Get all clubs
	if *verbose {
		fmt.Println("Fetching clubs...")
	}
	clubs, err := fetchAllClubs(client, *serverURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching clubs: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d clubs, fetching players...\n", len(clubs))

	// Initialize statistics
	stats := &YouthStatistics{
		ReferenceYear: *refYear,
		Districts:     make(map[string]*DistrictStats),
		Total:         newDistrictStats("TOTAL"),
		GeneratedAt:   time.Now().Format(time.RFC3339),
	}

	// Populate age groups and birth years
	for _, ag := range ageGroups {
		stats.AgeGroups = append(stats.AgeGroups, ag.Name)
	}
	for year := ageGroups[0].MinBirthYear; year <= *refYear-7; year++ {
		stats.BirthYears = append(stats.BirthYears, year)
	}

	// Create channels for worker pool
	clubChan := make(chan ClubResponse, len(clubs))
	resultChan := make(chan ClubResult, len(clubs))

	// Progress counter
	var processed int64

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for club := range clubChan {
				players, err := fetchClubPlayers(client, *serverURL, club.ID)
				resultChan <- ClubResult{Club: club, Players: players, Error: err}

				current := atomic.AddInt64(&processed, 1)
				if current%100 == 0 || current == int64(len(clubs)) {
					fmt.Printf("\rProcessing: %d/%d clubs...", current, len(clubs))
				}
			}
		}()
	}

	// Send clubs to workers
	go func() {
		for _, club := range clubs {
			clubChan <- club
		}
		close(clubChan)
	}()

	// Close result channel when workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	totalPlayers := 0
	youthPlayers := 0
	errorCount := 0

	for result := range resultChan {
		if result.Error != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "\nWarning: Error fetching players for club %s: %v\n", result.Club.ID, result.Error)
			}
			errorCount++
			continue
		}

		totalPlayers += len(result.Players)

		// Process each player
		for _, player := range result.Players {
			if player.BirthYear == nil {
				continue
			}

			birthYear := *player.BirthYear
			ageGroup := getAgeGroup(birthYear, ageGroups)
			if ageGroup == "" {
				continue // Not in youth age range
			}

			youthPlayers++

			// Get or create district stats
			district := result.Club.District
			if district == "" {
				district = "UNKNOWN"
			}
			districtStats := getOrCreateDistrictStats(stats.Districts, district)

			// Determine gender
			gender := normalizeGender(player.Gender)

			// Update statistics
			updateStats(districtStats, ageGroup, birthYear, gender)
			updateStats(stats.Total, ageGroup, birthYear, gender)
		}
	}

	fmt.Printf("\n\nProcessed %d total players, %d youth players (U08-U18)\n", totalPlayers, youthPlayers)
	if errorCount > 0 {
		fmt.Printf("Errors: %d clubs failed to fetch\n", errorCount)
	}
	fmt.Println()

	// Output results
	switch *outputFormat {
	case "json":
		outputJSON(stats)
	case "csv":
		outputCSV(stats, *byYear)
	default:
		outputTable(stats, *byYear)
	}
}

func createHTTPClient(serverURL string) *http.Client {
	transport := &http.Transport{}

	// Skip TLS verification for HTTPS
	if strings.HasPrefix(serverURL, "https://") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func calculateAgeGroups(refYear int) []AgeGroup {
	// German chess age groups: U08, U10, U12, U14, U16, U18
	// UXX means player must be under XX years old on January 1st of reference year
	return []AgeGroup{
		{Name: "U18", MinBirthYear: refYear - 17}, // Born 2009 or later for 2026
		{Name: "U16", MinBirthYear: refYear - 15}, // Born 2011 or later
		{Name: "U14", MinBirthYear: refYear - 13}, // Born 2013 or later
		{Name: "U12", MinBirthYear: refYear - 11}, // Born 2015 or later
		{Name: "U10", MinBirthYear: refYear - 9},  // Born 2017 or later
		{Name: "U08", MinBirthYear: refYear - 7},  // Born 2019 or later
	}
}

func getAgeGroup(birthYear int, ageGroups []AgeGroup) string {
	// Find the smallest age group the player qualifies for
	for i := len(ageGroups) - 1; i >= 0; i-- {
		if birthYear >= ageGroups[i].MinBirthYear {
			return ageGroups[i].Name
		}
	}
	return "" // Not in youth range
}

func fetchAllClubs(client *http.Client, serverURL string) ([]ClubResponse, error) {
	url := fmt.Sprintf("%s/api/v1/clubs/all", serverURL)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ClubsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

func fetchClubPlayers(client *http.Client, serverURL, clubID string) ([]PlayerResponse, error) {
	var allPlayers []PlayerResponse
	offset := 0
	limit := 500

	for {
		url := fmt.Sprintf("%s/api/v1/clubs/%s/players?limit=%d&offset=%d&active=true",
			serverURL, clubID, limit, offset)

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		var result PlayersAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		allPlayers = append(allPlayers, result.Data.Data...)

		// Check if we've fetched all players
		if len(result.Data.Data) < limit || offset+len(result.Data.Data) >= result.Data.Meta.Total {
			break
		}
		offset += limit
	}

	return allPlayers, nil
}

func normalizeGender(gender string) string {
	switch strings.ToLower(gender) {
	case "m", "male", "männlich":
		return "m"
	case "w", "f", "female", "weiblich":
		return "w"
	default:
		return "d"
	}
}

func newDistrictStats(name string) *DistrictStats {
	return &DistrictStats{
		District:   name,
		ByAgeGroup: make(map[string]*GenderStats),
		ByYear:     make(map[int]*GenderStats),
		Total:      &GenderStats{},
	}
}

func getOrCreateDistrictStats(districts map[string]*DistrictStats, name string) *DistrictStats {
	if stats, ok := districts[name]; ok {
		return stats
	}
	stats := newDistrictStats(name)
	districts[name] = stats
	return stats
}

func getOrCreateGenderStats(m map[string]*GenderStats, key string) *GenderStats {
	if stats, ok := m[key]; ok {
		return stats
	}
	stats := &GenderStats{}
	m[key] = stats
	return stats
}

func getOrCreateGenderStatsByYear(m map[int]*GenderStats, year int) *GenderStats {
	if stats, ok := m[year]; ok {
		return stats
	}
	stats := &GenderStats{}
	m[year] = stats
	return stats
}

func updateStats(ds *DistrictStats, ageGroup string, birthYear int, gender string) {
	// Update by age group
	ags := getOrCreateGenderStats(ds.ByAgeGroup, ageGroup)
	updateGenderStats(ags, gender)

	// Update by birth year
	bys := getOrCreateGenderStatsByYear(ds.ByYear, birthYear)
	updateGenderStats(bys, gender)

	// Update total
	updateGenderStats(ds.Total, gender)
}

func updateGenderStats(gs *GenderStats, gender string) {
	switch gender {
	case "m":
		gs.Male++
	case "w":
		gs.Female++
	default:
		gs.Other++
	}
	gs.Total++
}

func outputJSON(stats *YouthStatistics) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(stats)
}

func outputCSV(stats *YouthStatistics, byYear bool) {
	// Sort districts
	var districts []string
	for d := range stats.Districts {
		districts = append(districts, d)
	}
	sort.Strings(districts)

	if byYear {
		// Header
		fmt.Print("Bezirk;Geschlecht")
		for _, year := range stats.BirthYears {
			fmt.Printf(";%d", year)
		}
		fmt.Println(";Gesamt")

		// Data rows
		for _, district := range districts {
			ds := stats.Districts[district]
			outputCSVRowByYear(district, "m", ds, stats.BirthYears)
			outputCSVRowByYear(district, "w", ds, stats.BirthYears)
			outputCSVRowByYear(district, "Gesamt", ds, stats.BirthYears)
		}

		// Total row
		fmt.Println()
		outputCSVRowByYear("GESAMT", "m", stats.Total, stats.BirthYears)
		outputCSVRowByYear("GESAMT", "w", stats.Total, stats.BirthYears)
		outputCSVRowByYear("GESAMT", "Gesamt", stats.Total, stats.BirthYears)
	} else {
		// Header
		fmt.Print("Bezirk;Geschlecht")
		for _, ag := range stats.AgeGroups {
			fmt.Printf(";%s", ag)
		}
		fmt.Println(";Gesamt")

		// Data rows
		for _, district := range districts {
			ds := stats.Districts[district]
			outputCSVRowByAgeGroup(district, "m", ds, stats.AgeGroups)
			outputCSVRowByAgeGroup(district, "w", ds, stats.AgeGroups)
			outputCSVRowByAgeGroup(district, "Gesamt", ds, stats.AgeGroups)
		}

		// Total row
		fmt.Println()
		outputCSVRowByAgeGroup("GESAMT", "m", stats.Total, stats.AgeGroups)
		outputCSVRowByAgeGroup("GESAMT", "w", stats.Total, stats.AgeGroups)
		outputCSVRowByAgeGroup("GESAMT", "Gesamt", stats.Total, stats.AgeGroups)
	}
}

func outputCSVRowByAgeGroup(district, gender string, ds *DistrictStats, ageGroups []string) {
	fmt.Printf("%s;%s", district, gender)
	var total int
	for _, ag := range ageGroups {
		count := 0
		if gs, ok := ds.ByAgeGroup[ag]; ok {
			switch gender {
			case "m":
				count = gs.Male
			case "w":
				count = gs.Female
			case "Gesamt":
				count = gs.Total
			}
		}
		fmt.Printf(";%d", count)
		total += count
	}
	fmt.Printf(";%d\n", total)
}

func outputCSVRowByYear(district, gender string, ds *DistrictStats, years []int) {
	fmt.Printf("%s;%s", district, gender)
	var total int
	for _, year := range years {
		count := 0
		if gs, ok := ds.ByYear[year]; ok {
			switch gender {
			case "m":
				count = gs.Male
			case "w":
				count = gs.Female
			case "Gesamt":
				count = gs.Total
			}
		}
		fmt.Printf(";%d", count)
		total += count
	}
	fmt.Printf(";%d\n", total)
}

func outputTable(stats *YouthStatistics, byYear bool) {
	// Sort districts
	var districts []string
	for d := range stats.Districts {
		districts = append(districts, d)
	}
	sort.Strings(districts)

	if byYear {
		outputTableByYear(stats, districts)
	} else {
		outputTableByAgeGroup(stats, districts)
	}
}

func outputTableByAgeGroup(stats *YouthStatistics, districts []string) {
	// Calculate column widths
	districtWidth := 12
	for _, d := range districts {
		if len(d) > districtWidth {
			districtWidth = len(d)
		}
	}

	// Header
	fmt.Printf("%-*s  %-7s", districtWidth, "Bezirk", "Gender")
	for _, ag := range stats.AgeGroups {
		fmt.Printf("  %6s", ag)
	}
	fmt.Printf("  %7s\n", "Gesamt")

	// Separator
	fmt.Print(strings.Repeat("-", districtWidth+2))
	fmt.Print(strings.Repeat("-", 9))
	for range stats.AgeGroups {
		fmt.Print(strings.Repeat("-", 8))
	}
	fmt.Println(strings.Repeat("-", 9))

	// Data rows
	for _, district := range districts {
		ds := stats.Districts[district]
		printTableRowByAgeGroup(district, "m", ds, stats.AgeGroups, districtWidth)
		printTableRowByAgeGroup("", "w", ds, stats.AgeGroups, districtWidth)
		printTableRowByAgeGroup("", "Total", ds, stats.AgeGroups, districtWidth)
		fmt.Println()
	}

	// Grand total
	fmt.Print(strings.Repeat("=", districtWidth+2))
	fmt.Print(strings.Repeat("=", 9))
	for range stats.AgeGroups {
		fmt.Print(strings.Repeat("=", 8))
	}
	fmt.Println(strings.Repeat("=", 9))

	printTableRowByAgeGroup("TOTAL", "m", stats.Total, stats.AgeGroups, districtWidth)
	printTableRowByAgeGroup("", "w", stats.Total, stats.AgeGroups, districtWidth)
	printTableRowByAgeGroup("", "Total", stats.Total, stats.AgeGroups, districtWidth)
}

func printTableRowByAgeGroup(district, gender string, ds *DistrictStats, ageGroups []string, width int) {
	fmt.Printf("%-*s  %-7s", width, district, gender)
	var total int
	for _, ag := range ageGroups {
		count := 0
		if gs, ok := ds.ByAgeGroup[ag]; ok {
			switch gender {
			case "m":
				count = gs.Male
			case "w":
				count = gs.Female
			case "Total":
				count = gs.Total
			}
		}
		fmt.Printf("  %6d", count)
		total += count
	}
	fmt.Printf("  %7d\n", total)
}

func outputTableByYear(stats *YouthStatistics, districts []string) {
	// Calculate column widths
	districtWidth := 12
	for _, d := range districts {
		if len(d) > districtWidth {
			districtWidth = len(d)
		}
	}

	// Header
	fmt.Printf("%-*s  %-7s", districtWidth, "Bezirk", "Gender")
	for _, year := range stats.BirthYears {
		fmt.Printf("  %6d", year)
	}
	fmt.Printf("  %7s\n", "Gesamt")

	// Separator
	fmt.Print(strings.Repeat("-", districtWidth+2))
	fmt.Print(strings.Repeat("-", 9))
	for range stats.BirthYears {
		fmt.Print(strings.Repeat("-", 8))
	}
	fmt.Println(strings.Repeat("-", 9))

	// Data rows
	for _, district := range districts {
		ds := stats.Districts[district]
		printTableRowByYear(district, "m", ds, stats.BirthYears, districtWidth)
		printTableRowByYear("", "w", ds, stats.BirthYears, districtWidth)
		printTableRowByYear("", "Total", ds, stats.BirthYears, districtWidth)
		fmt.Println()
	}

	// Grand total
	fmt.Print(strings.Repeat("=", districtWidth+2))
	fmt.Print(strings.Repeat("=", 9))
	for range stats.BirthYears {
		fmt.Print(strings.Repeat("=", 8))
	}
	fmt.Println(strings.Repeat("=", 9))

	printTableRowByYear("TOTAL", "m", stats.Total, stats.BirthYears, districtWidth)
	printTableRowByYear("", "w", stats.Total, stats.BirthYears, districtWidth)
	printTableRowByYear("", "Total", stats.Total, stats.BirthYears, districtWidth)
}

func printTableRowByYear(district, gender string, ds *DistrictStats, years []int, width int) {
	fmt.Printf("%-*s  %-7s", width, district, gender)
	var total int
	for _, year := range years {
		count := 0
		if gs, ok := ds.ByYear[year]; ok {
			switch gender {
			case "m":
				count = gs.Male
			case "w":
				count = gs.Female
			case "Total":
				count = gs.Total
			}
		}
		fmt.Printf("  %6d", count)
		total += count
	}
	fmt.Printf("  %7d\n", total)
}
