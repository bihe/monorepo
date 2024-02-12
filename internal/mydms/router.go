package mydms

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/mydms/app/config"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/web"
	conf "golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/security"

	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
)

// HTTPHandlerOptions are used to configure the http handler setup
type HTTPHandlerOptions struct {
	BasePath  string
	ErrorPath string
	Config    config.AppConfig
	Version   string
	Build     string
}

const forbiddenPath = "/mydms/403"

// MakeHTTPHandler creates a new handler implementation which is used together with the HTTP server
func MakeHTTPHandler(docSvc document.Service, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	std, sec := setupRouter(opts, logger)

	templateHandler := &web.TemplateHandler{
		TemplateHandler: &handler.TemplateHandler{
			Logger:    logger,
			Env:       opts.Config.Environment,
			BasePath:  "/public",
			StartPage: "/mydms",
		},
		DocSvc:  docSvc,
		Version: opts.Version,
		Build:   opts.Build,
	}

	// use this for development purposes only!
	if opts.Config.Environment == conf.Development {
		devTokenHandler := develop.DevTokenHandler{
			Logger: logger,
			Env:    opts.Config.Environment,
		}
		std.Get("/gettoken", devTokenHandler.Index())
	}

	// server-side rendered paths
	// the following paths provide server-rendered UIs
	// /403 displays a page telling the user that access/permissions are missing
	std.Get(forbiddenPath, templateHandler.Show403())

	std.Mount("/", sec)

	// the routes for the templates
	sec.Mount("/mydms", func() http.Handler {
		r := chi.NewRouter()
		r.Get("/", templateHandler.DisplayDocuments())
		r.Put("/partial", templateHandler.DisplayDocumentsPartial())
		r.Post("/dialog/{id}", templateHandler.ShowEditDocumentDialog())
		r.Get("/list/{type}", templateHandler.SearchListItems())
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
		ErrorRedirect: forbiddenPath,
	}

	secureRouter = chi.NewRouter()
	secureRouter.Use(interceptor.HandleJWT)

	return
}