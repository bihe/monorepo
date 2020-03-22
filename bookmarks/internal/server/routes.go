package server

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"golang.binggl.net/commons/handler"
	"golang.binggl.net/commons/security"
)

// routes performs setup of middlewares and API handlers
func (s *Server) routes() {
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
		AllowedOrigins:   s.cors.Origins,
		AllowedMethods:   s.cors.Methods,
		AllowedHeaders:   s.cors.Headers,
		AllowCredentials: s.cors.Credentials,
		MaxAge:           s.cors.MaxAge,
	})
	r.Use(cors.Handler)

	s.setupRequestLogging()

	r.Get("/error", s.errorHandler.Call(s.errorHandler.HandleError))

	// serving static content
	handler.ServeStaticFile(r, "/favicon.ico", filepath.Join(s.basePath, "./assets/favicon.ico"))
	handler.ServeStaticDir(r, "/assets", http.Dir(filepath.Join(s.basePath, "./assets")))

	// this group "indicates" that all routes within this group use the JWT authentication
	r.Group(func(r chi.Router) {
		// authenticate and authorize users via JWT
		r.Use(security.NewJwtMiddleware(s.jwtOpts, s.cookieSettings).JwtContext)

		r.Get("/appinfo", s.appInfoAPI.Secure(s.appInfoAPI.HandleAppInfo))

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
	})

	s.router = r
}
