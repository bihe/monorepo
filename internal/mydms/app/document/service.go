package document

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/microcosm-cc/bluemonday"
)

// --------------------------------------------------------------------------
// Service Definition
// --------------------------------------------------------------------------

// Service defines the methods of the document logic
type Service interface {
	GetDocumentByID(id string) (d Document, err error)
}

// NewService returns a Service with all of the expected middlewares wired in.
func NewService(logger log.Logger, repo Repository) Service {
	var svc Service
	{
		svc = &documentService{
			repo:   repo,
			policy: bluemonday.UGCPolicy(),
		}
		svc = ServiceLoggingMiddleware(logger)(svc)
	}
	return svc
}

var (
	// ErrNotFound notifies the caller, that the given element cannot be found
	ErrNotFound = errors.New("not found")
)

// --------------------------------------------------------------------------
// Service implementation
// --------------------------------------------------------------------------

const jsonTimeLayout = "2006-01-02T15:04:05+07:00"

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ Service = &documentService{}
)

type documentService struct {
	repo   Repository
	policy *bluemonday.Policy
}

// GetDocumentByID returns a Document object for a specified id, or returns an error if the document is not found
func (s documentService) GetDocumentByID(id string) (d Document, err error) {
	var doc DocEntity
	if doc, err = s.repo.Get(id); err != nil {
		return Document{}, fmt.Errorf("could not find doucment by id: %s, %w", id, ErrNotFound)
	}
	return s.convertToDomain(doc), nil
}

func (s documentService) convertToDomain(d DocEntity) Document {
	var (
		tags    []string
		senders []string
		cre     string
		mod     string
	)

	p := d.PreviewLink
	preview := ""
	if p.Valid {
		preview = p.String
	} else {
		// missing DB value for preview-link
		preview = base64.StdEncoding.EncodeToString([]byte(d.FileName))
	}
	if d.TagList != "" {
		tags = strings.Split(d.TagList, ";")
	}
	if d.SenderList != "" {
		senders = strings.Split(d.SenderList, ";")
	}
	cre = d.Created.Format(jsonTimeLayout)
	if d.Modified.Valid {
		mod = d.Modified.Time.Format(jsonTimeLayout)
	}

	inv := ""
	if d.InvoiceNumber.Valid {
		inv = d.InvoiceNumber.String
	}
	doc := sanitize(s.policy, &Document{
		ID:            d.ID,
		Title:         d.Title,
		AltID:         d.AltID,
		Amount:        d.Amount,
		Created:       cre,
		Modified:      mod,
		FileName:      d.FileName,
		PreviewLink:   preview,
		Tags:          tags,
		Senders:       senders,
		InvoiceNumber: inv,
	})
	return *doc
}

func sanitize(policy *bluemonday.Policy, d *Document) *Document {
	doc := Document{
		ID:     d.ID,
		Amount: d.Amount,
	}
	doc.Title = policy.Sanitize(d.Title)
	doc.AltID = policy.Sanitize(d.AltID)
	doc.Created = policy.Sanitize(d.Created)
	doc.Modified = policy.Sanitize(d.Modified)
	doc.FileName = policy.Sanitize(d.FileName)
	doc.PreviewLink = policy.Sanitize(d.PreviewLink)
	doc.UploadToken = policy.Sanitize(d.UploadToken)
	doc.InvoiceNumber = policy.Sanitize(d.InvoiceNumber)

	for _, t := range d.Tags {
		doc.Tags = append(doc.Tags, policy.Sanitize(t))
	}
	for _, s := range d.Senders {
		doc.Senders = append(doc.Senders, policy.Sanitize(s))
	}
	return &doc
}
