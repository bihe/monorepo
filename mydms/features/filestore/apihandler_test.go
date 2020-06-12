package filestore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

// PATH/file.pdf
const validPath = "UEFUSC9maWxlLnBkZg=="
const invalidPath = "UEFUSC9ub25lLnBkZg=="

// SaveFile(file FileItem) error
// GetFile(filePath string) (FileItem, error)
type mockService struct {
	s3service
}

func (m *mockService) GetFile(filePath string) (FileItem, error) {
	if filePath == "PATH/file.pdf" {
		return FileItem{
			FileName:   "file.pdf",
			FolderName: "PATH",
			MimeType:   "application.pdf",
			Payload:    []byte(pdfPayload),
		}, nil
	}
	return FileItem{}, fmt.Errorf("could not get file")
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(NewService(S3Config{
		Region: "region",
		Bucket: "bucket",
		Key:    "key",
		Secret: "secret",
	}), log.New().WithField("mode", "test"))
	if h == nil {
		t.Errorf("could not create a new handler")
	}
}

func TestGetFiles(t *testing.T) {

	cases := []struct {
		Name string
		Path string
		Size int
		Err  bool
	}{
		{
			Name: "valid file",
			Path: "UEFUSC9maWxlLnBkZg==", // PATH/file.pdf
			Size: len([]byte(pdfPayload)),
			Err:  false,
		},
		{
			Name: "invalid file",
			Path: "UEFUSC9ub25lLnBkZg==", // PATH/none.pdf
			Err:  true,
		},
		{
			Name: "bad encoding",
			Path: "UEFUSC9maWxlLn___BkZg==",
			Err:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/?path="+tc.Path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := &Handler{fs: new(mockService), log: log.New().WithField("mode", "test")}

			err := h.GetFile(c)
			isErr := err != nil
			if tc.Err != isErr {
				t.Fatalf("error expected status: %v, got %v", tc.Err, isErr)
			}
			payload := rec.Body.Bytes()
			if len(payload) != tc.Size {
				t.Errorf("the returend payload does not match!")
			}
		})
	}

}
