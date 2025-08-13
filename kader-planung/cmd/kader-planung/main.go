package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/portal64/kader-planung/internal/api"
	"github.com/portal64/kader-planung/internal/export"
	"github.com/portal64/kader-planung/internal/processor"
	"github.com/portal64/kader-planung/internal/resume"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	clubPrefix     string
	outputFormat   string
	outputDir      string
	concurrency    int
	resumeFlag     bool
	checkpointFile string
	apiBaseURL     string
	timeout        int
	verbose        bool
)

var rootCmd = &cobra.Command{	Use:   "kader-planung",
	Short: "Generate comprehensive player roster reports from Portal64 API",
	Long: `Kader-Planung is a standalone application that generates detailed CSV/JSON/Excel 
reports with historical player data for club management purposes.

The application fetches data from the Portal64 REST API and produces reports containing
club information, player details, current ratings, historical ratings, and performance
statistics.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runKaderPlanung(); err != nil {
			logrus.Fatalf("Application failed: %v", err)
		}
	},
}

func init() {
	rootCmd.Flags().StringVar(&clubPrefix, "club-prefix", "", "Filter clubs by ID prefix (e.g., 'C0' for C0327, C0401, etc.)")
	rootCmd.Flags().StringVar(&outputFormat, "output-format", "csv", "Output format (csv|json|excel)")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output directory")
	rootCmd.Flags().IntVar(&concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent API requests")
	rootCmd.Flags().BoolVar(&resumeFlag, "resume", false, "Resume from previous run using checkpoint file")
	rootCmd.Flags().StringVar(&checkpointFile, "checkpoint-file", "", "Custom checkpoint file path")
	rootCmd.Flags().StringVar(&apiBaseURL, "api-base-url", "http://localhost:8080", "Base URL for Portal64 API")
	rootCmd.Flags().IntVar(&timeout, "timeout", 30, "API request timeout in seconds")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable detailed logging")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
func runKaderPlanung() error {
	// Configure logging
	setupLogging()
	
	logrus.Info("Starting Kader-Planung data collection...")
	
	// Validate arguments
	if err := validateArgs(); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	
	// Setup checkpoint file path
	if checkpointFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		checkpointFile = filepath.Join(outputDir, fmt.Sprintf("kader-planung-checkpoint-%s.json", timestamp))
	}
	
	// Initialize API client
	apiClient := api.NewClient(apiBaseURL, time.Duration(timeout)*time.Second)
	
	// Initialize resume manager
	resumeManager := resume.NewManager(checkpointFile)
	
	// Initialize processor
	proc := processor.New(apiClient, resumeManager, concurrency)	
	// Load or create checkpoint
	var checkpoint *resume.Checkpoint
	var err error
	
	if resumeFlag {
		logrus.Info("Attempting to resume from previous run...")
		checkpoint, err = resumeManager.LoadCheckpoint()
		if err != nil {
			logrus.Warnf("Could not load checkpoint: %v. Starting fresh.", err)
			checkpoint = createNewCheckpoint()
		} else {
			logrus.Infof("Resumed from checkpoint: %d/%d clubs processed", 
				checkpoint.Progress.ProcessedClubs, checkpoint.Progress.TotalClubs)
		}
	} else {
		checkpoint = createNewCheckpoint()
	}
	
	// Process data
	results, err := proc.ProcessKaderPlanung(checkpoint, clubPrefix)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}
	
	// Generate output filename
	timestamp := time.Now().Format("20060102-150405")
	prefix := clubPrefix
	if prefix == "" {
		prefix = "all"
	}	
	var filename string
	switch outputFormat {
	case "csv":
		filename = fmt.Sprintf("kader-planung-%s-%s.csv", prefix, timestamp)
	case "json":
		filename = fmt.Sprintf("kader-planung-%s-%s.json", prefix, timestamp)
	case "excel":
		filename = fmt.Sprintf("kader-planung-%s-%s.xlsx", prefix, timestamp)
	}
	
	outputPath := filepath.Join(outputDir, filename)
	
	// Export results
	exporter := export.New()
	if err := exporter.Export(results, outputPath, outputFormat); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	
	logrus.Infof("Export complete: %s (%d players)", outputPath, len(results))
	
	// Clean up checkpoint file on success
	if err := resumeManager.Cleanup(); err != nil {
		logrus.Warnf("Could not clean up checkpoint file: %v", err)
	}
	
	return nil
}
func setupLogging() {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
}

func validateArgs() error {
	// Validate output format
	validFormats := []string{"csv", "json", "excel"}
	found := false
	for _, format := range validFormats {
		if format == strings.ToLower(outputFormat) {
			outputFormat = strings.ToLower(outputFormat)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid output format: %s (valid: %v)", outputFormat, validFormats)
	}	
	// Validate concurrency
	if concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}
	
	// Validate timeout
	if timeout < 1 {
		return fmt.Errorf("timeout must be at least 1 second")
	}
	
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}
	
	return nil
}

func createNewCheckpoint() *resume.Checkpoint {
	return &resume.Checkpoint{
		Timestamp: time.Now(),
		Config: resume.Config{
			ClubPrefix:   clubPrefix,
			OutputFormat: outputFormat,
			Concurrency:  concurrency,
		},
		Progress: resume.Progress{
			TotalClubs:     0,
			ProcessedClubs: 0,
			CurrentPhase:   "initialization",
		},
		ProcessedItems: make([]resume.ProcessedItem, 0),
		PartialData:    make([]resume.PartialPlayerData, 0),
	}
}