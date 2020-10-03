package appinfo

import (
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

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// GetAppInfo implements the Service so that it can be used transparently
// the provided parameters are logged and the execution is passed on to the underlying/next service
func (mw loggingMiddleware) GetAppInfo(user *security.User) (ai AppInfo, err error) {
	defer shared.Log(mw.logger, "GetAppInfo", err)
	return mw.next.GetAppInfo(user)
}
