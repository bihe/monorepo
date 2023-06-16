package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/app"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/pkg/config"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"
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
func MakeHTTPHandler(app *bookmarks.Application, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	std, sec := setupRouter(opts, logger)
	appInfoHandler := AppInfoHandler{
		Logger:  logger,
		Version: opts.Version,
		Build:   opts.Build,
	}
	bookmarksHandler := BookmarksHandler{
		App: app,
	}

	// the following APIs have the base-URL /api/v1
	std.Mount("/api/v1", sec)

	// base routes
	sec.Mount("/", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/whoami", appInfoHandler.handleWhoAmI())
		return r
	}())

	// the bookmarks routes
	sec.Mount("/bookmarks", func() http.Handler {
		r := chi.NewRouter()
		// information about the app itself
		r.Get("/appinfo", appInfoHandler.handleGetAppInfo())
		// the real bookmark paths
		r.Put("/sortorder", bookmarksHandler.UpdateSortOrder())
		r.Delete("/{id}", bookmarksHandler.Delete())
		r.Get("/{id}", bookmarksHandler.GetBookmarkByID())
		r.Get("/bypath", bookmarksHandler.GetBookmarksByPath())
		r.Get("/allpaths", bookmarksHandler.GetAllPaths())
		r.Get("/folder", bookmarksHandler.GetBookmarksFolderByPath())
		r.Get("/byname", bookmarksHandler.GetBookmarksByName())
		r.Get("/fetch/{id}", bookmarksHandler.FetchAndForward())
		r.Get("/favicon/{id}", bookmarksHandler.GetFavicon())
		r.Get("/favicon/temp/{id}", bookmarksHandler.GetTempFavicon())
		r.Post("/favicon/temp/base", bookmarksHandler.CreateBaseURLFavicon())
		r.Post("/favicon/temp/url", bookmarksHandler.CreateFaviconFromCustomURL())

		r.Post("/", bookmarksHandler.Create())
		r.Put("/", bookmarksHandler.Update())
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

func encodeError(err error, logger logging.Logger, w http.ResponseWriter, r *http.Request) {
	if err == nil {
		panic("encodeError with nil error")
	}

	logger.ErrorRequest(fmt.Sprintf("got error during request: %v", err), r)

	t := "about:blank"
	var (
		pd            *pkgerr.ProblemDetail
		errNotFound   *app.NotFoundError
		errValidation *app.ValidationError
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
