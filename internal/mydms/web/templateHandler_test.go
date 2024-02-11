package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms"
	"golang.binggl.net/monorepo/internal/mydms/app/config"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	conf "golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
)

const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

var logger = logging.NewNop()

func handler(repo document.Repository) http.Handler {

	svc := document.NewService(logger,
		repo,
		nil, /* filestore.FileService */
		nil, /* upload.Cleint */
	)

	return mydms.MakeHTTPHandler(svc, logger, mydms.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		Config: config.AppConfig{
			BaseConfig: conf.BaseConfig{
				Cookies: conf.ApplicationCookies{
					Prefix: "core",
				},
				Security: conf.Security{
					CookieName: "jwt",
					JwtIssuer:  "issuer",
					JwtSecret:  "secret",
					Claim: conf.Claim{
						Name:  "A",
						URL:   "http://A",
						Roles: []string{"A"},
					},
				},
				Logging: conf.LogConfig{},
				Cors:    conf.CorsSettings{},
				Assets: conf.AssetSettings{
					AssetDir:    "./",
					AssetPrefix: "/public",
				},
			},
		},
		Version: "local",
		Build:   "testing",
	})
}

func memRepo(t *testing.T) (document.Repository, persistence.Connection) {
	con := persistence.NewConnForDb("sqlite3", "file::memory:")
	repo, err := document.NewRepository(con)
	if err != nil {
		t.Fatalf("cannot establish database connection: %v", err)
	}
	return repo, con
}

func addJwtAuth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
}

func Test_Page_403(t *testing.T) {
	repo, con := memRepo(t)
	defer con.Close()

	r := handler(repo)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mydms/403", nil)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func Test_Document_List(t *testing.T) {
	repo, con := memRepo(t)
	defer con.Close()

	r := handler(repo)

	// without auth we get an 403
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mydms", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("location")
	if location != "/mydms/403" {
		t.Errorf("expected a redirect to '%s' but got '%s'", "/mydms/403", location)
	}

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", location, nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// add the JWT auth
	// ------------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/mydms", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// this is an empty database - so we should receive the message "No results available!"
	payload := rec.Body.String()
	if !strings.Contains(payload, "No results available!") {
		t.Errorf("cannot find string in page result")
	}
}

func Test_Document_Partial(t *testing.T) {
	repo, con := memRepo(t)
	defer con.Close()

	formData := url.Values{}
	formData.Set("q", "")
	formData.Set("skip", "0")

	r := handler(repo)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/mydms/partial", strings.NewReader(formData.Encode()))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// this is an empty database - so we should receive the message "No results available!"
	payload := rec.Body.String()
	if !strings.Contains(payload, "No results available!") {
		t.Errorf("cannot find string in page result")
	}

	// provide wrong skip value
	formData = url.Values{}
	formData.Set("q", "")
	formData.Set("skip", "this_is_not_a_number")

	r = handler(repo)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/mydms/partial", strings.NewReader(formData.Encode()))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// negative skip-value
	formData = url.Values{}
	formData.Set("q", "")
	formData.Set("skip", "-100")

	r = handler(repo)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/mydms/partial", strings.NewReader(formData.Encode()))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// incorrect form data
	r = handler(repo)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/mydms/partial", nil)
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
