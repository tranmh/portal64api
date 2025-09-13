package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"somatogramm/internal/models"
)

type Exporter struct {
	OutputDir    string
	OutputFormat string
	MaxVersions  int
	Verbose      bool
}

func NewExporter(outputDir, outputFormat string, maxVersions int, verbose bool) *Exporter {
	return &Exporter{
		OutputDir:    outputDir,
		OutputFormat: outputFormat,
		MaxVersions:  maxVersions,
		Verbose:      verbose,
	}
}

func (e *Exporter) log(message string) {
	if e.Verbose {
		fmt.Printf("[EXPORT] %s\n", message)
	}
}

func (e *Exporter) ExportData(data map[string]models.SomatogrammData) error {
	e.log("Starting export process...")

	if err := os.MkdirAll(e.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")

	for gender, somatogramm := range data {
		if err := e.exportGenderData(gender, somatogramm, timestamp); err != nil {
			return fmt.Errorf("failed to export data for gender %s: %w", gender, err)
		}
	}

	if err := e.cleanupOldFiles(); err != nil {
		e.log(fmt.Sprintf("Warning: failed to cleanup old files: %v", err))
	}

	e.log("Export process completed successfully")
	return nil
}

func (e *Exporter) exportGenderData(gender string, data models.SomatogrammData, timestamp string) error {
	genderName := e.getGenderName(gender)

	switch e.OutputFormat {
	case "json":
		return e.exportJSON(genderName, data, timestamp)
	case "csv":
		return e.exportCSV(genderName, data, timestamp)
	default:
		return fmt.Errorf("unsupported output format: %s", e.OutputFormat)
	}
}

func (e *Exporter) exportJSON(genderName string, data models.SomatogrammData, timestamp string) error {
	filename := fmt.Sprintf("somatogramm-%s-%s.json", genderName, timestamp)
	filepath := filepath.Join(e.OutputDir, filename)

	e.log(fmt.Sprintf("Exporting JSON: %s", filename))

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	e.log(fmt.Sprintf("Successfully exported JSON: %s", filepath))
	return nil
}

func (e *Exporter) exportCSV(genderName string, data models.SomatogrammData, timestamp string) error {
	filename := fmt.Sprintf("somatogramm-%s-%s.csv", genderName, timestamp)
	filepath := filepath.Join(e.OutputDir, filename)

	e.log(fmt.Sprintf("Exporting CSV: %s", filename))

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"age", "sample_size", "avg_dwz", "median_dwz"}
	for p := 0; p <= 100; p++ {
		headers = append(headers, fmt.Sprintf("p%d", p))
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	ages := make([]int, 0, len(data.Percentiles))
	for ageStr := range data.Percentiles {
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			continue
		}
		ages = append(ages, age)
	}
	sort.Ints(ages)

	for _, age := range ages {
		percentileData := data.Percentiles[strconv.Itoa(age)]

		record := []string{
			strconv.Itoa(percentileData.Age),
			strconv.Itoa(percentileData.SampleSize),
			fmt.Sprintf("%.2f", percentileData.AvgDWZ),
			strconv.Itoa(percentileData.MedianDWZ),
		}

		for p := 0; p <= 100; p++ {
			if value, exists := percentileData.Percentiles[p]; exists {
				record = append(record, strconv.Itoa(value))
			} else {
				record = append(record, "")
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	e.log(fmt.Sprintf("Successfully exported CSV: %s", filepath))
	return nil
}

func (e *Exporter) getGenderName(gender string) string {
	switch gender {
	case "m":
		return "male"
	case "w":
		return "female"
	case "d":
		return "divers"
	default:
		return gender
	}
}

func (e *Exporter) cleanupOldFiles() error {
	if e.MaxVersions <= 0 {
		return nil
	}

	e.log(fmt.Sprintf("Cleaning up old files, keeping %d versions", e.MaxVersions))

	files, err := os.ReadDir(e.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	genderFiles := make(map[string][]os.DirEntry)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if len(name) < 20 || !file.Type().IsRegular() {
			continue
		}

		if name[:12] == "somatogramm-" {
			gender := ""
			if len(name) > 17 && name[12:17] == "male-" {
				gender = "male"
			} else if len(name) > 19 && name[12:19] == "female-" {
				gender = "female"
			} else if len(name) > 19 && name[12:19] == "divers-" {
				gender = "divers"
			}

			if gender != "" {
				genderFiles[gender] = append(genderFiles[gender], file)
			}
		}
	}

	for gender, fileList := range genderFiles {
		if len(fileList) <= e.MaxVersions {
			continue
		}

		sort.Slice(fileList, func(i, j int) bool {
			iInfo, _ := fileList[i].Info()
			jInfo, _ := fileList[j].Info()
			return iInfo.ModTime().After(jInfo.ModTime())
		})

		filesToDelete := fileList[e.MaxVersions:]
		for _, file := range filesToDelete {
			filePath := filepath.Join(e.OutputDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				e.log(fmt.Sprintf("Failed to delete old file %s: %v", filePath, err))
			} else {
				e.log(fmt.Sprintf("Deleted old file: %s", file.Name()))
			}
		}

		e.log(fmt.Sprintf("Cleaned up %d old files for gender: %s", len(filesToDelete), gender))
	}

	return nil
}

func (e *Exporter) ListFiles() ([]FileInfo, error) {
	files, err := os.ReadDir(e.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %w", err)
	}

	var fileInfos []FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if len(name) < 20 || !file.Type().IsRegular() {
			continue
		}

		if name[:12] == "somatogramm-" {
			info, err := file.Info()
			if err != nil {
				continue
			}

			fileInfos = append(fileInfos, FileInfo{
				Name:         name,
				Size:         info.Size(),
				ModTime:      info.ModTime(),
				IsJSON:       filepath.Ext(name) == ".json",
			})
		}
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime.After(fileInfos[j].ModTime)
	})

	return fileInfos, nil
}

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsJSON  bool      `json:"is_json"`
}