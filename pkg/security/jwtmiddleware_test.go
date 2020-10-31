package security

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"

	"github.com/stretchr/testify/assert"
)

const cookie = "cookie"
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4NzAzOTI2NDcsImp0aSI6IjZmYWQ1YzAwLWZlZTItNDU5Yy1hYmFkLTIwNDU3Y2ZmM2Q4YSIsImlhdCI6MTU1OTc4Nzg0NywiaXNzIjoiaXNzdWVyIiwic3ViIjoidXNlciIsIlR5cGUiOiJsb2dpbi5Vc2VyIiwiRGlzcGxheU5hbWUiOiJEaXNwbGF5IE5hbWUiLCJFbWFpbCI6ImEuYkBjLmRlIiwiVXNlcklkIjoiMTIzNDUiLCJVc2VyTmFtZSI6ImEuYkBjLmRlIiwiR2l2ZW5OYW1lIjoiRGlzcGxheSIsIlN1cm5hbWUiOiJOYW1lIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3Q6MzAwMHxyb2xlIl19.qUwvHXBmV_FuwLtykOnzu3AMbxSqrg82bQlAi3Nabyo"
const path = "/JWT"
const unmarshal = "could not unmarshal problemdetails: %v"

var jwtOpts = JwtOptions{
	JwtSecret:  "secret",
	JwtIssuer:  "issuer",
	CookieName: cookie,
	RequiredClaim: Claim{
		Name:  "claim",
		URL:   "http://localhost:3000",
		Roles: []string{"role"},
	},
	RedirectURL:   "/redirect",
	CacheDuration: "10m",
}

var cookieSettings = cookies.Settings{
	Domain: "localhost",
	Path:   "/",
	Secure: false,
}
var logger = logging.NewNop()

func getJWTMiddleware() *JwtMiddleware {
	return NewJwtMiddleware(jwtOpts, cookieSettings, logger)
}

func TestJWTMiddlewareWrongJWTOptions(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r := chi.NewRouter()
	wrongOpts := jwtOpts
	wrongOpts.CacheDuration = "wrong"
	jwt := NewJwtMiddleware(wrongOpts, cookieSettings, logger)

	req.AddCookie(&http.Cookie{Name: cookie, Value: token})

	assert.Panics(t, func() {
		r.Use(jwt.JwtContext)
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		})
		r.ServeHTTP(rec, req)
	}, "panics because of wrong duration!")
}

func TestJWTMiddlewareCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r := chi.NewRouter()
	jwt := getJWTMiddleware()

	req.AddCookie(&http.Cookie{Name: cookie, Value: token})
	r.Use(jwt.JwtContext)
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTMiddlewareCookieAndCache(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r := chi.NewRouter()
	jwt := getJWTMiddleware()

	req.AddCookie(&http.Cookie{Name: cookie, Value: token})
	r.Use(jwt.JwtContext)
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTMiddlewareBearer(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r := chi.NewRouter()
	jwt := getJWTMiddleware()

	req.Header.Add("Authorization", "Bearer "+token)
	r.Use(jwt.JwtContext)
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTMiddlewareNoToken(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r := chi.NewRouter()
	jwt := getJWTMiddleware()

	r.Use(jwt.JwtContext)
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf(unmarshal, err)
	}
	assert.Equal(t, http.StatusUnauthorized, p.Status)
}

func TestJWTMiddlewareWrongToken(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r := chi.NewRouter()
	jwt := getJWTMiddleware()

	req.Header.Add("Authorization", "Bearer "+"token")
	r.Use(jwt.JwtContext)
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf(unmarshal, err)
	}
	assert.Equal(t, http.StatusForbidden, p.Status)
}
