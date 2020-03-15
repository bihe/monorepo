package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bihe/mydms/internal/persistence"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

// Write(item Upload, a persistence.Atomic) (err error)
// Read(id string) (Upload, error)
// Delete(id string, a persistence.Atomic) (err error)
type mockRepository struct{}

func (m mockRepository) Write(item Upload, a persistence.Atomic) (err error) {
	if item.FileName == fileName {
		return nil
	}
	return fmt.Errorf("error")
}

func (m mockRepository) Read(id string) (Upload, error) {
	return Upload{}, nil
}

func (m mockRepository) Delete(id string, a persistence.Atomic) (err error) {
	return nil
}

func (m mockRepository) CreateAtomic() (persistence.Atomic, error) {
	return persistence.Atomic{}, nil
}

func setup(t *testing.T, config Config, formfield, file string) (echo.Context, *Handler, *httptest.ResponseRecorder) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(formfield, file)
	if err != nil {
		t.Fatalf(multipartErr, err)
	}
	io.Copy(part, bytes.NewBuffer([]byte(pdfPayload)))
	writer.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	ctype := writer.FormDataContentType()
	req.Header.Add(contentType, ctype)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rw := mockRepository{}

	tmp := config.UploadPath
	if config.UploadPath == "" {
		if os.PathSeparator == '\\' {
			tmp = os.Getenv("TEMP")
		} else {
			tmp = "/tmp"
		}
	}
	config.UploadPath = tmp
	return c, NewHandler(rw, Config{
		AllowedFileTypes: config.AllowedFileTypes,
		MaxUploadSize:    config.MaxUploadSize,
		UploadPath:       tmp,
	}), rec
}

func TestUpload(t *testing.T) {
	// Setup
	c, h, rec := setup(t, Config{
		AllowedFileTypes: []string{"png", "pdf"},
		MaxUploadSize:    10000,
	}, "file", fileName)

	if assert.NoError(t, h.UploadFile(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		var r Result
		err := json.Unmarshal(rec.Body.Bytes(), &r)
		if err != nil {
			t.Errorf("could not get valid json: %v", err)
		}
	}
}

func TestUploadFail(t *testing.T) {
	// Setup
	var err error

	// file size
	c, h, _ := setup(t, Config{
		AllowedFileTypes: []string{"png", "pdf"},
		MaxUploadSize:    1,
	}, "file", fileName)
	err = h.UploadFile(c)
	if err == nil {
		t.Errorf("expected error file type!")
	}

	// upload destination
	c, h, _ = setup(t, Config{
		AllowedFileTypes: []string{"png", "pdf"},
		MaxUploadSize:    1,
		UploadPath:       "/NOTAVAIL/",
	}, "file", fileName)
	err = h.UploadFile(c)
	if err == nil {
		t.Errorf("expected error file type!")
	}
}

func TestMissingUploadFile(t *testing.T) {
	// Setup
	var err error
	c, h, _ := setup(t, Config{
		AllowedFileTypes: []string{"png", "pdf"},
		MaxUploadSize:    1000,
	}, "FILE__", fileName)

	err = h.UploadFile(c)
	if err == nil {
		t.Errorf("expected error missing file!")
	}
}

func TestUploadPeristenceFail(t *testing.T) {
	// Setup
	var err error
	c, h, _ := setup(t, Config{
		AllowedFileTypes: []string{"png", "pdf"},
		MaxUploadSize:    1000,
	}, "file", "error.pdf")

	err = h.UploadFile(c)
	if err == nil {
		t.Errorf("expected error persistence!")
	}
}

func TestUploadFileTypeFail(t *testing.T) {
	// Setup
	var err error
	c, h, _ := setup(t, Config{
		AllowedFileTypes: []string{"png"},
		MaxUploadSize:    1000,
	}, "file", fileName)

	err = h.UploadFile(c)
	if err == nil {
		t.Errorf("expected error upload file type!")
	}
}
