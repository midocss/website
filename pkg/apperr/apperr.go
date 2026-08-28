// Package apperr defines the single error type every layer returns, so the HTTP
// layer can map any failure to a consistent JSON response without type
// switching on package-specific errors.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeBadRequest   Code = "bad_request"
	CodeValidation   Code = "validation_error"
	CodeUnauthorized Code = "unauthorized"
	CodeForbidden    Code = "forbidden"
	CodeNotFound     Code = "not_found"
	CodeConflict     Code = "conflict"
	CodeTooManyReqs  Code = "too_many_requests"
	CodeInternal     Code = "internal_error"
)

// FieldError describes one invalid input field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error carries an HTTP-mappable code, a client-safe message and an optional
// wrapped cause that is logged but never serialized.
type Error struct {
	Code    Code
	Message string
	Fields  []FieldError
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// WithCause attaches the underlying error for logging purposes.
func (e *Error) WithCause(err error) *Error {
	e.cause = err
	return e
}

// WithFields attaches per-field validation details.
func (e *Error) WithFields(fields ...FieldError) *Error {
	e.Fields = append(e.Fields, fields...)
	return e
}

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func BadRequest(message string) *Error   { return New(CodeBadRequest, message) }
func Validation(message string) *Error   { return New(CodeValidation, message) }
func Unauthorized(message string) *Error { return New(CodeUnauthorized, message) }
func Forbidden(message string) *Error    { return New(CodeForbidden, message) }
func NotFound(message string) *Error     { return New(CodeNotFound, message) }
func Conflict(message string) *Error     { return New(CodeConflict, message) }
func TooManyRequests(message string) *Error {
	return New(CodeTooManyReqs, message)
}

func Internal(err error) *Error {
	return New(CodeInternal, "an unexpected error occurred").WithCause(err)
}

// From converts any error into an *Error, defaulting to an internal error.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}

// HTTPStatus maps the error code to its HTTP status code.
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeValidation:
		return http.StatusUnprocessableEntity
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeTooManyReqs:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
