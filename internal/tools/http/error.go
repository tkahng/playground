package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

// ErrorDetailer returns error details for responses & debugging. This enables
// the use of custom error types. See `NewError` for more details.
type ErrorDetailer interface {
	ErrorDetail() *ErrorDetail
}

// ErrorDetail provides details about a specific error.
type ErrorDetail struct {
	// Message is a human-readable explanation of the error.
	Message string `json:"message,omitempty" doc:"Error message text"`

	// Location is a path-like string indicating where the error occurred.
	// It typically begins with `path`, `query`, `header`, or `body`. Example:
	// `body.items[3].tags` or `path.thing-id`.
	Location string `json:"location,omitempty" doc:"Where the error occurred, e.g. 'body.items[3].tags' or 'path.thing-id'"`

	// Value is the value at the given location, echoed back to the client
	// to help with debugging. This can be useful for e.g. validating that
	// the client didn't send extra whitespace or help when the client
	// did not log an outgoing request.
	Value any `json:"value,omitempty" doc:"The value at the given location"`
}

// Error returns the error message / satisfies the `error` interface. If a
// location and value are set, they will be included in the error message,
// otherwise just the message is returned.
func (e *ErrorDetail) Error() string {
	if e.Location == "" && e.Value == nil {
		return e.Message
	}
	return fmt.Sprintf("%s (%s: %v)", e.Message, e.Location, e.Value)
}

// ErrorDetail satisfies the `ErrorDetailer` interface.
func (e *ErrorDetail) ErrorDetail() *ErrorDetail {
	return e
}

// ErrorModel defines a basic error message model based on RFC 9457 Problem
// Details for HTTP APIs (https://datatracker.ietf.org/doc/html/rfc9457). It
// is augmented with an `errors` field of `huma.ErrorDetail` objects that
// can help provide exhaustive & descriptive errors.
//
//	err := &huma.ErrorModel{
//		Title: http.StatusText(http.StatusBadRequest),
//		Status http.StatusBadRequest,
//		Detail: "Validation failed",
//		Errors: []*huma.ErrorDetail{
//			&huma.ErrorDetail{
//				Message: "expected required property id to be present",
//				Location: "body.friends[0]",
//				Value: nil,
//			},
//			&huma.ErrorDetail{
//				Message: "expected boolean",
//				Location: "body.friends[1].active",
//				Value: 5,
//			},
//		},
//	}
type ErrorModel struct {
	// Type is a URI to get more information about the error type.
	Type string `json:"type,omitempty" format:"uri" default:"about:blank" example:"https://example.com/errors/example" doc:"A URI reference to human-readable documentation for the error."`

	// Title provides a short static summary of the problem. Huma will default this
	// to the HTTP response status code text if not present.
	Title string `json:"title,omitempty" example:"Bad Request" doc:"A short, human-readable summary of the problem type. This value should not change between occurrences of the error."`

	// Status provides the HTTP status code for client convenience. Huma will
	// default this to the response status code if unset. This SHOULD match the
	// response status code (though proxies may modify the actual status code).
	Status int `json:"status,omitempty" example:"400" doc:"HTTP status code"`

	// Detail is an explanation specific to this error occurrence.
	Detail string `json:"detail,omitempty" example:"Property foo is required but is missing." doc:"A human-readable explanation specific to this occurrence of the problem."`

	// Instance is a URI to get more info about this error occurrence.
	Instance string `json:"instance,omitempty" format:"uri" example:"https://example.com/error-log/abc123" doc:"A URI reference that identifies the specific occurrence of the problem."`

	// Errors provides an optional mechanism of passing additional error details
	// as a list.
	Errors []*ErrorDetail `json:"errors,omitempty" doc:"Optional list of individual error details"`
}

// Error satisfies the `error` interface. It returns the error's detail field.
func (e *ErrorModel) Error() string {
	return e.Detail
}

// Add an error to the `Errors` slice. If passed a struct that satisfies the
// `huma.ErrorDetailer` interface, then it is used, otherwise the error
// string is used as the error detail message.
//
//	err := &ErrorModel{ /* ... */ }
//	err.Add(&huma.ErrorDetail{
//		Message: "expected boolean",
//		Location: "body.friends[1].active",
//		Value: 5
//	})
func (e *ErrorModel) Add(err error) {
	if converted, ok := err.(ErrorDetailer); ok {
		e.Errors = append(e.Errors, converted.ErrorDetail())
		return
	}

	e.Errors = append(e.Errors, &ErrorDetail{Message: err.Error()})
}

// GetStatus returns the HTTP status that should be returned to the client
// for this error.
func (e *ErrorModel) GetStatus() int {
	return e.Status
}

func (e *ErrorModel) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.Status)
	return nil
}

type StatusError interface {
	GetStatus() int
	Error() string
	Render(w http.ResponseWriter, r *http.Request) error
}

func NewError(status int, msg string, errs ...error) StatusError {
	details := make([]*ErrorDetail, len(errs))
	for i := range errs {
		if converted, ok := errs[i].(ErrorDetailer); ok {
			details[i] = converted.ErrorDetail()
		} else {
			if errs[i] == nil {
				continue
			}
			details[i] = &ErrorDetail{Message: errs[i].Error()}
		}
	}
	return &ErrorModel{
		Status: status,
		Title:  http.StatusText(status),
		Detail: msg,
		Errors: details,
	}
}
