package errors

import (
	"fmt"
	"net/http"
)

// APIError represents an API error with status code and message
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

// Predefined errors
var (
	ErrNotFound = APIError{
		Code:    http.StatusNotFound,
		Message: "Resource not found",
	}
	
	ErrBadRequest = APIError{
		Code:    http.StatusBadRequest,
		Message: "Bad request",
	}
	
	ErrInternalServer = APIError{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}
	
	ErrUnauthorized = APIError{
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	}
	
	ErrForbidden = APIError{
		Code:    http.StatusForbidden,
		Message: "Forbidden",
	}
)

// NewAPIError creates a new API error
func NewAPIError(code int, message string, details ...string) APIError {
	err := APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) APIError {
	return APIError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) APIError {
	return APIError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// NewInternalServerError creates an internal server error
func NewInternalServerError(details string) APIError {
	return APIError{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
		Details: details,
	}
}
