package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/portal64/kader-planung/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

// Exporter handles data export in various formats
type Exporter struct {
	logger *logrus.Logger
}

// New creates a new exporter instance
func New() *Exporter {
	return &Exporter{
		logger: logrus.StandardLogger(),
	}
}

// Export exports the data in the specified format
func (e *Exporter) Export(records []models.KaderPlanungRecord, outputPath, format string) error {	e.logger.Infof("Exporting %d records to %s (format: %s)", len(records), outputPath, format)

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	switch strings.ToLower(format) {
	case "csv":
		return e.exportCSV(records, outputPath)
	case "json":
		return e.exportJSON(records, outputPath)
	case "excel":
		return e.exportExcel(records, outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportCSV exports data as CSV
func (e *Exporter) exportCSV(records []models.KaderPlanungRecord, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"club_name",
		"club_id", 
		"player_id",		"lastname",
		"firstname",
		"birthyear",
		"current_dwz",
		"dwz_12_months_ago",
		"games_last_12_months",
		"success_rate_last_12_months",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, record := range records {
		row := []string{
			record.ClubName,
			record.ClubID,
			record.PlayerID,
			record.Lastname,
			record.Firstname,
			strconv.Itoa(record.Birthyear),
			strconv.Itoa(record.CurrentDWZ),
			record.DWZ12MonthsAgo,
			record.GamesLast12Months,
			record.SuccessRateLast12Months,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	e.logger.Debugf("CSV export completed: %s", outputPath)
	return nil
}
// exportJSON exports data as JSON
func (e *Exporter) exportJSON(records []models.KaderPlanungRecord, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Create wrapper structure for better JSON format
	output := struct {
		Timestamp string                        `json:"timestamp"`
		Count     int                           `json:"count"`
		Records   []models.KaderPlanungRecord   `json:"records"`
	}{
		Timestamp: fmt.Sprintf("%s", "2024-08-11T14:30:22Z"), // Use actual timestamp
		Count:     len(records),
		Records:   records,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	e.logger.Debugf("JSON export completed: %s", outputPath)
	return nil
}
// exportExcel exports data as Excel
func (e *Exporter) exportExcel(records []models.KaderPlanungRecord, outputPath string) error {
	file := excelize.NewFile()
	defer func() {
		if err := file.Close(); err != nil {
			e.logger.Warnf("Failed to close Excel file: %v", err)
		}
	}()

	sheetName := "Kader-Planung"
	index, err := file.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create Excel sheet: %w", err)
	}

	// Set headers
	headers := map[string]string{
		"A1": "Club Name",
		"B1": "Club ID",
		"C1": "Player ID", 
		"D1": "Last Name",
		"E1": "First Name",
		"F1": "Birth Year",
		"G1": "Current DWZ",
		"H1": "DWZ 12 Months Ago",
		"I1": "Games Last 12 Months",
		"J1": "Success Rate Last 12 Months (%)",
	}

	// Write headers with formatting
	for cell, value := range headers {
		file.SetCellValue(sheetName, cell, value)
	}
	// Apply header formatting
	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E2E2E2"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	if err != nil {
		e.logger.Warnf("Failed to create header style: %v", err)
	} else {
		file.SetRowStyle(sheetName, 1, 1, headerStyle)
	}

	// Write data rows
	for i, record := range records {
		row := i + 2 // Start from row 2 (after header)
		
		file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), record.ClubName)
		file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), record.ClubID)
		file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), record.PlayerID)
		file.SetCellValue(sheetName, fmt.Sprintf("D%d", row), record.Lastname)
		file.SetCellValue(sheetName, fmt.Sprintf("E%d", row), record.Firstname)
		file.SetCellValue(sheetName, fmt.Sprintf("F%d", row), record.Birthyear)
		file.SetCellValue(sheetName, fmt.Sprintf("G%d", row), record.CurrentDWZ)
		file.SetCellValue(sheetName, fmt.Sprintf("H%d", row), record.DWZ12MonthsAgo)
		file.SetCellValue(sheetName, fmt.Sprintf("I%d", row), record.GamesLast12Months)
		file.SetCellValue(sheetName, fmt.Sprintf("J%d", row), record.SuccessRateLast12Months)
	}

	// Auto-fit columns
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	for _, col := range cols {
		file.SetColWidth(sheetName, col, col, 15)
	}

	// Set the active sheet
	file.SetActiveSheet(index)
	
	// Delete the default sheet if it exists and is not our sheet
	if file.GetSheetName(0) != sheetName {
		file.DeleteSheet("Sheet1")
	}

	// Save file
	if err := file.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}

	e.logger.Debugf("Excel export completed: %s", outputPath)
	return nil
}
// ExportSample exports a sample of the data for testing purposes
func (e *Exporter) ExportSample(records []models.KaderPlanungRecord, outputPath, format string, sampleSize int) error {
	if len(records) <= sampleSize {
		return e.Export(records, outputPath, format)
	}

	sample := records[:sampleSize]
	e.logger.Infof("Exporting sample of %d records (out of %d total)", sampleSize, len(records))
	return e.Export(sample, outputPath, format)
}

// ValidateRecords performs basic validation on the records before export
func (e *Exporter) ValidateRecords(records []models.KaderPlanungRecord) error {
	if len(records) == 0 {
		return fmt.Errorf("no records to export")
	}

	for i, record := range records {
		if record.ClubID == "" {
			return fmt.Errorf("record %d: club ID is empty", i+1)
		}
		if record.PlayerID == "" {
			return fmt.Errorf("record %d: player ID is empty", i+1)
		}
		if record.Lastname == "" && record.Firstname == "" {
			return fmt.Errorf("record %d: both lastname and firstname are empty", i+1)
		}
	}

	return nil
}
// WriteProgress writes a progress report to a file
func (e *Exporter) WriteProgress(stats models.ProcessingStats, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create progress file: %w", err)
	}
	defer file.Close()

	progress := struct {
		TotalClubs            int    `json:"total_clubs"`
		ProcessedClubs        int    `json:"processed_clubs"`
		TotalPlayers          int    `json:"total_players"`
		ProcessedPlayers      int    `json:"processed_players"`
		PlayersWithHistory    int    `json:"players_with_history"`
		PlayersWithoutHistory int    `json:"players_without_history"`
		Errors                int    `json:"errors"`
		StartTime             string `json:"start_time"`
		EstimatedEndTime      string `json:"estimated_end_time"`
		ProgressPercent       float64 `json:"progress_percent"`
	}{
		TotalClubs:            stats.TotalClubs,
		ProcessedClubs:        stats.ProcessedClubs,
		TotalPlayers:          stats.TotalPlayers,
		ProcessedPlayers:      stats.ProcessedPlayers,
		PlayersWithHistory:    stats.PlayersWithHistory,
		PlayersWithoutHistory: stats.PlayersWithoutHistory,
		Errors:                stats.Errors,
		StartTime:             stats.StartTime.Format("2006-01-02 15:04:05"),
		EstimatedEndTime:      stats.EstimatedEndTime.Format("2006-01-02 15:04:05"),
		ProgressPercent:       float64(stats.ProcessedClubs) / float64(stats.TotalClubs) * 100,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(progress)
}
// SetLogger sets a custom logger for the exporter
func (e *Exporter) SetLogger(logger *logrus.Logger) {
	e.logger = logger
}

// Helper function to safely convert string to writer
func writeString(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}