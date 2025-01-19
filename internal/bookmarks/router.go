package bookmarks

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/web"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/handler"
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

	templateHandler := &web.TemplateHandler{
		TemplateHandler: &handler.TemplateHandler{
			Logger:    logger,
			Env:       opts.Config.Environment,
			BasePath:  "/public",
			StartPage: "/bm",
		},
		App:     app,
		Version: opts.Version,
		Build:   opts.Build,
	}

	std.Mount("/", sec)

	// server-side rendered paths
	// the following paths provide server-rendered UIs
	// /403 displays a page telling the user that access/permissions are missing
	std.Get("/bm/403", templateHandler.Show403())

	// use this for development purposes only!
	develop.SetupDevTokenHandler(std, logger, opts.Config.Environment)

	// /bm performs a server-side rendering
	// this implementation is to supersede the client/angular-based search interaction
	sec.Mount("/bm", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/search", templateHandler.SearchBookmarks())
		r.Get("/partial/search", templateHandler.SearchBookmarksPartial())
		r.Get("/~*", templateHandler.GetBookmarksForPath())
		r.Get("/partial/~*", templateHandler.GetBookmarksForPathPartial())
		r.Get("/confirm/delete/{id}", templateHandler.DeleteConfirm())
		r.Delete("/delete/{id}", templateHandler.DeleteBookmark())
		r.Get("/favicon/edit", templateHandler.AvailableFaviconsDialog())
		r.Post("/favicon/page", templateHandler.FetchCustomFaviconFromPage())
		r.Post("/favicon/url", templateHandler.FetchCustomFaviconURL())
		r.Get("/favicon/{id}", templateHandler.GetFaviconByBookmarkID())
		r.Get("/favicon/raw/{id}", templateHandler.GetFaviconByID())
		r.Get("/favicon/temp/{id}", templateHandler.GetTempFaviconByID())
		r.Post("/favicon/upload", templateHandler.UploadCustomFavicon())
		r.Post("/sort", templateHandler.SortBookmarks())
		r.Get("/{id}", templateHandler.EditBookmarkDialog())
		r.Get("/", templateHandler.GetBookmarksForPath())
		r.Post("/", templateHandler.SaveBookmark())
		return r
	}())

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
		ErrorRedirect: "/bm/403",
	}
	secureRouter = chi.NewRouter()
	secureRouter.Use(interceptor.HandleJWT)

	return
}
