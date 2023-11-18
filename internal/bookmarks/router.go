package bookmarks

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/api"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/web"
	"golang.binggl.net/monorepo/pkg/config"
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
	appInfoHandler := api.AppInfoHandler{
		Logger:  logger,
		Version: opts.Version,
		Build:   opts.Build,
	}
	bookmarksHandler := api.BookmarksHandler{
		App: app,
	}
	templateHandler := &web.TemplateHandler{
		TemplateHandler: &handler.TemplateHandler{
			Logger:    logger,
			Env:       opts.Config.Environment,
			BasePath:  "/bm/public",
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
	if opts.Config.Environment == config.Development {
		devTokenHandler := develop.DevTokenHandler{
			Logger: logger,
			Env:    opts.Config.Environment,
		}
		std.Get("/gettoken", devTokenHandler.Index())
	}

	// /bm performs a server-side rendering
	// this implementation is to supersede the client/angular-based search interaction
	sec.Mount("/bm", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/search", templateHandler.SearchBookmarks())
		r.Get("/~*", templateHandler.GetBookmarksForPath())
		r.Get("/partial/~*", templateHandler.GetBookmarksForPathPartial())
		r.Get("/confirm/delete/{id}", templateHandler.DeleteConfirm())
		r.Delete("/delete/{id}", templateHandler.DeleteBookmark())
		r.Post("/favicon/page", templateHandler.FetchCustomFaviconFromPage())
		r.Post("/favicon/url", templateHandler.FetchCustomFaviconURL())
		r.Post("/sort", templateHandler.SortBookmarks())
		r.Get("/{id}", templateHandler.EditBookmarkDialog())
		r.Get("/", templateHandler.GetBookmarksForPath())
		r.Post("/", templateHandler.SaveBookmark())
		return r
	}())

	// the following APIs have the base-URL /api/v1
	sec.Get("/api/v1/whoami", appInfoHandler.HandleWhoAmI())
	// the bookmarks routes
	sec.Mount("/api/v1/bookmarks", func() http.Handler {
		r := chi.NewRouter()
		// information about the app itself
		r.Get("/appinfo", appInfoHandler.HandleGetAppInfo())
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
