package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
)

func TestErrorPage(t *testing.T) {
	r := chi.NewRouter()

	cookieSettings := cookies.Settings{
		Path:   "/",
		Domain: "localhost",
	}

	errHandler := &TemplateHandler{
		Handler: Handler{
			ErrRep: &errors.ErrorReporter{
				CookieSettings: cookieSettings,
				ErrorPath:      "error",
			},
		},
		Version:        "1",
		Build:          "2",
		CookieSettings: cookieSettings,
		TemplateDir:    "./",
		AppName:        "app",
	}

	r.Get("/error", errHandler.Call(errHandler.HandleError))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	req.AddCookie(&http.Cookie{
		Path:   cookieSettings.Path,
		Domain: cookieSettings.Domain,
		Name:   "_" + errors.FlashKeyError,
		Value:  "error",
	})
	req.AddCookie(&http.Cookie{
		Path:   cookieSettings.Path,
		Domain: cookieSettings.Domain,
		Name:   "_" + errors.FlashKeyInfo,
		Value:  "info",
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	if body == "" {
		t.Errorf("could not get the template!")
	}
}
