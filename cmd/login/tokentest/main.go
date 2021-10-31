package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.binggl.net/monorepo/pkg/security"
)

// the only purpose of this server is to write a cookie with the JWT token-payload
// to have a established authentication for testing-purposes
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		jwt := createToken()
		cookie := http.Cookie{
			Name:     getEnvOrDefault("JWT_COOKIE_NAME", "login_token"),
			Value:    jwt,
			MaxAge:   3600, /* exp in seconds */
			HttpOnly: true, // only let the api access those cookies
			Domain:   getEnvOrDefault("JWT_COOKIE_DOMAIN", "localhost"),
			Path:     "/",
			Secure:   false,
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, getEnvOrDefault("REDIRECT_URL", "http://localhost:3000/redirect"), http.StatusFound)
	})
	http.ListenAndServe(":"+getEnvOrDefault("PORT", "3000"), r)
}

func getEnvOrDefault(env, def string) string {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	return v
}

func createToken() string {
	var siteClaims []string
	siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", "core", "http://localhost:3001", "User;Admin"))
	siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", "mydms", "http://localhost:3002", "User"))
	siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", "bookmarks", "http://localhost:3003", "User"))
	siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", "crypter", "http://localhost:3004", "User"))
	siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", "onefrontend", "http://localhost:3000", "User"))

	claims := security.Claims{
		Type:        "login.User",
		DisplayName: "DisplayName",
		Email:       getEnvOrDefault("JWT_USER_EMAIL", "a.b@c.de"),
		UserID:      "UID",
		UserName:    getEnvOrDefault("JWT_USER_EMAIL", "a.b@c.de"),
		GivenName:   "Test",
		Surname:     "User",
		ProfileURL:  "https://lh3.googleusercontent.com/a-/AOh14Ghd0ONOWZ9mNBoUvtLhw4D1lPmGupZkggK6RahYMA",
		Claims:      siteClaims,
	}
	token, _ := security.CreateToken(getEnvOrDefault("JWT_ISSUER", "issuer"), []byte(getEnvOrDefault("JWT_SECRET", "secret")), 3600, claims)
	return token
}
