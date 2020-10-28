package document

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"golang.binggl.net/monorepo/pkg/security"
)

// MakeGetDocumentByIDEndpoint returns an endpoint to get a document by ID
func MakeGetDocumentByIDEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetDocumentByIDRequest)
		doc, err := s.GetDocumentByID(req.ID)
		return GetDocumentByIDResponse{Document: doc, Err: err}, nil
	}
}

// MakeDeleteDocumentByIDEndpoint returns an endpoint to delete a document by ID
func MakeDeleteDocumentByIDEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeleteDocumentByIDRequest)
		err = s.DeleteDocumentByID(req.ID)
		return DeleteDocumentByIDResponse{Err: err, ID: req.ID}, nil
	}
}

// MakeSearchListEndpoint returns an endpoint to search for lists {Tags, Senders}
func MakeSearchListEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SearchListRequest)

		var st SearchType
		switch req.SearchType {
		case Tags:
			st = TAGS
		case Senders:
			st = SENDERS
		}

		entries, err := s.SearchList(req.Name, st)
		return SearchListResponse{Err: err, Entries: entries}, nil
	}
}

// MakeSearchDocumentsEndpoint returns an endpoint to search for documents
func MakeSearchDocumentsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SearchDocumentsRequest)
		documents, err := s.SearchDocuments(req.Title, req.Tag, req.Sender, req.From, req.Until, req.Limit, req.Skip)
		return SearchDocumentsResponse{Err: err, Result: documents}, nil
	}
}

// MakeSaveDocumentEnpoint returns an endpoint to save a document
func MakeSaveDocumentEnpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SaveDocumentRequest)
		document, err := s.SaveDocument(req.Document, req.User)
		return SaveDocumentResponse{Err: err, ID: document.ID}, nil
	}
}

// --------------------------------------------------------------------------
// Requests / Responses
// --------------------------------------------------------------------------

// GetDocumentByIDRequest
// --------------------------------------------------------------------------

// GetDocumentByIDRequest combines the necessary parameters for a document request
type GetDocumentByIDRequest struct {
	ID string
}

// GetDocumentByIDResponse is the document response object
type GetDocumentByIDResponse struct {
	Err      error    `json:"err,omitempty"`
	Document Document `json:"document,omitempty"`
}

// Failed implements endpoint.Failer.
func (r GetDocumentByIDResponse) Failed() error { return r.Err }

// DeleteDocumentByIDRequest
// --------------------------------------------------------------------------

// DeleteDocumentByIDRequest holds an ID to delete a document
type DeleteDocumentByIDRequest struct {
	ID string
}

// DeleteDocumentByIDResponse is the response object of a delete operation
type DeleteDocumentByIDResponse struct {
	Err error  `json:"err,omitempty"`
	ID  string `json:"id,omitempty"`
}

// Failed implements endpoint.Failer.
func (r DeleteDocumentByIDResponse) Failed() error { return r.Err }

// SearchListRequest
// --------------------------------------------------------------------------

// ListType identifies a list
type ListType string

const (
	// Tags is a listtype holding tag-like entries
	Tags ListType = "Tags"
	// Senders is a listtype for email-like senders
	Senders ListType = "Senders"
)

// SearchListRequest searches the given name for a type of list {Tags, Senders}
type SearchListRequest struct {
	Name       string
	SearchType ListType
}

// SearchListResponse holds the result of the SearchListRequest or an error
type SearchListResponse struct {
	Err     error    `json:"err,omitempty"`
	Entries []string `json:"entries,omitempty"`
}

// Failed implements endpoint.Failer.
func (r SearchListResponse) Failed() error { return r.Err }

// SearchDocuments performs a search and returns paginated results
//SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDcoument, err error)

// SearchDocumentsRequest
// --------------------------------------------------------------------------

// SearchDocumentsRequest holds parameters to get a filtered list of documents
type SearchDocumentsRequest struct {
	// Title of the document, search for parts of the field
	Title string
	// Tag of a document, search for parts of the field
	Tag string
	// Sender of a document, search for parts of the field
	Sender string
	// From timestamp to search
	From time.Time
	// Until timestamp to search
	Until time.Time
	// Limit defines the max number of entries to return
	Limit int
	// Skip is used for paged-requests, skip #num entries
	Skip int
}

// SearchDocumentsResponse is the response of a search-request
type SearchDocumentsResponse struct {
	Err    error         `json:"err,omitempty"`
	Result PagedDocument `json:"result"`
}

// Failed implements endpoint.Failer.
func (r SearchDocumentsResponse) Failed() error { return r.Err }

// SaveDocumentRequest
// --------------------------------------------------------------------------

// SaveDocumentRequest holds a document which is either created or updated
type SaveDocumentRequest struct {
	Document Document
	User     security.User
}

// SaveDocumentResponse returns the document if the request was successfull
type SaveDocumentResponse struct {
	Err error  `json:"err,omitempty"`
	ID  string `json:"id,omitempty"`
}

// Failed implements endpoint.Failer.
func (r SaveDocumentResponse) Failed() error { return r.Err }
