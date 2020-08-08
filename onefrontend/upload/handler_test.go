package upload_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/onefrontend/upload"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/security"
)

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

const fileName = "test.pdf"
const multipartErr = "could not create multipart: %v"
const contentType = "Content-Type"

// common components necessary for handlers
var baseHandler = handler.Handler{
	ErrRep: &errors.ErrorReporter{
		CookieSettings: cookies.Settings{
			Path:   "/",
			Domain: "localhost",
		},
		ErrorPath: "error",
	},
	Log: logger,
}

type mockService struct {
	valErr bool
	srvErr bool
}

func (m *mockService) Save(file upload.File) (string, error) {
	if m.valErr {
		return "", fmt.Errorf("error, %w", upload.ErrValidation)
	}
	if m.srvErr {
		return "", fmt.Errorf("error")
	}
	return "id", nil
}

func (m *mockService) Read(id string) (upload.Upload, error) {
	if m.valErr {
		return upload.Upload{}, fmt.Errorf("error, %w", upload.ErrInvalidParameters)
	}
	if id == "id" {
		return upload.Upload{
			ID:       "id",
			FileName: fileName,
		}, nil
	}
	return upload.Upload{}, fmt.Errorf("error")
}

func (m *mockService) Delete(id string) error {
	if m.valErr {
		return fmt.Errorf("error, %w", upload.ErrInvalidParameters)
	}
	if id == "id" {
		return nil
	}
	return fmt.Errorf("error")
}

var _ upload.Service = &mockService{}

func jwtUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := security.NewContext(r.Context(), &security.User{
			Username:    "user",
			Email:       "a.b@c.de",
			DisplayName: "displayname",
			Roles:       []string{"role"},
			UserID:      "12345",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func router(svc upload.Service) *chi.Mux {
	handler := &upload.Handler{
		Handler: baseHandler,
		Service: svc,
	}
	r := chi.NewRouter()
	r.Use(jwtUser)
	r.Mount("/", handler.GetHandlers())
	return r
}

func Test_Handler_Save(t *testing.T) {
	// arrange
	r := router(&mockService{})

	url := "/file"
	rec := httptest.NewRecorder()
	body, ctype := getMultiPartPayload(t, "file")

	// act
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add(contentType, ctype)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusCreated, rec.Code)
	var response upload.ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.True(t, response.ID != "")
	assert.True(t, response.Message != "")
}

func Test_Handler_Save_Missing_File(t *testing.T) {
	// arrange
	r := router(&mockService{})

	url := "/file"
	rec := httptest.NewRecorder()
	body, ctype := getMultiPartPayload(t, "nofile")

	// act
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add(contentType, ctype)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 400, pd.Status)
}

func Test_Handler_Save_Validation_Err(t *testing.T) {
	// arrange
	r := router(&mockService{
		valErr: true,
	})

	url := "/file"
	rec := httptest.NewRecorder()
	body, ctype := getMultiPartPayload(t, "file")

	// act
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add(contentType, ctype)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 400, pd.Status)
}

func Test_Handler_Save_Server_Err(t *testing.T) {
	// arrange
	r := router(&mockService{
		srvErr: true,
	})

	url := "/file"
	rec := httptest.NewRecorder()
	body, ctype := getMultiPartPayload(t, "file")

	// act
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add(contentType, ctype)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 500, pd.Status)
}

func Test_Item_By_ID(t *testing.T) {
	// arrange
	r := router(&mockService{})

	url := "/id"
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	var response upload.ItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.True(t, response.ID != "")
	assert.Equal(t, fileName, response.FileName)
}

func Test_Item_By_ID_Parameter_Err(t *testing.T) {
	// arrange
	r := router(&mockService{
		valErr: true,
	})

	url := "/id"
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 400, pd.Status)
}

func Test_Item_By_ID_NotFound_Err(t *testing.T) {
	// arrange
	r := router(&mockService{})

	url := "/id1"
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 404, pd.Status)
}

func Test_Delete_Item_By_ID(t *testing.T) {
	// arrange
	r := router(&mockService{})

	url := "/id"
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("DELETE", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	var response upload.ResultResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.True(t, response.ID != "")
	assert.True(t, response.Message != "")
}

func Test_Delete_Item_By_ID_Parameter_Err(t *testing.T) {
	// arrange
	r := router(&mockService{})

	url := "/id1"
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("DELETE", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 404, pd.Status)
}

func Test_Delete_Item_By_ID_Validation_Err(t *testing.T) {
	// arrange
	r := router(&mockService{
		valErr: true,
	})

	url := "/id"
	rec := httptest.NewRecorder()

	// act
	req, _ := http.NewRequest("DELETE", url, nil)
	r.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var pd errors.ProblemDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &pd); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}

	assert.Equal(t, 400, pd.Status)
}

func getMultiPartPayload(t *testing.T, multiPartKey string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	part, err := writer.CreateFormFile(multiPartKey, fileName)
	if err != nil {
		t.Fatalf(multipartErr, err)
	}
	io.Copy(part, bytes.NewBuffer([]byte(pdfPayload)))
	ctype := writer.FormDataContentType()
	return body, ctype
}
