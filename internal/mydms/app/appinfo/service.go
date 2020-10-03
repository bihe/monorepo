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
