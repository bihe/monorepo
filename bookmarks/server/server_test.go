package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"golang.binggl.net/monorepo/bookmarks"
	"golang.binggl.net/monorepo/bookmarks/config"
	"golang.binggl.net/monorepo/pkg/handler"

	_ "github.com/jinzhu/gorm/dialects/sqlite" // use sqlite for testing
)

const userName = "username"

var logger = log.New().WithField("mode", "test")

var appCfg = config.AppConfig{
	Database: config.Database{
		Dialect:          "sqlite3",
		ConnectionString: ":memory:",
	},
	Cookies: config.ApplicationCookies{
		Domain: "localhost",
		Path:   "/",
		Prefix: "app",
		Secure: false,
	},
	Logging: config.LogConfig{
		LogLevel: "Debug",
		FilePath: "./",
	},
	Security: config.Security{
		JwtSecret:  "secret",
		JwtIssuer:  "issuer",
		CookieName: "cookie",
		Claim: config.Claim{
			Name:  "claim",
			URL:   "http://localhost:3000",
			Roles: []string{"role"},
		},
		LoginRedirect: "/redirect",
		CacheDuration: "10m",
	},
	AppName:           "bookmarks",
	Environment:       "Development",
	ErrorPath:         "/error",
	HostID:            "test",
	FaviconUploadPath: "./upload",
	DefaultFavicon:    "./favicon.ico ",
}

const cookie = "cookie"
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4NzAzOTI2NDcsImp0aSI6IjZmYWQ1YzAwLWZlZTItNDU5Yy1hYmFkLTIwNDU3Y2ZmM2Q4YSIsImlhdCI6MTU1OTc4Nzg0NywiaXNzIjoiaXNzdWVyIiwic3ViIjoidXNlciIsIlR5cGUiOiJsb2dpbi5Vc2VyIiwiRGlzcGxheU5hbWUiOiJEaXNwbGF5IE5hbWUiLCJFbWFpbCI6ImEuYkBjLmRlIiwiVXNlcklkIjoiMTIzNDUiLCJVc2VyTmFtZSI6ImEuYkBjLmRlIiwiR2l2ZW5OYW1lIjoiRGlzcGxheSIsIlN1cm5hbWUiOiJOYW1lIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3Q6MzAwMHxyb2xlIl19.qUwvHXBmV_FuwLtykOnzu3AMbxSqrg82bQlAi3Nabyo"

func TestPostAuthRedirectRoutes(t *testing.T) {
	// arrange
	srv := Create("../", appCfg, bookmarks.VersionInfo{}, logger)
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("GET", "/?valid=true&ref=/api/v1/bookmarks/fetch/fd249874-b8ee-44eb-942b-18328ec7530a", nil)
	req.AddCookie(&http.Cookie{Name: cookie, Value: token})

	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/api/v1/bookmarks/fetch/fd249874-b8ee-44eb-942b-18328ec7530a", rec.Header().Get("Location"))
}

func TestHealthChck(t *testing.T) {
	// arrange
	srv := Create("../", appCfg, bookmarks.VersionInfo{}, logger)
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("GET", "/hc", nil)
	req.AddCookie(&http.Cookie{Name: cookie, Value: token})

	srv.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result handler.HealthCheckResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, handler.OK, result.Status)
}
