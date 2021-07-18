// Package errors provides RFC7807 problem-details and negotiates HTML or JSON responses
// the HTML response is in the form of a redirect to an error page
package errors

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// --------------------------------------------------------------------------
// Error responses adhering to https://tools.ietf.org/html/rfc7807
// --------------------------------------------------------------------------

// ProblemDetail combines the fields defined in RFC7807
//
// "Note that both "type" and "instance" accept relative URIs; this means
// that they must be resolved relative to the document's base URI"
// swagger:model
type ProblemDetail struct {
	// Type is a URI reference [RFC3986] that identifies the
	// problem type.  This specification encourages that, when
	// dereferenced, it provide human-readable documentation for the problem
	Type string `json:"type"`
	// Title is a short, human-readable summary of the problem type
	Title string `json:"title"`
	// Status is the HTTP status code
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence of the problem
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence of the problem
	Instance string `json:"instance,omitempty"`
}

// --------------------------------------------------------------------------
// Specific Errors
// --------------------------------------------------------------------------

// NotFoundError is used when a given object cannot be found
type NotFoundError struct {
	Err     error
	Request *http.Request
}

// Error implements the error interface
func (e NotFoundError) Error() string {
	return fmt.Sprintf("the object for request '%s' cannot be found: %v", e.Request.RequestURI, e.Err)
}

// BadRequestError indicates that the client request cannot be fulfilled
type BadRequestError struct {
	Err     error
	Request *http.Request
}

// Error implements the error interface
func (e BadRequestError) Error() string {
	return fmt.Sprintf("the request '%s' cannot be fulfilled because: %v", e.Request.RequestURI, e.Err)
}

// ServerError is used when an unexpected situation occurred
type ServerError struct {
	Err     error
	Request *http.Request
}

// Error implements the error interface
func (e ServerError) Error() string {
	return fmt.Sprintf("the request '%s' resulted in an unexpected error: %v", e.Request.RequestURI, e.Err)
}

// SecurityError is used when something is not allowed
type SecurityError struct {
	Err     error
	Request *http.Request
	Status  uint
}

// Error implements the error interface
func (e SecurityError) Error() string {
	return fmt.Sprintf("the request '%s' is not allowed: %v", e.Request.RequestURI, e.Err)
}

// --------------------------------------------------------------------------
// Shortcuts for commen error responses
// --------------------------------------------------------------------------

const t = "about:blank"

// ErrBadRequest returns a http.StatusBadRequest
func ErrBadRequest(err BadRequestError) *ProblemDetail {
	return &ProblemDetail{
		Type:   t,
		Title:  "the request cannot be fulfilled",
		Status: http.StatusBadRequest,
		Detail: err.Error(),
	}
}

// ErrNotFound returns a http.StatusNotFound
func ErrNotFound(err NotFoundError) *ProblemDetail {
	return &ProblemDetail{
		Type:   t,
		Title:  "object cannot be found",
		Status: http.StatusNotFound,
		Detail: err.Error(),
	}
}

// ErrServerError returns a http.StatusInternalServerError
func ErrServerError(err ServerError) *ProblemDetail {
	return &ProblemDetail{
		Type:   t,
		Title:  "cannot service the request",
		Status: http.StatusInternalServerError,
		Detail: err.Error(),
	}
}

// ErrSecurityError returns a http.StatusForbidden
func ErrSecurityError(err SecurityError) *ProblemDetail {
	status := http.StatusForbidden
	if err.Status > 0 {
		status = int(err.Status)
	}
	return &ProblemDetail{
		Type:   t,
		Title:  "operation is not allowed",
		Status: status,
		Detail: err.Error(),
	}
}

// --------------------------------------------------------------------------
// Error handling
// --------------------------------------------------------------------------

// WriteError sets the correct status code and returns error information as a ProblemDetail
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		pd *ProblemDetail
	)

	// default error is the server-error
	if svrErr, ok := err.(ServerError); ok {
		pd = ErrServerError(svrErr)
	} else {
		pd = ErrServerError(ServerError{Err: err, Request: r})
	}

	if notfound, ok := err.(NotFoundError); ok {
		pd = ErrNotFound(notfound)
	}

	if badrequest, ok := err.(BadRequestError); ok {
		pd = ErrBadRequest(badrequest)
	}

	if security, ok := err.(SecurityError); ok {
		pd = ErrSecurityError(security)
	}

	status := http.StatusInternalServerError
	if pd.Status > 0 {
		status = pd.Status
	}
	writeProblemJSON(w, status, pd)
}

func writeProblemJSON(w http.ResponseWriter, code int, pd *ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(code)
	b, err := json.Marshal(pd)
	if err != nil {
		log.Printf("writeProblemJSON: could not marshal json %v", err)
	}
	_, err = w.Write(b)
	if err != nil {
		log.Printf("writeProblemJSON: could not write bytes using http.ResponseWriter: %v", err)
	}
}
