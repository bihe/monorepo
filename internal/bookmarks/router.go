package bookmarks

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/api"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/app/conf"
	"golang.binggl.net/monorepo/internal/bookmarks/web"
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
	appInfoHandler := api.AppInfoHandler{
		Logger:  logger,
		Version: opts.Version,
		Build:   opts.Build,
	}
	bookmarksHandler := api.BookmarksHandler{
		App: app,
	}
	templateHandler := web.TemplateHandler{
		Logger:  logger,
		Env:     opts.Config.Environment,
		App:     app,
		Version: opts.Version,
		Build:   opts.Build,
	}

	std.Mount("/", sec)

	// server-side rendered paths
	// the following paths provide server-rendered UIs
	// /403 displays a page telling the user that acces/permissions are missing
	std.Get("/403", templateHandler.Show403())

	// use this for development purposes only!
	if opts.Config.Environment == config.Development {
		devTokenHandler := web.DevTokenHandler{
			Logger: logger,
			Env:    opts.Config.Environment,
		}
		std.Get("/gettoken", devTokenHandler.Index())
	}

	// /bookmarks/search performs a server-side search and renders the results
	// this supersedes the client/frontend-based search interaction
	sec.Get("/bookmarks/search", templateHandler.SearchBookmarks())

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
				pkgerr.WriteError(w, r, pkgerr.SecurityError{
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
