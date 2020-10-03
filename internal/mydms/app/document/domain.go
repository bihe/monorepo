package document

// Document represents a document entity
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
