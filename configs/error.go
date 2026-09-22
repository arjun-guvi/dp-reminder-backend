package configs

import "fmt"

// AppError represents a custom application error
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code int, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Common error constructors
func ErrNotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return NewAppError(40401, message, 404)
}

func ErrUnauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return NewAppError(40101, message, 401)
}

func ErrForbidden(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return NewAppError(40301, message, 403)
}

func ErrBadRequest(message string) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return NewAppError(40001, message, 400)
}

func ErrInternal(message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError(50001, message, 500)
}

func ErrValidation(message string) *AppError {
	if message == "" {
		message = "Validation error"
	}
	return NewAppError(42201, message, 422)
}
