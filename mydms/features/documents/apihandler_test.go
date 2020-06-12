package documents

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.binggl.net/monorepo/mydms/features/upload"
	"golang.binggl.net/monorepo/mydms/persistence"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

const invalidJSON = "could not get valid json: %v"
const couldNotSave = "could not save documents: %v"
const errorUnmarshal = "could not unmarshal json: %v"
const errExp = "error expected"
const ID = "id"
const notExists = "!exists"
const noDelete = "!delete"
const noResult = "!result"
const noFileDelete = "!fileDelete"

var logger = log.New().WithField("mode", "test")
var errRaise = fmt.Errorf("error")

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

var uploadConfig = upload.Config{
	AllowedFileTypes: []string{"png", "pdf"},
	MaxUploadSize:    1,
	UploadPath:       "/tmp/",
}

func TestActionResults(t *testing.T) {
	if None.mapToString() != "None" {
		t.Errorf("None != None")
	}
	if Created.mapToString() != "Created" {
		t.Errorf("Created != Created")
	}
	if Updated.mapToString() != "Updated" {
		t.Errorf("Updated != Updated")
	}
	if Deleted.mapToString() != "Deleted" {
		t.Errorf("Deleted != Deleted")
	}
	var a ActionResult
	a = Error
	if a.mapToString() != "Error" {
		t.Errorf("Error != Error")
	}

	if a.mapToAction("None") != None {
		t.Errorf("None != None")
	}
	if a.mapToAction("Created") != Created {
		t.Errorf("Created != Created")
	}
	if a.mapToAction("Updated") != Updated {
		t.Errorf("Updated != Updated")
	}
	if a.mapToAction("Deleted") != Deleted {
		t.Errorf("Deleted != Deleted")
	}
	if a.mapToAction("Error") != Error {
		t.Errorf("Error != Error")
	}

	a = Created
	b, err := a.MarshalJSON()
	if err != nil {
		t.Errorf("could not marshall json: %v", err)
	}
	if string(b) != "\"Created\"" {
		t.Errorf("wrong json marshalling")
	}
	aa := new(ActionResult)
	err = aa.UnmarshalJSON([]byte("\"Deleted\""))
	if err != nil {
		t.Errorf("cannot unmarshall: %v", err)
	}
	if *aa != Deleted {
		t.Errorf("wrong json unmarshalling")
	}

	err = aa.UnmarshalJSON([]byte("__Deleted\""))
	if err == nil {
		t.Errorf(expectedErr)
	}
}

func TestGetDocumentByID(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mdr := &mockRepository{}
	svc := &mockFileService{}
	repos := Repositories{
		DocRepo: mdr,
	}
	h := NewHandler(repos, svc, uploadConfig, logger)

	e.GET("/:id", h.GetDocumentByID) // this is necessary to supply parameters
	c := e.NewContext(req, rec)
	c.SetParamNames(ID)
	c.SetParamValues(ID)

	err := h.GetDocumentByID(c)
	if err != nil {
		t.Errorf("cannot get document by id: %v", err)
	}

	assert.Equal(t, http.StatusOK, rec.Code)
	var doc Document
	err = json.Unmarshal(rec.Body.Bytes(), &doc)
	if err != nil {
		t.Errorf(invalidJSON, err)
	}

	// error
	c = e.NewContext(req, rec)

	mdr = &mockRepository{}
	repos = Repositories{
		DocRepo: mdr,
	}
	h = NewHandler(repos, svc, uploadConfig, logger)

	err = h.GetDocumentByID(c)
	if err == nil {
		t.Errorf(errExp)
	}
}

func TestDeleteDocumentByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	con := persistence.NewFromDB(dbx)

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()

	mdr := newDocRepo(con)
	svc := newFileService()
	repos := Repositories{
		DocRepo: mdr,
	}

	h := NewHandler(repos, svc, uploadConfig, logger)

	e.GET("/:id", h.DeleteDocumentByID) // this is necessary to supply parameters
	c := e.NewContext(req, rec)
	c.SetParamNames(ID)
	c.SetParamValues(ID)

	// straight
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = h.DeleteDocumentByID(c)
	if err != nil {
		t.Errorf("cannot delete document by id: %v", err)
	}

	// start transaction failes
	c = e.NewContext(req, rec)
	failmdr := newDocRepo(con)
	failmdr.fail = true
	faileRepo := Repositories{
		DocRepo: failmdr,
	}
	failH := NewHandler(faileRepo, svc, uploadConfig, logger)
	err = failH.DeleteDocumentByID(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error exists
	mock.ExpectBegin()
	mock.ExpectRollback()

	c = e.NewContext(req, rec)
	c.SetParamNames(ID)
	c.SetParamValues(notExists)
	err = h.DeleteDocumentByID(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error delete
	mock.ExpectBegin()
	mock.ExpectRollback()

	c = e.NewContext(req, rec)
	c.SetParamNames(ID)
	c.SetParamValues(noDelete)
	err = h.DeleteDocumentByID(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error no file delete
	mock.ExpectBegin()
	mock.ExpectRollback()
	svc.callCount = 0
	svc.errMap[1] = errRaise
	h = NewHandler(repos, svc, uploadConfig, logger)

	c = e.NewContext(req, rec)
	c.SetParamNames(ID)
	c.SetParamValues(noFileDelete)
	err = h.DeleteDocumentByID(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestSearchDocuments(t *testing.T) {
	// Setup
	e := echo.New()

	q := make(url.Values)
	q.Set("skip", "0")
	q.Set("limit", "20")
	q.Set("from", time.Now().Format(time.RFC3339))

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mdr := &mockRepository{}
	svc := &mockFileService{}

	repos := Repositories{
		DocRepo: mdr,
	}
	h := NewHandler(repos, svc, uploadConfig, logger)

	// success
	err := h.SearchDocuments(c)
	if err != nil {
		t.Errorf("cannot search for documents: %v", err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
	var pd PagedDcoument
	err = json.Unmarshal(rec.Body.Bytes(), &pd)
	if err != nil {
		t.Errorf(errorUnmarshal, err)
	}
	assert.Equal(t, 2, pd.TotalEntries)
	assert.Equal(t, 2, len(pd.Documents))

	// error
	q = make(url.Values)
	q.Set("skip", "-")
	q.Set("limit", "-")
	q.Set("title", noResult)

	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = h.SearchDocuments(c)
	if err == nil {
		t.Errorf(errExp)
	}
}

func TestSaveUpdateDocument(t *testing.T) {
	// Setup
	e := echo.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	con := persistence.NewFromDB(dbx)

	updateJSON := `{
  "id":"f03756a1-59ad-426f-9e59-d1ee227edf0d",
  "title":"Test",
  "fileName":"/2019_09_07/test.pdf",
  "alternativeId":"UP5XqwA3",
  "previewLink":
  "LzIwMTlfMDlfMDcvaW52b2ljZS5wZGY=",
  "amount":116,
  "created":"2019-09-07T11:26:56",
  "modified":null,
  "tags":["Tag1","Tag2"],
  "senders":["Sender1"],
  "uploadFileToken":"-",
  "invoicenumber": "12345"
}`

	newReq := func(reader *strings.Reader) (request *http.Request, recorder *httptest.ResponseRecorder, context echo.Context) {
		request = httptest.NewRequest(http.MethodPost, "/", reader)
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder = httptest.NewRecorder()
		context = e.NewContext(request, recorder)
		return
	}

	_, rec, c := newReq(strings.NewReader(updateJSON))
	docRepo := newDocRepo(con)
	uploadRepo := newUploadRepo()
	svc := newFileService()
	repos := Repositories{
		DocRepo:    docRepo,
		UploadRepo: uploadRepo,
	}
	h := NewHandler(repos, svc, uploadConfig, logger)

	// update success
	mock.ExpectBegin()
	mock.ExpectCommit()
	err = h.SaveDocument(c)
	if err != nil {
		t.Errorf(couldNotSave, err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
	var result Result
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Errorf(errorUnmarshal, err)
	}
	assert.Equal(t, Updated, result.Result)

	// error get document - create new
	_, rec, c = newReq(strings.NewReader(updateJSON))

	docRepo.callCount = 0
	docRepo.errMap[2] = errRaise
	h = NewHandler(repos, svc, uploadConfig, logger)
	mock.ExpectBegin()
	mock.ExpectCommit()
	err = h.SaveDocument(c)
	if err != nil {
		t.Errorf(couldNotSave, err)
	}
	delete(docRepo.errMap, 1)

	// error save document
	_, rec, c = newReq(strings.NewReader(updateJSON))

	docRepo.callCount = 0
	docRepo.errMap[3] = errRaise
	h = NewHandler(repos, svc, uploadConfig, logger)
	mock.ExpectBegin()
	mock.ExpectRollback()
	err = h.SaveDocument(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestSaveNewDocument(t *testing.T) {
	// Setup
	e := echo.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf(fatalErr, err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "mysql")
	con := persistence.NewFromDB(dbx)

	initialJSON := `{
  "title":"Test",
  "fileName":"test.pdf",
  "amount":116,
  "created":"2019-09-07T11:26:56",
  "modified":null,
  "tags":["Tag1","Tag2"],
  "senders":["Sender1","Sender2"],
  "uploadFileToken":"ABC",
  "invoicenumber": "12345"
}`
	doError := fmt.Errorf("error")

	newReq := func() (request *http.Request, recorder *httptest.ResponseRecorder, context echo.Context) {
		request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(initialJSON))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder = httptest.NewRecorder()
		context = e.NewContext(request, recorder)
		return
	}

	_, rec, c := newReq()
	docRepo := newDocRepo(con)
	uploadRepo := newUploadRepo()
	svc := newFileService()
	repos := Repositories{
		DocRepo:    docRepo,
		UploadRepo: uploadRepo,
	}

	// ------------------------------------------------------------------
	// save a random file for the upload-logic!
	tempPath := getTempPath()
	uploadFile := "ABC.pdf"
	uploadConfig.UploadPath = tempPath
	uploadFile = filepath.Join(uploadConfig.UploadPath, uploadFile)
	ioutil.WriteFile(uploadFile, []byte(pdfPayload), 0644)
	defer func() {
		os.Remove(uploadFile)
	}()
	// ------------------------------------------------------------------

	h := NewHandler(repos, svc, uploadConfig, logger)

	// insert success
	mock.ExpectBegin()
	mock.ExpectCommit()
	err = h.SaveDocument(c)
	if err != nil {
		t.Errorf(couldNotSave, err)
	}
	assert.Equal(t, http.StatusCreated, rec.Code)
	var result Result
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Errorf(errorUnmarshal, err)
	}
	assert.Equal(t, Created, result.Result)

	// error processUploadFile
	ioutil.WriteFile(uploadFile, []byte(pdfPayload), 0644)
	_, rec, c = newReq()

	mock.ExpectBegin()
	mock.ExpectRollback()
	uploadRepo.callCount = 0
	uploadRepo.errMap[1] = doError
	h = NewHandler(repos, svc, uploadConfig, logger)
	err = h.SaveDocument(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error uploadfile
	ioutil.WriteFile(uploadFile, []byte(pdfPayload), 0644)
	_, rec, c = newReq()

	mock.ExpectBegin()
	mock.ExpectRollback()
	uploadConfig.UploadPath = "--"
	h = NewHandler(repos, svc, uploadConfig, logger)
	err = h.SaveDocument(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error save backend
	ioutil.WriteFile(uploadFile, []byte(pdfPayload), 0644)
	_, rec, c = newReq()

	mock.ExpectBegin()
	mock.ExpectRollback()
	uploadConfig.UploadPath = tempPath
	svc.callCount = 0
	svc.errMap[1] = doError
	h = NewHandler(repos, svc, uploadConfig, logger)
	err = h.SaveDocument(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error uploadrepo delete - no rollback herer
	ioutil.WriteFile(uploadFile, []byte(pdfPayload), 0644)
	_, rec, c = newReq()

	mock.ExpectBegin()
	mock.ExpectCommit()
	uploadConfig.UploadPath = tempPath
	uploadRepo.callCount = 0
	delete(uploadRepo.errMap, 1)
	uploadRepo.errMap[2] = doError
	h = NewHandler(repos, svc, uploadConfig, logger)
	err = h.SaveDocument(c)
	if err != nil {
		t.Errorf(couldNotSave, err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func TestSearchList(t *testing.T) {
	var (
		rec *httptest.ResponseRecorder
		c   echo.Context
	)

	// Setup
	e := echo.New()

	reqURL := "/?name=search"

	newReq := func(param, value string, h func(c echo.Context) (err error)) (request *http.Request, recorder *httptest.ResponseRecorder, context echo.Context) {
		request = httptest.NewRequest(http.MethodGet, reqURL, nil)
		recorder = httptest.NewRecorder()

		e.GET("/:type", h) // this is necessary to supply parameters
		context = e.NewContext(request, recorder)
		context.SetParamNames(param)
		context.SetParamValues(value)

		return
	}

	mdr := &mockRepository{}
	svc := &mockFileService{}

	repos := Repositories{
		DocRepo: mdr,
	}
	h := NewHandler(repos, svc, uploadConfig, logger)
	_, rec, c = newReq("type", "tags", h.SearchList)

	var result SearchResult

	// success
	err := h.SearchList(c)
	if err != nil {
		t.Errorf("cannot search for lists: %v", err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Errorf(errorUnmarshal, err)
	}
	assert.Equal(t, 2, result.Length)
	assert.Equal(t, 2, len(result.Result))

	// senders
	_, rec, c = newReq("type", "senders", h.SearchList)
	err = h.SearchList(c)
	if err != nil {
		t.Errorf("cannot search for lists: %v", err)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	if err != nil {
		t.Errorf(errorUnmarshal, err)
	}
	assert.Equal(t, 2, result.Length)
	assert.Equal(t, 2, len(result.Result))

	// wrong search-type
	_, rec, c = newReq("type", "something", h.SearchList)

	err = h.SearchList(c)
	if err == nil {
		t.Errorf(errExp)
	}

	// error result
	mdr.fail = true
	repos = Repositories{
		DocRepo: mdr,
	}
	h = NewHandler(repos, svc, uploadConfig, logger)
	_, rec, c = newReq("type", "tags", h.SearchList)

	err = h.SearchList(c)
	if err == nil {
		t.Errorf(errExp)
	}

}

func getTempPath() string {
	dir, _ := ioutil.TempDir("", "mydms")
	return dir
}
