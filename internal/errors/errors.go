package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a custom application error with classification.
// This allows for:
// - Consistent error handling across the application
// - Automatic HTTP status code mapping
// - Better error messages to users
// - Easier error categorization for monitoring
type AppError struct {
	Code       string // Machine-readable error code (e.g., "INVALID_INPUT")
	Message    string // Human-readable error message
	StatusCode int    // HTTP status code to return
	Err        error  // Underlying error (for wrapping)
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows errors.Is and errors.As to work
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error codes for consistent error handling across the application
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeRateLimited    = "RATE_LIMITED"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeServiceUnavail = "SERVICE_UNAVAILABLE"
	ErrCodeBadRequest     = "BAD_REQUEST"
)

// Pre-defined error constructors for common scenarios

// NewValidationError creates a validation error (400)
// Use when user input is invalid
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// NewNotFoundError creates a not found error (404)
// Use when a requested resource doesn't exist
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		Err:        nil,
	}
}

// NewUnauthorizedError creates an unauthorized error (401)
// Use when authentication is required but not provided
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        nil,
	}
}

// NewForbiddenError creates a forbidden error (403)
// Use when user is authenticated but lacks permissions
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Err:        nil,
	}
}

// NewRateLimitError creates a rate limit error (429)
// Use when user has exceeded rate limits
func NewRateLimitError(retryAfter int) *AppError {
	return &AppError{
		Code:       ErrCodeRateLimited,
		Message:    fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", retryAfter),
		StatusCode: http.StatusTooManyRequests,
		Err:        nil,
	}
}

// NewInternalError creates an internal server error (500)
// Use when an unexpected error occurs
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewServiceUnavailableError creates a service unavailable error (503)
// Use when a required service (DB, API) is down
func NewServiceUnavailableError(service string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeServiceUnavail,
		Message:    fmt.Sprintf("%s is temporarily unavailable", service),
		StatusCode: http.StatusServiceUnavailable,
		Err:        err,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetStatusCode extracts the HTTP status code from an error
// Returns 500 if error is not an AppError
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetErrorCode extracts the error code from an error
// Returns ErrCodeInternal if error is not an AppError
func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeInternal
}
