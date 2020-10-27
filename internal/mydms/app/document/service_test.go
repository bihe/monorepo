package document_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/pkg/security"
)

var logger = log.NewLogfmtLogger(os.Stderr)

const expectations = "there were unfulfilled expectations: %s"

func Test_GetDocumentByID(t *testing.T) {
	svc := document.NewService(logger, &mockRepository{}, nil, nil)
	doc, err := svc.GetDocumentByID("id")
	if err != nil {
		t.Errorf("could not get document by id 'id'; %v", err)
	}
	assert.Equal(t, "id", doc.ID)

	// no document found
	doc, err = svc.GetDocumentByID("")
	if err == nil {
		t.Errorf("should return ErrNotFound")
	}
}

func Test_DeleteDocumentByID(t *testing.T) {
	c, db, mock := GetMockConn(t)
	defer db.Close()

	fileSvc := newFileService()
	svc := document.NewService(logger, newDocRepo(c), fileSvc, nil)

	// straight
	mock.ExpectBegin()
	mock.ExpectCommit()

	err := svc.DeleteDocumentByID("id")
	if err != nil {
		t.Errorf("could not delete document by id 'id'; %v", err)
	}

	// cannot find by ID
	mock.ExpectBegin()
	mock.ExpectRollback()

	err = svc.DeleteDocumentByID("!exists")
	if err == nil {
		t.Errorf("error expected for delete without found ID")
	}

	// cannot delete
	mock.ExpectBegin()
	mock.ExpectRollback()

	err = svc.DeleteDocumentByID("!delete")
	if err == nil {
		t.Errorf("error expected for delete")
	}

	// cannot delete backend fileservice
	mock.ExpectBegin()
	mock.ExpectRollback()

	fileSvc.callCount = 0
	fileSvc.errMap[1] = fmt.Errorf("error")

	err = svc.DeleteDocumentByID("id")
	if err == nil {
		t.Errorf("error expected for delete")
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}

func Test_SearchDocuments(t *testing.T) {
	svc := document.NewService(logger, &mockRepository{}, nil, nil)
	pd, err := svc.SearchDocuments("", "", "", time.Now(), time.Time{}, 0, 0)
	if err != nil {
		t.Errorf("error searching documents; %v", err)
	}
	assert.Equal(t, 2, pd.TotalEntries)

	_, err = svc.SearchDocuments("!result", "", "", time.Now(), time.Time{}, 20, 0)
	if err == nil {
		t.Errorf("search error expected, nothing found")
	}
}

func Test_SearchList(t *testing.T) {
	svc := document.NewService(logger, &mockRepository{}, nil, nil)
	l, err := svc.SearchList("name", document.SENDERS)
	if err != nil {
		t.Errorf("error searching for list '%s'; %v", "name", err)
	}
	assert.True(t, len(l) == 2)

	l, err = svc.SearchList("name", document.TAGS)
	if err != nil {
		t.Errorf("error searching for list '%s'; %v", "name", err)
	}
	assert.True(t, len(l) == 2)

	// fail
	svc = document.NewService(logger, &mockRepository{fail: true}, nil, nil)
	_, err = svc.SearchList("name", document.SENDERS)
	if err == nil {
		t.Errorf("error expected")
	}
}

func Test_SaveDocument(t *testing.T) {
	c, db, mock := GetMockConn(t)
	defer db.Close()

	fileSvc := newFileService()
	svc := document.NewService(logger, newDocRepo(c), fileSvc, &mockUploadClient{})

	// test a blank / new document
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectCommit()

	doc, err := svc.SaveDocument(document.Document{
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err != nil {
		t.Errorf("could not save document: %v", err)
	}
	// clean input via policy
	assert.Equal(t, "<span>Title</span>", doc.Title)

	// test an exiting document
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectCommit()

	doc, err = svc.SaveDocument(document.Document{
		ID:          "id",
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err != nil {
		t.Errorf("could not save document: %v", err)
	}
	// clean input via policy
	assert.Equal(t, "<span>Title</span>", doc.Title)

	// failed to load the existing document with the given ID
	// resulting in a new document
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectCommit()

	repo := newDocRepo(c)
	// att: call-count is not 0-based
	// repo.CreateAtomic is the first call
	repo.errMap[2] = fmt.Errorf("cannt get document by ID")
	svc = document.NewService(logger, repo, fileSvc, &mockUploadClient{})

	doc, err = svc.SaveDocument(document.Document{
		ID:          "id",
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err != nil {
		t.Errorf("could not save document: %v", err)
	}
	// clean input via policy
	assert.Equal(t, "<span>Title</span>", doc.Title)

	// failed to load save the document
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectRollback()

	repo = newDocRepo(c)
	// att: call-count is not 0-based
	// repo.CreateAtomic is the first call, repo.Get the second call
	repo.errMap[3] = fmt.Errorf("cannt get document by ID")
	svc = document.NewService(logger, repo, fileSvc, &mockUploadClient{})

	doc, err = svc.SaveDocument(document.Document{
		ID:          "id",
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err == nil {
		t.Errorf("error expected")
	}

	// missing upload file token / or filename
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectRollback()

	repo = newDocRepo(c)
	// att: call-count is not 0-based
	// repo.CreateAtomic is the first call, repo.Get the second call
	repo.errMap[3] = fmt.Errorf("cannt get document by ID")
	svc = document.NewService(logger, repo, fileSvc, &mockUploadClient{})

	doc, err = svc.SaveDocument(document.Document{
		ID:      "id",
		Title:   "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:    []string{"A", "B"},
		Senders: []string{"Sender"},
	}, security.User{})
	if err == nil {
		t.Errorf("error expected")
	}

	// failed to get file from upload-client
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectRollback()

	fileSvc = newFileService()
	svc = document.NewService(logger, newDocRepo(c), fileSvc, &mockUploadClient{getFail: true})

	doc, err = svc.SaveDocument(document.Document{
		ID:          "id",
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err == nil {
		t.Errorf("error expected")
	}

	// filesvc failed to save
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectRollback()

	fileSvc = newFileService()
	fileSvc.errMap[1] = fmt.Errorf("could not save uploaded file")
	svc = document.NewService(logger, newDocRepo(c), fileSvc, &mockUploadClient{getFail: false})

	doc, err = svc.SaveDocument(document.Document{
		ID:          "id",
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err == nil {
		t.Errorf("error expected")
	}

	// failed to delete the uploaded file
	// but this error is just loged, not returned
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectCommit()

	fileSvc = newFileService()
	svc = document.NewService(logger, newDocRepo(c), fileSvc, &mockUploadClient{deleteFail: true})

	doc, err = svc.SaveDocument(document.Document{
		ID:          "id",
		UploadToken: "token",
		Title:       "<span onclick=\"alert(Hello)\">Title</span>",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err != nil {
		t.Errorf("%v", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}
