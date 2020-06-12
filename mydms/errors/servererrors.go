package errors

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// --------------------------------------------------------------------------
// Error responses adhering to https://tools.ietf.org/html/rfc7807
// --------------------------------------------------------------------------

// ProblemDetail combines the fields defined in RFC7807
//
// "Note that both "type" and "instance" accept relative URIs; this means
// that they must be resolved relative to the document's base URI"
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

// RedirectError is a specific error indicating a necessary redirect
type RedirectError struct {
	Err     error
	Request *http.Request
	Status  int
	URL     string
}

// Error implements the error interface
func (e RedirectError) Error() string {
	return fmt.Sprintf("the request '%s' resulted in a redirect to: '%s', error: %v", e.Request.RequestURI, e.URL, e.Err)
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

// ErrRedirectError returns a http.StatusTemporaryRedirect
func ErrRedirectError(err RedirectError) *ProblemDetail {
	return &ProblemDetail{
		Type:     t,
		Title:    "authentication requires a redirect",
		Status:   err.Status,
		Detail:   err.Error(),
		Instance: err.URL,
	}
}

// CustomErrorHandler centrally handles errors of API handlers
func CustomErrorHandler(err error, c echo.Context) {
	var e *ProblemDetail
	// decide based on error-type what to do
	// the basic distinction is just between a browser request (requesting HTML)
	// and API usage, where JSON is returned
	content := NegotiateContent(c)

	if notfound, ok := err.(NotFoundError); ok {
		e = ErrNotFound(notfound)
		_ = c.JSON(e.Status, e)
		return
	}

	if badrequest, ok := err.(BadRequestError); ok {
		e = ErrBadRequest(badrequest)
		_ = c.JSON(e.Status, e)
		return
	}

	if redirect, ok := err.(RedirectError); ok {
		e = ErrRedirectError(redirect)
		switch content {
		case HTML:
			_ = c.Redirect(http.StatusTemporaryRedirect, redirect.URL)
		default:
			status := http.StatusTemporaryRedirect
			if e.Status > 0 {
				status = e.Status
			}
			_ = c.JSON(status, e)
		}
		return

	}

	// any other case we just print an internal server error
	e = ErrServerError(ServerError{Err: err, Request: c.Request()})
	_ = c.JSON(e.Status, e)
}
