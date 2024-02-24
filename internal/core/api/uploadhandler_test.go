package api_test

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"mime/multipart"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"golang.binggl.net/monorepo/internal/core/api"
// 	"golang.binggl.net/monorepo/internal/core/app/upload"
// 	pkgerr "golang.binggl.net/monorepo/pkg/errors"
// )

// func uploadHandler() http.Handler {
// 	return handlerWith(&handlerOps{
// 		uploadSvc: &mockService{},
// 	})
// }

// func uploadHandlerSvc(valErr, srvErr bool) http.Handler {
// 	return handlerWith(&handlerOps{
// 		uploadSvc: &mockService{
// 			valErr: valErr,
// 			srvErr: srvErr,
// 		},
// 	})
// }

// // mock service implementation for upload.Service
// type mockService struct {
// 	valErr bool
// 	srvErr bool
// }

// func (m *mockService) Save(file upload.File) (string, error) {
// 	if m.valErr {
// 		return "", fmt.Errorf("error, %w", upload.ErrValidation)
// 	}
// 	if m.srvErr {
// 		return "", fmt.Errorf("error")
// 	}
// 	return "id", nil
// }

// func (m *mockService) Read(id string) (upload.Upload, error) {
// 	if m.valErr {
// 		return upload.Upload{}, fmt.Errorf("error, %w", upload.ErrInvalidParameters)
// 	}
// 	if id == "id" {
// 		return upload.Upload{
// 			ID:       "id",
// 			FileName: fileName,
// 		}, nil
// 	}
// 	return upload.Upload{}, fmt.Errorf("error")
// }

// func (m *mockService) Delete(id string) error {
// 	if m.valErr {
// 		return fmt.Errorf("error, %w", upload.ErrInvalidParameters)
// 	}
// 	if id == "id" {
// 		return nil
// 	}
// 	return fmt.Errorf("error")
// }

// var _ upload.Service = &mockService{}

// func Test_Upload_Item_By_ID_Auth_Fail(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/upload/1", nil)
// 	req.Header.Add("Accept", "application/json")

// 	// act
// 	uploadHandler().ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusUnauthorized, rec.Code)
// }

// func Test_Upload_Item_By_ID_Not_Found(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/upload/1", nil)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandler().ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusNotFound, rec.Code)
// }

// // rather small PDF payload
// // https://stackoverflow.com/questions/17279712/what-is-the-smallest-possible-valid-pdf
// const pdfPayload = `%PDF-1.0
// 1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj 2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj 3 0 obj<</Type/Page/MediaBox[0 0 3 3]>>endobj
// xref
// 0 4
// 0000000000 65535 f
// 0000000010 00000 n
// 0000000053 00000 n
// 0000000102 00000 n
// trailer<</Size 4/Root 1 0 R>>
// startxref
// 149
// %EOF
// `

// const fileName = "test.pdf"
// const multipartErr = "could not create multipart: %v"
// const contentType = "Content-Type"

// func Test_Upload_Save_File(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	body, ctype := getMultiPartPayload(t, "file")
// 	req, _ := http.NewRequest("POST", "/api/v1/upload/file", body)
// 	req.Header.Add(contentType, ctype)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandler().ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusCreated, rec.Code)
// 	var response api.UploadResult
// 	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
// 		t.Errorf("could not unmarshal: %v", err)
// 	}
// 	assert.True(t, response.ID != "")
// 	assert.True(t, response.Message != "")
// }

// func Test_Upload_Save_File_Missing(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	body, ctype := getMultiPartPayload(t, "file_missing")
// 	req, _ := http.NewRequest("POST", "/api/v1/upload/file", body)
// 	req.Header.Add(contentType, ctype)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandler().ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// 	var response pkgerr.ProblemDetail
// 	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
// 		t.Errorf("could not unmarshal: %v", err)
// 	}
// 	assert.Equal(t, http.StatusBadRequest, response.Status)
// }

// func Test_Upload_Save_File_Error(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	body, ctype := getMultiPartPayload(t, "file")
// 	req, _ := http.NewRequest("POST", "/api/v1/upload/file", body)
// 	req.Header.Add(contentType, ctype)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandlerSvc(false, true).ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
// 	var response pkgerr.ProblemDetail
// 	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
// 		t.Errorf("could not unmarshal: %v", err)
// 	}
// 	assert.Equal(t, http.StatusInternalServerError, response.Status)
// }

// func Test_Upload_Delete(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	req, _ := http.NewRequest("DELETE", "/api/v1/upload/id", nil)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandler().ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusOK, rec.Code)
// }

// func Test_Upload_Delete_Error(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	req, _ := http.NewRequest("DELETE", "/api/v1/upload/id", nil)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandlerSvc(true, false).ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
// }

// func Test_Upload_Delete_Not_Found(t *testing.T) {
// 	// arrange
// 	rec := httptest.NewRecorder()
// 	req, _ := http.NewRequest("DELETE", "/api/v1/upload/wrongID", nil)
// 	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

// 	// act
// 	uploadHandlerSvc(false, false).ServeHTTP(rec, req)

// 	// assert
// 	assert.Equal(t, http.StatusNotFound, rec.Code)
// }

// func getMultiPartPayload(t *testing.T, multiPartKey string) (*bytes.Buffer, string) {
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	defer writer.Close()

// 	part, err := writer.CreateFormFile(multiPartKey, fileName)
// 	if err != nil {
// 		t.Fatalf(multipartErr, err)
// 	}
// 	io.Copy(part, bytes.NewBuffer([]byte(pdfPayload)))
// 	ctype := writer.FormDataContentType()
// 	return body, ctype
// }
