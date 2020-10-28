package document

import (
	"time"

	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/security"
)

// ServiceMiddleware describes a service (as opposed to endpoint) middleware.
// it is used to intercept the method execution and perform actions before/after the
// serivce method execution
type ServiceMiddleware func(Service) Service

// ServiceLoggingMiddleware takes a logger as a dependency
// and returns a ServiceLoggingMiddleware.
func ServiceLoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

// compile guard for Service implementation
var (
	_ Service = &loggingMiddleware{}
)

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) GetDocumentByID(id string) (d Document, err error) {
	shared.Log(mw.logger, "GetDocumentByID", err, "param:ID", id)
	defer shared.Log(mw.logger, "GetDocumentByID-End", err)
	return mw.next.GetDocumentByID(id)
}

func (mw loggingMiddleware) DeleteDocumentByID(id string) (err error) {
	shared.Log(mw.logger, "DeleteDocumentByID", err, "param:ID", id)
	defer shared.Log(mw.logger, "DeleteDocumentByID-End", err)
	return mw.next.DeleteDocumentByID(id)
}

func (mw loggingMiddleware) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDocument, err error) {
	shared.Log(mw.logger, "SearchDocuments", err, "param:title", title, "param:tag", tag, "param:sender", sender, "param:from", from, "param:until", until, "param:limit", limit, "param:skip", skip)
	defer shared.Log(mw.logger, "SearchDocuments-End", err)
	return mw.next.SearchDocuments(title, tag, sender, from, until, limit, skip)
}

func (mw loggingMiddleware) SearchList(name string, st SearchType) (l []string, err error) {
	shared.Log(mw.logger, "SearchList", err, "param:name", name, "param:st", st)
	defer shared.Log(mw.logger, "SearchList-End", err)
	return mw.next.SearchList(name, st)
}

func (mw loggingMiddleware) SaveDocument(doc Document, user security.User) (d Document, err error) {
	shared.Log(mw.logger, "SaveDocument", err, "param:doc", doc, "param:user", user, "param.doc.filename", doc.FileName, "param.doc.uploadtoken", doc.UploadToken)
	defer shared.Log(mw.logger, "SaveDocument-Err", err)
	return mw.next.SaveDocument(doc, user)
}
