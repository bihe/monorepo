package document

import (
	"fmt"
	"time"

	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// ServiceMiddleware describes a service (as opposed to endpoint) middleware.
// it is used to intercept the method execution and perform actions before/after the
// serivce method execution
type ServiceMiddleware func(Service) Service

// ServiceLoggingMiddleware takes a logger as a dependency
// and returns a ServiceLoggingMiddleware.
func ServiceLoggingMiddleware(logger logging.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

// compile guard for Service implementation
var (
	_ Service = &loggingMiddleware{}
)

type loggingMiddleware struct {
	logger logging.Logger
	next   Service
}

func (mw loggingMiddleware) GetDocumentByID(id string) (d Document, err error) {
	mw.logger.Info("GetDocumentByID", logging.LogV("param:ID", id))
	defer mw.logger.Info("called GetDocumentByID", logging.ErrV(err))
	return mw.next.GetDocumentByID(id)
}

func (mw loggingMiddleware) DeleteDocumentByID(id string) (err error) {
	mw.logger.Info("DeleteDocumentByID", logging.LogV("param:ID", id))
	defer mw.logger.Info("called DeleteDocumentByID", logging.ErrV(err))
	return mw.next.DeleteDocumentByID(id)
}

func (mw loggingMiddleware) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDocument, err error) {
	mw.logger.Info("SearchDocuments", logging.LogV("param:title", title),
		logging.LogV("param:tag", tag),
		logging.LogV("param:sender", sender),
		logging.LogV("param:from", from.String()),
		logging.LogV("param:until", until.String()),
		logging.LogV("param:limit", fmt.Sprintf("%d", limit)),
		logging.LogV("param:skip", fmt.Sprintf("%d", skip)),
	)
	defer mw.logger.Info("called SearchDocuments", logging.ErrV(err))
	return mw.next.SearchDocuments(title, tag, sender, from, until, limit, skip)
}

func (mw loggingMiddleware) SearchList(name string, st SearchType) (l []string, err error) {
	mw.logger.Info("DeleteDocumentByID", logging.LogV("param:name", name), logging.LogV("param:searchtype", st.String()))
	defer mw.logger.Info("called SearchList", logging.ErrV(err))
	return mw.next.SearchList(name, st)
}

func (mw loggingMiddleware) SaveDocument(doc Document, user security.User) (d Document, err error) {
	mw.logger.Info("SaveDocument", logging.LogV("param:doc", doc.String()),
		logging.LogV("param:user", user.String()),
		logging.LogV("param.doc.filename", doc.FileName),
		logging.LogV("param.doc.uploadtoken", doc.UploadToken),
	)
	defer mw.logger.Info("called SaveDocument", logging.ErrV(err))
	return mw.next.SaveDocument(doc, user)
}
