package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code represents an application-level error code.
type Code int

const (
	CodeUnknown     Code = 0
	CodeNotFound    Code = 1
	CodeInvalidArg  Code = 2
	CodeInternal    Code = 3
	CodeConflict    Code = 4
	CodeUnauthenticated Code = 5
)

// AppError is the application error type that carries an error code,
// a human-readable message and an optional cause.
type AppError struct {
	Code    Code
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

// New creates a new AppError with the given code and message.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an existing error with an AppError.
func Wrap(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// HTTPStatus maps an error code to an HTTP status code.
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case CodeNotFound:
			return http.StatusNotFound
		case CodeInvalidArg:
			return http.StatusBadRequest
		case CodeConflict:
			return http.StatusConflict
		case CodeUnauthenticated:
			return http.StatusUnauthorized
		case CodeInternal:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

// Is wraps errors.Is.
var Is = errors.Is

// As wraps errors.As.
var As = errors.As
