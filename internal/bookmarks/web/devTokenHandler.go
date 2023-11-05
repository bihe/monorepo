package web

import (
	"fmt"
	"net/http"
	"os"

	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

type DevTokenHandler struct {
	Logger logging.Logger
	Env    config.Environment
}

func (d DevTokenHandler) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d.Env != config.Development {
			w.WriteHeader(http.StatusConflict)
			return
		}

		jwt := createToken()

		cookie := http.Cookie{
			Name:  "login_token",
			Value: jwt,
		}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(jwt))
	}
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

func getEnvOrDefault(env, def string) string {
	v := os.Getenv(env)
	if v == "" {
		return def
	}
	return v
}
