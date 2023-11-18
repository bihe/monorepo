package core

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/cookies"

	"golang.binggl.net/monorepo/internal/core/api"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/upload"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
)

// HTTPHandlerOptions are used to configure the http handler setup
type HTTPHandlerOptions struct {
	BasePath  string
	ErrorPath string
	Config    conf.AppConfig
	Version   string
	Build     string
}

// MakeHTTPHandler creates a new handler implementation which is used together with the HTTP server
func MakeHTTPHandler(oidcSvc oidc.Service, siteSvc sites.Service, uploadSvc upload.Service, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	std, sec := setupRouter(opts, logger)
	oidcHandler := api.OidcHandler{
		OidcSvc: oidcSvc,
		Logger:  logger,
		Cookies: cookies.NewAppCookie(cookies.Settings{
			Path:   opts.Config.Cookies.Path,
			Domain: opts.Config.Cookies.Domain,
			Secure: opts.Config.Cookies.Secure,
			Prefix: opts.Config.Cookies.Prefix,
		}),
		JwtCookieName: opts.Config.Security.CookieName,
		JwtExpiryDays: opts.Config.Security.Expiry,
	}
	sitesHandler := api.SitesHandler{
		SitesSvc: siteSvc,
		Logger:   logger,
	}

	uploadHandler := &api.UploadHandler{
		Service: uploadSvc,
		Logger:  logger,
	}

	appInfoHandler := api.AppInfoHandler{
		Logger:  logger,
		Version: opts.Version,
		Build:   opts.Build,
	}

	// the following APIs have the base-URL /api/v1
	std.Mount("/api/v1", sec)
	sec.Mount("/", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/appinfo", appInfoHandler.HandleGetAppInfo())
		r.Get("/whoami", appInfoHandler.HandleWhoAmI())
		return r
	}())
	sec.Mount("/sites", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", sitesHandler.HandleGetSitesForUser())
		r.Get("/users/{siteName}", sitesHandler.HandleGetUsersForSite())
		r.Post("/", sitesHandler.HandleSaveSitesForUser())
		return r
	}())
	sec.Mount("/upload", func() http.Handler {
		r := chi.NewRouter()
		r.Post("/file", uploadHandler.Upload())
		r.Get("/{id}", uploadHandler.GetItemByID())
		r.Delete("/{id}", uploadHandler.DeleteItemByID())
		return r
	}())

	// call on the ROOT path
	std.Get("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	std.Mount("/oidc", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/start", oidcHandler.HandlePrepIntOIDCRedirect())
		r.Get(oidc.OIDCInitiateRoutingPath, oidcHandler.HandleGetExtOIDCRedirect())
		r.Get("/signin", oidcHandler.HandleLoginOIDC(oidcHandler.JwtCookieName, oidcHandler.JwtExpiryDays))
		r.Get("/auth/flow", oidcHandler.HandleAuthFlow())
		return r
	}())

	return std
}

func setupRouter(opts HTTPHandlerOptions, logger logging.Logger) (router chi.Router, apiRouter chi.Router) {
	router = server.SetupBasicRouter(opts.BasePath, opts.Config.Cookies, opts.Config.Cors, opts.Config.Assets, logger)
	apiRouter = server.SetupSecureAPIRouter(opts.ErrorPath, config.Security{
		JwtIssuer:     opts.Config.Security.JwtIssuer,
		JwtSecret:     opts.Config.Security.JwtSecret,
		CookieName:    opts.Config.Security.CookieName,
		LoginRedirect: opts.Config.Security.LoginRedirect,
		Claim: config.Claim{
			Name:  opts.Config.Security.Claim.Name,
			URL:   opts.Config.Security.Claim.URL,
			Roles: opts.Config.Security.Claim.Roles,
		},
		CacheDuration: opts.Config.Security.CacheDuration,
	}, config.ApplicationCookies{
		Path:   opts.Config.Cookies.Path,
		Domain: opts.Config.Cookies.Domain,
		Secure: opts.Config.Cookies.Secure,
		Prefix: opts.Config.Cookies.Prefix,
	}, logger)
	return
}
