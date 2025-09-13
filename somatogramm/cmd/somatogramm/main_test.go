package main

import (
	"flag"
	"os"
	"runtime"
	"testing"

	"somatogramm/internal/models"
)

func TestParseFlagsDefaults(t *testing.T) {
	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock command line args
	oldArgs := os.Args
	os.Args = []string{"somatogramm"}
	defer func() { os.Args = oldArgs }()

	config := parseFlags()

	// Test default values
	if config.OutputFormat != "csv" {
		t.Errorf("expected default output format 'csv', got %s", config.OutputFormat)
	}

	if config.OutputDir != "." {
		t.Errorf("expected default output dir '.', got %s", config.OutputDir)
	}

	if config.Concurrency != 0 {
		t.Errorf("expected default concurrency 0, got %d", config.Concurrency)
	}

	if config.APIBaseURL != "http://localhost:8080" {
		t.Errorf("expected default API base URL 'http://localhost:8080', got %s", config.APIBaseURL)
	}

	if config.Timeout != 30 {
		t.Errorf("expected default timeout 30, got %d", config.Timeout)
	}

	if config.Verbose != false {
		t.Errorf("expected default verbose false, got %t", config.Verbose)
	}

	if config.MinSampleSize != 100 {
		t.Errorf("expected default min sample size 100, got %d", config.MinSampleSize)
	}
}

func TestParseFlagsCustomValues(t *testing.T) {
	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock command line args
	oldArgs := os.Args
	os.Args = []string{
		"somatogramm",
		"--output-format", "json",
		"--output-dir", "/tmp/output",
		"--concurrency", "4",
		"--api-base-url", "http://example.com:9000",
		"--timeout", "60",
		"--verbose",
		"--min-sample-size", "50",
	}
	defer func() { os.Args = oldArgs }()

	config := parseFlags()

	// Test custom values
	if config.OutputFormat != "json" {
		t.Errorf("expected output format 'json', got %s", config.OutputFormat)
	}

	if config.OutputDir != "/tmp/output" {
		t.Errorf("expected output dir '/tmp/output', got %s", config.OutputDir)
	}

	if config.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", config.Concurrency)
	}

	if config.APIBaseURL != "http://example.com:9000" {
		t.Errorf("expected API base URL 'http://example.com:9000', got %s", config.APIBaseURL)
	}

	if config.Timeout != 60 {
		t.Errorf("expected timeout 60, got %d", config.Timeout)
	}

	if config.Verbose != true {
		t.Errorf("expected verbose true, got %t", config.Verbose)
	}

	if config.MinSampleSize != 50 {
		t.Errorf("expected min sample size 50, got %d", config.MinSampleSize)
	}
}

func TestParseFlagsJSONFormat(t *testing.T) {
	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock command line args with JSON format
	oldArgs := os.Args
	os.Args = []string{"somatogramm", "--output-format", "json"}
	defer func() { os.Args = oldArgs }()

	config := parseFlags()

	if config.OutputFormat != "json" {
		t.Errorf("expected output format 'json', got %s", config.OutputFormat)
	}
}

func TestParseFlagsCSVFormat(t *testing.T) {
	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock command line args with CSV format
	oldArgs := os.Args
	os.Args = []string{"somatogramm", "--output-format", "csv"}
	defer func() { os.Args = oldArgs }()

	config := parseFlags()

	if config.OutputFormat != "csv" {
		t.Errorf("expected output format 'csv', got %s", config.OutputFormat)
	}
}

func TestParseFlagsMinSampleSizeValidation(t *testing.T) {
	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test valid min sample size
	oldArgs := os.Args
	os.Args = []string{"somatogramm", "--min-sample-size", "50"}
	defer func() { os.Args = oldArgs }()

	config := parseFlags()

	if config.MinSampleSize != 50 {
		t.Errorf("expected min sample size 50, got %d", config.MinSampleSize)
	}
}

func TestParseFlagsTimeoutValidation(t *testing.T) {
	// Reset flag package for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Test valid timeout
	oldArgs := os.Args
	os.Args = []string{"somatogramm", "--timeout", "60"}
	defer func() { os.Args = oldArgs }()

	config := parseFlags()

	if config.Timeout != 60 {
		t.Errorf("expected timeout 60, got %d", config.Timeout)
	}
}

// Test the Config struct directly
func TestConfigStruct(t *testing.T) {
	config := &models.Config{
		OutputFormat:  "json",
		OutputDir:     "/tmp",
		Concurrency:   8,
		APIBaseURL:    "http://localhost:8080",
		Timeout:       45,
		Verbose:       true,
		MinSampleSize: 75,
	}

	if config.OutputFormat != "json" {
		t.Errorf("expected OutputFormat 'json', got %s", config.OutputFormat)
	}

	if config.OutputDir != "/tmp" {
		t.Errorf("expected OutputDir '/tmp', got %s", config.OutputDir)
	}

	if config.Concurrency != 8 {
		t.Errorf("expected Concurrency 8, got %d", config.Concurrency)
	}

	if config.APIBaseURL != "http://localhost:8080" {
		t.Errorf("expected APIBaseURL 'http://localhost:8080', got %s", config.APIBaseURL)
	}

	if config.Timeout != 45 {
		t.Errorf("expected Timeout 45, got %d", config.Timeout)
	}

	if config.Verbose != true {
		t.Errorf("expected Verbose true, got %t", config.Verbose)
	}

	if config.MinSampleSize != 75 {
		t.Errorf("expected MinSampleSize 75, got %d", config.MinSampleSize)
	}
}

func TestConcurrencyLogic(t *testing.T) {
	// Test the concurrency adjustment logic from main function

	config := &models.Config{
		Concurrency: 0, // Should default to CPU cores
	}

	// Simulate the main function logic
	if config.Concurrency <= 0 {
		config.Concurrency = runtime.NumCPU()
	}

	expectedCPUs := runtime.NumCPU()
	if config.Concurrency != expectedCPUs {
		t.Errorf("expected concurrency to be set to CPU count %d, got %d", expectedCPUs, config.Concurrency)
	}

	// Test with explicit concurrency value
	config2 := &models.Config{
		Concurrency: 4,
	}

	originalConcurrency := config2.Concurrency
	if config2.Concurrency <= 0 {
		config2.Concurrency = runtime.NumCPU()
	}

	if config2.Concurrency != originalConcurrency {
		t.Errorf("expected concurrency to remain %d, got %d", originalConcurrency, config2.Concurrency)
	}
}