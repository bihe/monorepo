// Package server defines the HTTP server and performs the setup for the API
package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql" // import the mysql driver

	"golang.binggl.net/monorepo/internal/login"
	"golang.binggl.net/monorepo/internal/login/config"
	"golang.binggl.net/monorepo/internal/login/persistence"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"

	"golang.binggl.net/monorepo/internal/login/api"

	per "golang.binggl.net/monorepo/pkg/persistence"
)

// --------------------------------------------------------------------------
// internal logic to setup the server
// --------------------------------------------------------------------------

// Server struct defines the basic layout of a HTTP API server
type Server struct {
	basePath       string
	cookieSettings cookies.Settings
	jwtOpts        security.JwtOptions
	router         chi.Router
	api            api.Login
	appInfoAPI     *handler.AppInfoHandler
	environment    config.Environment
	logConfig      config.LogConfig
	cors           config.CorsSettings
	log            logging.Logger
}

// Create instantiates a new Server instance
func Create(basePath string, config config.AppConfig, version login.VersionInfo, logger logging.Logger) *Server {
	base, err := filepath.Abs(basePath)
	if err != nil {
		panic(fmt.Sprintf("cannot resolve basepath '%s', %v", basePath, err))
	}
	con := per.NewConn(config.Database.ConnectionString)
	repo, err := persistence.NewRepository(con)
	if err != nil {
		panic(fmt.Sprintf("could not create a repository: %v", err))
	}

	jwtOpts := security.JwtOptions{
		JwtSecret:  config.Security.JwtSecret,
		JwtIssuer:  config.Security.JwtIssuer,
		CookieName: config.Security.CookieName,
		RequiredClaim: security.Claim{
			Name:  config.Security.Claim.Name,
			URL:   config.Security.Claim.URL,
			Roles: config.Security.Claim.Roles,
		},
		RedirectURL:   config.Security.LoginRedirect,
		CacheDuration: config.Security.CacheDuration,
	}
	cookieSettings := cookies.Settings{
		Path:   config.AppCookies.Path,
		Domain: config.AppCookies.Domain,
		Secure: config.AppCookies.Secure,
		Prefix: config.AppCookies.Prefix,
	}
	errorReporter := &errors.ErrorReporter{
		CookieSettings: cookieSettings,
		ErrorPath:      "/error",
	}
	baseHandler := handler.Handler{
		ErrRep: errorReporter,
		Log:    logger,
	}

	appInfo := &handler.AppInfoHandler{
		Handler: baseHandler,
		Version: version.Version,
		Build:   version.Build,
	}

	srv := Server{
		basePath:       base,
		cookieSettings: cookieSettings,
		jwtOpts:        jwtOpts,

		environment: config.Environment,
		logConfig:   config.Logging,
		cors:        config.Cors,
		api:         api.New(base, baseHandler, cookieSettings, version, config.OIDC, config.Security, repo, logger),
		appInfoAPI:  appInfo,
		log:         logger,
	}
	srv.routes()
	return &srv
}

// ServeHTTP turns the server into a http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
