package mydms_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/persistence"
	"golang.binggl.net/monorepo/pkg/security"

	pkgerr "golang.binggl.net/monorepo/pkg/errors"

	_ "github.com/mattn/go-sqlite3" // use sqlite for testing
)

// --------------------------------------------------------------------------
// mocks for transport tests
// --------------------------------------------------------------------------

// FileService --------------------------------------------------------------

// rather small PDF payload
// https://stackoverflow.com/questions/17279712/what-is-the-smallest-possible-valid-pdf
const pdfPayload = `%PDF-1.0
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj 2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj 3 0 obj<</Type/Page/MediaBox[0 0 3 3]>>endobj
xref
0 4
0000000000 65535 f
0000000010 00000 n
0000000053 00000 n
0000000102 00000 n
trailer<</Size 4/Root 1 0 R>>
startxref
149
%EOF
`

type mockFileService struct {
	fail bool
}

func (m *mockFileService) InitClient() (err error) {
	return nil
}

func (m *mockFileService) SaveFile(file filestore.FileItem) (err error) {
	return nil
}

func (m *mockFileService) GetFile(filePath string) (item filestore.FileItem, err error) {
	if m.fail == true {
		return filestore.FileItem{}, fmt.Errorf("error")
	}
	return filestore.FileItem{
		MimeType: "application/pdf",
		Payload:  []byte(pdfPayload),
	}, nil
}

func (m *mockFileService) DeleteFile(filePath string) (err error) {
	return nil
}

var _ filestore.FileService = &mockFileService{}

// UploadClient -------------------------------------------------------------

type mockUploadClient struct {
	fail bool
}

func (m *mockUploadClient) Get(id, authToken string) (upload.Upload, error) {
	if m.fail {
		return upload.Upload{}, fmt.Errorf("error")
	}
	return upload.Upload{}, nil
}

func (m *mockUploadClient) Delete(id, authToken string) error {
	if m.fail {
		return fmt.Errorf("error")
	}
	return nil
}

var _ upload.Client = &mockUploadClient{}

// AppInfoService -----------------------------------------------------------

type mockAppInfoService struct {
	fail bool
}

func (m *mockAppInfoService) GetAppInfo(user *security.User) (ai appinfo.AppInfo, err error) {
	if m.fail {
		return appinfo.AppInfo{}, fmt.Errorf("error")
	}
	return appinfo.AppInfo{
		UserInfo: appinfo.UserInfo{
			DisplayName: user.DisplayName,
		},
		VersionInfo: appinfo.VersionInfo{
			Version: "1.0",
		},
	}, nil
}

var _ appinfo.Service = &mockAppInfoService{}

// DocumentService ----------------------------------------------------------

type mockDocumentService struct {
	fail bool
}

func (s *mockDocumentService) GetDocumentByID(id string) (d document.Document, err error) {
	if s.fail {
		return document.Document{}, fmt.Errorf("error")
	}
	return document.Document{
		ID: "id",
	}, nil
}

func (s *mockDocumentService) DeleteDocumentByID(id string) (err error) {
	if s.fail {
		return fmt.Errorf("error")
	}
	return nil
}

func (s *mockDocumentService) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p document.PagedDocument, err error) {
	if s.fail {
		return document.PagedDocument{}, fmt.Errorf("error")
	}
	return document.PagedDocument{
		Documents: []document.Document{
			{
				ID: "id",
			},
		},
		TotalEntries: 1,
	}, nil
}

func (s *mockDocumentService) SearchList(name string, st document.SearchType) (l []string, err error) {
	if s.fail {
		return nil, fmt.Errorf("error")
	}
	return []string{"one", "two"}, nil
}

func (s *mockDocumentService) SaveDocument(doc document.Document, user security.User) (d document.Document, err error) {
	if s.fail {
		return document.Document{}, fmt.Errorf("error")
	}
	return doc, nil
}

var _ document.Service = &mockDocumentService{}

// --------------------------------------------------------------------------
// Configuration and common test elements
// --------------------------------------------------------------------------

const jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MDIxNDIwNjYsImp0aSI6IjhjMmRlNzIzLTUwOTItNDQxOC05ZTcyLTZkNDlkYjZiMGI1ZSIsImlhdCI6MTYwMTUzNzI2NiwiaXNzIjoidGVzdCIsInN1YiI6ImEuYkBjLmRlIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJEaXNwbGF5TmFtZSI6IlRlc3QgVXNlciIsIkVtYWlsIjoiYS5iQGMuZGUiLCJVc2VySWQiOiIxMjM0NSIsIlVzZXJOYW1lIjoiVXNlcm5hbWUiLCJHaXZlbk5hbWUiOiJUZXN0IiwiU3VybmFtZSI6IlVzZXIiLCJDbGFpbXMiOlsidGVzdHxodHRwOi8vbG9jYWxob3N0L3Rlc3R8cm9sZUEiXX0.nVynmKxh8RN1iNuwNmDd47pHrH25nVRcssC80-PFXLs"
const jwt_different_role = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MDIxNDIwNjYsImp0aSI6IjhjMmRlNzIzLTUwOTItNDQxOC05ZTcyLTZkNDlkYjZiMGI1ZSIsImlhdCI6MTYwMTUzNzI2NiwiaXNzIjoidGVzdCIsInN1YiI6ImEuYkBjLmRlIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJEaXNwbGF5TmFtZSI6IlRlc3QgVXNlciIsIkVtYWlsIjoiYS5iQGMuZGUiLCJVc2VySWQiOiIxMjM0NSIsIlVzZXJOYW1lIjoiVXNlcm5hbWUiLCJHaXZlbk5hbWUiOiJUZXN0IiwiU3VybmFtZSI6IlVzZXIiLCJDbGFpbXMiOlsidGVzdHxodHRwOi8vbG9jYWxob3N0L3Rlc3R8ZGlmZmVyZW50Um9sZSJdfQ.RRg7Bdk0_PlCDwUhzDa__2HMXQ1WlfJkAi6VNaiwVTQ"

var logger = logging.NewNop()

var jwtConfig = config.Security{
	CacheDuration: "10s",
	CookieName:    "login_token",
	JwtIssuer:     "test",
	JwtSecret:     "test",
	LoginRedirect: "http://localhost/redirect",
	Claim: config.Claim{
		Name:  "test",
		URL:   "http://localhost/test",
		Roles: []string{"roleA"},
	},
}

var assetConfig = config.AssetSettings{
	AssetDir:    "./",
	AssetPrefix: "/assets",
}

type handlerOps struct {
	aiSvc  appinfo.Service
	docSvc document.Service
	fsSvc  filestore.FileService
}

func handler() http.Handler {
	return handlerWith(nil)
}

func handlerWith(ops *handlerOps) http.Handler {
	var (
		ai appinfo.Service
		ds document.Service
		fs filestore.FileService
	)

	ai = &mockAppInfoService{}
	ds = &mockDocumentService{}
	fs = &mockFileService{}

	if ops != nil {
		if ops.aiSvc != nil {
			ai = ops.aiSvc
		}
		if ops.docSvc != nil {
			ds = ops.docSvc
		}
		if ops.fsSvc != nil {
			fs = ops.fsSvc
		}
	}

	endpoints := mydms.MakeServerEndpoints(ai, ds, fs, logger)
	apiSrv := mydms.MakeHTTPHandler(endpoints, logger, mydms.HTTPHandlerOptions{
		BasePath:     "./",
		ErrorPath:    "/error",
		AssetConfig:  assetConfig,
		CookieConfig: config.ApplicationCookies{},
		CorsConfig:   config.CorsSettings{},
		JWTConfig:    jwtConfig,
	})
	return apiSrv
}

func addAuth(r *http.Request) {
	r.AddCookie(&http.Cookie{Name: "login_token", Value: jwt})
}

func addAuthToken(r *http.Request, token string) {
	r.AddCookie(&http.Cookie{Name: "login_token", Value: token})
}

// --------------------------------------------------------------------------
// Test cases
// --------------------------------------------------------------------------

// AppInfo ------------------------------------------------------------------

const appinfo_route = "/api/v1/appinfo"

func Test_GetAppInfo(t *testing.T) {
	var ai appinfo.AppInfo

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appinfo_route, nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &ai); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, "1.0", ai.VersionInfo.Version)
}

func Test_GetAppInfo_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appinfo_route, nil)
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		aiSvc: &mockAppInfoService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
}

func Test_GetAppInfo_MissingAuth(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appinfo_route, nil)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 401, pd.Status)
}

func Test_GetAppInfo_MissingPermissions(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appinfo_route, nil)
	addAuthToken(req, jwt_different_role)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, http.StatusUnauthorized, pd.Status)
}

// GetDocumentByID ----------------------------------------------------------

const doc_by_id_route = "/api/v1/documents/ID"

func Test_GetDocumentByID(t *testing.T) {
	var doc document.Document

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", doc_by_id_route, nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, "id", doc.ID)
}

func Test_GetDocumentByID_Invalid_Param(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/documents/", nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
}

func Test_GetDocumentByID_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", doc_by_id_route, nil)
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		docSvc: &mockDocumentService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
}

// DeleteDocumentByID -------------------------------------------------------

const delete_doc_by_id_route = "/api/v1/documents/ID"

func Test_DeleteDocumentByID(t *testing.T) {
	var result document.Result

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", delete_doc_by_id_route, nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, document.Deleted, result.ActionResult)
}

func Test_DeleteDocumentByID_Invalid_Param(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/documents/", nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
}

func Test_DeleteDocumentByID_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", delete_doc_by_id_route, nil)
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		docSvc: &mockDocumentService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
}

// SearchList ---------------------------------------------------------------

const search_list_route = "/api/v1/documents/%s/search"

func Test_SearchList(t *testing.T) {
	var result mydms.EntriesResult

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf(search_list_route, "tags"), nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 2, result.Lenght)

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf(search_list_route, "senders"), nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 2, result.Lenght)
}

func Test_SearchList_Invalid_URL(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf(search_list_route, ""), nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
}

func Test_SearchList_Invalid_Param(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf(search_list_route, "something_else"), nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
}

func Test_SearchList_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf(search_list_route, "tags"), nil)
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		docSvc: &mockDocumentService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
}

// SearchDocuments ----------------------------------------------------------

const search_documents_route = "/api/v1/documents/search"

func Test_SearchDocuments(t *testing.T) {
	var result document.PagedDocument

	var url string
	url = search_documents_route
	url = url + "?title=a&tag=b&limit=10&skip=0"

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 1, result.TotalEntries)
}

func Test_SearchDocuments_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	var url string
	url = search_documents_route
	url = url + "?title=a&tag=b&limit=10&skip=0"

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		docSvc: &mockDocumentService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
}

// SaveDocument -------------------------------------------------------------

const save_document_route = "/api/v1/documents"

func Test_SaveDocument(t *testing.T) {
	var result document.Document

	// arrange
	rec := httptest.NewRecorder()
	payload := `{
		"id": "ID",
		"title": "Title",
		"fileName": "FileName",
		"senders": ["sender"],
		"tags": ["tag"]
	}`
	req, _ := http.NewRequest("POST", save_document_route, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	// invalid document supplied
	rec = httptest.NewRecorder()
	var pd pkgerr.ProblemDetail
	req, _ = http.NewRequest("POST", save_document_route, strings.NewReader("{}"))
	req.Header.Add("Content-Type", "application/json")
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
}

func Test_SaveDocument_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	payload := `{
		"id": "ID",
		"title": "Title",
		"fileName": "FileName",
		"senders": ["sender"],
		"tags": ["tag"]
	}`
	req, _ := http.NewRequest("POST", save_document_route, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		docSvc: &mockDocumentService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 500, pd.Status)
}

// GetFile ------------------------------------------------------------------

const get_file_route = "/api/v1/file"

func Test_GetFile(t *testing.T) {
	var result []byte

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf(get_file_route+"?path=%s", base64.StdEncoding.EncodeToString([]byte("/a"))), nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	result = rec.Body.Bytes()
	assert.True(t, len(result) > 0)
}

func Test_GetFile_Validation(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// not base64

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", get_file_route+"?path=/a", nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)

	// missing path

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", get_file_route, nil)
	addAuth(req)

	// act
	handler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 400, pd.Status)
}

func Test_GetFile_Fail(t *testing.T) {
	var pd pkgerr.ProblemDetail

	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf(get_file_route+"?path=%s", base64.StdEncoding.EncodeToString([]byte("/a"))), nil)
	addAuth(req)

	// act
	handlerWith(&handlerOps{
		fsSvc: &mockFileService{fail: true},
	}).ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 500, pd.Status)
}

// ----------------------------------------------------------------------------------------------------------
// E2E-like test
// use the real service and the real repository to create a document, read a document and delete a document
// ----------------------------------------------------------------------------------------------------------

func e2eHandler(repo document.Repository) http.Handler {
	var (
		ai appinfo.Service
		ds document.Service
		fs filestore.FileService
		uc upload.Client
	)

	ai = &mockAppInfoService{}
	fs = &mockFileService{}
	uc = &mockUploadClient{}
	ds = document.NewService(logger, repo, fs, uc)

	endpoints := mydms.MakeServerEndpoints(ai, ds, fs, logger)
	apiSrv := mydms.MakeHTTPHandler(endpoints, logger, mydms.HTTPHandlerOptions{
		BasePath:     "./",
		ErrorPath:    "/error",
		AssetConfig:  assetConfig,
		CookieConfig: config.ApplicationCookies{},
		CorsConfig:   config.CorsSettings{},
		JWTConfig:    jwtConfig,
	})
	return apiSrv
}

func getRepo() document.Repository {
	// persistence store && application version
	con := persistence.NewConnForDb("sqlite3", "file:test.db?cache=shared&mode=memory")
	f, err := ioutil.ReadFile("./ddl_test.sql")
	if err != nil {
		panic("cannot read ddl_test.sql")
	}
	_, err = con.DB.Exec(string(f))
	if err != nil {
		panic("cannot setup sqlite!")
	}

	repo, err := document.NewRepository(con)
	if err != nil {
		panic(fmt.Sprintf("cannot establish database connection: %v", err))
	}
	return repo
}

func Test_DocumentCRUD(t *testing.T) {
	repo := getRepo()
	handler := e2eHandler(repo)

	// -------------------- create a new document -----------------------

	var result document.IDResult

	// arrange
	rec := httptest.NewRecorder()
	payload := `{
		"id": "ID",
		"title": "Title",
		"fileName": "FileName",
		"senders": ["sender"],
		"tags": ["tag"]
	}`
	req, _ := http.NewRequest("POST", save_document_route, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	addAuth(req)

	// act
	handler.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.True(t, result.ID != "")

	// -------------------- read the new document -----------------------

	var doc document.Document

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/documents/"+result.ID, nil)
	addAuth(req)

	// act
	handler.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	// -------------------- update the document -------------------------

	// arrange
	rec = httptest.NewRecorder()
	payload = `{
		"id": "ID",
		"title": "Title (updated)",
		"fileName": "FileName (updated)",
		"senders": ["sender1"],
		"tags": ["tag1"]
	}`
	req, _ = http.NewRequest("POST", save_document_route, strings.NewReader(strings.Replace(payload, "ID", result.ID, 1)))
	req.Header.Add("Content-Type", "application/json")
	addAuth(req)

	// act
	handler.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.True(t, result.ID != "")

	// -------------------- delete the document -------------------------

	var dr document.Result

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/documents/"+result.ID, nil)
	addAuth(req)

	// act
	handler.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &dr); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, document.Deleted, dr.ActionResult)

	// ----------------- read the document again ------------------------

	var pd pkgerr.ProblemDetail

	// arrange
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/documents/"+result.ID, nil)
	addAuth(req)

	// act
	handler.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, 404, pd.Status)

}
