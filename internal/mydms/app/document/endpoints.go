package document

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// MakeGetDocumentByIDEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetDocumentByIDEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetDocumentByIDRequest)
		doc, err := s.GetDocumentByID(req.ID)
		return GetDocumentByIDResponse{Document: doc, Err: err}, nil
	}
}

// --------------------------------------------------------------------------
// Requests / Responses
// --------------------------------------------------------------------------

// GetDocumentByIDRequest combines the necessary parameters for a document request
type GetDocumentByIDRequest struct {
	ID string
}

// GetDocumentByIDResponse is the document response object
type GetDocumentByIDResponse struct {
	Document Document
	Err      error `json:"err,omitempty"`
}

// Failed implements endpoint.Failer.
func (r GetDocumentByIDResponse) Failed() error { return r.Err }
