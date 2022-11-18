package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/security"

	"golang.binggl.net/monorepo/internal/bookmarks-new/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks-new/app/store"
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
func MakeHTTPHandler(store store.Repository, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	std, sec := setupRouter(opts, logger)
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
	// sec.Mount("/sites", func() http.Handler {
	// 	r := chi.NewRouter()
	// 	r.Get("/", sitesHandler.handleGetSitesForUser())
	// 	r.Get("/users/{siteName}", sitesHandler.handleGetUsersForSite())
	// 	r.Post("/", sitesHandler.handleSaveSitesForUser())
	// 	return r
	// }())
	// sec.Mount("/upload", func() http.Handler {
	// 	r := chi.NewRouter()
	// 	r.Post("/file", uploadHandler.Upload())
	// 	r.Get("/{id}", uploadHandler.GetItemByID())
	// 	r.Delete("/{id}", uploadHandler.DeleteItemByID())
	// 	return r
	// }())

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
