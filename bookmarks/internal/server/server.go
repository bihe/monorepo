// Package server implements the API backend of the bookmark application
package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/jinzhu/gorm"
	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"

	"github.com/bihe/bookmarks/internal"
	"github.com/bihe/bookmarks/internal/config"
	"github.com/bihe/bookmarks/internal/server/api"
	"github.com/bihe/bookmarks/internal/store"
	"github.com/go-chi/chi"

	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/mysql" // use mysql
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
	errorHandler   *handler.TemplateHandler
	appInfoAPI     *handler.AppInfoHandler
	bookmarkAPI    *api.BookmarksAPI
	log            *log.Entry
}

// Create instantiates a new Server instance
func Create(basePath string, config config.AppConfig, version internal.VersionInfo, logger *log.Entry) *Server {
	base, err := filepath.Abs(basePath)
	if err != nil {
		panic(fmt.Sprintf("cannot resolve basepath '%s', %v", basePath, err))
	}

	// setup repository
	// ------------------------------------------------------------------
	con, err := gorm.Open(config.Database.Dialect, config.Database.ConnectionString)
	if err != nil {
		panic(fmt.Sprintf("cannot create database connection: %v", err))
	}
	repository := store.Create(con)

	// setup handlers for API
	// ------------------------------------------------------------------
	cookieSettings := cookies.Settings{
		Path:   config.Cookies.Path,
		Domain: config.Cookies.Domain,
		Secure: config.Cookies.Secure,
		Prefix: config.Cookies.Prefix,
	}
	errorReporter := &errors.ErrorReporter{
		CookieSettings: cookieSettings,
		ErrorPath:      config.ErrorPath,
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

	templatePath := filepath.Join(basePath, "templates")
	errHandler := &handler.TemplateHandler{
		Handler:        baseHandler,
		Version:        version.Version,
		Build:          version.Build,
		CookieSettings: cookieSettings,
		AppName:        config.AppName,
		TemplateDir:    templatePath,
	}
	bookmarkAPI := &api.BookmarksAPI{
		Handler:        baseHandler,
		Repository:     repository,
		BasePath:       basePath,
		FaviconPath:    config.FaviconUploadPath,
		DefaultFavicon: config.DefaultFavicon,
	}

	// server combines setting and handlers to form the backend
	// ------------------------------------------------------------------

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

	srv := Server{
		basePath:       base,
		jwtOpts:        jwtOptions,
		cookieSettings: cookieSettings,
		logConfig:      config.Logging,
		cors:           config.Cors,
		environment:    config.Environment,
		appInfoAPI:     appInfo,
		errorHandler:   errHandler,
		bookmarkAPI:    bookmarkAPI,
		log:            logger,
	}
	srv.routes()
	return &srv
}

// ServeHTTP turns the server into a http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
