// Package server implements the API backend of the bookmark application
package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/config"
	"golang.binggl.net/monorepo/internal/bookmarks/server/api"
	"golang.binggl.net/monorepo/internal/bookmarks/store"

	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

// Server struct defines the basic layout of a HTTP API server
type Server struct {
	router         chi.Router
	basePath       string
	jwtOpts        security.JwtOptions
	cookieSettings cookies.Settings
	logConfig      config.LogConfig
	cors           config.CorsSettings
	environment    config.Environment
	logger         logging.Logger
	errorHandler   *handler.TemplateHandler
	appInfoAPI     *handler.AppInfoHandler
	healthCheck    *handler.HealthCheckHandler
	bookmarkAPI    *api.BookmarksAPI
}

type healthCheck struct {
	timeout        uint
	repository     store.Repository
	version, build string
}

func (h *healthCheck) Check(user security.User) (handler.HealthCheck, error) {
	var (
		status handler.Status
		msg    string
		err    error
	)

	status = handler.OK
	msg = "service is healthy"

	if err = h.repository.CheckStoreConnectivity(h.timeout); err != nil {
		status = handler.Error
		msg = err.Error()
	}

	return handler.HealthCheck{
		Status:    status,
		Message:   msg,
		TimeStamp: time.Now().UTC(),
		Version:   fmt.Sprintf("%s-%s", h.version, h.build),
	}, nil
}

// Create instantiates a new Server instance
func Create(basePath string, config config.AppConfig, version bookmarks.VersionInfo, logger logging.Logger) *Server {
	base, err := filepath.Abs(basePath)
	if err != nil {
		panic(fmt.Sprintf("cannot resolve basepath '%s', %v", basePath, err))
	}

	// setup repository
	// ------------------------------------------------------------------
	// persistence store && application version
	con, err := gorm.Open(sqlite.Open(config.Database.ConnectionString), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection: %v", err))
	}
	repository := store.Create(con, logger)

	// setup handlers for API
	// ------------------------------------------------------------------
	cookieSettings := cookies.Settings{
		Path:   config.Cookies.Path,
		Domain: config.Cookies.Domain,
		Secure: config.Cookies.Secure,
		Prefix: config.Cookies.Prefix,
	}
	baseHandler := handler.Handler{
		Log: logger,
	}

	// std. application information
	appInfo := &handler.AppInfoHandler{
		Handler: baseHandler,
		Version: version.Version,
		Build:   version.Build,
	}

	// a dynamic html template
	templatePath := filepath.Join(basePath, "templates")
	errHandler := &handler.TemplateHandler{
		Handler:        baseHandler,
		Version:        version.Version,
		Build:          version.Build,
		CookieSettings: cookieSettings,
		AppName:        config.AppName,
		TemplateDir:    templatePath,
	}

	// a /hc
	hc := &handler.HealthCheckHandler{
		Handler: baseHandler,
		Checker: &healthCheck{
			repository: repository,
			build:      version.Build,
			version:    version.Version,
			timeout:    30, // timeout of 30seconds
		},
	}

	// the API itself
	bookmarkAPI := &api.BookmarksAPI{
		Handler:        baseHandler,
		Repository:     repository,
		BasePath:       basePath,
		FaviconPath:    config.FaviconUploadPath,
		DefaultFavicon: config.DefaultFavicon,
	}

	// server combines setting and handlers to form the backend
	jwtOptions := security.JwtOptions{
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
		ErrorPath:     config.ErrorPath,
	}

	// finally all ingredients for the server-struct
	srv := Server{
		basePath:       base,
		jwtOpts:        jwtOptions,
		cookieSettings: cookieSettings,
		logConfig:      config.Logging,
		cors:           config.Cors,
		environment:    config.Environment,
		appInfoAPI:     appInfo,
		healthCheck:    hc,
		errorHandler:   errHandler,
		bookmarkAPI:    bookmarkAPI,
		logger:         logger,
	}
	srv.routes()
	return &srv
}

// ServeHTTP turns the server into a http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
