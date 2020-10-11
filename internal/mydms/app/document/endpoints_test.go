package document_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
)

// type Service interface {
// 	// GetDocumentByID returns a document object specified by the given id
// 	GetDocumentByID(id string) (d Document, err error)
// 	// DeleteDocumentByID deletes a document specified by the given id
//	DeleteDocumentByID(id string) (err error)
// 	// SearchDocuments performs a search and returns paginated results
// 	SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDcoument, err error)
// }
type mockService struct {
	fail bool
}

func (s mockService) GetDocumentByID(id string) (d document.Document, err error) {
	if s.fail {
		return document.Document{}, fmt.Errorf("error")
	}
	return document.Document{
		ID: "id",
	}, nil
}

func (s mockService) DeleteDocumentByID(id string) (err error) {
	if s.fail {
		return fmt.Errorf("error")
	}
	return nil
}

func (s mockService) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p document.PagedDcoument, err error) {
	if s.fail {
		return document.PagedDcoument{}, fmt.Errorf("error")
	}
	return document.PagedDcoument{}, nil
}

func Test_Endpoints(t *testing.T) {
	e := document.MakeGetDocumentByIDEndpoint(mockService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.GetDocumentByIDRequest{ID: "id"})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.GetDocumentByIDResponse)
	if !ok {
		t.Error("response is not of type 'document.GetDocumentByIDResponse'")
	}
	assert.Equal(t, "id", resp.Document.ID)
}

func Test_Endpoints_Fail(t *testing.T) {
	e := document.MakeGetDocumentByIDEndpoint(mockService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.GetDocumentByIDRequest{ID: "id"})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.GetDocumentByIDResponse)
	if !ok {
		t.Error("response is not of type 'document.GetDocumentByIDResponse'")
	}

	if resp.Err == nil {
		t.Error("error expected")
	}
	eString := resp.Failed().Error()
	assert.NotEmpty(t, eString)
}
