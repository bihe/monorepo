package appinfo

import (
	"github.com/go-kit/kit/log"

	"golang.binggl.net/monorepo/pkg/security"
)

// --------------------------------------------------------------------------
// Service Definition
// --------------------------------------------------------------------------

// Service defines the appinfo methods
type Service interface {
	// GetAppInfo returns meta-data of the application and the currently loged-in user
	GetAppInfo(user *security.User) (ai AppInfo, err error)
}

// NewService returns a Service with all of the expected middlewares wired in.
func NewService(logger log.Logger, version, build string) Service {
	var svc Service
	{
		svc = &appInfoService{
			version: version,
			build:   build,
		}
		svc = ServiceLoggingMiddleware(logger)(svc)
	}
	return svc
}

// --------------------------------------------------------------------------
// Service Middleware
// --------------------------------------------------------------------------

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
	defer func() {
		mw.logger.Log("method", "GetAppInfo", "err", err)
	}()
	// pass on to the real service
	return mw.next.GetAppInfo(user)
}

// --------------------------------------------------------------------------
// Service implementation
// --------------------------------------------------------------------------

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ Service = &appInfoService{}
)

type appInfoService struct {
	version string
	build   string
}

func (a *appInfoService) GetAppInfo(user *security.User) (ai AppInfo, err error) {
	meta := AppInfo{
		UserInfo: UserInfo{
			DisplayName: user.DisplayName,
			UserID:      user.UserID,
			UserName:    user.Username,
			Email:       user.Email,
			Roles:       user.Roles,
		},
		VersionInfo: VersionInfo{
			Version:     a.version,
			BuildNumber: a.build,
		},
	}
	return meta, nil
}
