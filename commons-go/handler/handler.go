// Package handler provides common functions needed to create API handlers as well as some utilities
package handler // import "golang.binggl.net/commons/handler"

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/security"
)

// Handler defines common handler logic
type Handler struct {
	// ErrRep is used to send errors according to the users accept headers
	ErrRep *errors.ErrorReporter
}

// LogFieldNameFunction defines the field-name for structured logging of cuntions
const LogFieldNameFunction = "function"

// LogFunction uses the function name to perform structured logging with a field
func LogFunction(name string) *log.Entry {
	return log.WithField(LogFieldNameFunction, name)
}

// Secure wraps handlers to have a common signature
// a User is retrieved from the context and a possible error from the handler function is processed
func (h *Handler) Secure(f func(user security.User, w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(security.UserKey)
		if u == nil {
			LogFunction("handler.Secure").Errorf("user is not available in context!")
			h.ErrRep.Negotiate(w, r, fmt.Errorf("user is not available in context"))
			return
		}
		user := r.Context().Value(security.UserKey).(*security.User)
		if err := f(*user, w, r); err != nil {
			LogFunction("handler.Secure").Errorf("error during API call %v\n", err)
			h.ErrRep.Negotiate(w, r, err)
			return
		}
	})
}

// Call wraps handlers to have a common signature
func (h *Handler) Call(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			LogFunction("handler.Call").Errorf("error during API call %v\n", err)
			h.ErrRep.Negotiate(w, r, err)
			return
		}
	})
}
