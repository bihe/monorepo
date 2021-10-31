package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// Status enumerates the available states
// swagger:enum Status
type Status string

const (
	// OK indicates all is good
	OK Status = "Ok"
	// Error indicates an error
	Error Status = "Error"
)

// HealthCheck returns information about the health of a service
// swagger:model
type HealthCheck struct {
	Status    Status    `json:"status"`
	Message   string    `json:"message"`
	Version   string    `json:"version"`
	TimeStamp time.Time `json:"timestamp"`
}

// String prints the request
func (h HealthCheck) String() string {
	return fmt.Sprintf("%s (%s)", h.Status, h.TimeStamp)
}

// --------------------------------------------------------------------------
// Request and Response objects using go-chi render
// --------------------------------------------------------------------------

// HealthCheckResponse returns a JSON repsonse
type HealthCheckResponse struct {
	*HealthCheck
}

// Render the specific response
func (a HealthCheckResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}

// --------------------------------------------------------------------------
// Interface definitions
// --------------------------------------------------------------------------

// HealthChecker defines an interface which is used to create HealthCheck result
type HealthChecker interface {
	// Check performs the service health-check
	Check(user security.User) (HealthCheck, error)
}

// --------------------------------------------------------------------------
// Handler implementation
// --------------------------------------------------------------------------

// HealthCheckHandler is responsnible to return a JSON formated health-check result
type HealthCheckHandler struct {
	Handler
	Checker HealthChecker
}

// GetHandler returns the health-check handler
func (h *HealthCheckHandler) GetHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.check())
	return r
}

// --------------------------------------------------------------------------
// internal handler methods
// --------------------------------------------------------------------------

// check wraps the getHealth handler to have a common signature
func (h *HealthCheckHandler) check() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := security.UserFromContext(r.Context())
		if !ok || user == nil {
			h.Log.Error("check: user is not available in context!")
			errors.WriteError(w, r, errors.SecurityError{Err: fmt.Errorf("user is not available in context"), Request: r})
			return
		}

		if err := h.getHealth(*user, w, r, h.Checker); err != nil {
			h.Log.Error("check: error in health-check function", logging.ErrV(fmt.Errorf("error during health-check call %v", err)))
			errors.WriteError(w, r, err)
			return
		}
	})
}

// getHealth returns health-check info about the service
func (h *HealthCheckHandler) getHealth(user security.User, w http.ResponseWriter, r *http.Request, checker HealthChecker) error {
	h.Log.Debug("check for health")
	health, err := checker.Check(user)
	if err != nil {
		h.Log.Error("getHealth: error from Check function", logging.ErrV(fmt.Errorf("error during health-check call %v", err)))
		return err
	}
	h.Log.Info(fmt.Sprintf("health: %s", health))
	if health.Status != OK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	return render.Render(w, r, HealthCheckResponse{HealthCheck: &health})
}
