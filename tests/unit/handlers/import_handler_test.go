package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"portal64api/internal/api/handlers"
	"portal64api/internal/models"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockImportService implements the ImportService interface for testing
type MockImportService struct {
	mock.Mock
}

func (m *MockImportService) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockImportService) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockImportService) TriggerManualImport() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockImportService) GetStatus() *models.ImportStatus {
	args := m.Called()
	return args.Get(0).(*models.ImportStatus)
}

func (m *MockImportService) GetLogs(limit int) []models.ImportLogEntry {
	args := m.Called(limit)
	return args.Get(0).([]models.ImportLogEntry)
}

func (m *MockImportService) TestConnection() error {
	args := m.Called()
	return args.Error(0)
}

func TestImportHandler_GetStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockStatus     *models.ImportStatus
		expectedCode   int
		expectedStatus string
	}{
		{
			name: "idle status",
			mockStatus: &models.ImportStatus{
				Status:        "idle",
				Progress:      0,
				CurrentStep:   "",
				StartedAt:     nil,
				CompletedAt:   nil,
				LastSuccess:   nil,
				NextScheduled: timePtr(time.Now().Add(24 * time.Hour)),
				RetryCount:    0,
				MaxRetries:    3,
				Error:         "",
				SkipReason:    "",
				FilesInfo:     nil,
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "idle",
		},
		{
			name: "running status",
			mockStatus: &models.ImportStatus{
				Status:        "running",
				Progress:      45,
				CurrentStep:   "downloading_files",
				StartedAt:     timePtr(time.Now().Add(-10 * time.Minute)),
				CompletedAt:   nil,
				LastSuccess:   timePtr(time.Now().Add(-24 * time.Hour)),
				NextScheduled: timePtr(time.Now().Add(24 * time.Hour)),
				RetryCount:    0,
				MaxRetries:    3,
				Error:         "",
				SkipReason:    "",
				FilesInfo: &models.ImportFilesInfo{
					RemoteFiles: []models.FileMetadata{
						{
							Filename: "mvdsb_20250806.zip",
							Size:     1024000,
							ModTime:  time.Now().Add(-1 * time.Hour),
							IsNewer:  true,
						},
					},
					LastImported: []models.FileMetadata{
						{
							Filename: "mvdsb_20250805.zip",
							Size:     1000000,
							ModTime:  time.Now().Add(-25 * time.Hour),
						},
					},
					Downloaded: []string{},
					Extracted:  []string{},
					Imported:   []string{},
				},
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "running",
		},
		{
			name: "success status",
			mockStatus: &models.ImportStatus{
				Status:        "success",
				Progress:      100,
				CurrentStep:   "",
				StartedAt:     timePtr(time.Now().Add(-30 * time.Minute)),
				CompletedAt:   timePtr(time.Now().Add(-5 * time.Minute)),
				LastSuccess:   timePtr(time.Now().Add(-5 * time.Minute)),
				NextScheduled: timePtr(time.Now().Add(24 * time.Hour)),
				RetryCount:    0,
				MaxRetries:    3,
				Error:         "",
				SkipReason:    "",
				FilesInfo: &models.ImportFilesInfo{
					RemoteFiles: []models.FileMetadata{
						{
							Filename: "mvdsb_20250806.zip",
							Size:     1024000,
							ModTime:  time.Now().Add(-1 * time.Hour),
							IsNewer:  true,
						},
					},
					LastImported: []models.FileMetadata{
						{
							Filename: "mvdsb_20250806.zip",
							Size:     1024000,
							ModTime:  time.Now().Add(-1 * time.Hour),
						},
					},
					Downloaded: []string{"mvdsb_20250806.zip", "portal64_bdw_20250806.zip"},
					Extracted:  []string{"mvdsb_dump.sql", "portal64_bdw_dump.sql"},
					Imported:   []string{"mvdsb", "portal64_bdw"},
				},
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "success",
		},
		{
			name: "failed status",
			mockStatus: &models.ImportStatus{
				Status:        "failed",
				Progress:      25,
				CurrentStep:   "downloading_files",
				StartedAt:     timePtr(time.Now().Add(-15 * time.Minute)),
				CompletedAt:   timePtr(time.Now().Add(-2 * time.Minute)),
				LastSuccess:   timePtr(time.Now().Add(-48 * time.Hour)),
				NextScheduled: timePtr(time.Now().Add(24 * time.Hour)),
				RetryCount:    2,
				MaxRetries:    3,
				Error:         "Connection timeout while downloading files",
				SkipReason:    "",
				FilesInfo: &models.ImportFilesInfo{
					RemoteFiles: []models.FileMetadata{
						{
							Filename: "mvdsb_20250806.zip",
							Size:     1024000,
							ModTime:  time.Now().Add(-1 * time.Hour),
							IsNewer:  true,
						},
					},
					LastImported: []models.FileMetadata{
						{
							Filename: "mvdsb_20250804.zip",
							Size:     980000,
							ModTime:  time.Now().Add(-49 * time.Hour),
						},
					},
					Downloaded: []string{},
					Extracted:  []string{},
					Imported:   []string{},
				},
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "failed",
		},
		{
			name: "skipped status",
			mockStatus: &models.ImportStatus{
				Status:        "skipped",
				Progress:      100,
				CurrentStep:   "",
				StartedAt:     timePtr(time.Now().Add(-5 * time.Minute)),
				CompletedAt:   timePtr(time.Now().Add(-3 * time.Minute)),
				LastSuccess:   timePtr(time.Now().Add(-24 * time.Hour)),
				NextScheduled: timePtr(time.Now().Add(24 * time.Hour)),
				RetryCount:    0,
				MaxRetries:    3,
				Error:         "",
				SkipReason:    "no_newer_files_available",
				FilesInfo: &models.ImportFilesInfo{
					RemoteFiles: []models.FileMetadata{
						{
							Filename: "mvdsb_20250805.zip",
							Size:     1000000,
							ModTime:  time.Now().Add(-25 * time.Hour),
							IsNewer:  false,
						},
					},
					LastImported: []models.FileMetadata{
						{
							Filename: "mvdsb_20250805.zip",
							Size:     1000000,
							ModTime:  time.Now().Add(-25 * time.Hour),
						},
					},
					Downloaded: []string{},
					Extracted:  []string{},
					Imported:   []string{},
				},
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockImportService)
			mockService.On("GetStatus").Return(tt.mockStatus)

			// Create handler
			handler := handlers.NewImportHandler(mockService)

			// Setup router
			router := gin.New()
			router.GET("/api/v1/import/status", handler.GetImportStatus)

			// Create request
			req, err := http.NewRequest(http.MethodGet, "/api/v1/import/status", nil)
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				var response models.ImportStatus
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedStatus, response.Status)
				assert.Equal(t, tt.mockStatus.Progress, response.Progress)
				assert.Equal(t, tt.mockStatus.CurrentStep, response.CurrentStep)
				assert.Equal(t, tt.mockStatus.RetryCount, response.RetryCount)
				assert.Equal(t, tt.mockStatus.MaxRetries, response.MaxRetries)
				assert.Equal(t, tt.mockStatus.Error, response.Error)
				assert.Equal(t, tt.mockStatus.SkipReason, response.SkipReason)

				// Verify time fields
				if tt.mockStatus.StartedAt != nil {
					assert.NotNil(t, response.StartedAt)
				}
				if tt.mockStatus.CompletedAt != nil {
					assert.NotNil(t, response.CompletedAt)
				}
				if tt.mockStatus.LastSuccess != nil {
					assert.NotNil(t, response.LastSuccess)
				}
				if tt.mockStatus.NextScheduled != nil {
					assert.NotNil(t, response.NextScheduled)
				}

				// Verify files info
				if tt.mockStatus.FilesInfo != nil {
					assert.NotNil(t, response.FilesInfo)
					assert.Equal(t, len(tt.mockStatus.FilesInfo.RemoteFiles), len(response.FilesInfo.RemoteFiles))
					assert.Equal(t, len(tt.mockStatus.FilesInfo.LastImported), len(response.FilesInfo.LastImported))
					assert.Equal(t, len(tt.mockStatus.FilesInfo.Downloaded), len(response.FilesInfo.Downloaded))
					assert.Equal(t, len(tt.mockStatus.FilesInfo.Extracted), len(response.FilesInfo.Extracted))
					assert.Equal(t, len(tt.mockStatus.FilesInfo.Imported), len(response.FilesInfo.Imported))
				}
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestImportHandler_StartImport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceError   error
		expectedCode   int
		expectedError  string
		expectedMessage string
	}{
		{
			name:            "successful start",
			serviceError:    nil,
			expectedCode:    http.StatusOK,
			expectedMessage: "Manual import started",
		},
		{
			name:          "import already running",
			serviceError:  errors.New("import is already running"),
			expectedCode:  http.StatusConflict,
			expectedError: "Import is already running",
		},
		{
			name:          "service disabled",
			serviceError:  errors.New("import service is disabled"),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Import service is disabled",
		},
		{
			name:          "service error",
			serviceError:  errors.New("database connection failed"),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockImportService)
			mockService.On("TriggerManualImport").Return(tt.serviceError)

			// Create handler
			handler := handlers.NewImportHandler(mockService)

			// Setup router
			router := gin.New()
			router.POST("/api/v1/import/start", handler.StartManualImport)

			// Create request
			req, err := http.NewRequest(http.MethodPost, "/api/v1/import/start", bytes.NewBuffer([]byte("{}")))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedCode == http.StatusOK {
				assert.Contains(t, response, "message")
				assert.Equal(t, tt.expectedMessage, response["message"])
				assert.Contains(t, response, "started_at")
			} else {
				assert.Contains(t, response, "error")
				assert.Contains(t, response["error"].(string), tt.expectedError)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestImportHandler_GetLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		mockLogs     []models.ImportLogEntry
		expectedCode int
		expectedLogs int
	}{
		{
			name:         "empty logs",
			mockLogs:     []models.ImportLogEntry{},
			expectedCode: http.StatusOK,
			expectedLogs: 0,
		},
		{
			name: "multiple log entries",
			mockLogs: []models.ImportLogEntry{
				{
					Timestamp: time.Now().Add(-10 * time.Minute),
					Level:     "INFO",
					Message:   "Import process started",
					Step:      "initialization",
				},
				{
					Timestamp: time.Now().Add(-8 * time.Minute),
					Level:     "INFO",
					Message:   "Connecting to portal.svw.info",
					Step:      "downloading",
				},
				{
					Timestamp: time.Now().Add(-7 * time.Minute),
					Level:     "INFO",
					Message:   "Downloaded mvdsb_20250806.zip (1.2MB)",
					Step:      "downloading",
				},
				{
					Timestamp: time.Now().Add(-6 * time.Minute),
					Level:     "WARN",
					Message:   "Large file detected, extraction may take longer",
					Step:      "extracting",
				},
				{
					Timestamp: time.Now().Add(-5 * time.Minute),
					Level:     "INFO",
					Message:   "Successfully extracted 2 SQL files",
					Step:      "extracting",
				},
				{
					Timestamp: time.Now().Add(-3 * time.Minute),
					Level:     "INFO",
					Message:   "Database import completed for mvdsb",
					Step:      "importing",
				},
				{
					Timestamp: time.Now().Add(-1 * time.Minute),
					Level:     "INFO",
					Message:   "Import completed successfully",
					Step:      "cleanup",
				},
			},
			expectedCode: http.StatusOK,
			expectedLogs: 7,
		},
		{
			name: "logs with errors",
			mockLogs: []models.ImportLogEntry{
				{
					Timestamp: time.Now().Add(-10 * time.Minute),
					Level:     "INFO",
					Message:   "Import process started",
					Step:      "initialization",
				},
				{
					Timestamp: time.Now().Add(-8 * time.Minute),
					Level:     "ERROR",
					Message:   "Failed to connect to SCP server",
					Step:      "downloading",
				},
				{
					Timestamp: time.Now().Add(-7 * time.Minute),
					Level:     "INFO",
					Message:   "Retrying connection...",
					Step:      "downloading",
				},
				{
					Timestamp: time.Now().Add(-6 * time.Minute),
					Level:     "ERROR",
					Message:   "Connection failed after 3 attempts",
					Step:      "downloading",
				},
				{
					Timestamp: time.Now().Add(-5 * time.Minute),
					Level:     "ERROR",
					Message:   "Import failed",
					Step:      "cleanup",
				},
			},
			expectedCode: http.StatusOK,
			expectedLogs: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(MockImportService)
			mockService.On("GetLogs", mock.AnythingOfType("int")).Return(tt.mockLogs)

			// Create handler
			handler := handlers.NewImportHandler(mockService)

			// Setup router
			router := gin.New()
			router.GET("/api/v1/import/logs", handler.GetImportLogs)

			// Create request
			req, err := http.NewRequest(http.MethodGet, "/api/v1/import/logs", nil)
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "logs")
			logs := response["logs"].([]interface{})
			assert.Equal(t, tt.expectedLogs, len(logs))

			// Verify log entry structure
			if len(logs) > 0 {
				firstLog := logs[0].(map[string]interface{})
				assert.Contains(t, firstLog, "timestamp")
				assert.Contains(t, firstLog, "level")
				assert.Contains(t, firstLog, "message")
				assert.Contains(t, firstLog, "step")

				// Verify specific values for first log entry
				if len(tt.mockLogs) > 0 {
					expectedLog := tt.mockLogs[0]
					assert.Equal(t, expectedLog.Level, firstLog["level"])
					assert.Equal(t, expectedLog.Message, firstLog["message"])
					assert.Equal(t, expectedLog.Step, firstLog["step"])
				}
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestImportHandler_ServiceNotAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		endpoint     string
		method       string
		body         string
		expectedCode int
	}{
		{
			name:         "status endpoint with nil service",
			endpoint:     "/api/v1/import/status",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusServiceUnavailable,
		},
		{
			name:         "start endpoint with nil service",
			endpoint:     "/api/v1/import/start",
			method:       http.MethodPost,
			body:         "{}",
			expectedCode: http.StatusServiceUnavailable,
		},
		{
			name:         "logs endpoint with nil service",
			endpoint:     "/api/v1/import/logs",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with nil service (simulating disabled import)
			handler := handlers.NewImportHandler(nil)

			// Setup router
			router := gin.New()
			switch tt.endpoint {
			case "/api/v1/import/status":
				router.GET(tt.endpoint, handler.GetImportStatus)
			case "/api/v1/import/start":
				router.POST(tt.endpoint, handler.StartManualImport)
			case "/api/v1/import/logs":
				router.GET(tt.endpoint, handler.GetImportLogs)
			}

			// Create request
			var req *http.Request
			var err error
			if tt.body != "" {
				req, err = http.NewRequest(tt.method, tt.endpoint, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, tt.endpoint, nil)
			}
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response, "error")
			assert.Contains(t, response["error"].(string), "Import service is not available")
		})
	}
}

func TestImportHandler_HTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockImportService)
	mockService.On("GetStatus").Return(&models.ImportStatus{Status: "idle"})
	mockService.On("TriggerManualImport").Return(nil)
	mockService.On("GetLogs", mock.AnythingOfType("int")).Return([]models.ImportLogEntry{})

	handler := handlers.NewImportHandler(mockService)

	tests := []struct {
		name           string
		endpoint       string
		allowedMethod  string
		forbiddenMethod string
		expectedAllowed int
		expectedForbidden int
	}{
		{
			name:            "status endpoint methods",
			endpoint:        "/api/v1/import/status",
			allowedMethod:   http.MethodGet,
			forbiddenMethod: http.MethodPost,
			expectedAllowed: http.StatusOK,
			expectedForbidden: http.StatusNotFound,
		},
		{
			name:            "start endpoint methods",
			endpoint:        "/api/v1/import/start",
			allowedMethod:   http.MethodPost,
			forbiddenMethod: http.MethodGet,
			expectedAllowed: http.StatusOK,
			expectedForbidden: http.StatusNotFound,
		},
		{
			name:            "logs endpoint methods",
			endpoint:        "/api/v1/import/logs",
			allowedMethod:   http.MethodGet,
			forbiddenMethod: http.MethodDelete,
			expectedAllowed: http.StatusOK,
			expectedForbidden: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			switch tt.endpoint {
			case "/api/v1/import/status":
				router.GET(tt.endpoint, handler.GetImportStatus)
			case "/api/v1/import/start":
				router.POST(tt.endpoint, handler.StartManualImport)
			case "/api/v1/import/logs":
				router.GET(tt.endpoint, handler.GetImportLogs)
			}

			// Test allowed method
			var req *http.Request
			var err error
			if tt.allowedMethod == http.MethodPost {
				req, err = http.NewRequest(tt.allowedMethod, tt.endpoint, bytes.NewBufferString("{}"))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.allowedMethod, tt.endpoint, nil)
			}
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedAllowed, w.Code)

			// Test forbidden method
			req, err = http.NewRequest(tt.forbiddenMethod, tt.endpoint, nil)
			require.NoError(t, err)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedForbidden, w.Code)
		})
	}
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestImportHandler_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockImportService)
	mockService.On("GetStatus").Return(&models.ImportStatus{Status: "idle"})
	mockService.On("GetLogs", mock.AnythingOfType("int")).Return([]models.ImportLogEntry{})

	handler := handlers.NewImportHandler(mockService)

	tests := []struct {
		name         string
		endpoint     string
		method       string
		expectedType string
	}{
		{
			name:         "status endpoint content type",
			endpoint:     "/api/v1/import/status",
			method:       http.MethodGet,
			expectedType: "application/json",
		},
		{
			name:         "logs endpoint content type",
			endpoint:     "/api/v1/import/logs",
			method:       http.MethodGet,
			expectedType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			switch tt.endpoint {
			case "/api/v1/import/status":
				router.GET(tt.endpoint, handler.GetImportStatus)
			case "/api/v1/import/logs":
				router.GET(tt.endpoint, handler.GetImportLogs)
			}

			// Create request
			req, err := http.NewRequest(tt.method, tt.endpoint, nil)
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, req)

			// Assert content type
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), tt.expectedType)

			// Verify response is valid JSON
			var response interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		})
	}
}
