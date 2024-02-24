package document_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/upload"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

var logger = logging.NewNop()

const expectations = "there were unfulfilled expectations: %s"

var uploadSettings = upload.ServiceOptions{
	Logger:           logger,
	Store:            upload.NewStore("/tmp"),
	MaxUploadSize:    500000,
	AllowedFileTypes: []string{"pdf"},
}
var uploadSvc = upload.NewService(uploadSettings)

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

const unencryptedPDF = "../../../../testdata/unencrypted.pdf"

func Test_SaveDocument(t *testing.T) {
	c, db, mock := GetMockConn(t)
	defer db.Close()

	fileSvc := newFileService()
	svc := document.NewService(logger, newDocRepo(c), fileSvc, uploadSvc)

	// test a blank / new document
	// ------------------------------------------------------------------
	mock.ExpectBegin()
	mock.ExpectCommit()

	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	var b bytes.Buffer // A Buffer needs no initialization.
	if _, err := b.Write(payload); err != nil {
		t.Fatalf("could not write payload to buffer: %v", err)
	}

	// check that the supported file-types is case insensitive
	var id string
	if id, err = uploadSvc.Save(upload.File{
		File:     &b,
		MimeType: "application/pdf",
		Name:     "unencrypted.PDF",
		Size:     int64(len(payload)),
	}); err != nil {
		t.Fatalf("could not write file: %v", err)
	}

	doc, err := svc.SaveDocument(document.Document{
		UploadToken: id,
		Title:       "New-Document",
		Tags:        []string{"A", "B"},
		Senders:     []string{"Sender"},
	}, security.User{})
	if err != nil {
		t.Errorf("could not save document: %v", err)
	}
	// clean input via policy
	assert.Equal(t, "New-Document", doc.Title)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf(expectations, err)
	}
}
