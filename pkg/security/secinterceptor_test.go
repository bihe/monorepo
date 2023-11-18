package security_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

func setup() security.SecInterceptor {
	options := security.JwtOptions{
		CacheDuration: "30s",
		CookieName:    "jwt",
		ErrorPath:     "/error",
		JwtIssuer:     "issuer",
		JwtSecret:     "secret",
		RedirectURL:   "login_redirect",
		RequiredClaim: security.Claim{
			Name:  "A",
			URL:   "http://A",
			Roles: []string{"A"},
		},
	}
	return security.SecInterceptor{
		Log:           logging.NewNop(),
		Auth:          security.NewJWTAuthorization(options, false),
		Options:       options,
		ErrorRedirect: "/error",
	}
}

func Test_Std_SecInterceptor(t *testing.T) {
	sec := setup()
	r := chi.NewRouter()
	r.Use(sec.HandleJWT)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/error", rec.Header().Get("Location"))
}

func Test_Api_SecInterceptor(t *testing.T) {
	sec := setup()
	r := chi.NewRouter()
	r.Use(sec.HandleJWT)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Accept", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var payload pkgerr.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, http.StatusUnauthorized, payload.Status)
}

func Test_SecInterceptor_Success(t *testing.T) {
	sec := setup()
	r := chi.NewRouter()
	r.Use(sec.HandleJWT)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
