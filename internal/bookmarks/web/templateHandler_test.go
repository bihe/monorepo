package web_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	bm "golang.binggl.net/monorepo/internal/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/internal/bookmarks/web"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

var logger = logging.NewNop()

type handlerOps struct {
	version string
	build   string
	app     *bookmarks.Application
}

func bookmarkHandler(bRepo store.BookmarkRepository, fRepo store.FaviconRepository) http.Handler {
	ops := handlerOps{
		version: "1.0",
		build:   "today",
		app: &bookmarks.Application{
			Logger:        logger,
			BookmarkStore: bRepo,
			FavStore:      fRepo,
			FaviconPath:   "/tmp",
		},
	}
	return bm.MakeHTTPHandler(ops.app, logger, bm.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		Config: conf.AppConfig{
			BaseConfig: config.BaseConfig{
				Cookies: config.ApplicationCookies{
					Prefix: "core",
				},
				Security: config.Security{
					CookieName: "jwt",
					JwtIssuer:  "issuer",
					JwtSecret:  "secret",
					Claim: config.Claim{
						Name:  "A",
						URL:   "http://A",
						Roles: []string{"A"},
					},
				},
				Logging: config.LogConfig{},
				Cors:    config.CorsSettings{},
				Assets: config.AssetSettings{
					AssetDir:    "./",
					AssetPrefix: "/static",
				},
			},
		},
		Version: ops.version,
		Build:   ops.build,
	})
}

func repositories(t *testing.T) (store.BookmarkRepository, store.FaviconRepository, *sql.DB) {
	var (
		DB  *gorm.DB
		err error
	)
	if DB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&store.Bookmark{}, &store.Favicon{})
	db, err := DB.DB()
	if err != nil {
		t.Fatalf("cannot access database handle: %v", err)
	}
	logger := logging.NewNop()
	return store.CreateBookmarkRepo(DB, logger), store.CreateFaviconRepo(DB, logger), db
}

func addJwtAuth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
}

func Test_Page_403(t *testing.T) {
	repo, fRepo, db := repositories(t)
	defer db.Close()

	r := bookmarkHandler(repo, fRepo)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bm/403", nil)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func Test_Bookmark_Search(t *testing.T) {
	repo, fRepo, db := repositories(t)
	defer db.Close()

	r := bookmarkHandler(repo, fRepo)

	// without authentication the redirect should be sent
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bm/search", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	if rec.Header().Get("Location") != "/bm/403" {
		t.Errorf("expected redirect to '403' got '%s'", rec.Header().Get("Location"))
	}

	// add jwt auth
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/bm/search", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	payload := rec.Body.String()
	if !strings.Contains(payload, "Search Bookmarks!") {
		t.Errorf("cannot find string in page result")
	}

	if !strings.Contains(payload, "searching for") {
		t.Errorf("cannot find string in page result")
	}
}

func Test_Bookmark_ForPath(t *testing.T) {
	repo, fRepo, db := repositories(t)
	defer db.Close()

	r := bookmarkHandler(repo, fRepo)

	// without authentication the redirect should be sent
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bm/~", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	if rec.Header().Get("Location") != "/bm/403" {
		t.Errorf("expected redirect to '403' got '%s'", rec.Header().Get("Location"))
	}

	// add jwt auth
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/bm/~", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	payload := rec.Body.String()
	if !strings.Contains(payload, "Bookmarks!") {
		t.Errorf("cannot find string in page result")
	}

	if !strings.Contains(payload, "path") {
		t.Errorf("cannot find string in page result")
	}
}

func Test_EllipsisValue(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bm/~", nil)
	el := web.GetEllipsisValues(req)

	// for the standard case without anything the standard ellipsis values should be returned
	if el.PathLen != web.StdEllipsis.PathLen &&
		el.NodeLen != web.StdEllipsis.NodeLen &&
		el.FolderLen != web.StdEllipsis.FolderLen {
		t.Errorf("expected the std ellipsis - got something different")
	}

	// we have a viewport cookie
	req, _ = http.NewRequest("GET", "/bm/~", nil)
	req.AddCookie(&http.Cookie{
		Name:  "viewport",
		Value: "1024:768",
	})

	el = web.GetEllipsisValues(req)
	if el.PathLen != web.StdEllipsis.PathLen &&
		el.NodeLen != web.StdEllipsis.NodeLen &&
		el.FolderLen != web.StdEllipsis.FolderLen {
		t.Errorf("expected the std ellipsis - got something different")
	}

	// we have a mobile viewport
	req, _ = http.NewRequest("GET", "/bm/~", nil)
	req.AddCookie(&http.Cookie{
		Name:  "viewport",
		Value: "320:560",
	})

	el = web.GetEllipsisValues(req)
	if el.PathLen != web.MobileEllipsis.PathLen &&
		el.NodeLen != web.MobileEllipsis.NodeLen &&
		el.FolderLen != web.MobileEllipsis.FolderLen {
		t.Errorf("expected the std ellipsis - got something different")
	}
}
