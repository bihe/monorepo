package server

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"golang.binggl.net/monorepo/pkg/handler"
	"golang.binggl.net/monorepo/pkg/security"
)

// routes performs setup of middlewares and API handlers
func (s *Server) routes() {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(handler.NewLoggerMiddleware(s.log).LoggerContext)
	// use the default list of "compressable" content-type
	r.Use(middleware.NewCompressor(5).Handler)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// setup cors for single frontend
	cors := cors.New(cors.Options{
		AllowedOrigins:   s.cors.Origins,
		AllowedMethods:   s.cors.Methods,
		AllowedHeaders:   s.cors.Headers,
		AllowCredentials: s.cors.Credentials,
		MaxAge:           s.cors.MaxAge,
	})
	r.Use(cors.Handler)
	r.Get("/error", s.errorHandler.Call(s.errorHandler.HandleError))

	// serving static content
	handler.ServeStaticFile(r, "/favicon.ico", filepath.Join(s.basePath, "./assets/favicon.ico"))
	handler.ServeStaticDir(r, "/assets", http.Dir(filepath.Join(s.basePath, "./assets")))

	// this group "indicates" that all routes within this group use the JWT authentication
	r.Group(func(r chi.Router) {
		// authenticate and authorize users via JWT
		r.Use(security.NewJwtMiddleware(s.jwtOpts, s.cookieSettings, s.log).JwtContext)

		r.Mount("/hc", s.healthCheck.GetHandler())
		r.Mount("/appinfo", s.appInfoAPI.GetHandler())

		// group API methods together
		r.Route("/api/v1/bookmarks", func(r chi.Router) {
			r.Post("/", s.bookmarkAPI.Secure(s.bookmarkAPI.Create))
			r.Put("/", s.bookmarkAPI.Secure(s.bookmarkAPI.Update))
			r.Put("/sortorder", s.bookmarkAPI.Secure(s.bookmarkAPI.UpdateSortOrder))
			r.Delete("/{id}", s.bookmarkAPI.Secure(s.bookmarkAPI.Delete))
			r.Get("/{id}", s.bookmarkAPI.Secure(s.bookmarkAPI.GetBookmarkByID))
			r.Get("/bypath", s.bookmarkAPI.Secure(s.bookmarkAPI.GetBookmarksByPath))
			r.Get("/allpaths", s.bookmarkAPI.Secure(s.bookmarkAPI.GetAllPaths))
			r.Get("/folder", s.bookmarkAPI.Secure(s.bookmarkAPI.GetBookmarksFolderByPath))
			r.Get("/byname", s.bookmarkAPI.Secure(s.bookmarkAPI.GetBookmarksByName))
			r.Get("/mostvisited/{num}", s.bookmarkAPI.Secure(s.bookmarkAPI.GetMostVisited))
			r.Get("/fetch/{id}", s.bookmarkAPI.Secure(s.bookmarkAPI.FetchAndForward))
			r.Get("/favicon/{id}", s.bookmarkAPI.Secure(s.bookmarkAPI.GetFavicon))
		})

		// swagger
		handler.ServeStaticDir(r, "/swagger", http.Dir(filepath.Join(s.basePath, "./assets/swagger")))

		// auth-redirect sometimes creates routes with the
		// syntax https:/.../?valid=true&ref=/api/v1/...
		// redirect to the /api path
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// determine if well-known parameters are supplied
			valid := r.URL.Query().Get("valid")
			ref := r.URL.Query().Get("ref")
			if valid == "" || ref == "" {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// the ref path needs to be known, valid know paths are
			// /api, /api/v1/bookmarks/ ...
			if valid == "true" && (strings.HasPrefix(ref, "/api/v1/bookmarks") || strings.HasPrefix(ref, "/api")) {
				http.Redirect(w, r, ref, http.StatusFound)
				return
			}

			// default to the frontend
			w.WriteHeader(http.StatusNotFound)
		})
	})

	s.router = r
}
