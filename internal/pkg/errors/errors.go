package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound      = New(CodeNotFound, "resource not found")
	ErrConflict      = New(CodeConflict, "resource already exists")
	ErrValidation    = New(CodeValidation, "validation failed")
	ErrUnauthorized  = New(CodeUnauthorized, "unauthorized")
	ErrForbidden     = New(CodeForbidden, "forbidden")
	ErrInternal      = New(CodeInternal, "internal server error")
	ErrBadRequest    = New(CodeBadRequest, "bad request")
	ErrUnavailable   = New(CodeUnavailable, "service unavailable")
	ErrTimeout       = New(CodeTimeout, "request timeout")
)

type Code string

const (
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeInternal     Code = "INTERNAL_ERROR"
	CodeBadRequest   Code = "BAD_REQUEST"
	CodeUnavailable  Code = "UNAVAILABLE"
	CodeTimeout      Code = "TIMEOUT"
)

type AppError struct {
	Code       Code              `json:"code"`
	Message    string            `json:"message"`
	HTTPStatus int               `json:"-"`
	Details    map[string]string `json:"details,omitempty"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatusFromCode(code),
	}
}

func Wrap(err error, code Code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatusFromCode(code),
		Err:        err,
	}
}

func WithDetails(err *AppError, details map[string]string) *AppError {
	clone := *err
	clone.Details = details
	return &clone
}

func WithMessage(err *AppError, message string) *AppError {
	clone := *err
	clone.Message = message
	return &clone
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func httpStatusFromCode(code Code) int {
	switch code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeValidation, CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
