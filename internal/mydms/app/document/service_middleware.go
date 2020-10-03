package document

import "github.com/go-kit/kit/log"

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

// GetDocumentByID implements the Service so that it can be used transparently
// the provided parameters are logged and the execution is passed on to the underlying/next service
func (mw loggingMiddleware) GetDocumentByID(id string) (d Document, err error) {
	defer func() {
		mw.logger.Log("method", "GetDocumentByID", "err", err, "id", id)
	}()
	// pass on to the real service
	return mw.next.GetDocumentByID(id)
}
