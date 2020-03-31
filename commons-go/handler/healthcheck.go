package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"golang.binggl.net/commons"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/security"
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

// GetRouter returns the health-check handler mounted to /hc
func (h *HealthCheckHandler) GetRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/hc", h.check())
	return r
}

// --------------------------------------------------------------------------
// internal handler methods
// --------------------------------------------------------------------------

// Check wraps handlers to have a common signature
func (h *HealthCheckHandler) check() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := security.UserFromContext(r.Context())
		if !ok || user == nil {
			commons.LogWithReq(r, h.Log, "handler.Check").Errorf("user is not available in context!")
			h.ErrRep.Negotiate(w, r, errors.SecurityError{Err: fmt.Errorf("user is not available in context"), Request: r})
			return
		}

		if err := h.getHealth(*user, w, r, h.Checker); err != nil {
			commons.LogWithReq(r, h.Log, "handler.Check").Errorf("error during health-check call %v\n", err)
			h.ErrRep.Negotiate(w, r, errors.ServerError{Err: err, Request: r})
			return
		}
	})
}

// getHealth returns health-check info about the service
func (h *HealthCheckHandler) getHealth(user security.User, w http.ResponseWriter, r *http.Request, check HealthChecker) error {
	commons.LogWithReq(r, h.Log, "handler.GetHealth").Debugf("check for health")
	health, err := check.Check(user)
	if err != nil {
		commons.LogWithReq(r, h.Log, "handler.GetHealth").Errorf("error during health-check call %v\n", err)
		return fmt.Errorf("error during health-check: %v", err)
	}
	commons.LogWithReq(r, h.Log, "handler.GetHealth").Infof("health: %s", health)
	return render.Render(w, r, HealthCheckResponse{HealthCheck: &health})
}
