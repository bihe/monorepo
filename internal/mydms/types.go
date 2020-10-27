package mydms

import (
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/mydms/app/document"
)

// EntriesResult holds a list of strings
// swagger:model
type EntriesResult struct {
	Lenght  int      `json:"length,omitempty"`
	Entries []string `json:"result,omitempty"`
}

// --------------------------------------------------------------------------
// DocumentRequest
// --------------------------------------------------------------------------

// DocumentRequestSwagger is used for swagger docu
// swagger:parameters SaveDocument
type DocumentRequestSwagger struct {
	// In: body
	Body document.Document
}

// DocumentRequest is the request payload for the Document data model.
type DocumentRequest struct {
	*document.Document
}

// Bind assigns the the provided data to a DocumentRequest
func (d *DocumentRequest) Bind(r *http.Request) error {
	// a Document is nil if no Document fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	if d.Document == nil {
		return fmt.Errorf("missing required Document fields")
	}
	return nil
}

// String returns a string representation of a DocumentRequest
func (d *DocumentRequest) String() string {
	return fmt.Sprintf("Document: '%s' (Id: '%s')", d.Title, d.ID)
}
