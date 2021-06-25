package gway_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/gway"
	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"

	pkgerr "golang.binggl.net/monorepo/pkg/errors"
)

// --------------------------------------------------------------------------
//   Mock Service implementation
// --------------------------------------------------------------------------

type mockService struct {
	fail bool
}

func (m *mockService) PrepIntOIDCRedirect() (url string, state string) {
	return "initial-redirect", "state"
}

func (m *mockService) GetExtOIDCRedirect(state string) (url string, err error) {
	if m.fail {
		return "", fmt.Errorf("error")
	}
	return "oidc-redirect", nil
}

func (m *mockService) LoginOIDC(state, oidcState, oidcCode string) (token, url string, err error) {
	return "token", "redirect", nil
}

func (m *mockService) LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error) {
	return "token", "redirect", nil
}

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

var logger = logging.NewNop()

type handlerOps struct {
	oidcSvc oidc.Service
}

func handler() http.Handler {
	return handlerWith(nil)
}

func handlerWith(ops *handlerOps) http.Handler {
	var svc oidc.Service
	svc = &mockService{}

	if ops != nil && ops.oidcSvc != nil {
		svc = ops.oidcSvc
	}

	return gway.MakeHTTPHandler(svc, logger, gway.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		JWTConfig: config.Security{},
		CookieConfig: config.ApplicationCookies{
			Prefix: "gway",
		},
		CorsConfig: config.CorsSettings{},
		AssetConfig: config.AssetSettings{
			AssetDir:    "./",
			AssetPrefix: "/static",
		},
	})
}

// --------------------------------------------------------------------------
//   The Test-Cases itself
// --------------------------------------------------------------------------

func Test_HandleOIDC_Prepare(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/start-oidc", nil)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/initial-redirect", rec.Header().Get("Location"))
}

func Test_HandleOIDC_Redirect(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", oidc.OIDCInitiateURL, nil)
	req.AddCookie(&http.Cookie{Name: "gway_state", Value: "state"})

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/oidc-redirect", rec.Header().Get("Location"))

	// ---- missing state ----

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", oidc.OIDCInitiateURL, nil)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	var pd pkgerr.ProblemDetail
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
	assert.NotEmpty(t, pd.Detail)

	// ---- general error ----

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", oidc.OIDCInitiateURL, nil)
	req.AddCookie(&http.Cookie{Name: "gway_state", Value: "state"})

	// act
	handlerWith(&handlerOps{
		oidcSvc: &mockService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
	assert.NotEmpty(t, pd.Detail)

}
