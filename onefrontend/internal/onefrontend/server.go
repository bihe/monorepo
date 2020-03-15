// Package onefrontend implements a web-server for the unified frontend components
package onefrontend

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bihe/onefrontend/internal/onefrontend/config"
	"github.com/bihe/onefrontend/internal/onefrontend/types"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"
)

// Server configures a HTTP server
type Server struct {
	Cookies        config.CookieSettings
	LogConfig      config.LogConfig
	JWTSecurity    config.JwtSettings
	Cors           config.CorsSettings
	BasePath       string
	AssetDir       string
	AssetPrefix    string
	FrontendDir    string
	FrontendPrefix string
	ErrorPath      string
	StartURL       string
	Environment    string
	Version        types.VersionInfo

	router chi.Router
}

// MapRoutes defines and configures the application routes
func (s *Server) MapRoutes() {
	// basic handler building blocks
	cookieSettings := cookies.Settings{
		Path:   s.Cookies.Path,
		Domain: s.Cookies.Domain,
		Secure: s.Cookies.Secure,
		Prefix: "one",
	}
	errorReporter := &errors.ErrorReporter{
		CookieSettings: cookieSettings,
		ErrorPath:      s.ErrorPath,
	}
	baseHandler := handler.Handler{
		ErrRep: errorReporter,
	}

	// handler responsible returning application information
	appInfo := &handler.AppInfoHandler{
		Handler: baseHandler,
		Version: s.Version.Version,
		Build:   s.Version.Build,
	}
	// if we need to display errors in HTML
	templatePath := filepath.Join(s.BasePath, "templates")
	errHandler := &handler.TemplateHandler{
		Handler:        baseHandler,
		Version:        s.Version.Version,
		Build:          s.Version.Build,
		CookieSettings: cookieSettings,
		TemplateDir:    templatePath,
		AppName:        "one.binggl.net",
	}

	// JWT authentication settings
	jwtOptions := security.JwtOptions{
		JwtSecret:  s.JWTSecurity.JwtSecret,
		JwtIssuer:  s.JWTSecurity.JwtIssuer,
		CookieName: s.JWTSecurity.CookieName,
		RequiredClaim: security.Claim{
			Name:  s.JWTSecurity.Claim.Name,
			URL:   s.JWTSecurity.Claim.URL,
			Roles: s.JWTSecurity.Claim.Roles,
		},
		RedirectURL:   s.JWTSecurity.LoginRedirect,
		CacheDuration: s.JWTSecurity.CacheDuration,
		ErrorPath:     s.ErrorPath,
	}

	// the routing
	// ------------------------------------------------------------------
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// setup cors for single frontend
	cors := cors.New(cors.Options{
		AllowedOrigins:   s.Cors.AllowedOrigins,
		AllowedMethods:   s.Cors.AllowedMethods,
		AllowedHeaders:   s.Cors.AllowedHeaders,
		AllowCredentials: s.Cors.AllowCredentials,
		MaxAge:           s.Cors.MaxAge,
	})
	r.Use(cors.Handler)

	s.setupRequestLogging()

	r.Get("/error", errHandler.Call(errHandler.HandleError))

	// serving static content
	handler.ServeStaticFile(r, "/favicon.ico", filepath.Join(s.BasePath, s.AssetDir, "favicon.ico"))
	handler.ServeStaticDir(r, s.AssetPrefix, http.Dir(filepath.Join(s.BasePath, s.AssetDir)))

	// this group "indicates" that all routes within this group use the JWT authentication
	r.Group(func(r chi.Router) {
		// authenticate and authorize users via JWT
		r.Use(security.NewJwtMiddleware(jwtOptions, cookieSettings).JwtContext)

		r.Get("/appinfo", appInfo.Secure(appInfo.HandleAppInfo))

		// the SPA
		handler.ServeStaticDir(r, s.FrontendPrefix, http.Dir(filepath.Join(s.BasePath, s.FrontendDir)))
	})

	// either redirect to the frontend root or use a supplied ref-path
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// determine if well-known parameters are supplied
		valid := r.URL.Query().Get("valid")
		ref := r.URL.Query().Get("ref")
		if valid == "" || ref == "" {
			// perform a "normal" redirect to the frontend
			http.Redirect(w, r, s.FrontendPrefix, http.StatusFound)
			return
		}

		// the ref path needs to be known, valid know paths are
		// /ui, /ui/bookmarks/ ...
		if valid == "true" && (strings.HasPrefix(ref, "/ui/bookmarks") || strings.HasPrefix(ref, "/ui")) {
			http.Redirect(w, r, ref, http.StatusFound)
			return
		}

		// default to the frontend
		http.Redirect(w, r, s.FrontendPrefix, http.StatusFound)
	})

	s.router = r
}

// ServeHTTP turns the server into a http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// use the go-chi logger middleware and redirect request logging to a file
func (s *Server) setupRequestLogging() {

	if s.Environment != "Development" {
		var file *os.File
		file, err := os.OpenFile(s.LogConfig.RequestPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("cannot use filepath '%s' as a logfile: %v", s.LogConfig.RequestPath, err))
		}
		middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger:  log.New(file, "", log.LstdFlags),
			NoColor: true,
		})
	}
}
