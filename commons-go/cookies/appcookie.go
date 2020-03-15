// Package cookies is responsible to read and write cookies
// based on a supplied configuration
package cookies // import "golang.binggl.net/commons/cookies"

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

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
		Domain:   a.Settings.Domain,
		Path:     a.Settings.Path,
		MaxAge:   expirySec,
		Secure:   a.Settings.Secure,
		HttpOnly: true, // only let the api access those cookies
	}
	http.SetCookie(w, &cookie)
}

// Del removes the cookie by setting MaxAge to < 0
func (a *AppCookie) Del(name string, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     a.cookieName(name),
		Value:    "",
		Domain:   a.Settings.Domain,
		Path:     a.Settings.Path,
		MaxAge:   -1,
		Secure:   a.Settings.Secure,
		HttpOnly: true, // only let the api access those cookies
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
