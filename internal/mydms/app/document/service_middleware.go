package document

import (
	"time"

	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
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
	defer shared.Log(mw.logger, "GetDocumentByID", err, "param:ID", id)
	return mw.next.GetDocumentByID(id)
}

func (mw loggingMiddleware) DeleteDocumentByID(id string) (err error) {
	defer shared.Log(mw.logger, "DeleteDocumentByID", err, "param:ID", id)
	return mw.next.DeleteDocumentByID(id)
}

func (mw loggingMiddleware) SearchDocuments(title, tag, sender string, from, until time.Time, limit, skip int) (p PagedDcoument, err error) {
	defer shared.Log(mw.logger, "SearchDocuments", err, "param:title", title, "param:tag", tag, "param:sender", sender, "param:from", from, "param:until", until, "param:limit", limit, "param:skip", skip)
	return mw.next.SearchDocuments(title, tag, sender, from, until, limit, skip)
}
