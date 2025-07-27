package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"portal64api/internal/models"
	"portal64api/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ParseSearchParams parses search parameters from gin context
func ParseSearchParams(c *gin.Context) models.SearchRequest {
	return ParseSearchParamsWithDefaults(c, "name", "asc")
}

// ParseSearchParamsWithDefaults parses search parameters with custom defaults
func ParseSearchParamsWithDefaults(c *gin.Context, defaultSortBy, defaultSortOrder string) models.SearchRequest {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Validate limit (max 100)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}

	return models.SearchRequest{
		Query:       c.Query("query"),
		Limit:       limit,
		Offset:      offset,
		SortBy:      c.DefaultQuery("sort_by", defaultSortBy),
		SortOrder:   c.DefaultQuery("sort_order", defaultSortOrder),
		FilterBy:    c.Query("filter_by"),
		FilterValue: c.Query("filter_value"),
	}
}

// GeneratePlayerID generates a player ID in format C0101-1014
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
		SendJSONResponse(c, http.StatusInternalServerError, 
			errors.NewInternalServerError("Failed to generate CSV"))
		return
	}
}

// writeCSV writes data to CSV writer using reflection
func writeCSV(writer *csv.Writer, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}

	if v.Len() == 0 {
		return nil
	}

	// Get headers from first element
	firstElement := v.Index(0)
	if firstElement.Kind() == reflect.Ptr {
		firstElement = firstElement.Elem()
	}

	headers := getCSVHeaders(firstElement)
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data rows
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i)
		if element.Kind() == reflect.Ptr {
			element = element.Elem()
		}

		row := getCSVRow(element)
		if err := writer.Write(row); err != nil {
			return err
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
				// Handle pointer types recursively
				elem := fieldValue.Elem()
				if elem.Kind() == reflect.Struct {
					// For time.Time and other structs, convert to string
					if stringer, ok := fieldValue.Interface().(fmt.Stringer); ok {
						cellValue = stringer.String()
					} else {
						jsonBytes, _ := json.Marshal(fieldValue.Interface())
						cellValue = string(jsonBytes)
					}
				} else {
					cellValue = fmt.Sprintf("%v", elem.Interface())
				}
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
