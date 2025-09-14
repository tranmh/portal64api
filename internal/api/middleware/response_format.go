package middleware

import (
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResponseFormat middleware handles content negotiation for JSON/CSV responses
func ResponseFormat() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original response writer
		originalWriter := c.Writer

		// Create a custom response writer that captures the response
		responseCapture := &ResponseCapture{
			ResponseWriter: originalWriter,
			context:        c,
		}
		c.Writer = responseCapture

		c.Next()
	}
}

// ResponseCapture captures response data for format conversion
type ResponseCapture struct {
	gin.ResponseWriter
	context *gin.Context
}

// Write intercepts the response and handles format conversion
func (rc *ResponseCapture) Write(data []byte) (int, error) {
	// Check if client wants CSV format
	acceptHeader := rc.context.GetHeader("Accept")
	if strings.Contains(acceptHeader, "text/csv") {
		return rc.writeCSV(data)
	}
	
	// Default to JSON
	return rc.ResponseWriter.Write(data)
}

// writeCSV converts JSON response to CSV format
func (rc *ResponseCapture) writeCSV(jsonData []byte) (int, error) {
	// Parse JSON data
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		// If JSON parsing fails, return original data
		return rc.ResponseWriter.Write(jsonData)
	}

	// Convert to CSV
	csvData, err := convertToCSV(data)
	if err != nil {
		// If CSV conversion fails, return original JSON
		return rc.ResponseWriter.Write(jsonData)
	}

	// Set CSV headers
	rc.Header().Set("Content-Type", "text/csv; charset=utf-8")
	rc.Header().Set("Content-Disposition", "attachment; filename=export.csv")

	return rc.ResponseWriter.Write([]byte(csvData))
}

// convertToCSV converts interface{} data to CSV string
func convertToCSV(data interface{}) (string, error) {
	var records [][]string
	
	switch v := data.(type) {
	case map[string]interface{}:
		// Handle single object or object with data array
		if dataArray, exists := v["data"]; exists {
			if arr, ok := dataArray.([]interface{}); ok {
				records = extractRecordsFromArray(arr)
			}
		} else if members, exists := v["members"]; exists {
			if arr, ok := members.([]interface{}); ok {
				records = extractRecordsFromArray(arr)
			}
		} else if tournaments, exists := v["tournaments"]; exists {
			if arr, ok := tournaments.([]interface{}); ok {
				records = extractRecordsFromArray(arr)
			}
		} else if clubs, exists := v["clubs"]; exists {
			if arr, ok := clubs.([]interface{}); ok {
				records = extractRecordsFromArray(arr)
			}
		} else {
			// Single object
			records = extractRecordsFromObject(v)
		}
	case []interface{}:
		// Handle direct array
		records = extractRecordsFromArray(v)
	default:
		// Fallback for other types
		records = [][]string{{"error"}, {"Unsupported data format for CSV export"}}
	}

	// Convert records to CSV string
	var csvBuilder strings.Builder
	csvWriter := csv.NewWriter(&csvBuilder)
	csvWriter.Comma = ';' // Use semicolon separator for German Excel compatibility
	
	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return "", err
		}
	}
	
	csvWriter.Flush()
	return csvBuilder.String(), csvWriter.Error()
}

// extractRecordsFromArray extracts CSV records from an array of objects
func extractRecordsFromArray(arr []interface{}) [][]string {
	if len(arr) == 0 {
		return [][]string{{"No data available"}}
	}

	// Get headers from first object
	firstObj, ok := arr[0].(map[string]interface{})
	if !ok {
		return [][]string{{"error"}, {"Invalid data format"}}
	}

	headers := getObjectKeys(firstObj)
	records := [][]string{headers}

	// Extract data rows
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			row := make([]string, len(headers))
			for i, header := range headers {
				if value, exists := obj[header]; exists {
					row[i] = formatValue(value)
				}
			}
			records = append(records, row)
		}
	}

	return records
}

// extractRecordsFromObject extracts CSV records from a single object
func extractRecordsFromObject(obj map[string]interface{}) [][]string {
	headers := getObjectKeys(obj)
	records := [][]string{headers}
	
	row := make([]string, len(headers))
	for i, header := range headers {
		if value, exists := obj[header]; exists {
			row[i] = formatValue(value)
		}
	}
	records = append(records, row)
	
	return records
}

// getObjectKeys returns sorted keys from a map
func getObjectKeys(obj map[string]interface{}) []string {
	keys := make([]string, 0, len(obj))
	
	// Prioritize common fields for better CSV layout
	priorityFields := []string{"id", "player_id", "club_id", "tournament_id", "name", "firstname", "rating", "date"}
	
	// Add priority fields first if they exist
	for _, field := range priorityFields {
		if _, exists := obj[field]; exists {
			keys = append(keys, field)
		}
	}
	
	// Add remaining fields
	for key := range obj {
		found := false
		for _, existing := range keys {
			if existing == key {
				found = true
				break
			}
		}
		if !found {
			keys = append(keys, key)
		}
	}
	
	return keys
}

// formatValue converts interface{} to string for CSV
func formatValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', 2, 64)
	case bool:
		return strconv.FormatBool(v)
	case map[string]interface{}:
		// For nested objects, extract key information
		if name, exists := v["name"]; exists {
			return formatValue(name)
		}
		// Fallback to JSON representation for complex objects
		if jsonStr, err := json.Marshal(v); err == nil {
			return string(jsonStr)
		}
		return "complex_object"
	case []interface{}:
		// For arrays, join simple values or count complex ones
		if len(v) == 0 {
			return ""
		}
		if len(v) > 5 {
			return strconv.Itoa(len(v)) + "_items"
		}
		var strValues []string
		for _, item := range v {
			strValues = append(strValues, formatValue(item))
		}
		return strings.Join(strValues, ";")
	default:
		// Use reflection as fallback
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr && rv.IsNil() {
			return ""
		}
		return strings.Trim(strings.ReplaceAll(rv.String(), "\n", " "), "<>")
	}
}
