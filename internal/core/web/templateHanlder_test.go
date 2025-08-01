package web_test

import (
	"net/http"

	"golang.binggl.net/monorepo/internal/core"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
)

var logger = logging.NewNop()

// for testing the templateHanlder we mock the OIDCService

type mockOIDCService struct{}

func (*mockOIDCService) PrepIntOIDCRedirect() (url string, state string) {
	return "", ""
}

func (*mockOIDCService) GetExtOIDCRedirect(state string) (url string, err error) {
	return "", nil
}

func (*mockOIDCService) LoginOIDC(state, oidcState, oidcCode string) (token, url string, err error) {
	return "", "", nil
}

func (*mockOIDCService) LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error) {
	return "", "", nil
}

// compile guard
var _ oidc.Service = &mockOIDCService{}

func templateHandler(siteSvc sites.Service) http.Handler {

	return core.MakeHTTPHandler(&mockOIDCService{}, siteSvc, logger, core.HTTPHandlerOptions{
		BasePath:  "./",
		ErrorPath: "/error",
		Version:   "1.0",
		Build:     "today",
		Config: conf.AppConfig{
			BaseConfig: config.BaseConfig{
				Cookies: config.ApplicationCookies{
					Prefix: "core",
				},
				Security: config.Security{
					CookieName: "jwt",
					JwtIssuer:  "issuer",
					JwtSecret:  "secret",
					Claim: config.Claim{
						Name:  "A",
						URL:   "http://A",
						Roles: []string{"A"},
					},
				},
				Logging: config.LogConfig{},
				Cors:    config.CorsSettings{},
				Assets: config.AssetSettings{
					AssetDir:    "./",
					AssetPrefix: "/static",
				},
			},
			Security: conf.Security{
				CookieName: "jwt",
				JwtIssuer:  "issuer",
				JwtSecret:  "secret",
				Expiry:     1000,
				Claim: config.Claim{
					Name:  "A",
					URL:   "http://A",
					Roles: []string{"A"},
				},
			},
		},
	})
}
