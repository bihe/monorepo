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
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	conf "golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
)

const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

// --------------------------------------------------------------------------
// MOCK: filestore.FileService
// --------------------------------------------------------------------------

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
	errMap    map[int]error
	callCount int
}

func newFileService() *mockFileService {
	return &mockFileService{
		errMap: make(map[int]error),
	}
}

// compile time assertions for our service
var (
	_ filestore.FileService = &mockFileService{}
)

func (m *mockFileService) InitClient() (err error) {
	m.callCount++
	return m.errMap[m.callCount]
}

func (m *mockFileService) SaveFile(file filestore.FileItem) error {
	m.callCount++
	return m.errMap[m.callCount]
}

func (m *mockFileService) GetFile(filePath string) (filestore.FileItem, error) {
	m.callCount++
	return filestore.FileItem{
		FileName:   "test.pdf",
		FolderName: "PATH",
		MimeType:   "application/pdf",
		Payload:    []byte(pdfPayload),
	}, m.errMap[m.callCount]
}

func (m *mockFileService) DeleteFile(filePath string) error {
	m.callCount++
	return m.errMap[m.callCount]
}

var logger = logging.NewNop()

func handler(repo document.Repository) http.Handler {

	fileStore := newFileService()
	uploadStore := upload.NewStore("/tmp")
	uploadSvc := upload.NewService(upload.ServiceOptions{
		Logger: logger,
		Store:  uploadStore,
	})
	svc := document.NewService(logger,
		repo,
		fileStore, /* filestore.FileService */
		uploadSvc, /* upload.Service */
	)

	return mydms.MakeHTTPHandler(svc, uploadSvc, fileStore, logger, mydms.HTTPHandlerOptions{
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

func memRepo(t *testing.T) (document.Repository, shared.Connection) {
	con := shared.NewConnForDb("sqlite3", "file::memory:")
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
	req, _ := http.NewRequest("PUT", "/mydms/partial/list", strings.NewReader(formData.Encode()))
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
	req, _ = http.NewRequest("PUT", "/mydms/partial/list", strings.NewReader(formData.Encode()))
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
	req, _ = http.NewRequest("PUT", "/mydms/partial/list", strings.NewReader(formData.Encode()))
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// incorrect form data
	r = handler(repo)
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/mydms/partial/list", nil)
	addJwtAuth(req)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func Test_ShowEditDocumentDialog(t *testing.T) {
	repo, con := memRepo(t)
	defer con.Close()

	r := handler(repo)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/mydms/dialog/NEW", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// with empty/new ID we will receive a dialog to create a document
	payload := rec.Body.String()
	if !strings.Contains(payload, "Create Document") {
		t.Errorf("cannot find string in page result")
	}

	// a unknown ID still provides a dialog to create a NEW document
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/mydms/dialog/abc-def-ghi", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	payload = rec.Body.String()
	if !strings.Contains(payload, "Create Document") {
		t.Errorf("cannot find string in page result")
	}
}

func Test_SearchListItemsForAutocomplete(t *testing.T) {
	repo, con := memRepo(t)
	defer con.Close()

	r := handler(repo)
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mydms/list/tags", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	// empty database
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// missing list-type

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/mydms/list/senders", nil)
	addJwtAuth(req)

	r.ServeHTTP(rec, req)
	// empty database
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// // the response is a JSON array
	// type listItem struct {
	// 	Value string `json:"value,omitempty"`
	// 	Label string `json:"label,omitempty"`
	// }
	// listItems := make([]listItem, 0)
	// err := json.Unmarshal(rec.Body.Bytes(), &listItems)
	// if err != nil {
	// 	t.Errorf("could not get JSON structure: %v", err)
	// }

}
