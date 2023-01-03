package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/bookmarks-new/api"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
)

var errRaised = fmt.Errorf("error")

const fail = true
const pass = false

// the bookmarkHandler is the api-logic for bookmarks
func bookmarkHandler(fail bool) http.Handler {
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

func TestGetBookmarkByID(t *testing.T) {
	// arrange
	url := "/api/v1/bookmarks/ID"
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandler(pass).ServeHTTP(rec, req)

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
	bookmarkHandler(fail).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no id--------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/bookmarks/", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandler(pass).ServeHTTP(rec, req)

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

func TestGetBookmarkByPath(t *testing.T) {
	// arrange
	reqURL := "/api/v1/bookmarks/bypath"
	q := make(url.Values)
	q.Set("path", "/")
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	bookmarkHandler(pass).ServeHTTP(rec, req)

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
	bookmarkHandler(fail).ServeHTTP(rec, req)

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
	bookmarkHandler(fail).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
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
