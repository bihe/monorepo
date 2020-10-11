package document

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/microcosm-cc/bluemonday"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/persistence"
)

// --------------------------------------------------------------------------
// Service Definition
// --------------------------------------------------------------------------

// Service defines the methods of the document logic
type Service interface {
	// GetDocumentByID returns a document object specified by the given id
	GetDocumentByID(id string) (d Document, err error)
	// DeleteDocumentByID deletes a document specified by the given id
	DeleteDocumentByID(id string) (err error)
	// SearchDocuments performs a search and returns paginated results
	SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDcoument, err error)
}

// NewService returns a Service with all of the expected middlewares wired in.
func NewService(logger log.Logger, repo Repository, fileSvc filestore.FileService) Service {
	var svc Service
	{
		svc = &documentService{
			repo:    repo,
			policy:  bluemonday.UGCPolicy(),
			logger:  logger,
			fileSvc: fileSvc,
		}
		svc = ServiceLoggingMiddleware(logger)(svc)
	}
	return svc
}

// --------------------------------------------------------------------------
// Service implementation
// --------------------------------------------------------------------------

const jsonTimeLayout = "2006-01-02T15:04:05+07:00"

// compile time assertions for our service
var (
	_ Service = &documentService{}
)

type documentService struct {
	repo    Repository
	policy  *bluemonday.Policy
	fileSvc filestore.FileService
	logger  log.Logger
}

// GetDocumentByID returns a Document object for a specified id, or returns an error if the document is not found
func (s documentService) GetDocumentByID(id string) (d Document, err error) {
	var doc DocEntity
	if doc, err = s.repo.Get(id); err != nil {
		return Document{}, shared.ErrNotFound(fmt.Sprintf("could not find doucment by id: %s", id))
	}
	return s.convertToDomain(doc), nil
}

// DeleteDocumentByID deletes a document identified by the id. the method returns an error if no document is availalbe for the given id
func (s documentService) DeleteDocumentByID(id string) (err error) {
	atomic, err := s.repo.CreateAtomic()
	if err != nil {
		return
	}
	// complete the atomic method
	defer func() {
		err = persistence.HandleTX(true, &atomic, err)
	}()

	fileName, err := s.repo.Exists(id, atomic)
	if err != nil {
		shared.Log(s.logger, "document.DeleteDocumentByID", err, fmt.Sprintf("the document '%s' is not available, %v", id, err))
		return shared.ErrNotFound(fmt.Sprintf("document '%s' not available", id))
	}

	err = s.repo.Delete(id, atomic)
	if err != nil {
		shared.Log(s.logger, "document.DeleteDocumentByID", err, fmt.Sprintf("error during delete operation of '%s', %v", id, err))
		return fmt.Errorf("could not delete '%s', %v", id, err)
	}

	// also remove the file payload stored in the backend store
	err = s.fileSvc.DeleteFile(fileName)
	if err != nil {
		shared.Log(s.logger, "document.DeleteDocumentByID", err, fmt.Sprintf("could not delete file in backend store '%s', %v", fileName, err))
		return fmt.Errorf("could not delete '%s', %v", id, err)
	}
	return nil
}

// SearchDocuments performs a search and returns paginated results
func (s documentService) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDcoument, err error) {
	var (
		order []OrderBy
		pd    PagedDcoument
	)

	// defaults
	if limit == 0 {
		limit = 20
	}
	orderByTitle := OrderBy{Field: "title", Order: ASC}
	orderByCreated := OrderBy{Field: "created", Order: DESC}

	docs, err := s.repo.Search(DocSearch{
		Title:  title,
		Tag:    tag,
		Sender: sender,
		From:   from,
		Until:  until,
		Limit:  limit,
		Skip:   skip,
	}, append(order, orderByCreated, orderByTitle))
	if err != nil {
		shared.Log(s.logger, "document.SearchDocuments", fmt.Errorf("search resulted in an error; %v", err))
		return pd, fmt.Errorf("cannot search for documents; %v", err)
	}

	pd.Documents = s.convertList(docs.Documents)
	pd.TotalEntries = docs.Count

	return pd, nil
}

// --------------------------------------------------------------------------
// internal helpers
// --------------------------------------------------------------------------

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
	doc := s.sanitize(&Document{
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

func (s documentService) convertList(ds []DocEntity) []Document {
	var (
		doc  Document
		docs []Document
	)
	for _, d := range ds {
		doc = s.convertToDomain(d)
		docs = append(docs, doc)
	}
	return docs
}

func (s documentService) sanitize(d *Document) *Document {
	doc := Document{
		ID:     d.ID,
		Amount: d.Amount,
	}
	doc.Title = s.policy.Sanitize(d.Title)
	doc.AltID = s.policy.Sanitize(d.AltID)
	doc.Created = s.policy.Sanitize(d.Created)
	doc.Modified = s.policy.Sanitize(d.Modified)
	doc.FileName = s.policy.Sanitize(d.FileName)
	doc.PreviewLink = s.policy.Sanitize(d.PreviewLink)
	doc.UploadToken = s.policy.Sanitize(d.UploadToken)
	doc.InvoiceNumber = s.policy.Sanitize(d.InvoiceNumber)

	for _, tag := range d.Tags {
		doc.Tags = append(doc.Tags, s.policy.Sanitize(tag))
	}
	for _, sender := range d.Senders {
		doc.Senders = append(doc.Senders, s.policy.Sanitize(sender))
	}
	return &doc
}
