package document_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/pkg/security"
)

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

func (s mockService) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p document.PagedDocument, err error) {
	if s.fail {
		return document.PagedDocument{}, fmt.Errorf("error")
	}
	return document.PagedDocument{
		Documents: []document.Document{
			{
				ID: "id",
			},
		},
		TotalEntries: 1,
	}, nil
}

func (s mockService) SearchList(name string, st document.SearchType) (l []string, err error) {
	if s.fail {
		return nil, fmt.Errorf("error")
	}
	return []string{"one", "two"}, nil
}

func (s mockService) SaveDocument(doc document.Document, user security.User) (d document.Document, err error) {
	if s.fail {
		return document.Document{}, fmt.Errorf("error")
	}
	return doc, nil
}

// --------------------------------------------------------------------------
// MakeGetDocumentByIDEndpoint
// --------------------------------------------------------------------------

func Test_Endpoint_GetDocumentByID(t *testing.T) {
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

func Test_Endpoint_GetDocumentByID_Fail(t *testing.T) {
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

// --------------------------------------------------------------------------
// MakeDeleteDocumentByIDEndpoint
// --------------------------------------------------------------------------

func Test_Endpoint_DeleteDocumentByID(t *testing.T) {
	e := document.MakeDeleteDocumentByIDEndpoint(mockService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.DeleteDocumentByIDRequest{ID: "id"})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.DeleteDocumentByIDResponse)
	if !ok {
		t.Error("response is not of type 'document.DeleteDocumentByIDResponse'")
	}
	if resp.Err != nil {
		t.Errorf("the response returned error: %v", resp.Err)
	}
	assert.Equal(t, "id", resp.ID)
}

func Test_Endpoint_DeleteDocumentByID_Fail(t *testing.T) {
	e := document.MakeDeleteDocumentByIDEndpoint(mockService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.DeleteDocumentByIDRequest{ID: "id"})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.DeleteDocumentByIDResponse)
	if !ok {
		t.Error("response is not of type 'document.DeleteDocumentByIDResponse'")
	}
	if resp.Err == nil {
		t.Error("error expected")
	}
	eString := resp.Failed().Error()
	assert.NotEmpty(t, eString)
}

// --------------------------------------------------------------------------
// MakeSearchListEndpoint
// --------------------------------------------------------------------------

func Test_Endpoint_MakeSearchListEndpoint(t *testing.T) {
	e := document.MakeSearchListEndpoint(mockService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.SearchListRequest{Name: "one", SearchType: document.Tags})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.SearchListResponse)
	if !ok {
		t.Error("response is not of type 'document.SearchListResponse'")
	}
	if resp.Err != nil {
		t.Errorf("the response returned error: %v", resp.Err)
	}
	assert.Equal(t, 2, len(resp.Entries))

	// use the endpoint
	r, err = e(context.TODO(), document.SearchListRequest{Name: "one", SearchType: document.Senders})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok = r.(document.SearchListResponse)
	if !ok {
		t.Error("response is not of type 'document.SearchListResponse'")
	}
	if resp.Err != nil {
		t.Errorf("the response returned error: %v", resp.Err)
	}
	assert.Equal(t, 2, len(resp.Entries))
}

func Test_Endpoint_MakeSearchListEndpoint_Fail(t *testing.T) {
	e := document.MakeSearchListEndpoint(mockService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.SearchListRequest{Name: "one", SearchType: document.Tags})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.SearchListResponse)
	if !ok {
		t.Error("response is not of type 'document.SearchListResponse'")
	}
	if resp.Err == nil {
		t.Error("error expected")
	}
	eString := resp.Failed().Error()
	assert.NotEmpty(t, eString)
}

// --------------------------------------------------------------------------
// MakeSearchDocumentsEndpoint
// --------------------------------------------------------------------------

func Test_Endpoint_SearchDocuments(t *testing.T) {
	e := document.MakeSearchDocumentsEndpoint(mockService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.SearchDocumentsRequest{})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.SearchDocumentsResponse)
	if !ok {
		t.Error("response is not of type 'document.SearchDocumentsResponse'")
	}
	if resp.Err != nil {
		t.Errorf("the response returned error: %v", resp.Err)
	}
	assert.Equal(t, 1, resp.Result.TotalEntries)
	assert.Equal(t, "id", resp.Result.Documents[0].ID)
}

func Test_Endpoint_SearchDocuments_Fail(t *testing.T) {
	e := document.MakeSearchDocumentsEndpoint(mockService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.SearchDocumentsRequest{})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.SearchDocumentsResponse)
	if !ok {
		t.Error("response is not of type 'document.SearchDocumentsResponse'")
	}
	if resp.Err == nil {
		t.Error("error expected")
	}
	eString := resp.Failed().Error()
	assert.NotEmpty(t, eString)
}

// --------------------------------------------------------------------------
// MakeSaveDocumentEnpoint
// --------------------------------------------------------------------------

func Test_Endpoint_SaveDocument(t *testing.T) {
	e := document.MakeSaveDocumentEnpoint(mockService{})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.SaveDocumentRequest{
		Document: document.Document{
			ID: "id",
		},
		User: security.User{
			UserID: "1",
		},
	})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.SaveDocumentResponse)
	if !ok {
		t.Error("response is not of type 'document.SaveDocumentResponse'")
	}
	if resp.Err != nil {
		t.Errorf("the response returned error: %v", resp.Err)
	}
	assert.Equal(t, "id", resp.ID)
}

func Test_Endpoint_SaveDocument_Fail(t *testing.T) {
	e := document.MakeSaveDocumentEnpoint(mockService{fail: true})
	if e == nil {
		t.Errorf("could not crate endpoint")
	}
	// use the endpoint
	r, err := e(context.TODO(), document.SaveDocumentRequest{
		Document: document.Document{
			ID: "id",
		},
		User: security.User{
			UserID: "1",
		},
	})
	if err != nil {
		t.Errorf("could not get response from endpoint; %v", err)
	}
	resp, ok := r.(document.SaveDocumentResponse)
	if !ok {
		t.Error("response is not of type 'document.SaveDocumentResponse'")
	}
	if resp.Err == nil {
		t.Error("error expected")
	}
	eString := resp.Failed().Error()
	assert.NotEmpty(t, eString)
}
