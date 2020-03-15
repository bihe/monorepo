package handler // import "golang.binggl.net/commons/handler"

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	"golang.binggl.net/commons/security"
)

const (
	version = "1"
	build   = "2"
)

var handler = &AppInfoHandler{
	Handler: Handler{
		ErrRep: &errors.ErrorReporter{
			CookieSettings: cookies.Settings{
				Prefix: "test",
			},
		},
	},
	Version: version,
	Build:   build,
}

func TestGetAppInfo(t *testing.T) {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), security.UserKey, &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/appinfo", handler.Secure(handler.HandleAppInfo))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/appinfo", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var m Meta
	err := json.Unmarshal(rec.Body.Bytes(), &m)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}

	assert.Equal(t, "1", m.Version.Version)
	assert.Equal(t, "2", m.Version.Build)
}

func TestGetAppInfoNilUser(t *testing.T) {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), security.UserKey, nil)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/appinfo", handler.Secure(handler.HandleAppInfo))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/appinfo", nil)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, http.StatusInternalServerError, p.Status)
}
