package document

import "fmt"

// ActionResult is a code specifying a specific outcome/result
// swagger:enum ActionResult
type ActionResult string

const (
	// None is the default result
	None ActionResult = "none"
	// Saved indicates that an item was saved
	Saved ActionResult = "saved"
	// Deleted indicates that an item was deleted
	Deleted ActionResult = "deleted"
	// Error indicates any error
	Error = 99
)

// Result is a generic result object
// swagger:model
type Result struct {
	Message      string       `json:"message"`
	ActionResult ActionResult `json:"result"`
}

// IDResult is a generic result which returns an ID
// swagger:model
type IDResult struct {
	Result
	ID string `json:"id,omitempty"`
}

// Document represents a document entity
// swagger:model
type Document struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	AltID         string   `json:"alternativeId"`
	Amount        float32  `json:"amount"`
	Created       string   `json:"created"`
	Modified      string   `json:"modified,omitempty"`
	FileName      string   `json:"fileName"`
	PreviewLink   string   `json:"previewLink,omitempty"`
	UploadToken   string   `json:"uploadFileToken,omitempty"`
	Tags          []string `json:"tags"`
	Senders       []string `json:"senders"`
	InvoiceNumber string   `json:"invoiceNumber,omitempty"`
}

func (d Document) String() string {
	return fmt.Sprintf("%s (ID: %s)", d.Title, d.ID)
}

// PagedDocument represents a paged result
// swagger:model
type PagedDocument struct {
	Documents    []Document `json:"documents"`
	TotalEntries int        `json:"totalEntries"`
}
