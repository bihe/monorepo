package document

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"golang.binggl.net/monorepo/internal/common/upload"
	"golang.binggl.net/monorepo/internal/mydms/app/filestore"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.binggl.net/monorepo/pkg/text"
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
	SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDocument, err error)
	// SearchList searches for senders or tags
	SearchList(name string, st SearchType) (l []string, err error)
	// SaveDocument receives a document and stores it
	// either creates a new document or updates an existing one
	SaveDocument(doc Document, user security.User) (d Document, err error)
}

// NewService returns a Service with all of the expected middlewares wired in.
func NewService(logger logging.Logger, repo Repository, fileSvc filestore.FileService, uploadSvc upload.Service) Service {
	var svc Service
	{
		svc = &documentService{
			repo:      repo,
			policy:    bluemonday.UGCPolicy(),
			logger:    logger,
			fileSvc:   fileSvc,
			uploadSvc: uploadSvc,
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
	repo      Repository
	policy    *bluemonday.Policy
	fileSvc   filestore.FileService
	uploadSvc upload.Service
	logger    logging.Logger
}

// GetDocumentByID returns a Document object for a specified id, or returns an error if the document is not found
func (s documentService) GetDocumentByID(id string) (d Document, err error) {
	var doc DocEntity
	if doc, err = s.repo.Get(id); err != nil {
		s.logger.Error("GetDocumentByID: repository error", logging.ErrV(fmt.Errorf("could not find doucment by id: '%s', %v", id, err)))
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
		err = shared.HandleTX(true, &atomic, err)
	}()

	fileName, err := s.repo.Exists(id, atomic)
	if err != nil {
		s.logger.Error("DeleteDocumentByID: error in repository", logging.ErrV(fmt.Errorf("the document '%s' is not available, %v", id, err)))
		return shared.ErrNotFound(fmt.Sprintf("document '%s' not available", id))
	}

	err = s.repo.Delete(id, atomic)
	if err != nil {
		s.logger.Error("DeleteDocumentByID: error in repository", logging.ErrV(fmt.Errorf("error during delete operation of '%s', %v", id, err)))
		return fmt.Errorf("could not delete '%s', %v", id, err)
	}

	// also remove the file payload stored in the backend store
	err = s.fileSvc.DeleteFile(fileName)
	if err != nil {
		s.logger.Error("DeleteDocumentByID: error in file-service", logging.ErrV(fmt.Errorf("could not delete file in backend store '%s', %v", fileName, err)))
		return fmt.Errorf("could not delete '%s', %v", id, err)
	}
	return nil
}

// SearchDocuments performs a search and returns paginated results
func (s documentService) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDocument, err error) {
	var (
		order []OrderBy
		pd    PagedDocument
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
		s.logger.Error("SearchDocuments: repository error", logging.ErrV(fmt.Errorf("search resulted in an error; %v", err)))
		return pd, fmt.Errorf("cannot search for documents; %v", err)
	}

	pd.Documents = s.convertList(docs.Documents)
	pd.TotalEntries = docs.Count

	return pd, nil
}

// SearchList searches for senders or tags
func (s documentService) SearchList(name string, st SearchType) (l []string, err error) {
	result, err := s.repo.SearchLists(name, st)
	if err != nil {
		s.logger.Error("could not search the list", logging.ErrV(fmt.Errorf("error searching for '%s'; %v", name, err)))
		return nil, fmt.Errorf("could not search for '%s': %v", name, err)
	}
	return result, nil
}

// SaveDocument receives a document and stores it
// either creation a new document or updating an existing
func (s documentService) SaveDocument(doc Document, user security.User) (d Document, err error) {
	var (
		docE DocEntity
	)

	atomic, err := s.repo.CreateAtomic()
	if err != nil {
		s.logger.Error("SaveDocument: error in repository", logging.ErrV(fmt.Errorf("could not start tx; %v", err)))
		return
	}

	// complete the atomic method
	defer func() {
		err = shared.HandleTX(true, &atomic, err)
	}()

	cleanDoc := s.sanitize(&doc)
	d = *cleanDoc

	filename, err := s.processUploadFile(d.UploadToken, d.FileName)
	if err != nil {
		s.logger.Error("SaveDocument: upload-processing error", logging.ErrV(fmt.Errorf("could not process the uploaded file, %v", err)))
		return
	}
	if filename == "" {
		s.logger.Error("SaveDocument: missing filename", logging.ErrV(fmt.Errorf("processUploadFile did not return an error, but the filename is empty")))
		return d, fmt.Errorf("no filename is available for the document")
	}
	d.FileName = filename

	tagList := strings.Join(d.Tags, ";")
	senderList := strings.Join(d.Senders, ";")

	if d.ID == "" {
		docE = initDocument(&d, senderList, tagList)
	} else {
		// supplied ID needs to be checked if exists
		docE, err = s.repo.Get(d.ID)
		if err != nil {
			s.logger.Info(fmt.Sprintf("SaveDocument: cannot find document by ID '%s' - create a new entry, %v", d.ID, err))
			docE = initDocument(&d, senderList, tagList)
		} else {
			s.logger.Info(fmt.Sprintf("SaveDocument: will update existing document ID '%s'", d.ID))
			docE.Title = d.Title
			docE.FileName = d.FileName
			docE.PreviewLink = sql.NullString{String: text.EncBase64SafePath(d.FileName), Valid: true}
			docE.Amount = d.Amount
			docE.SenderList = senderList
			docE.TagList = tagList
			docE.InvoiceNumber = sql.NullString{String: d.InvoiceNumber, Valid: true}
		}
	}

	docE, err = s.repo.Save(docE, atomic)
	if err != nil {
		s.logger.Error("error saving document with repository", logging.ErrV(fmt.Errorf("could not save document: %v", err)))
		return d, fmt.Errorf("error while saving document: %v", err)
	}

	err = s.uploadSvc.Delete(d.UploadToken)
	if err != nil {
		// this error is ignored, does not invalidate the overall operation
		s.logger.Warn("unable to delete uploaded file", logging.ErrV(fmt.Errorf("could not delete the upload-item by id '%s', %v", d.UploadToken, err)))
	}

	return s.convertToDomain(docE), nil
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
		preview = text.EncBase64SafePath(d.FileName)
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

func (s documentService) processUploadFile(uploadToken, fileName string) (string, error) {
	if uploadToken == "" || uploadToken == "-" {
		return fileName, nil
	}
	u, err := s.uploadSvc.Read(uploadToken)
	if err != nil {
		s.logger.Error("upload returned error", logging.ErrV(fmt.Errorf("could not read upload-file for token '%s', %v", uploadToken, err)))
		return "", fmt.Errorf("upload token error: %v", err)
	}
	s.logger.Info(fmt.Sprintf("use uploaded file identified by token '%s'", uploadToken))

	now := time.Now().UTC()
	folder := now.Format("2006_01_02")
	s.logger.Info(fmt.Sprintf("got upload file '%s' with payload size '%d'!", u.FileName, len(u.Payload)))

	item := filestore.FileItem{
		FileName:   fileName,
		FolderName: folder,
		MimeType:   u.MimeType,
		Payload:    u.Payload,
	}
	err = s.fileSvc.SaveFile(item)
	if err != nil {
		s.logger.Error("unable to save file", logging.ErrV(fmt.Errorf("could not save file '%s', %v", u.FileName, err)))
		return "", fmt.Errorf("error while saving file: %v", err)
	}

	return fmt.Sprintf("/%s/%s", folder, fileName), nil
}

func initDocument(d *Document, sList, tList string) DocEntity {
	return DocEntity{
		Title:         d.Title,
		FileName:      d.FileName,
		PreviewLink:   sql.NullString{String: text.EncBase64SafePath(d.FileName), Valid: true},
		Amount:        d.Amount,
		SenderList:    sList,
		TagList:       tList,
		InvoiceNumber: sql.NullString{String: d.InvoiceNumber, Valid: true},
	}
}
