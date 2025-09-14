package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"time"

	"somatogramm/internal/api"
	"somatogramm/internal/export"
	"somatogramm/internal/models"
	"somatogramm/internal/processor"
)

func main() {
	config := parseFlags()

	if config.Verbose {
		fmt.Printf("Starting somatogramm with config: %+v\n", config)
	}


	client := api.NewClient(config.APIBaseURL, time.Duration(config.Timeout)*time.Second, config.Verbose, config.Concurrency)

	// Check if debug mode is enabled
	debugMode := config.DebugAge > 0 && config.DebugGender != ""
	
	if debugMode {
		if config.Verbose {
			fmt.Printf("Debug mode enabled: age=%d, gender=%s, output=%s\n", 
				config.DebugAge, config.DebugGender, config.DebugOutput)
		}
		
		if err := runDebugMode(client, config); err != nil {
			log.Fatalf("Debug mode failed: %v", err)
		}
		
		if config.Verbose {
			fmt.Println("Debug mode completed successfully")
		}
		return
	}

	// Normal somatogramm processing
	players, err := client.FetchAllPlayers()
	if err != nil {
		log.Fatalf("Failed to fetch players: %v", err)
	}

	if config.Verbose {
		fmt.Printf("Fetched %d valid players\n", len(players))
	}

	proc := processor.NewProcessor(config.MinSampleSize, config.Verbose)

	processedData, err := proc.ProcessPlayers(players)
	if err != nil {
		log.Fatalf("Failed to process players: %v", err)
	}

	if config.Verbose {
		fmt.Printf("Processed data for %d genders\n", len(processedData))
	}

	exporter := export.NewExporter(config.OutputDir, config.OutputFormat, 7, config.Verbose)

	if err := exporter.ExportData(processedData); err != nil {
		log.Fatalf("Failed to export data: %v", err)
	}

	if config.Verbose {
		fmt.Println("Somatogramm generation completed successfully")
	}
}

func parseFlags() *models.Config {
	config := &models.Config{}

	flag.StringVar(&config.OutputFormat, "output-format", "csv", "Output format (csv|json)")
	flag.StringVar(&config.OutputDir, "output-dir", ".", "Output directory")
	flag.IntVar(&config.Concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent API requests (default: number of CPU cores)")
	flag.StringVar(&config.APIBaseURL, "api-base-url", "http://localhost:8080", "Base URL for Portal64 API")
	flag.IntVar(&config.Timeout, "timeout", 30, "API request timeout in seconds")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable detailed logging")
	flag.IntVar(&config.MinSampleSize, "min-sample-size", 10, "Minimum players per age/gender group")

	// Debug mode flags
	flag.IntVar(&config.DebugAge, "debug-age", 0, "Debug mode: filter players by exact age (disables normal processing)")
	flag.StringVar(&config.DebugGender, "debug-gender", "", "Debug mode: filter players by gender (m/w/d)")
	flag.StringVar(&config.DebugOutput, "debug-output", "debug-players.csv", "Debug mode: output CSV filename")

	flag.Parse()

	// Check if debug mode is requested
	debugMode := config.DebugAge > 0 || config.DebugGender != ""
	
	if debugMode {
		// Validate debug parameters
		if config.DebugAge <= 0 {
			fmt.Fprintf(os.Stderr, "Error: debug-age must be greater than 0, got: %d\n", config.DebugAge)
			flag.Usage()
			os.Exit(1)
		}
		
		if config.DebugGender == "" {
			fmt.Fprintf(os.Stderr, "Error: debug-gender is required in debug mode\n")
			flag.Usage()
			os.Exit(1)
		}
		
		if config.DebugGender != "m" && config.DebugGender != "w" && config.DebugGender != "d" {
			fmt.Fprintf(os.Stderr, "Error: debug-gender must be 'm', 'w', or 'd', got: %s\n", config.DebugGender)
			flag.Usage()
			os.Exit(1)
		}
		
		if config.DebugOutput == "" {
			config.DebugOutput = "debug-players.csv"
		}
		
		return config // Skip normal validation in debug mode
	}

	if config.OutputFormat != "csv" && config.OutputFormat != "json" {
		fmt.Fprintf(os.Stderr, "Error: output-format must be 'csv' or 'json', got: %s\n", config.OutputFormat)
		flag.Usage()
		os.Exit(1)
	}

	if config.MinSampleSize < 1 {
		fmt.Fprintf(os.Stderr, "Error: min-sample-size must be at least 1, got: %d\n", config.MinSampleSize)
		flag.Usage()
		os.Exit(1)
	}

	if config.Timeout < 1 {
		fmt.Fprintf(os.Stderr, "Error: timeout must be at least 1 second, got: %d\n", config.Timeout)
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func runDebugMode(client *api.Client, config *models.Config) error {
	// Fetch all players using the same API as normal mode
	allPlayers, err := client.FetchAllPlayers()
	if err != nil {
		return fmt.Errorf("failed to fetch players: %w", err)
	}

	if config.Verbose {
		fmt.Printf("Fetched %d total players\n", len(allPlayers))
	}

	// Filter and process debug players
	debugRecords, err := processDebugPlayers(allPlayers, config)
	if err != nil {
		return fmt.Errorf("failed to process debug players: %w", err)
	}

	if config.Verbose {
		fmt.Printf("Found %d players matching debug criteria (age=%d, gender=%s)\n", 
			len(debugRecords), config.DebugAge, config.DebugGender)
	}

	// Export to CSV
	outputPath := filepath.Join(config.OutputDir, config.DebugOutput)
	if err := exportDebugCSV(debugRecords, outputPath, config.Verbose); err != nil {
		return fmt.Errorf("failed to export debug CSV: %w", err)
	}

	if config.Verbose {
		fmt.Printf("Debug data exported to: %s\n", outputPath)
	}

	return nil
}

func processDebugPlayers(allPlayers []models.Player, config *models.Config) ([]models.DebugPlayerRecord, error) {
	currentYear := time.Now().Year()
	var debugRecords []models.DebugPlayerRecord

	for _, player := range allPlayers {
		// Apply same validation filters as normal processing
		if player.CurrentDWZ <= 0 {
			continue
		}

		if player.BirthYear == nil {
			continue
		}

		age := currentYear - *player.BirthYear
		if age < 4 || age > 100 {
			continue
		}

		// Filter by debug criteria
		if age != config.DebugAge {
			continue
		}

		if player.Gender != config.DebugGender {
			continue
		}

		// Create debug record
		record := models.DebugPlayerRecord{
			PlayerID:      player.ID,
			FirstName:     player.Firstname,
			LastName:      player.Name,
			ClubID:        player.ClubID,
			ClubName:      player.ClubName,
			BirthYear:     *player.BirthYear,
			CalculatedAge: age,
			CurrentDWZ:    player.CurrentDWZ,
			Gender:        player.Gender,
		}

		debugRecords = append(debugRecords, record)
	}

	// Sort by DWZ (highest first)
	sort.Slice(debugRecords, func(i, j int) bool {
		return debugRecords[i].CurrentDWZ > debugRecords[j].CurrentDWZ
	})

	return debugRecords, nil
}

func exportDebugCSV(records []models.DebugPlayerRecord, outputPath string, verbose bool) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write UTF-8 BOM to ensure proper encoding for German umlauts
	if _, err := file.Write([]byte("\xEF\xBB\xBF")); err != nil {
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	writer := csv.NewWriter(file)
	writer.Comma = ';' // Use semicolon separator for German Excel compatibility
	defer writer.Flush()

	// Write header
	header := []string{
		"player_id",
		"first_name", 
		"last_name",
		"club_id",
		"club_name",
		"birth_year",
		"calculated_age",
		"current_dwz",
		"gender",
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, record := range records {
		row := []string{
			record.PlayerID,
			record.FirstName,
			record.LastName,
			record.ClubID,
			record.ClubName,
			strconv.Itoa(record.BirthYear),
			strconv.Itoa(record.CalculatedAge),
			strconv.Itoa(record.CurrentDWZ),
			record.Gender,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	if verbose {
		fmt.Printf("Exported %d debug records\n", len(records))
	}

	return nil
}