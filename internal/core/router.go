package core

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/core/api"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/upload"
	"golang.binggl.net/monorepo/internal/core/web"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
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

	templateHandler := web.CreateTemplateHandler(logger, opts.Config.Environment, opts.Config.Service.SiteApiURL)

	// the page to show that no access is possible
	std.Get("/403", templateHandler.Show403())

	// use this for development purposes only!
	if opts.Config.Environment == config.Development {
		devTokenHandler := web.DevTokenHandler{
			Logger: logger,
			Env:    opts.Config.Environment,
		}
		std.Get("/gettoken", devTokenHandler.Index())
	}

	std.Mount("/oidc", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/start", oidcHandler.HandlePrepIntOIDCRedirect())
		r.Get(oidc.OIDCInitiateRoutingPath, oidcHandler.HandleGetExtOIDCRedirect())
		r.Get("/signin", oidcHandler.HandleLoginOIDC(oidcHandler.JwtCookieName, oidcHandler.JwtExpiryDays))
		r.Get("/auth/flow", oidcHandler.HandleAuthFlow())
		return r
	}())

	// the frontend part is located under /ui
	sec.Mount("/ui", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", templateHandler.Index())

		// sites
		r.Get("/sites", templateHandler.GetSites())
		return r
	}())

	// the following APIs have the base-URL /api/v1
	sec.Mount("/api/v1", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/appinfo", appInfoHandler.HandleGetAppInfo())
		r.Get("/whoami", appInfoHandler.HandleWhoAmI())
		return r
	}())
	sec.Mount("/api/v1/sites", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", sitesHandler.HandleGetSitesForUser())
		r.Get("/users/{siteName}", sitesHandler.HandleGetUsersForSite())
		r.Post("/", sitesHandler.HandleSaveSitesForUser())
		return r
	}())
	sec.Mount("/api/v1/upload", func() http.Handler {
		r := chi.NewRouter()
		r.Post("/file", uploadHandler.Upload())
		r.Get("/{id}", uploadHandler.GetItemByID())
		r.Delete("/{id}", uploadHandler.DeleteItemByID())
		return r
	}())

	std.Mount("/", sec)

	return std
}

func setupRouter(opts HTTPHandlerOptions, logger logging.Logger) (router chi.Router, secureRouter chi.Router) {
	router = server.SetupBasicRouter(opts.BasePath,
		opts.Config.Cookies,
		opts.Config.Cors,
		opts.Config.Assets,
		logger)

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
	jwtSecurity := security.NewJwtMiddleware(jwtOptions, logger)
	jwtAuth := security.NewJWTAuthorization(jwtOptions, true)

	interceptor := secInterceptor{
		log:     logger,
		jwt:     jwtSecurity,
		auth:    jwtAuth,
		options: jwtOptions,
	}

	secureRouter = chi.NewRouter()
	secureRouter.Use(interceptor.handleJWT)

	return
}

// secInterceptor is used the redirect to an error page if we can assume that no api-client
// requests the URL
type secInterceptor struct {
	log     logging.Logger
	jwt     *security.JwtMiddleware
	auth    *security.JWTAuthorization
	options security.JwtOptions
}

func (u secInterceptor) handleJWT(next http.Handler) http.Handler {
	jwtAuth := u.auth
	options := u.options
	logger := u.log
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := security.Validate(r, jwtAuth, options, logger)
		if err != nil {
			u.log.Debug("custom JWT validation in secInterceptor!")
			acc := r.Header.Get("Accept")
			if acc == "application/json" {
				u.log.Debug("the client requests JSON, provide json error-message")
				errors.WriteError(w, r, errors.SecurityError{
					Err:     fmt.Errorf("security error because of missing JWT token"),
					Request: r,
					Status:  http.StatusUnauthorized,
				})
				return
			}
			// when the Accept header is not pure 'application/json' provide a human-readable response
			http.Redirect(w, r, "/403", http.StatusFound)
			return

		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
