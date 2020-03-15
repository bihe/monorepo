package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bihe/bookmarks/internal/store"
	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	_ "github.com/jinzhu/gorm/dialects/sqlite" // use sqlite for testing
)

// --------------------------------------------------------------------------
// test helpers
// --------------------------------------------------------------------------

const userName = "username"

var raisedError = fmt.Errorf("error")

// common components necessary for handlers
var baseHandler = handler.Handler{
	ErrRep: &errors.ErrorReporter{
		CookieSettings: cookies.Settings{
			Path:   "/",
			Domain: "localhost",
		},
		ErrorPath: "error",
	},
}

func jwtUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), security.UserKey, &security.User{
			Username:    userName,
			Email:       "a.b@c.de",
			DisplayName: "displayname",
			Roles:       []string{"role"},
			UserID:      "12345",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --------------------------------------------------------------------------
// test-code
// --------------------------------------------------------------------------

type MockRepository struct {
	mockRepository
	fail bool
}

func (r *MockRepository) GetBookmarkById(id, username string) (store.Bookmark, error) {
	if r.fail {
		return store.Bookmark{}, raisedError
	}
	return store.Bookmark{
		DisplayName: "displayname",
		AccessCount: 1,
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

func TestGetBookmarkById(t *testing.T) {
	// arrange
	r := chi.NewRouter()
	r.Use(jwtUser)
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{},
	}
	url := "/api/v1/bookmarks/{id}"
	function := bookmarkAPI.Secure(bookmarkAPI.GetBookmarkByID)
	rec := httptest.NewRecorder()
	var bm Bookmark

	// act
	r.Get(url, function)
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bm); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, bm.ID, "ID")
	assert.Equal(t, bm.DisplayName, "displayname")
	assert.Equal(t, bm.Path, "/")
	assert.Equal(t, bm.URL, "http://url")
	assert.Equal(t, bm.Type, Node)

	// fail -------------------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{fail: true}
	rec = httptest.NewRecorder()

	r.Get(url, function)
	req, _ = http.NewRequest("GET", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no id--------------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{}
	rec = httptest.NewRecorder()

	r.Get("/api/v1/bookmarks/", function)
	req, _ = http.NewRequest("GET", "/api/v1/bookmarks/", nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetBookmarksByPath(path, username string) ([]store.Bookmark, error) {
	if r.fail {
		return make([]store.Bookmark, 0), raisedError
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		AccessCount: 1,
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
	r := chi.NewRouter()
	r.Use(jwtUser)
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{},
	}
	reqUrl := "/api/v1/bookmarks/bypath"
	function := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksByPath)
	rec := httptest.NewRecorder()
	var bl BookmarkList

	q := make(url.Values)
	q.Set("path", "/")

	// act
	r.Get(reqUrl, function)
	req, _ := http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 1, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail repository---------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{fail: true}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 0, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail no path------------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetFolderByPath(path, username string) (store.Bookmark, error) {
	if r.fail {
		return store.Bookmark{}, raisedError
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		AccessCount: 1,
		ChildCount:  0,
		Created:     time.Now().UTC(),
		ID:          "ID",
		Path:        "/",
		Type:        store.Folder,
		UserName:    username,
	}
	return bm, nil
}

func TestGetBookmarkFolderBypath(t *testing.T) {
	// arrange
	r := chi.NewRouter()
	r.Use(jwtUser)
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{},
	}
	reqUrl := "/api/v1/bookmarks/folder"
	function := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksFolderByPath)
	rec := httptest.NewRecorder()
	var br BookmarkResult

	// root folder ------------------------------------------------------
	q := make(url.Values)
	q.Set("path", "/")

	// act
	r.Get(reqUrl, function)
	req, _ := http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &br); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, true, br.Success)
	assert.Equal(t, Folder, br.Value.Type)

	// other folder -----------------------------------------------------
	q = make(url.Values)
	q.Set("path", "/Folder")

	bookmarkAPI.Repository = &MockRepository{}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, true, br.Success)
	assert.Equal(t, Folder, br.Value.Type)

	// fail repository---------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{fail: true}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// fail no path------------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetBookmarksByName(name, username string) ([]store.Bookmark, error) {
	if r.fail {
		return make([]store.Bookmark, 0), raisedError
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		AccessCount: 1,
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

func TestGetBookmarkByName(t *testing.T) {
	// arrange
	r := chi.NewRouter()
	r.Use(jwtUser)
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{},
	}
	reqUrl := "/api/v1/bookmarks/byname"
	function := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksByName)
	rec := httptest.NewRecorder()
	var bl BookmarkList

	q := make(url.Values)
	q.Set("name", "displayname")

	// act
	r.Get(reqUrl, function)
	req, _ := http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 1, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail repository---------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{fail: true}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl+"?"+q.Encode(), nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 0, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail no name -----------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{}
	rec = httptest.NewRecorder()

	r.Get(reqUrl, function)
	req, _ = http.NewRequest("GET", reqUrl, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (r *MockRepository) GetMostRecentBookmarks(username string, limit int) ([]store.Bookmark, error) {
	if r.fail {
		return make([]store.Bookmark, 0), raisedError
	}

	bm := store.Bookmark{
		DisplayName: "displayname",
		AccessCount: 1,
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

func TestGetMostVisited(t *testing.T) {
	// arrange
	r := chi.NewRouter()
	r.Use(jwtUser)
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{},
	}
	defUrl := "/api/v1/bookmarks/mostvisited/{num}"
	reqUrl := "/api/v1/bookmarks/mostvisited/1"
	function := bookmarkAPI.Secure(bookmarkAPI.GetMostVisited)
	rec := httptest.NewRecorder()
	var bl BookmarkList

	// act
	r.Get(defUrl, function)
	req, _ := http.NewRequest("GET", reqUrl, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 1, bl.Count)
	assert.Equal(t, true, bl.Success)

	// fail repository---------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{fail: true}
	rec = httptest.NewRecorder()

	r.Get(defUrl, function)
	req, _ = http.NewRequest("GET", reqUrl, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 0, bl.Count)
	assert.Equal(t, true, bl.Success)

	// default num ------------------------------------------------------
	bookmarkAPI.Repository = &MockRepository{fail: true}
	rec = httptest.NewRecorder()

	r.Get(defUrl, function)
	req, _ = http.NewRequest("GET", "/api/v1/bookmarks/mostvisited/0", nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &bl); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 0, bl.Count)
	assert.Equal(t, true, bl.Success)
}

func repository(t *testing.T) (store.Repository, *gorm.DB) {
	var (
		DB  *gorm.DB
		err error
	)
	if DB, err = gorm.Open("sqlite3", ":memory:"); err != nil {
		t.Fatalf("cannot create database connection: %v", err)
	}
	// Migrate the schema
	DB.AutoMigrate(&store.Bookmark{})

	DB.LogMode(true)
	return store.Create(DB), DB
}

func (r *MockRepository) InUnitOfWork(fn func(repo store.Repository) error) error {
	if r.fail {
		return raisedError
	}
	return nil
}

func TestCreateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// arrange
	r := chi.NewRouter()
	r.Use(jwtUser)
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}
	defUrl := "/api/v1/bookmarks"
	function := bookmarkAPI.Secure(bookmarkAPI.Create)

	mockAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{fail: true},
	}

	// testcases
	tt := []struct {
		name     string
		payload  string
		status   int
		response string
		function http.HandlerFunc
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
			function: function,
		},
		{
			name: "create folder",
			payload: `{
				"displayName": "Folder",
				"path": "/",
				"type": "Folder"
			}`,
			status:   http.StatusCreated,
			response: "",
			function: function,
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
			function: function,
		},
		{
			name:     "invalid",
			payload:  "{}",
			status:   http.StatusBadRequest,
			response: "",
			function: function,
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
			function: mockAPI.Secure(mockAPI.Create),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var result ResultResponse

			// act
			r.Post(defUrl, tc.function)
			req, _ := http.NewRequest("POST", defUrl, strings.NewReader(tc.payload))
			req.Header.Add("Content-Type", "application/json")
			r.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Errorf("could not unmarshal: %v", err)
			}
		})
	}
}

func TestUpdateBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}
	url := "/api/v1/bookmarks"

	mockAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{fail: true},
	}

	// validation
	// ---------------------------------------------------------------
	tt := []struct {
		name     string
		payload  string
		status   int
		function http.HandlerFunc
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
			function: mockAPI.Secure(mockAPI.Update),
		},
		{
			name:     "invalid",
			payload:  "{}",
			status:   http.StatusBadRequest,
			function: mockAPI.Secure(mockAPI.Update),
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
			function: mockAPI.Secure(mockAPI.Update),
		},
	}
	for _, tc := range tt {
		r := chi.NewRouter()
		r.Use(jwtUser)
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var result ResultResponse

			// act
			r.Put(url, tc.function)
			req, _ := http.NewRequest("PUT", url, strings.NewReader(tc.payload))
			req.Header.Add("Content-Type", "application/json")
			r.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Errorf("could not unmarshal: %v", err)
			}
		})
	}

	create := bookmarkAPI.Secure(bookmarkAPI.Create)
	update := bookmarkAPI.Secure(bookmarkAPI.Update)
	get := bookmarkAPI.Secure(bookmarkAPI.GetBookmarkByID)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post(url, create)
	r.Put(url, update)
	r.Get(url+"/{id}", get)

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
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var result ResultResponse
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
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url, strings.NewReader(fmt.Sprintf(payload, id)))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

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
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// fetch the updated item
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/"+id, nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmark BookmarkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmark); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, "Node_updated", bookmark.DisplayName)
}

func TestUpdateBookmarkHierarchy(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}
	url := "/api/v1/bookmarks"

	// mockAPI := &BookmarksAPI{
	// 	Handler:    baseHandler,
	// 	Repository: &MockRepository{fail: true},
	// }

	create := bookmarkAPI.Secure(bookmarkAPI.Create)
	update := bookmarkAPI.Secure(bookmarkAPI.Update)
	getBookmark := bookmarkAPI.Secure(bookmarkAPI.GetBookmarkByID)
	getFolder := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksFolderByPath)
	getByPath := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksByPath)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post(url, create)
	r.Put(url, update)
	r.Get(url+"/{id}", getBookmark)
	r.Get(url+"/folder", getFolder)
	r.Get(url+"/bypath", getByPath)

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
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder1
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder1", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder2
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder2", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// get the folder /Folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path="+"/Folder", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmarkFolder BookmarResultResponse
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
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var response ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, id, response.Value)

	// check that the path of the child-folders has changed as well
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/bypath?path="+"/Folder_updated", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmarkList BookmarkListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmarkList); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 2, len(bookmarkList.Value))
	assert.Equal(t, "/Folder_updated", bookmarkList.Value[0].Path)
}

func TestDeleteBookmark(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}
	url := "/api/v1/bookmarks"

	// validation
	// ---------------------------------------------------------------
	mockAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{fail: true},
	}

	tt := []struct {
		name     string
		status   int
		id       string
		function http.HandlerFunc
	}{
		{
			name:     "no id",
			status:   http.StatusNotFound,
			function: mockAPI.Secure(mockAPI.Delete),
		},
		{
			name:     "repository error",
			id:       "id",
			status:   http.StatusInternalServerError,
			function: mockAPI.Secure(mockAPI.Delete),
		},
	}
	for _, tc := range tt {
		r := chi.NewRouter()
		r.Use(jwtUser)
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			// act
			r.Delete(url+"/{id}", tc.function)
			req, _ := http.NewRequest("DELETE", url+"/"+tc.id, nil)
			r.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
		})
	}

	create := bookmarkAPI.Secure(bookmarkAPI.Create)
	delete := bookmarkAPI.Secure(bookmarkAPI.Delete)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post(url, create)
	r.Delete(url+"/{id}", delete)

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
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var result ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id := result.Value

	// delete the created item
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url+"/"+id, nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDeleteBookmarkHierarchy(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}
	url := "/api/v1/bookmarks"

	create := bookmarkAPI.Secure(bookmarkAPI.Create)
	delete := bookmarkAPI.Secure(bookmarkAPI.Delete)
	getBookmark := bookmarkAPI.Secure(bookmarkAPI.GetBookmarkByID)
	getFolder := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksFolderByPath)
	getByPath := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksByPath)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post(url, create)
	r.Delete(url+"/{id}", delete)
	r.Get(url+"/{id}", getBookmark)
	r.Get(url+"/folder", getFolder)
	r.Get(url+"/bypath", getByPath)

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
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder1
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "Folder1", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// get the folder /Folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/folder?path="+"/Folder", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var bookmarkFolder BookmarResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &bookmarkFolder); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	id := bookmarkFolder.Value.ID

	// delete the created item
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", url+"/"+id, nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateSortOrder(t *testing.T) {

	repo, db := repository(t)
	defer db.Close()

	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}
	url := "/api/v1/bookmarks"

	mockAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{fail: true},
	}

	// validation
	// ---------------------------------------------------------------
	tt := []struct {
		name     string
		payload  string
		status   int
		function http.HandlerFunc
	}{
		{
			name:     "invalid",
			payload:  "{}",
			status:   http.StatusBadRequest,
			function: mockAPI.Secure(mockAPI.UpdateSortOrder),
		},
		{
			name: "unbalanced",
			payload: `{
				"ids": ["1","2"],
				"sortOrder": [1,2,3]
			}`,
			status:   http.StatusBadRequest,
			function: mockAPI.Secure(mockAPI.UpdateSortOrder),
		},
	}
	for _, tc := range tt {
		r := chi.NewRouter()
		r.Use(jwtUser)
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var result ResultResponse

			// act
			r.Put(url+"/sortorder", tc.function)
			req, _ := http.NewRequest("PUT", url+"/sortorder", strings.NewReader(tc.payload))
			req.Header.Add("Content-Type", "application/json")
			r.ServeHTTP(rec, req)

			// assert
			assert.Equal(t, tc.status, rec.Code)
			if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
				t.Errorf("could not unmarshal: %v", err)
			}
		})
	}

	create := bookmarkAPI.Secure(bookmarkAPI.Create)
	updateSortOrder := bookmarkAPI.Secure(bookmarkAPI.UpdateSortOrder)
	get := bookmarkAPI.Secure(bookmarkAPI.GetBookmarkByID)
	getByPath := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksByPath)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post(url, create)
	r.Put(url+"/sortorder", updateSortOrder)
	r.Get(url+"/{id}", get)
	r.Get(url+"/bypath", getByPath)

	payload := `{
		"displayName": "%s",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	var result ResultResponse
	// create A
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(payload, "A")))
	req.Header.Add("Content-Type", "application/json")
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
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var list BookmarkListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 2, len(list.Value))
	assert.Equal(t, "A", list.Value[0].DisplayName)
	assert.Equal(t, "B", list.Value[1].DisplayName)

	// change the sorting
	// ---------------------------------------------------------------
	newOrder := BookmarksSortOrder{
		IDs:       []string{id1, id2},
		SortOrder: []int{10, 1},
	}
	bytesPayload, _ := json.Marshal(newOrder)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url+"/sortorder", bytes.NewReader(bytesPayload))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	// get the bookmarks again
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", url+"/bypath?path=/", nil)
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
	newOrder = BookmarksSortOrder{
		IDs:       []string{"A"},
		SortOrder: []int{1},
	}
	bytesPayload, _ = json.Marshal(newOrder)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", url+"/sortorder", bytes.NewReader(bytesPayload))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetFavicon(t *testing.T) {
	mockAPI := &BookmarksAPI{
		Handler:        baseHandler,
		Repository:     &MockRepository{fail: false},
		BasePath:       "../../../",
		FaviconPath:    "./assets",
		DefaultFavicon: "./assets/favicon.ico",
	}
	getFavicon := mockAPI.Secure(mockAPI.GetFavicon)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Get("/", getFavicon)
	r.Get("/{id}", getFavicon)

	// get the favicon
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/id", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, len(rec.Body.Bytes()) > 0)

	// no ID
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// fail lookup
	// ---------------------------------------------------------------
	mockAPI = &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{fail: true},
	}
	getFavicon = mockAPI.Secure(mockAPI.GetFavicon)
	r = chi.NewRouter()
	r.Use(jwtUser)
	r.Get("/{id}", getFavicon)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/id", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// wrong favicon
	// ---------------------------------------------------------------
	mockAPI = &BookmarksAPI{
		Handler:        baseHandler,
		Repository:     &MockRepository{fail: false},
		BasePath:       "../../../",
		FaviconPath:    "./_a/",
		DefaultFavicon: "./assets/favicon.ico",
	}
	getFavicon = mockAPI.Secure(mockAPI.GetFavicon)
	r = chi.NewRouter()
	r.Use(jwtUser)
	r.Get("/{id}", getFavicon)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/id", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchAndForward(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	bookmarkAPI := &BookmarksAPI{
		Handler:        baseHandler,
		Repository:     repo,
		BasePath:       "../../../",
		FaviconPath:    "./assets",
		DefaultFavicon: "./assets/favicon.ico",
	}
	getFavicon := bookmarkAPI.Secure(bookmarkAPI.FetchAndForward)
	create := bookmarkAPI.Secure(bookmarkAPI.Create)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post("/", create)
	r.Get("/", getFavicon)
	r.Get("/{id}", getFavicon)

	payload := `{
		"displayName": "Node",
		"path": "/",
		"type": "Node",
		"url": "http://url"
	}`

	// create one item
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var result ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	id := result.Value

	// fetch the URL
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/"+id, nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "http://url", rec.Header().Get("location"))

	// no ID
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// create a folder
	// ---------------------------------------------------------------
	payload = `{
		"displayName": "Node",
		"path": "/",
		"type": "Folder"
	}`

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	id = result.Value

	// error folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/"+id, nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// error wrong ID
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/any", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func (r *MockRepository) GetAllPaths(username string) ([]string, error) {
	if r.fail {
		return nil, raisedError
	}

	return []string{"/", "/a", "/b"}, nil
}

func TestBookmarkAllPaths(t *testing.T) {
	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{},
	}
	function := bookmarkAPI.Secure(bookmarkAPI.GetAllPaths)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Get("/", function)

	// get the paths
	// ---------------------------------------------------------------
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var result BookmarksPathsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}

	assert.Equal(t, 3, result.Count)
	assert.Equal(t, "/", result.Paths[0])

	// fail
	// ---------------------------------------------------------------
	bookmarkAPI = &BookmarksAPI{
		Handler:    baseHandler,
		Repository: &MockRepository{fail: true},
	}
	function = bookmarkAPI.Secure(bookmarkAPI.GetAllPaths)

	r = chi.NewRouter()
	r.Use(jwtUser)
	r.Get("/", function)

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMoveBookmarks(t *testing.T) {
	repo, db := repository(t)
	defer db.Close()

	// arrange
	bookmarkAPI := &BookmarksAPI{
		Handler:    baseHandler,
		Repository: repo,
	}

	create := bookmarkAPI.Secure(bookmarkAPI.Create)
	update := bookmarkAPI.Secure(bookmarkAPI.Update)
	getBookmark := bookmarkAPI.Secure(bookmarkAPI.GetBookmarkByID)
	getFolder := bookmarkAPI.Secure(bookmarkAPI.GetBookmarksFolderByPath)

	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Post("/", create)
	r.Put("/", update)
	r.Get("/{id}", getBookmark)
	r.Get("/folder", getFolder)

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
	req, _ := http.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(payload, "Folder", "/")))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder1
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(payload, "Folder1", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// /Folder/Folder2
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/", strings.NewReader(fmt.Sprintf(payload, "Folder2", "/Folder")))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// retrieve the /Folder
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/folder?path=/Folder", nil)
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result BookmarResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, "/", result.Value.Path)
	assert.Equal(t, "Folder", result.Value.DisplayName)
	assert.Equal(t, 2, result.Value.ChildCount)

	// retrieve the /Folder/Folder1
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/folder?path=/Folder/Folder1", nil)
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
	req, _ = http.NewRequest("PUT", "/", bytes.NewReader(payloadBytes))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var bm ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &bm); err != nil {
		t.Errorf("could not unmarshal body: %v", err)
	}
	assert.Equal(t, folder.ID, bm.Value)

	// retrieve the /Folder again
	// ---------------------------------------------------------------
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/folder?path=/Folder", nil)
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
	req, _ = http.NewRequest("PUT", "/", bytes.NewReader(payloadBytes))
	req.Header.Add("Content-Type", "application/json")
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

func (m *mockRepository) GetBookmarkById(id, username string) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockRepository) GetFolderByPath(path, username string) (store.Bookmark, error) {
	return store.Bookmark{}, nil
}

func (m *mockRepository) GetAllPaths(username string) ([]string, error) {
	return nil, nil
}
