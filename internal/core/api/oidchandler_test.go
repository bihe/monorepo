package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/oidc"

	pkgerr "golang.binggl.net/monorepo/pkg/errors"
)

// --------------------------------------------------------------------------
//   Mock Service implementation
// --------------------------------------------------------------------------

type mockOidcService struct {
	fail bool
}

func (m *mockOidcService) PrepIntOIDCRedirect() (url string, state string) {
	return "/oidc/redirect", "state"
}

func (m *mockOidcService) GetExtOIDCRedirect(state string) (url string, err error) {
	if m.fail {
		return "", fmt.Errorf("error")
	}
	return "/oidc_provider-redirect", nil
}

func (m *mockOidcService) LoginOIDC(state, oidcState, oidcCode string) (token, url string, err error) {
	if m.fail {
		return "", "", fmt.Errorf("error")
	}
	return "token", "/redirect", nil
}

func (m *mockOidcService) LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error) {
	return "token", redirectURL, nil
}

func oidcHandler() http.Handler {
	return handlerWith(&handlerOps{
		oidcSvc: &mockOidcService{},
	})
}

// --------------------------------------------------------------------------
//   The Test-Cases itself
// --------------------------------------------------------------------------

func Test_HandleOIDC_Prepare(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/oidc/start", nil)

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/oidc/redirect", rec.Header().Get("Location"))
}

func Test_HandleOIDC_Auth_Flow(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/oidc/auth/flow?~site=A&~url=http://a", nil)

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/oidc/redirect", rec.Header().Get("Location"))

	header := http.Header{}
	cookies := rec.Header().Values("Set-Cookie")
	for _, c := range cookies {
		header.Add("Cookie", c)
	}
	r := http.Request{Header: header}

	// state cookie
	c, err := r.Cookie("core_state")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	// auth_flow cookie
	c, err = r.Cookie("core_auth_flow")

	// state cookie
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "A|http://a", c.Value)

	// ---- missing parameters ----
	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/oidc/auth/flow?~site=&~url=", nil)

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func Test_HandleOIDC_Redirect(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", oidc.OIDCInitiateURL, nil)
	req.AddCookie(&http.Cookie{Name: "core_state", Value: "state"})

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/oidc_provider-redirect", rec.Header().Get("Location"))

	// ---- missing state ----

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", oidc.OIDCInitiateURL, nil)

	// act
	oidcHandler().ServeHTTP(rec, req)

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
	req.AddCookie(&http.Cookie{Name: "core_state", Value: "state"})

	// act
	handlerWith(&handlerOps{
		oidcSvc: &mockOidcService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
	assert.NotEmpty(t, pd.Detail)
}

func Test_HandleOIDC_Login(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/oidc/signin", nil)
	req.AddCookie(&http.Cookie{Name: "core_state", Value: "state"})

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/redirect", rec.Header().Get("Location"))

	// ---- general error ----

	var pd pkgerr.ProblemDetail
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/oidc/signin", nil)
	req.AddCookie(&http.Cookie{Name: "core_state", Value: "state"})

	// act
	handlerWith(&handlerOps{
		oidcSvc: &mockOidcService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
	assert.NotEmpty(t, pd.Detail)

	// ---- missing state ----

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/oidc/signin", nil)

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
	assert.NotEmpty(t, pd.Detail)

	// TODO: test failed login!
}

func Test_HandleOIDC_Login_AuthFlow(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/oidc/signin", nil)
	req.AddCookie(&http.Cookie{Name: "core_state", Value: "state"})
	req.AddCookie(&http.Cookie{Name: "core_auth_flow", Value: "a|http://redirectA"})

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "http://redirectA", rec.Header().Get("Location"))

	// ---- wrong auth-flow ----

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/oidc/signin", nil)
	req.AddCookie(&http.Cookie{Name: "core_state", Value: "state"})
	req.AddCookie(&http.Cookie{Name: "core_auth_flow", Value: "http://redirectA"})

	// act
	oidcHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
