// Package constants centralizes error codes and user-facing messages.
package constants

import "errors"

const (
	CodeSuccess         = 0
	CodeBadRequest      = 40000
	CodeUnauthorized    = 40100
	CodeForbidden       = 40300
	CodeNotFound        = 40400
	CodeConflict        = 40900
	CodeTooManyRequests = 42900
	CodeInternal        = 50000
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("resource conflict")
	ErrInvalidInput = errors.New("invalid input")
)

// AppError carries a business error code and a user-facing message.
type AppError struct {
	Code    int
	Message string
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError builds an AppError.
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
