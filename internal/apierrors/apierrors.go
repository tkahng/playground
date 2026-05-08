package apierrors

import (
	"encoding/json"
	"net/http"
)

// AppError is a domain error carrying an HTTP status code.
// It implements huma.StatusError so Huma uses the correct status code
// and formats the response body automatically — no handler changes needed.
type AppError struct {
	status  int
	message string
	cause   error
}

func (e *AppError) Error() string  { return e.message }
func (e *AppError) GetStatus() int { return e.status }
func (e *AppError) Unwrap() error  { return e.cause }

func (e *AppError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}{
		Title:  http.StatusText(e.status),
		Status: e.status,
		Detail: e.message,
	})
}

func NotFound(message string) *AppError {
	return &AppError{status: http.StatusNotFound, message: message}
}

func Conflict(message string) *AppError {
	return &AppError{status: http.StatusConflict, message: message}
}

func BadRequest(message string) *AppError {
	return &AppError{status: http.StatusBadRequest, message: message}
}

func Unauthorized(message string) *AppError {
	return &AppError{status: http.StatusUnauthorized, message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{status: http.StatusForbidden, message: message}
}

// Gone is for resources that existed but are no longer valid (e.g. expired invitations).
func Gone(message string) *AppError {
	return &AppError{status: http.StatusGone, message: message}
}
