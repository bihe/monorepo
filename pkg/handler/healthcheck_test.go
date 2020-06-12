package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/security"
)

var hc = &HealthCheckHandler{
	Handler: Handler{
		ErrRep: &errors.ErrorReporter{
			CookieSettings: cookies.Settings{
				Prefix: "test",
			},
		},
		Log: log.New().WithField("mode", "test"),
	},
	Checker: simpleChecker{},
}

type simpleChecker struct {
	fail bool
}

func (s simpleChecker) Check(user security.User) (HealthCheck, error) {
	if s.fail {
		return HealthCheck{}, fmt.Errorf("error")
	}

	return HealthCheck{
		Status:    OK,
		Message:   "all is working",
		Version:   "1",
		TimeStamp: time.Now().UTC(),
	}, nil
}

var user = security.User{
	Username:    "username",
	Email:       "a.b@c.de",
	DisplayName: "displayname",
	Roles:       []string{"role"},
	UserID:      "12345",
}

func TestHealthCheck(t *testing.T) {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Mount("/hc", hc.GetHandler())

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hc", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var h HealthCheck
	err := json.Unmarshal(rec.Body.Bytes(), &h)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}

	assert.Equal(t, OK, h.Status)
	assert.Equal(t, "all is working", h.Message)
	assert.Equal(t, "1", h.Version)
	assert.True(t, h.TimeStamp.After(time.Now().Add(time.Duration(-1)*time.Minute)))
}

func TestHealthCheckNoUser(t *testing.T) {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	hc.Checker = simpleChecker{fail: true}
	r.Mount("/hc", hc.GetHandler())

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hc", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var pd errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &pd)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}

	assert.Equal(t, http.StatusInternalServerError, pd.Status)
}

func TestHealthCheckFail(t *testing.T) {
	r := chi.NewRouter()

	r.Mount("/hc", hc.GetHandler())

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hc", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var pd errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &pd)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}

	assert.Equal(t, http.StatusForbidden, pd.Status)
}
