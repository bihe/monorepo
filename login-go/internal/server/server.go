// Package server defines the HTTP server and performs the setup for the API
package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql" // import the mysql driver

	"github.com/bihe/login-go/internal"
	"github.com/bihe/login-go/internal/config"
	"github.com/bihe/login-go/internal/persistence"
	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"

	"github.com/bihe/login-go/internal/api"

	log "github.com/sirupsen/logrus"
	per "golang.binggl.net/commons/persistence"
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
	log            *log.Entry
}

// Create instantiates a new Server instance
func Create(basePath string, config config.AppConfig, version internal.VersionInfo, logEntry *log.Entry) *Server {
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
		Log:    logEntry,
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
		api:         api.New(base, baseHandler, cookieSettings, version, config.OIDC, config.Security, repo, logEntry),
		appInfoAPI:  appInfo,
		log:         logEntry,
	}
	srv.routes()
	return &srv
}

// ServeHTTP turns the server into a http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
