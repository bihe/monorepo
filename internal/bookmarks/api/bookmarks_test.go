package api_test

import (
	"bytes"
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
	"golang.binggl.net/monorepo/internal/bookmarks/api"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/store"
	"golang.binggl.net/monorepo/pkg/logging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var errRaised = fmt.Errorf("error")

const failRepository = true
const pass = false

const TestFaviconPath = "./assets"

// the bookmarkHandlerMock is the api-logic for bookmarks
func bookmarkHandlerMock(fail bool) http.Handler {
	return handlerWith(&handlerOps{
		version: "1.0",
		build:   "today",
		app: &bookmarks.Application{
			Logger:      logger,
			Store:       &MockRepository{fail: fail},
			FaviconPath: "",
		},
	})
}

// the bookmarkHandler uses the supplied repository
func bookmarkHandler(repo store.Repository) http.Handler {
	return handlerWith(&handlerOps{
		version: "1.0",
		build:   "today",
		app: &bookmarks.Application{
			Logger:      logger,
			Store:       repo,
			FaviconPath: "",
		},
	})
}

func addJwtAuth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
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
	addJwtAuth(req)

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
	addJwtAuth(req)

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no id--------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/bookmarks/", nil)
	addJwtAuth(req)

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
	addJwtAuth(req)

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
	addJwtAuth(req)

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
	addJwtAuth(req)

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
	addJwtAuth(req)
	var br api.BookmarkFolderResult

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
	addJwtAuth(req)
	q = make(url.Values)
	q.Set("path", "/Folder")

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	assert.Equal(t, true, br.Success)
	assert.Equal(t, bookmarks.Folder, br.Value.Type)

	// fail repository---------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL+"?"+q.Encode(), nil)
	addJwtAuth(req)
	q = make(url.Values)
	q.Set("path", "/Folder")

	// act
	bookmarkHandlerMock(failRepository).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no path------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL, nil)
	addJwtAuth(req)

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
	addJwtAuth(req)
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
	addJwtAuth(req)

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
	addJwtAuth(req)

	// act
	bookmarkHandlerMock(pass).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetAllPaths(username string) ([]string, error) {
	if r.fail {
		return nil, errRaised
	}

	return []string{"/", "/a", "/b"}, nil
}

func Test_BookmarkAllPaths(t *testing.T) {
	reqURL := "/api/v1/bookmarks/allpaths"

	// get the paths
	// ---------------------------------------------------------------
	r := bookmarkHandlerMock(pass)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", reqURL, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var result api.BookmarksPaths
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}

	assert.Equal(t, 3, result.Count)
	assert.Equal(t, "/", result.Paths[0])

	// fail
	// ---------------------------------------------------------------
	r = bookmarkHandlerMock(failRepository)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", reqURL, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
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
			addJwtAuth(req)
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
			addJwtAuth(req)
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
	addJwtAuth(req)
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
	addJwtAuth(req)
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
	addJwtAuth(req)
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
	addJwtAuth(req)

	// act
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var bm bookmarks.Bookmark
	if err := json.Unmarshal(rec.Body.Bytes(), &bm); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, "Node_updated", bm.DisplayName)
}

func Test_UpdateBookmarkHierarchy(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	url := "/api/v1/bookmarks"

	payload := `{
		"displayName": "%s",
		"path": "%s",
		"type": "Folder"
	}`

	// create folder hierarchy
	// /Folder
	//     /Folder1
	//     /folder2
	// ---------------------------------------------------------------

	// /Folder
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder", "/")))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/json")
	bookmarkHandler(repo).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder1
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder1", "/Folder")))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/json")
	bookmarkHandler(repo).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder2
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder2", "/Folder")))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/json")
	bookmarkHandler(repo).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// get the folder /Folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path="+"/Folder", nil)
	addJwtAuth(req)
	bookmarkHandler(repo).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmarkFolder api.BookmarkFolderResult
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmarkFolder); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	// update the folder
	// ---------------------------------------------------------------
	bm := bookmarkFolder.Value
	bm.DisplayName = bm.DisplayName + "_updated"
	id := bm.ID

	payloadBytes, err := json.Marshal(bm)
	if err != nil {
		t.Errorf("cannot marshall payload: %v", err)
	}

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url, bytes.NewReader(payloadBytes))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/json")
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response api.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, id, response.Value)

	// check that the path of the child-folders has changed as well
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/bypath?path="+"/Folder_updated", nil)
	addJwtAuth(req)
	bookmarkHandler(repo).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmarkList api.BookmarkList
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmarkList); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 2, len(bookmarkList.Value))
	assert.Equal(t, "/Folder_updated", bookmarkList.Value[0].Path)
}

func Test_DeleteBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	url := "/api/v1/bookmarks"

	// validation of standard logic
	// ---------------------------------------------------------------
	tt := []struct {
		name     string
		status   int
		id       string
		function http.Handler
	}{
		{
			name:     "no id",
			status:   http.StatusMethodNotAllowed,
			function: bookmarkHandler(repo),
		},
		{
			name:     "repository error",
			id:       "id",
			status:   http.StatusInternalServerError,
			function: bookmarkHandlerMock(failRepository),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", url+"/"+tc.id, nil)
			addJwtAuth(req)
			req.Header.Add("Content-Type", "application/json")

			tc.function.ServeHTTP(rec, req)
			// assert
			assert.Equal(t, tc.status, rec.Code)
		})
	}

	payload := `{
		"displayName": "Node",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	// create one item
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/json")
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var result api.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id := result.Value

	// delete the created item
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url+"/"+id, nil)
	addJwtAuth(req)
	bookmarkHandler(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func Test_DeleteBookmarkHierarchy(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	url := "/api/v1/bookmarks"
	r := bookmarkHandler(repo)

	// create folder hierarchy
	// /Folder
	//     /Folder1
	// ---------------------------------------------------------------

	payload := `{
		"displayName": "%s",
		"path": "%s",
		"type": "Folder"
	}`

	// /Folder
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder", "/")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder1
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder1", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// get the folder /Folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path="+"/Folder", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmarkFolder api.BookmarkFolderResult
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmarkFolder); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id := bookmarkFolder.Value.ID

	// delete the created item
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url+"/"+id, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func Test_UpdateSortOrder(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	url := "/api/v1/bookmarks"

	// validation of basic logic
	// ---------------------------------------------------------------
	tt := []struct {
		name     string
		payload  string
		status   int
		function http.Handler
	}{
		{
			name:     "invalid",
			payload:  "{}",
			status:   http.StatusBadRequest,
			function: bookmarkHandlerMock(pass),
		},
		{
			name: "unbalanced",
			payload: `{
				"ids": ["1","2"],
				"sortOrder": [1,2,3]
			}`,
			status:   http.StatusBadRequest,
			function: bookmarkHandlerMock(pass),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var result api.Result

			// act
			req, _ := http.NewRequest("PUT", url+"/sortorder", strings.NewReader(tc.payload))
			req.Header.Add("Content-Type", "application/json")
			addJwtAuth(req)
			tc.function.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Errorf("could not unmarshal: %v", err)
			}
		})
	}

	payload := `{
		"displayName": "%s",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	var result api.Result
	r := bookmarkHandler(repo)

	// create A
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "A")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id1 := result.Value

	// create B
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "B")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id2 := result.Value

	// get the bookmarks
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/bypath?path=/", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var list api.BookmarkList
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 2, len(list.Value))
	assert.Equal(t, "A", list.Value[0].DisplayName)
	assert.Equal(t, "B", list.Value[1].DisplayName)

	// change the sorting
	// ---------------------------------------------------------------
	newOrder := api.BookmarksSortOrderRequest{
		BookmarksSortOrder: &bookmarks.BookmarksSortOrder{
			IDs:       []string{id1, id2},
			SortOrder: []int{10, 1},
		},
	}
	bytesPayload, _ := json.Marshal(newOrder)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url+"/sortorder", bytes.NewReader(bytesPayload))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	// get the bookmarks again
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/bypath?path=/", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 2, len(list.Value))
	assert.Equal(t, "B", list.Value[0].DisplayName)
	assert.Equal(t, "A", list.Value[1].DisplayName)

	// change the sorting, but with wrong ID
	// ---------------------------------------------------------------
	newOrder = api.BookmarksSortOrderRequest{
		BookmarksSortOrder: &bookmarks.BookmarksSortOrder{
			IDs:       []string{"A"},
			SortOrder: []int{1},
		},
	}
	bytesPayload, _ = json.Marshal(newOrder)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url+"/sortorder", bytes.NewReader(bytesPayload))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func Test_GetFavicon(t *testing.T) {
	url := "/api/v1/bookmarks/favicon/"
	r := bookmarkHandlerMock(pass)

	// get the default favicon
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url+"ID", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, len(rec.Body.Bytes()) > 0)

	// no ID
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail lookup
	// ---------------------------------------------------------------
	r = bookmarkHandlerMock(failRepository)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"ID", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func Test_FetchAndForward(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	r := bookmarkHandler(repo)
	url := "/api/v1/bookmarks"

	payload := `{
		"displayName": "Node",
		"path": "/",
		"type": "Node",
		"url": "http://url",
		"highlight": 1
	}`

	// create one item
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var result api.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	id := result.Value

	// retrieve the created item
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/"+id, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	var bookmark bookmarks.Bookmark
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmark); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, 1, bookmark.Highlight)

	// fetch the URL
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/fetch/"+id, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "http://url", rec.Header().Get("location"))

	// retrieve the created item again - the highlight should be 0
	// fetch and forward resets the highlight value
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/"+id, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmark); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, 0, bookmark.Highlight)

	// no ID
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/fetch", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// create a folder
	// ---------------------------------------------------------------
	payload = `{
		"displayName": "Node",
		"path": "/",
		"type": "Folder"
	}`

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	id = result.Value

	// error folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/fetch/"+id, nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// error wrong ID
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/fetch/any", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func Test_MoveBookmarks(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()
	r := bookmarkHandler(repo)
	url := "/api/v1/bookmarks"

	// create folder hierarchy
	// /Folder
	//     /Folder1
	// ---------------------------------------------------------------
	payload := `{
		"displayName": "%s",
		"path": "%s",
		"type": "Folder"
	}`

	// /Folder
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder", "/")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder1
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder1", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder2
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder2", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// retrieve the /Folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path=/Folder", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result api.BookmarkFolderResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, "/", result.Value.Path)
	assert.Equal(t, "Folder", result.Value.DisplayName)
	assert.Equal(t, 2, result.Value.ChildCount)

	// retrieve the /Folder/Folder1
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path=/Folder/Folder1", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, "/Folder", result.Value.Path)
	assert.Equal(t, "Folder1", result.Value.DisplayName)

	// change the path of the folder
	// ---------------------------------------------------------------
	folder := result.Value
	folder.Path = "/"
	payloadBytes, err := json.Marshal(folder)
	if err != nil {
		t.Errorf("cannot marshall folder: %v", err)
	}

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url+"/", bytes.NewReader(payloadBytes))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var bm api.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &bm); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, folder.ID, bm.Value)

	// retrieve the /Folder again
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path=/Folder", nil)
	addJwtAuth(req)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, "/", result.Value.Path)
	assert.Equal(t, "Folder", result.Value.DisplayName)
	assert.Equal(t, 1, result.Value.ChildCount)

	// change the path of the folder to itself
	// ---------------------------------------------------------------
	folder = result.Value
	folder.Path = "/Folder"
	payloadBytes, err = json.Marshal(folder)
	if err != nil {
		t.Errorf("cannot marshall folder: %v", err)
	}

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url+"/", bytes.NewReader(payloadBytes))
	req.Header.Add("Content-Type", "application/json")
	addJwtAuth(req)
	r.ServeHTTP(rec, req)

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
