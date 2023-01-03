package api_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/bookmarks-new/api"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var errRaised = fmt.Errorf("error")

const failRepository = true
const pass = false

// the bookmarkHandlerMock is the api-logic for bookmarks
func bookmarkHandlerMock(fail bool) http.Handler {
	return handlerWith(&handlerOps{
		version: "1.0",
		build:   "today",
		app: &bookmarks.Application{
			Logger:         logger,
			Store:          &MockRepository{fail: fail},
			FaviconPath:    "",
			DefaultFavicon: "",
		},
	})
}

// the bookmarkHandler uses the supplied repository
func bookmarkHandler(repo store.Repository) http.Handler {
	return handlerWith(&handlerOps{
		version: "1.0",
		build:   "today",
		app: &bookmarks.Application{
			Logger:         logger,
			Store:          repo,
			FaviconPath:    "",
			DefaultFavicon: "",
		},
	})
}

// A MockRepository which provides functions for the specific test-case
type MockRepository struct {
	// the default implementation is available below
	// the specific methods are implemented for the test-case
	mockRepository
	fail bool
}

func (r *MockRepository) GetBookmarkByID(id, username string) (store.Bookmark, error) {
	if r.fail {
		return store.Bookmark{}, errRaised
	}
	return store.Bookmark{
		DisplayName: "displayname",
		Highlight:   0,
		ChildCount:  0,
		Created:     time.Now().UTC(),
		ID:          "ID",
		Path:        "/",
		Type:        store.Node,
		URL:         "http://url",
		UserName:    username,
		Favicon:     "favicon.ico",
	}, nil
}

func Test_GetBookmarkByID(t *testing.T) {
	// arrange
	url := "/api/v1/bookmarks/ID"
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	var bm bookmarks.Bookmark
	assert.Equal(t, 200, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bm); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, bm.ID, "ID")
	assert.Equal(t, bm.DisplayName, "displayname")
	assert.Equal(t, bm.Path, "/")
	assert.Equal(t, bm.URL, "http://url")
	assert.Equal(t, bm.Type, bookmarks.Node)

	// fail -------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no id--------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/bookmarks/", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func (r *MockRepository) GetBookmarksByPath(path, username string) ([]store.Bookmark, error) {
	if r.fail {
		return make([]store.Bookmark, 0), errRaised
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		Highlight:   0,
		ChildCount:  0,
		Created:     time.Now().UTC(),
		ID:          "ID",
		Path:        "/",
		Type:        store.Node,
		URL:         "http://url",
		UserName:    username,
	}
	var bms []store.Bookmark
	return append(bms, bm), nil
}

func Test_GetBookmarkByPath(t *testing.T) {
	// arrange
	reqURL := "/api/v1/bookmarks/bypath"
	q := make(url.Values)
	q.Set("path", "/")
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	var bl api.BookmarkList
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 1, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail repository---------------------------------------------------
	// if there is an error an empty list is returned
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 0, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail no path------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetFolderByPath(path, username string) (store.Bookmark, error) {
	if r.fail {
		return store.Bookmark{}, errRaised
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		Highlight:   0,
		ChildCount:  0,
		Created:     time.Now().UTC(),
		ID:          "ID",
		Path:        "/",
		Type:        store.Folder,
		UserName:    username,
	}
	return bm, nil
}

func Test_GetBookmarkFolderBypath(t *testing.T) {
	// arrange
	reqURL := "/api/v1/bookmarks/folder"
	q := make(url.Values)
	q.Set("path", "/")

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	var br api.BookmarkResult

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// root folder ------------------------------------------------------
	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &br); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, true, br.Success)
	assert.Equal(t, bookmarks.Folder, br.Value.Type)

	// other folder -----------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	q = make(url.Values)
	q.Set("path", "/Folder")

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	assert.Equal(t, true, br.Success)
	assert.Equal(t, bookmarks.Folder, br.Value.Type)

	// fail repository---------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	q = make(url.Values)
	q.Set("path", "/Folder")

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no path------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetBookmarksByName(name, username string) ([]store.Bookmark, error) {
	if r.fail {
		return make([]store.Bookmark, 0), errRaised
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		Highlight:   0,
		ChildCount:  0,
		Created:     time.Now().UTC(),
		ID:          "ID",
		Path:        "/",
		Type:        store.Node,
		URL:         "http://url",
		UserName:    username,
	}
	var bms []store.Bookmark
	return append(bms, bm), nil
}

func Test_GetBookmarkByName(t *testing.T) {
	// arrange
	reqURL := "/api/v1/bookmarks/byname"
	q := make(url.Values)
	q.Set("name", "displayname")

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	var bl api.BookmarkList

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 1, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail repository---------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 0, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail no name -----------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func repository(t *testing.T) (store.Repository, *sql.DB) {
	var (
		DB  *gorm.DB
		err error
	)
	if DB, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{}); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&store.Bookmark{})
	db, err := DB.DB()
	if err != nil {
		t.Fatalf("cannot access database handle: %v", err)
	}
	logger := logging.NewNop()
	return store.Create(DB, logger), db
}

// we need InUnitOfWork to fail the mockRepository
func (r *MockRepository) InUnitOfWork(fn func(repo store.Repository) error) error {
	if r.fail {
		return errRaised
	}
	return nil
}

func Test_CreateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// testcases
	tt := []struct {
		name     string
		payload  string
		status   int
		response string
		function http.Handler
	}{
		{
			name: "create node",
			payload: `{
				"displayName": "Node",
				"path": "/",
				"type": "Node",
				"url": "http://url"
			}`,
			status:   http.StatusCreated,
			response: "",
			function: bookmarkHandler(repo),
		},
		{
			name: "create folder",
			payload: `{
				"displayName": "Folder",
				"path": "/",
				"type": "Folder",
				"customFavicon": "https://getbootstrap.com/docs/4.5/assets/img/favicons/favicon.ico"
			}`,
			status:   http.StatusCreated,
			response: "",
			function: bookmarkHandler(repo),
		},
		{
			name: "no path",
			payload: `{
				"displayName": "DisplayName",
				"type": "Node",
				"url": "http://url"
			}`,
			status:   http.StatusBadRequest,
			response: "",
			function: bookmarkHandler(repo),
		},
		{
			name:     "invalid",
			payload:  "{}",
			status:   http.StatusBadRequest,
			response: "",
			function: bookmarkHandler(repo),
		},
		{
			name: "repository error",
			payload: `{
				"displayName": "Node",
				"path": "/",
				"type": "Node",
				"url": "http://url"
			}`,
			status:   http.StatusInternalServerError,
			response: "",
			function: bookmarkHandlerMock(failRepository),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			reqURL := "/api/v1/bookmarks"
			var result api.Result
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", reqURL, strings.NewReader(tc.payload))
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
			req.Header.Add("Content-Type", "application/json")

			// act
			tc.function.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Errorf("could not unmarshal: %v", err)
			}
		})
	}
}

func Test_UpdateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	reqURL := "/api/v1/bookmarks"

	// basic logic validation
	// ---------------------------------------------------------------
	tt := []struct {
		name     string
		payload  string
		status   int
		function http.Handler
	}{
		{
			name: "no path",
			payload: `{
				"id": "id",
				"displayName": "DisplayName",
				"type": "Node",
				"url": "http://url"
			}`,
			status:   http.StatusBadRequest,
			function: bookmarkHandler(repo),
		},
		{
			name:     "invalid",
			payload:  "{}",
			status:   http.StatusBadRequest,
			function: bookmarkHandler(repo),
		},
		{
			name: "repository error",
			payload: `{
				"id": "id",
				"displayName": "Node",
				"path": "/",
				"type": "Node",
				"url": "http://url"
			}`,
			status:   http.StatusInternalServerError,
			function: bookmarkHandlerMock(failRepository),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			var result api.Result
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("PUT", reqURL, strings.NewReader(tc.payload))
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
			req.Header.Add("Content-Type", "application/json")

			// act
			tc.function.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Errorf("could not unmarshal: %v", err)
			}
		})
	}

	// ---- test Create+Update+Read ----

	payload := `{
		"displayName": "Node",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	// create one item
	// ---------------------------------------------------------------

	var result api.Result

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	req.Header.Add("Content-Type", "application/json")

	// act
	bookmarkHandler(repo).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusCreated, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id := result.Value

	// update the created item
	// ---------------------------------------------------------------
	payload = `{
		"id": "%s",
		"displayName": "Node_updated",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", reqURL, strings.NewReader(fmt.Sprintf(payload, id)))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	req.Header.Add("Content-Type", "application/json")

	// act
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// update the WRONG item
	// ---------------------------------------------------------------
	payload = `{
		"id": "MISSING_ID",
		"displayName": "Node_updated",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", reqURL, strings.NewReader(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
	req.Header.Add("Content-Type", "application/json")

	// act
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// fetch the updated item
	// ---------------------------------------------------------------

	// arrange
	reqURL = "/api/v1/bookmarks/" + id
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var bm bookmarks.Bookmark
	if err := json.Unmarshal(rec.Body.Bytes(), &bm); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, "Node_updated", bm.DisplayName)
}

// --------------------------------------------------------------------------
// mock repository implementation
// --------------------------------------------------------------------------

// create a mock implementation of the store repository
// tests can use the mock to implement their desired behavior
type mockRepository struct{}

var _ store.Repository = (*mockRepository)(nil)

func (m *mockRepository) CheckStoreConnectivity(timeOut uint) (err error) {
	return nil
}

func (m *mockRepository) InUnitOfWork(fn func(repo store.Repository) error) error {
	return nil
}

func (m *mockRepository) Create(item store.Bookmark) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockRepository) Update(item store.Bookmark) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockRepository) Delete(item store.Bookmark) error {
	return nil
}

func (m *mockRepository) DeletePath(path, username string) error {
	return nil
}

func (m *mockRepository) GetAllBookmarks(username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockRepository) GetBookmarksByPath(path, username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockRepository) GetBookmarksByPathStart(path, username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockRepository) GetBookmarksByName(name, username string) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockRepository) GetMostRecentBookmarks(username string, limit int) ([]store.Bookmark, error) {
	return nil, nil
}

func (m *mockRepository) GetPathChildCount(path, username string) ([]store.NodeCount, error) {
	return nil, nil
}

func (m *mockRepository) GetBookmarkByID(id, username string) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockRepository) GetFolderByPath(path, username string) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockRepository) GetAllPaths(username string) ([]string, error) {
	return nil, nil
}
