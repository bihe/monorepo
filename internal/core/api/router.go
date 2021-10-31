package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/cookies"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/security"

	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/shared"
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
	oidcHandler := OidcHandler{
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
	sitesHandler := SitesHandler{
		SitesSvc: siteSvc,
		Logger:   logger,
	}

	uploadHandler := &UploadHandler{
		Service: uploadSvc,
		Logger:  logger,
	}

	appInfoHandler := AppInfoHandler{
		Logger:  logger,
		Version: opts.Version,
		Build:   opts.Build,
	}

	// the following APIs have the base-URL /api/v1
	std.Mount("/api/v1", sec)
	sec.Mount("/", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/appinfo", appInfoHandler.handleGetAppInfo())
		r.Get("/whoami", appInfoHandler.handleWhoAmI())
		return r
	}())
	sec.Mount("/sites", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", sitesHandler.handleGetSitesForUser())
		r.Get("/users/{siteName}", sitesHandler.handleGetUsersForSite())
		r.Post("/", sitesHandler.handleSaveSitesForUser())
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
		r.Get("/start", oidcHandler.handlePrepIntOIDCRedirect())
		r.Get(oidc.OIDCInitiateRoutingPath, oidcHandler.handleGetExtOIDCRedirect())
		r.Get("/signin", oidcHandler.handleLoginOIDC(oidcHandler.JwtCookieName, oidcHandler.JwtExpiryDays))
		r.Get("/auth/flow", oidcHandler.handleAuthFlow())
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

// --------------------------------------------------------------------------
//   Internal helper functions
// --------------------------------------------------------------------------

func encodeError(err error, logger logging.Logger, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}

	t := "about:blank"
	var (
		pd            *pkgerr.ProblemDetail
		errNotFound   *shared.NotFoundError
		errValidation *shared.ValidationError
		errSecurity   *shared.SecurityError
	)
	if errors.As(err, &errNotFound) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "object cannot be found",
			Status: http.StatusNotFound,
			Detail: errNotFound.Error(),
		}
	} else if errors.As(err, &errValidation) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "error in parameter-validaton",
			Status: http.StatusBadRequest,
			Detail: errValidation.Error(),
		}
	} else if errors.As(err, &errSecurity) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "security error",
			Status: http.StatusForbidden,
			Detail: errSecurity.Error(),
		}
	} else {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "cannot service the request",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
	}
	writeProblemJSON(logger, w, pd)
}

func respondJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func writeProblemJSON(logger logging.Logger, w http.ResponseWriter, pd *pkgerr.ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(pd.Status)
	b, err := json.Marshal(pd)
	if err != nil {
		logger.Error(fmt.Sprintf("could not marshal json %v", err))
	}
	_, err = w.Write(b)
	if err != nil {
		logger.Error(fmt.Sprintf("could not write bytes using http.ResponseWriter: %v", err))
	}
}

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the sucurity context user is not available!")
	}
	return user
}
