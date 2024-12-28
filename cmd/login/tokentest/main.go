package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/develop"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
)

// the only purpose of this server is to write a cookie with the JWT token-payload
// to have a established authentication for testing-purposes
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	develop.SetupTokenRedirect(r, logging.NewNop(), config.Development)
	server.PrintServerBanner("tokentest", "1.0.0", "local", "localhost", "localhost:3000")
	http.ListenAndServe(":"+getEnvOrDefault("PORT", "3000"), r)
}

func getEnvOrDefault(env, def string) string {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	return v
}
