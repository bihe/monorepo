package api_test

import (
	"net/http"

	"golang.binggl.net/monorepo/internal/core"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/upload"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
)

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

var logger = logging.NewNop()

type handlerOps struct {
	oidcSvc   oidc.Service
	siteSvc   sites.Service
	uploadSvc upload.Service
	version   string
	build     string
}

func handlerWith(ops *handlerOps) http.Handler {
	return core.MakeHTTPHandler(ops.oidcSvc, ops.siteSvc, ops.uploadSvc, logger, core.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		Config: conf.AppConfig{
			BaseConfig: config.BaseConfig{
				Cookies: config.ApplicationCookies{
					Prefix: "core",
				},
				Security: config.Security{},
				Logging:  config.LogConfig{},
				Cors:     config.CorsSettings{},
				Assets: config.AssetSettings{
					AssetDir:    "./",
					AssetPrefix: "/static",
				},
			},
			Security: conf.Security{
				CookieName: "jwt",
				Expiry:     10,
				JwtIssuer:  "issuer",
				JwtSecret:  "secret",
				Claim: config.Claim{
					Name:  "A",
					URL:   "http://A",
					Roles: []string{"A"},
				},
			},
			Upload: conf.UploadSettings{
				AllowedFileTypes: []string{"png"},
				MaxUploadSize:    100000,
				UploadPath:       "/tmp",
			},
		},
		Version: ops.version,
		Build:   ops.build,
	})
}
