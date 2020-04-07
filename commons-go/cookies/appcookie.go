// Package cookies is responsible to read and write cookies
// based on a supplied configuration
package cookies // import "golang.binggl.net/commons/cookies"

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// CookieSameSite specifies the cookie SameSite mode
type CookieSameSite int

const (
	// CookieSameSiteDefault is the default mode
	CookieSameSiteDefault CookieSameSite = 1
	// CookieSameSiteLax is SameSiteLaxMode
	CookieSameSiteLax CookieSameSite = 2
	// CookieSameSiteStrict is SameSiteStrictMode
	CookieSameSiteStrict CookieSameSite = 3
	// CookieSameSiteNone is SameSiteNoneMode
	CookieSameSiteNone CookieSameSite = 4
)

// Settings defines parameters for cookies used for HTML-based errors
type Settings struct {
	Path         string
	Domain       string
	Secure       bool
	Prefix       string
	SameSiteMode CookieSameSite
}

// AppCookie is responsible for writing, reading cookies
type AppCookie struct {
	Settings Settings
}

// NewAppCookie returns a new instance of type AppCookie
func NewAppCookie(c Settings) *AppCookie {
	return &AppCookie{Settings: c}
}

// Set create a cookie with the given name and value and using the provided settings
func (a *AppCookie) Set(name, value string, expirySec int, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     a.cookieName(name),
		Value:    value,
		MaxAge:   expirySec,
		HttpOnly: true, // only let the api access those cookies
		SameSite: http.SameSite(a.Settings.SameSiteMode),
		Domain:   a.Settings.Domain,
		Path:     a.Settings.Path,
		Secure:   a.Settings.Secure,
	}
	http.SetCookie(w, &cookie)
}

// Del removes the cookie by setting MaxAge to < 0
func (a *AppCookie) Del(name string, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     a.cookieName(name),
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true, // only let the api access those cookies
		SameSite: http.SameSite(a.Settings.SameSiteMode),
		Domain:   a.Settings.Domain,
		Path:     a.Settings.Path,
		Secure:   a.Settings.Secure,
	}
	http.SetCookie(w, &cookie)
}

// Get retrieves a cookie value
func (a *AppCookie) Get(name string, r *http.Request) string {
	var (
		cookie *http.Cookie
		err    error
	)
	cookieName := a.cookieName(name)
	if cookie, err = r.Cookie(cookieName); err != nil {
		log.WithField("func", "cookies.Get").Debugf("could not read cookie '%s': %v\n", cookieName, err)
		return ""
	}
	return cookie.Value
}

func (a *AppCookie) cookieName(name string) string {
	return fmt.Sprintf("%s_%s", a.Settings.Prefix, name)
}
