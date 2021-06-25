package gway

import (
	"net/http"

	"golang.binggl.net/monorepo/pkg/cookies"

	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
)

const StateCookie = "state"

// HTTPHandlerOptions are used to configure the handler-logic of the HTTP server
type HTTPHandlerOptions struct {
	BasePath     string
	ErrorPath    string
	JWTConfig    config.Security
	CookieConfig config.ApplicationCookies
	CorsConfig   config.CorsSettings
	AssetConfig  config.AssetSettings
}

// MakeHTTPHandler creates a new handler implementation which is used together with the HTTP server
func MakeHTTPHandler(oidcSvc oidc.Service, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	cookie := cookies.NewAppCookie(cookies.Settings{
		Path:   opts.CookieConfig.Path,
		Domain: opts.CookieConfig.Domain,
		Secure: opts.CookieConfig.Secure,
		Prefix: opts.CookieConfig.Prefix,
	})

	r := server.SetupBasicRouter(opts.BasePath, opts.CookieConfig, opts.CorsConfig, opts.AssetConfig, logger)
	//apiRouter := server.SetupSecureAPIRouter(opts.ErrorPath, opts.JWTConfig, opts.CookieConfig, logger)

	r.Get("/start-oidc", handlePrepIntOIDCRedirect(oidcSvc, logger, cookie))
	r.Get(oidc.OIDCInitiateURL, handleGetExtOIDCRedirect(oidcSvc, logger, cookie))

	return r
}
