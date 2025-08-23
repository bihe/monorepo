package develop

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

const DEV_TOKEN_PATH = "/gettoken"
const DEV_USER_PROFILE_IMAGE = "/user_profile_image.png"

// SetupDevTokenHandler is used during development to create a JWT token
func SetupDevTokenHandler(router chi.Router, logger logging.Logger, env config.Environment) {
	if env != config.Development {
		return
	}

	devHandler := devTokenHandler{
		Logger: logger,
		Env:    env,
	}
	router.Get(DEV_TOKEN_PATH, devHandler.index())
	router.Get(DEV_USER_PROFILE_IMAGE, devHandler.profileImage())

}

// SetupTokenRedirect is used for integration testing to simulate a proper login
func SetupTokenRedirect(router chi.Router, logger logging.Logger, env config.Environment) {
	if env != config.Development {
		return
	}

	devHandler := devTokenHandler{
		Logger: logger,
		Env:    env,
	}
	router.Get(DEV_USER_PROFILE_IMAGE, devHandler.profileImage())
	router.Get(DEV_TOKEN_PATH, devHandler.redirect())
}

// devTokenHandler is used during development to create a JWT token
type devTokenHandler struct {
	Logger logging.Logger
	Env    config.Environment
}

func createJWTCookie(env config.Environment, w http.ResponseWriter) string {
	if env != config.Development {
		return ""
	}
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
	return jwt
}

func (d devTokenHandler) index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwt := createJWTCookie(d.Env, w)

		// write the relevant cookie for development
		cookie := http.Cookie{
			Name:     "login_token",
			Value:    jwt,
			HttpOnly: true, // only let the api access those cookies
			Domain:   "localhost",
			Path:     "/",
			Secure:   false,
		}
		http.SetCookie(w, &cookie)

		content := `<html>
<body>
	<textarea rows="10" cols="100">%s</textarea>
	<br/>
	<br/>
	<a href="%s">Goto Start</a>
</body>
</html>`
		referrer := r.Header.Get("Referer")
		if referrer != "" {
			// get the "base-path"
			parts := strings.Split(referrer, "/")
			referrer = fmt.Sprintf("http://%s/%s", r.Host, parts[3])
		}
		if referrer == "" {
			referrer = "/"
		}
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(content, jwt, referrer)))
	}
}

func (d devTokenHandler) redirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		createJWTCookie(d.Env, w)
		http.Redirect(w, r, getEnvOrDefault("REDIRECT_URL", "http://localhost:3000/redirect"), http.StatusFound)
	}
}

//go:embed CloneWhite.png
var profileImage []byte

func (d devTokenHandler) profileImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if d.Env != config.Development {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(profileImage)))
		w.Header().Add("Last-Modified", time.Now().Format(http.TimeFormat))
		w.WriteHeader(http.StatusOK)
		w.Write(profileImage)
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
		ProfileURL:  DEV_USER_PROFILE_IMAGE,
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
