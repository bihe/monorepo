package filestore

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// MakeGetFileEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetFileEndpoint(f FileService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetFilePathRequest)
		item, err := f.GetFile(req.Path)
		return GetFilePathResponse{
			MimeType: item.MimeType,
			Payload:  item.Payload,
			Err:      err,
		}, nil
	}
}

// --------------------------------------------------------------------------
// Requests / Responses
// --------------------------------------------------------------------------

// GetFilePathRequest is used to retrieve a file by a given path
type GetFilePathRequest struct {
	Path string
}

// GetFilePathResponse returns a file payload
type GetFilePathResponse struct {
	MimeType string `json:"mimeType,omitempty"`
	Payload  []byte `json:"payload,omitempty"`
	Err      error  `json:"err,omitempty"`
}

// Failed implements endpoint.Failer.
func (r GetFilePathResponse) Failed() error { return r.Err }
