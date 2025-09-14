package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/portal64/kader-planung/internal/statistics"
	"github.com/sirupsen/logrus"
)

// OutputFormat defines the different output formats available
type OutputFormat int

const (
	// Traditional Kader-Planung formats
	CSVDetailed OutputFormat = iota
	JSONDetailed
	ExcelDetailed

	// Statistical formats (from Somatogramm)
	CSVStatistical
	JSONStatistical

	// Combined formats
	CSVCombined
	JSONCombined
)

// String returns the string representation of an OutputFormat
func (f OutputFormat) String() string {
	switch f {
	case CSVDetailed:
		return "csv-detailed"
	case JSONDetailed:
		return "json-detailed"
	case ExcelDetailed:
		return "excel-detailed"
	case CSVStatistical:
		return "csv-statistical"
	case JSONStatistical:
		return "json-statistical"
	case CSVCombined:
		return "csv-combined"
	case JSONCombined:
		return "json-combined"
	default:
		return "csv-detailed"
	}
}

// ParseOutputFormat parses a string into an OutputFormat
func ParseOutputFormat(format string) OutputFormat {
	switch strings.ToLower(format) {
	case "csv", "csv-detailed":
		return CSVDetailed
	case "json", "json-detailed":
		return JSONDetailed
	case "excel", "excel-detailed":
		return ExcelDetailed
	case "csv-statistical":
		return CSVStatistical
	case "json-statistical":
		return JSONStatistical
	case "csv-combined":
		return CSVCombined
	case "json-combined":
		return JSONCombined
	default:
		return CSVDetailed
	}
}

// UnifiedExportConfig holds configuration for unified export
type UnifiedExportConfig struct {
	OutputDir         string
	Format            OutputFormat
	IncludeStatistics bool
	Prefix            string // File name prefix
	Timestamp         bool   // Include timestamp in filename
}

// UnifiedExporter handles exporting data in various formats
type UnifiedExporter struct {
	config *UnifiedExportConfig
	logger *logrus.Logger
}

// NewUnifiedExporter creates a new unified exporter
func NewUnifiedExporter(config *UnifiedExportConfig, logger *logrus.Logger) *UnifiedExporter {
	return &UnifiedExporter{
		config: config,
		logger: logger,
	}
}

// ExportResult contains information about exported files
type ExportResult struct {
	Files       []string      `json:"files"`
	TotalRecords int          `json:"total_records"`
	ExportTime   time.Duration `json:"export_time"`
}

// Export exports the processing results in the configured format
func (e *UnifiedExporter) Export(
	records []models.KaderPlanungRecord,
	statisticalData map[string]statistics.SomatogrammData,
	mode string,
) (*ExportResult, error) {
	startTime := time.Now()

	var files []string
	var err error

	switch e.config.Format {
	case CSVDetailed:
		files, err = e.exportCSVDetailed(records, mode)
	case JSONDetailed:
		files, err = e.exportJSONDetailed(records, mode)
	case CSVStatistical:
		files, err = e.exportCSVStatistical(statisticalData)
	case JSONStatistical:
		files, err = e.exportJSONStatistical(statisticalData)
	case CSVCombined:
		files, err = e.exportCSVCombined(records, statisticalData, mode)
	case JSONCombined:
		files, err = e.exportJSONCombined(records, statisticalData, mode)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", e.config.Format.String())
	}

	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	result := &ExportResult{
		Files:        files,
		TotalRecords: len(records),
		ExportTime:   duration,
	}

	e.logger.Infof("Export completed in %v, created %d files", duration, len(files))
	return result, nil
}

// exportCSVDetailed exports detailed records to CSV (traditional Kader-Planung format)
func (e *UnifiedExporter) exportCSVDetailed(records []models.KaderPlanungRecord, mode string) ([]string, error) {
	if len(records) == 0 {
		return []string{}, nil
	}

	filename := e.generateFilename(fmt.Sprintf("kader-planung-%s", mode), "csv")
	filepath := filepath.Join(e.config.OutputDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Write UTF-8 BOM to ensure proper encoding for German umlauts
	if _, err := file.Write([]byte("\xEF\xBB\xBF")); err != nil {
		return nil, fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	writer := csv.NewWriter(file)
	writer.Comma = ';' // Use semicolon separator for German Excel compatibility
	defer writer.Flush()

	// Write header
	header := []string{
		"club_id_prefix_1", "club_id_prefix_2", "club_id_prefix_3", "club_name", "club_id",
		"player_id", "pkz", "lastname", "firstname", "birthyear", "gender",
		"current_dwz", "list_ranking", "dwz_12_months_ago", "games_last_12_months",
		"success_rate_last_12_months", "somatogram_percentile", "dwz_age_relation",
	}

	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, record := range records {
		row := []string{
			record.ClubIDPrefix1,
			record.ClubIDPrefix2,
			record.ClubIDPrefix3,
			record.ClubName,
			record.ClubID,
			record.PlayerID,
			record.PKZ,
			record.Lastname,
			record.Firstname,
			strconv.Itoa(record.Birthyear),
			record.Gender,
			strconv.Itoa(record.CurrentDWZ),
			strconv.Itoa(record.ListRanking),
			record.DWZ12MonthsAgo,
			record.GamesLast12Months,
			record.SuccessRateLast12Months,
			record.SomatogramPercentile,
			record.DWZAgeRelation,
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	e.logger.Infof("Exported %d detailed records to %s", len(records), filename)
	return []string{filename}, nil
}

// exportJSONDetailed exports detailed records to JSON
func (e *UnifiedExporter) exportJSONDetailed(records []models.KaderPlanungRecord, mode string) ([]string, error) {
	if len(records) == 0 {
		return []string{}, nil
	}

	filename := e.generateFilename(fmt.Sprintf("kader-planung-%s", mode), "json")
	filepath := filepath.Join(e.config.OutputDir, filename)

	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"generated_at":  time.Now(),
			"mode":         mode,
			"total_records": len(records),
		},
		"records": records,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write JSON file: %w", err)
	}

	e.logger.Infof("Exported %d detailed records to %s", len(records), filename)
	return []string{filename}, nil
}

// exportCSVStatistical exports statistical data to CSV files (one per gender)
func (e *UnifiedExporter) exportCSVStatistical(statisticalData map[string]statistics.SomatogrammData) ([]string, error) {
	var files []string

	for gender, data := range statisticalData {
		if len(data.Percentiles) == 0 {
			continue
		}

		genderName := e.getGenderName(gender)
		filename := e.generateFilename(fmt.Sprintf("kader-planung-statistical-%s", genderName), "csv")
		filepath := filepath.Join(e.config.OutputDir, filename)

		file, err := os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create CSV file for gender %s: %w", gender, err)
		}
		defer file.Close()

		// Write UTF-8 BOM to ensure proper encoding for German umlauts
		if _, err := file.Write([]byte("\xEF\xBB\xBF")); err != nil {
			return nil, fmt.Errorf("failed to write UTF-8 BOM: %w", err)
		}

		writer := csv.NewWriter(file)
		writer.Comma = ';' // Use semicolon separator for German Excel compatibility
		defer writer.Flush()

		// Write metadata header
		metadataRows := [][]string{
			{"# Somatogramm Statistical Data"},
			{"# Generated:", data.Metadata.GeneratedAt.Format("2006-01-02 15:04:05")},
			{"# Gender:", genderName},
			{"# Total Players:", strconv.Itoa(data.Metadata.TotalPlayers)},
			{"# Valid Age Groups:", strconv.Itoa(data.Metadata.ValidAgeGroups)},
			{"# Min Sample Size:", strconv.Itoa(data.Metadata.MinSampleSize)},
			{""},
		}

		for _, row := range metadataRows {
			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write metadata: %w", err)
			}
		}

		// Write header for percentile data
		header := []string{"age", "sample_size", "avg_dwz", "median_dwz"}
		for p := 0; p <= 100; p += 5 { // Every 5th percentile
			header = append(header, fmt.Sprintf("p%d", p))
		}

		if err := writer.Write(header); err != nil {
			return nil, fmt.Errorf("failed to write header: %w", err)
		}

		// Sort ages for consistent output
		var ages []int
		for ageStr := range data.Percentiles {
			age, _ := strconv.Atoi(ageStr)
			ages = append(ages, age)
		}
		sort.Ints(ages)

		// Write data rows
		for _, age := range ages {
			ageStr := strconv.Itoa(age)
			percentileData := data.Percentiles[ageStr]

			row := []string{
				ageStr,
				strconv.Itoa(percentileData.SampleSize),
				fmt.Sprintf("%.1f", percentileData.AvgDWZ),
				strconv.Itoa(percentileData.MedianDWZ),
			}

			// Add percentile values
			for p := 0; p <= 100; p += 5 {
				if val, exists := percentileData.Percentiles[p]; exists {
					row = append(row, strconv.Itoa(val))
				} else {
					row = append(row, "")
				}
			}

			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write data row: %w", err)
			}
		}

		files = append(files, filename)
		e.logger.Infof("Exported statistical data for gender %s to %s (%d age groups)", genderName, filename, len(ages))
	}

	return files, nil
}

// exportJSONStatistical exports statistical data to JSON files (one per gender)
func (e *UnifiedExporter) exportJSONStatistical(statisticalData map[string]statistics.SomatogrammData) ([]string, error) {
	var files []string

	for gender, data := range statisticalData {
		if len(data.Percentiles) == 0 {
			continue
		}

		genderName := e.getGenderName(gender)
		filename := e.generateFilename(fmt.Sprintf("kader-planung-statistical-%s", genderName), "json")
		filepath := filepath.Join(e.config.OutputDir, filename)

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for gender %s: %w", gender, err)
		}

		if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write JSON file for gender %s: %w", gender, err)
		}

		files = append(files, filename)
		e.logger.Infof("Exported statistical data for gender %s to %s", genderName, filename)
	}

	return files, nil
}

// exportCSVCombined exports both detailed records and statistical data
func (e *UnifiedExporter) exportCSVCombined(
	records []models.KaderPlanungRecord,
	statisticalData map[string]statistics.SomatogrammData,
	mode string,
) ([]string, error) {
	var files []string

	// Export detailed records
	if len(records) > 0 {
		detailedFiles, err := e.exportCSVDetailed(records, mode)
		if err != nil {
			return nil, fmt.Errorf("failed to export detailed CSV: %w", err)
		}
		files = append(files, detailedFiles...)
	}

	// Export statistical data
	if len(statisticalData) > 0 {
		statisticalFiles, err := e.exportCSVStatistical(statisticalData)
		if err != nil {
			return nil, fmt.Errorf("failed to export statistical CSV: %w", err)
		}
		files = append(files, statisticalFiles...)
	}

	return files, nil
}

// exportJSONCombined exports both detailed records and statistical data to JSON
func (e *UnifiedExporter) exportJSONCombined(
	records []models.KaderPlanungRecord,
	statisticalData map[string]statistics.SomatogrammData,
	mode string,
) ([]string, error) {
	filename := e.generateFilename(fmt.Sprintf("kader-planung-combined-%s", mode), "json")
	filepath := filepath.Join(e.config.OutputDir, filename)

	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"generated_at":     time.Now(),
			"mode":            mode,
			"total_records":   len(records),
			"has_statistics":  len(statisticalData) > 0,
		},
	}

	if len(records) > 0 {
		data["records"] = records
	}

	if len(statisticalData) > 0 {
		data["statistical_data"] = statisticalData
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal combined JSON: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write combined JSON file: %w", err)
	}

	e.logger.Infof("Exported combined data to %s", filename)
	return []string{filename}, nil
}

// generateFilename generates a filename with optional timestamp
func (e *UnifiedExporter) generateFilename(base, extension string) string {
	if e.config.Timestamp {
		timestamp := time.Now().Format("20060102-150405")
		return fmt.Sprintf("%s-%s.%s", base, timestamp, extension)
	}
	return fmt.Sprintf("%s.%s", base, extension)
}

// getGenderName converts gender code to readable name
func (e *UnifiedExporter) getGenderName(gender string) string {
	switch gender {
	case "m":
		return "male"
	case "w":
		return "female"
	case "d":
		return "divers"
	default:
		return "unknown"
	}
}