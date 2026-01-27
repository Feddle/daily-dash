package domain

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNetworkFailure     = errors.New("network failure")
	ErrTimeout            = errors.New("request timeout")
	ErrParseFailure       = errors.New("failed to parse response")
	ErrInvalidResponse    = errors.New("invalid response format")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// APIError represents an error from an API call
type APIError struct {
	Service    string
	Operation  string
	StatusCode int
	Err        error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s %s failed (status %d): %v", e.Service, e.Operation, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("%s %s failed: %v", e.Service, e.Operation, e.Err)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError
func NewAPIError(service, operation string, statusCode int, err error) *APIError {
	return &APIError{
		Service:    service,
		Operation:  operation,
		StatusCode: statusCode,
		Err:        err,
	}
}
