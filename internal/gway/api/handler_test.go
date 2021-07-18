package api_test

import (
	"net/http"

	"golang.binggl.net/monorepo/internal/gway/api"
	"golang.binggl.net/monorepo/internal/gway/app/conf"
	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/internal/gway/app/sites"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
)

// --------------------------------------------------------------------------
//   Boilerplate to make the tests happen
// --------------------------------------------------------------------------

var logger = logging.NewNop()

type handlerOps struct {
	oidcSvc oidc.Service
	siteSvc sites.Service
}

func handlerWith(ops *handlerOps) http.Handler {
	return api.MakeHTTPHandler(ops.oidcSvc, ops.siteSvc, logger, api.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		Config: conf.AppConfig{
			BaseConfig: config.BaseConfig{
				Cookies: config.ApplicationCookies{
					Prefix: "gway",
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
		},
	})
}
