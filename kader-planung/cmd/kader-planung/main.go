package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	outputDir      string
	concurrency    int
	resumeFlag     bool
	checkpointFile string
	apiBaseURL     string
	timeout        int
	verbose        bool
	// Somatogram integration parameters
	minSampleSize  int
)

var rootCmd = &cobra.Command{
	Use:   "kader-planung",
	Short: "Generate comprehensive player roster reports with somatogram percentiles",
	Long: `Kader-Planung generates detailed CSV reports with player data, statistical analysis,
and somatogram percentile rankings for chess club management purposes.

The application fetches data from the Portal64 REST API and produces CSV reports containing:
- Club information and player details
- Current DWZ ratings and historical data (12 months)
- Game statistics and success rates
- Somatogram percentiles (Germany-wide rankings by age/gender)

All reports are generated in CSV format optimized for German Excel compatibility.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runKaderPlanung(); err != nil {
			logrus.Fatalf("Application failed: %v", err)
		}
	},
}

func init() {
	// Essential parameters
	rootCmd.Flags().StringVar(&clubPrefix, "club-prefix", "", "Filter clubs by ID prefix (e.g., 'C0' for C0327, C0401, etc.)")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output directory")
	rootCmd.Flags().IntVar(&concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent API requests")
	rootCmd.Flags().BoolVar(&resumeFlag, "resume", false, "Resume from previous run using checkpoint file")
	rootCmd.Flags().StringVar(&checkpointFile, "checkpoint-file", "", "Custom checkpoint file path")
	rootCmd.Flags().StringVar(&apiBaseURL, "api-base-url", "http://localhost:8080", "Base URL for Portal64 API")
	rootCmd.Flags().IntVar(&timeout, "timeout", 30, "API request timeout in seconds")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable detailed logging")
	
	// Somatogram percentile calculation
	rootCmd.Flags().IntVar(&minSampleSize, "min-sample-size", 10, "Minimum sample size for somatogram percentile accuracy")
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

	logrus.Info("Starting Kader-Planung data collection with somatogram percentiles...")

	// Validate arguments
	if err := validateArgs(); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	// Hardcode processing mode to hybrid (always include both historical data and somatogram percentiles)
	mode, err := processor.ParseProcessingMode("hybrid")
	if err != nil {
		return fmt.Errorf("invalid processing mode: %w", err)
	}

	// Setup checkpoint file path
	var checkpoint *resume.Checkpoint
	var resumeManager *resume.Manager

	if checkpointFile == "" {
		timestamp := time.Now().Format("20060102-150405")
		checkpointFile = filepath.Join(outputDir, fmt.Sprintf("kader-planung-checkpoint-%s.json", timestamp))
	}

	// Initialize resume manager
	resumeManager = resume.NewManager(checkpointFile)

	// Load or create checkpoint
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

	// Initialize API client with concurrency support
	apiClient := api.NewClientWithConcurrency(apiBaseURL, time.Duration(timeout)*time.Second, concurrency)

	// Configure unified processor (always hybrid mode with statistics enabled)
	config := &processor.UnifiedProcessorConfig{
		Mode:            mode,
		MinSampleSize:   minSampleSize,
		EnableStatistics: true, // Always enable for somatogram percentiles
		ClubPrefix:      clubPrefix,
		Concurrency:     concurrency,
		Verbose:         verbose,
	}

	// Initialize unified processor
	unifiedProcessor := processor.NewUnifiedProcessor(apiClient, resumeManager, config)
	if verbose {
		unifiedProcessor.SetLogger(logrus.StandardLogger())
	}

	// Process data
	result, err := unifiedProcessor.ProcessData(checkpoint)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}

	logrus.Infof("Processing completed: %d players, %d clubs, took %v",
		result.TotalPlayers, result.ProcessedClubs, result.ProcessingTime)

	// Configure unified export (always CSV with statistics)
	exportFormat := export.ParseOutputFormat("csv")
	exportConfig := &export.UnifiedExportConfig{
		OutputDir:         outputDir,
		Format:           exportFormat,
		IncludeStatistics: true, // Always include somatogram percentiles
		Timestamp:        true,  // Always include timestamp
	}

	// Initialize unified exporter
	exporter := export.NewUnifiedExporter(exportConfig, logrus.StandardLogger())

	// Export results
	exportResult, err := exporter.Export(result.Records, result.StatisticalData, mode.String())
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Report export results
	logrus.Infof("Export completed in %v, created %d files:", exportResult.ExportTime, len(exportResult.Files))
	for _, file := range exportResult.Files {
		logrus.Infof("  - %s", file)
	}

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
	// Validate minimum sample size
	if minSampleSize < 1 {
		return fmt.Errorf("minimum sample size must be at least 1")
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
			OutputFormat: "csv", // Always CSV
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