// Package onefrontend implements a web-server for the unified frontend components
package onefrontend

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.binggl.net/monorepo/onefrontend/config"
	"golang.binggl.net/monorepo/onefrontend/types"

	log "github.com/sirupsen/logrus"
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
	Environment    config.Environment
	Version        types.VersionInfo
	Log            *log.Entry
	router         chi.Router
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
		Log:    s.Log,
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
	r.Use(handler.NewLoggerMiddleware(s.Log).LoggerContext)
	// use the default list of "compressable" content-type
	r.Use(middleware.NewCompressor(5).Handler)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// setup cors for single frontend
	cors := cors.New(cors.Options{
		AllowedOrigins:   s.Cors.Origins,
		AllowedMethods:   s.Cors.Methods,
		AllowedHeaders:   s.Cors.Headers,
		AllowCredentials: s.Cors.Credentials,
		MaxAge:           s.Cors.MaxAge,
	})
	r.Use(cors.Handler)
	r.Get("/error", errHandler.Call(errHandler.HandleError))

	// serving static content
	handler.ServeStaticFile(r, "/favicon.ico", filepath.Join(s.BasePath, s.AssetDir, "favicon.ico"))
	handler.ServeStaticDir(r, s.AssetPrefix, http.Dir(filepath.Join(s.BasePath, s.AssetDir)))

	// this group "indicates" that all routes within this group use the JWT authentication
	r.Group(func(r chi.Router) {
		// authenticate and authorize users via JWT
		r.Use(security.NewJwtMiddleware(jwtOptions, cookieSettings, s.Log).JwtContext)

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
