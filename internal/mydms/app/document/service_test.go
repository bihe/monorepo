package document_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
)

var logger = log.NewLogfmtLogger(os.Stderr)

const expectations = "there were unfulfilled expectations: %s"

func Test_GetDocumentByID(t *testing.T) {
	svc := document.NewService(logger, &mockRepository{}, nil)
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
	svc := document.NewService(logger, newDocRepo(c), fileSvc)

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

	var mockFs *mockFileService
	mockFs = fileSvc.(*mockFileService)
	mockFs.callCount = 0
	mockFs.errMap[1] = fmt.Errorf("error")

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
	svc := document.NewService(logger, &mockRepository{}, nil)
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
