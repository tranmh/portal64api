package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"portal64api/internal/models"
	"portal64api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ParseSearchParams parses search parameters from gin context
func ParseSearchParams(c *gin.Context) (models.SearchRequest, error) {
	return ParseSearchParamsWithDefaults(c, "name", "asc")
}

// ParseSearchParamsWithDefaults parses search parameters with custom defaults
func ParseSearchParamsWithDefaults(c *gin.Context, defaultSortBy, defaultSortOrder string) (models.SearchRequest, error) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return models.SearchRequest{}, errors.NewBadRequestError("Invalid limit parameter")
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return models.SearchRequest{}, errors.NewBadRequestError("Invalid offset parameter")
	}

	// Validate limit (max 100, min 1)
	if limit > 100 {
		return models.SearchRequest{}, errors.NewBadRequestError("Limit cannot exceed 100")
	}
	if limit < 1 {
		return models.SearchRequest{}, errors.NewBadRequestError("Limit must be at least 1")
	}

	// Validate offset (cannot be negative)
	if offset < 0 {
		return models.SearchRequest{}, errors.NewBadRequestError("Offset cannot be negative")
	}

	// Validate sort order
	sortOrder := c.DefaultQuery("sort_order", defaultSortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		return models.SearchRequest{}, errors.NewBadRequestError("Sort order must be 'asc' or 'desc'")
	}

	return models.SearchRequest{
		Query:       c.Query("query"),
		Limit:       limit,
		Offset:      offset,
		SortBy:      c.DefaultQuery("sort_by", defaultSortBy),
		SortOrder:   sortOrder,
		FilterBy:    c.Query("filter_by"),
		FilterValue: c.Query("filter_value"),
	}, nil
}

// ValidateClubID validates a club ID format (e.g., C0101)
func ValidateClubID(clubID string) error {
	if clubID == "" {
		return errors.NewBadRequestError("Club ID cannot be empty")
	}
	
	// Club ID should start with 'C' followed by 4 digits
	if len(clubID) != 5 || clubID[0] != 'C' {
		return errors.NewBadRequestError("Invalid club ID format (expected: C0101)")
	}
	
	// Check if the last 4 characters are digits
	for i := 1; i < 5; i++ {
		if clubID[i] < '0' || clubID[i] > '9' {
			return errors.NewBadRequestError("Invalid club ID format (expected: C0101)")
		}
	}
	
	return nil
}

// ValidatePlayerID validates a player ID format (e.g., C0101-1014)
func ValidatePlayerID(playerID string) error {
	if playerID == "" {
		return errors.NewBadRequestError("Player ID cannot be empty")
	}
	
	parts := strings.Split(playerID, "-")
	if len(parts) != 2 {
		return errors.NewBadRequestError("Invalid player ID format (expected: C0101-1014)")
	}
	
	// Validate club part
	if err := ValidateClubID(parts[0]); err != nil {
		return errors.NewBadRequestError("Invalid player ID format (expected: C0101-1014)")
	}
	
	// Validate person ID part
	if _, err := strconv.ParseUint(parts[1], 10, 32); err != nil {
		return errors.NewBadRequestError("Invalid player ID format (expected: C0101-1014)")
	}
	
	return nil
}

// ValidateTournamentID validates a tournament ID format (e.g., C529-K00-HT1)
func ValidateTournamentID(tournamentID string) error {
	if tournamentID == "" {
		return errors.NewBadRequestError("Tournament ID cannot be empty")
	}
	
	parts := strings.Split(tournamentID, "-")
	if len(parts) != 3 {
		return errors.NewBadRequestError("Invalid tournament ID format (expected: C529-K00-HT1)")
	}
	
	// First part: should start with 'C' followed by digits
	if len(parts[0]) < 2 || parts[0][0] != 'C' {
		return errors.NewBadRequestError("Invalid tournament ID format (expected: C529-K00-HT1)")
	}
	
	// Validate each part has some basic structure
	for i, part := range parts {
		if len(part) < 2 {
			return errors.NewBadRequestError("Invalid tournament ID format (expected: C529-K00-HT1)")
		}
		// First part should have 'C' + digits, others can be alphanumeric
		if i == 0 {
			for j := 1; j < len(part); j++ {
				if part[j] < '0' || part[j] > '9' {
					return errors.NewBadRequestError("Invalid tournament ID format (expected: C529-K00-HT1)")
				}
			}
		}
	}
	
	return nil
}
func GeneratePlayerID(vkz string, personID uint) string {
	return fmt.Sprintf("%s-%d", vkz, personID)
}

// ParsePlayerID parses a player ID into VKZ and person ID
func ParsePlayerID(playerID string) (string, uint, error) {
	parts := strings.Split(playerID, "-")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid player ID format: %s", playerID)
	}

	personID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("invalid person ID in player ID: %s", playerID)
	}

	return parts[0], uint(personID), nil
}

// SendJSONResponse sends a JSON response
func SendJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	response := models.Response{
		Success: statusCode < 400,
		Data:    data,
	}

	if statusCode >= 400 {
		if err, ok := data.(errors.APIError); ok {
			response.Error = err.Message
		} else {
			response.Error = fmt.Sprintf("%v", data)
		}
		response.Data = nil
	}

	c.JSON(statusCode, response)
}

// SendCSVResponse sends a CSV response
func SendCSVResponse(c *gin.Context, filename string, data interface{}) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Convert data to CSV
	if err := writeCSV(writer, data); err != nil {
		// Reset headers and send JSON error instead
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "")
		SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to generate CSV: " + err.Error()))
		return
	}
}

// writeCSV writes data to CSV writer using reflection
func writeCSV(writer *csv.Writer, data interface{}) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("data pointer is nil")
		}
		v = v.Elem()
	}

	// Handle wrapped response structures that contain Data field
	if v.Kind() == reflect.Struct {
		// Check if it has a Data field (for paginated responses)
		dataField := v.FieldByName("Data")
		if dataField.IsValid() && dataField.Kind() == reflect.Slice {
			v = dataField
		} else {
			// Single struct - convert to slice
			sliceType := reflect.SliceOf(v.Type())
			slice := reflect.MakeSlice(sliceType, 1, 1)
			slice.Index(0).Set(v)
			v = slice
		}
	}

	if v.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice or struct with Data slice field")
	}

	if v.Len() == 0 {
		// Empty data - write just headers if possible
		return nil
	}

	// Get headers from first element
	firstElement := v.Index(0)
	if firstElement.Kind() == reflect.Ptr {
		if firstElement.IsNil() {
			return fmt.Errorf("first element is nil")
		}
		firstElement = firstElement.Elem()
	}

	if firstElement.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs")
	}

	headers := getCSVHeaders(firstElement)
	if len(headers) == 0 {
		return fmt.Errorf("no valid fields found for CSV headers")
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write data rows
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i)
		if element.Kind() == reflect.Ptr {
			if element.IsNil() {
				continue // Skip nil elements
			}
			element = element.Elem()
		}

		if element.Kind() != reflect.Struct {
			continue // Skip non-struct elements
		}

		row := getCSVRow(element)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return nil
}

// getCSVHeaders extracts headers from struct using json tags
func getCSVHeaders(v reflect.Value) []string {
	t := v.Type()
	headers := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove options like omitempty
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		headers = append(headers, jsonTag)
	}

	return headers
}

// getCSVRow extracts row data from struct
func getCSVRow(v reflect.Value) []string {
	t := v.Type()
	row := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove options like omitempty
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		fieldValue := v.Field(i)
		var cellValue string

		// Handle different field types more robustly
		switch fieldValue.Kind() {
		case reflect.String:
			cellValue = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			cellValue = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			cellValue = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			cellValue = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.Bool:
			cellValue = strconv.FormatBool(fieldValue.Bool())
		case reflect.Ptr:
			if fieldValue.IsNil() {
				cellValue = ""
			} else {
				// Handle pointer types
				elem := fieldValue.Elem()
				switch elem.Kind() {
				case reflect.Struct:
					// Special handling for time.Time
					if elem.Type().String() == "time.Time" {
						if t, ok := fieldValue.Interface().(*time.Time); ok {
							cellValue = t.Format("2006-01-02 15:04:05")
						} else {
							cellValue = ""
						}
					} else {
						// For other structs, try to marshal to JSON
						if jsonBytes, err := json.Marshal(fieldValue.Interface()); err == nil {
							cellValue = string(jsonBytes)
						} else {
							cellValue = fmt.Sprintf("%v", fieldValue.Interface())
						}
					}
				default:
					cellValue = fmt.Sprintf("%v", elem.Interface())
				}
			}
		case reflect.Struct:
			// Special handling for time.Time
			if fieldValue.Type().String() == "time.Time" {
				if t, ok := fieldValue.Interface().(time.Time); ok {
					cellValue = t.Format("2006-01-02 15:04:05")
				} else {
					cellValue = ""
				}
			} else {
				// For other structs, try to marshal to JSON
				if jsonBytes, err := json.Marshal(fieldValue.Interface()); err == nil {
					cellValue = string(jsonBytes)
				} else {
					cellValue = fmt.Sprintf("%v", fieldValue.Interface())
				}
			}
		case reflect.Slice, reflect.Array:
			// Handle slices/arrays by marshaling to JSON
			if jsonBytes, err := json.Marshal(fieldValue.Interface()); err == nil {
				cellValue = string(jsonBytes)
			} else {
				cellValue = fmt.Sprintf("%v", fieldValue.Interface())
			}
		default:
			cellValue = fmt.Sprintf("%v", fieldValue.Interface())
		}

		row = append(row, cellValue)
	}

	return row
}

// HandleResponse handles both JSON and CSV responses based on Accept header
func HandleResponse(c *gin.Context, data interface{}, filename string) {
	accept := c.GetHeader("Accept")
	format := c.Query("format")

	if format == "csv" || strings.Contains(accept, "text/csv") {
		SendCSVResponse(c, filename, data)
	} else {
		SendJSONResponse(c, http.StatusOK, data)
	}
}
