// Package handler provides common functions needed to create API handlers as well as some utilities
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// Handler defines common handler logic
type Handler struct {
	// Log is the supplied log-handler
	Log logging.Logger
}

// Secure wraps handlers to have a common signature
// a User is retrieved from the context and a possible error from the handler function is processed
func (h *Handler) Secure(f func(user security.User, w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := security.UserFromContext(r.Context())
		if !ok || user == nil {
			h.Log.InfoRequest("user is not available in context!", r)
			errors.WriteError(w, r, fmt.Errorf("user is not available in context"))
			return
		}
		if err := f(*user, w, r); err != nil {
			h.Log.Error("Secure: function returned an error", logging.ErrV(fmt.Errorf("error during API call %v", err)))
			errors.WriteError(w, r, err)
			return
		}
	})
}

// Call wraps handlers to have a common signature
func (h *Handler) Call(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			h.Log.Error("Call: function returned an error", logging.ErrV(fmt.Errorf("error during API call %v", err)))
			errors.WriteError(w, r, err)
			return
		}
	})
}

// RespondJSON returns a JSON formatted payload
func RespondJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
