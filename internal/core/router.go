package core

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/security"

	"golang.binggl.net/monorepo/internal/common/crypter"
	"golang.binggl.net/monorepo/internal/core/api"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/web"
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
func MakeHTTPHandler(oidcSvc oidc.Service, siteSvc sites.Service, cryptSvc crypter.EncryptionService, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
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

	templateHandler := &web.TemplateHandler{
		TemplateHandler: &handler.TemplateHandler{
			Logger:    logger,
			Env:       opts.Config.Environment,
			BasePath:  "/public",
			StartPage: "/sites",
		},
		SiteSvc:    siteSvc,
		CrypterSvc: cryptSvc,
		Version:    opts.Version,
		Build:      opts.Build,
	}

	// use this for development purposes only!
	develop.SetupDevTokenHandler(std, logger, opts.Config.Environment)

	// server-side rendered paths
	// the following paths provide server-rendered UIs
	// /403 displays a page telling the user that access/permissions are missing
	std.Get("/core/403", templateHandler.Show403())

	// the OIDC logic and handshake to authenticate with external auth system
	std.Mount("/oidc", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/start", oidcHandler.HandlePrepIntOIDCRedirect())
		r.Get(oidc.OIDCInitiateRoutingPath, oidcHandler.HandleGetExtOIDCRedirect())
		r.Get("/signin", oidcHandler.HandleLoginOIDC(oidcHandler.JwtCookieName, oidcHandler.JwtExpiryDays))
		r.Get("/auth/flow", oidcHandler.HandleAuthFlow())
		return r
	}())

	std.Mount("/", sec)

	// mount the server-side rendering paths
	sec.Mount("/sites", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", templateHandler.DisplaySites())
		r.Get("/edit", templateHandler.ShowEditSites())
		r.Post("/", templateHandler.SaveSites())
		return r
	}())

	// add the handlers for additional paths
	sec.Mount("/crypter", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", templateHandler.DisplayAgeStartPage())
		r.Get("/search", templateHandler.DisplayAgeStartPage())
		r.Post("/", templateHandler.PerformAgeAction())
		r.Put("/toast", templateHandler.DisplayToastNotification())
		return r
	}())

	// call on the ROOT path
	std.Get("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	std.NotFound(templateHandler.Show404())

	return std
}

func setupRouter(opts HTTPHandlerOptions, logger logging.Logger) (router chi.Router, secureRouter chi.Router) {
	router = server.SetupBasicRouter(opts.BasePath, opts.Config.Cookies, opts.Config.Cors, opts.Config.Assets, logger)

	// add a middleware to "catch" security errors and present a human-readable form
	// if the client requests "application/json" just use the the problem-json format
	jwtOptions := security.JwtOptions{
		CacheDuration: opts.Config.Security.CacheDuration,
		CookieName:    opts.Config.Security.CookieName,
		ErrorPath:     opts.ErrorPath,
		JwtIssuer:     opts.Config.Security.JwtIssuer,
		JwtSecret:     opts.Config.Security.JwtSecret,
		RedirectURL:   opts.Config.Security.LoginRedirect,
		RequiredClaim: security.Claim{
			Name:  opts.Config.Security.Claim.Name,
			URL:   opts.Config.Security.Claim.URL,
			Roles: opts.Config.Security.Claim.Roles,
		},
	}
	jwtAuth := security.NewJWTAuthorization(jwtOptions, true)
	interceptor := security.SecInterceptor{
		Log:           logger,
		Auth:          jwtAuth,
		Options:       jwtOptions,
		ErrorRedirect: "/core/403",
	}

	secureRouter = chi.NewRouter()
	secureRouter.Use(interceptor.HandleJWT)

	return
}
