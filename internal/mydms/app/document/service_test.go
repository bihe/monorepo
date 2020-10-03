package document_test

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
)

var logger = log.NewLogfmtLogger(os.Stderr)

func Test_GetDocumentByID(t *testing.T) {
	svc := document.NewService(logger, &mockRepository{})
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
