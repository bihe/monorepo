package gway

import (
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/internal/gway/app/shared"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/logging"
)

/*
type Service interface {
	// LoginSiteOIDC performs a login via OIDC but checks if the user has permissions for the given site and redirects to the defined url
	LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error)
}
*/

const oidcStateParam = "state"
const oidcCodeParam = "code"

func handlePrepIntOIDCRedirect(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, state := svc.PrepIntOIDCRedirect()
		c.Set(StateCookie, state, 60, w)
		logger.Debug("begin with state", logging.LogV("state_value", state), logging.LogV("cookie_name", StateCookie))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func handleGetExtOIDCRedirect(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := c.Get(StateCookie, r)
		logger.Debug("retrieve state cookie", logging.LogV("cookie_name", StateCookie), logging.LogV("cookie_value", state))
		if state == "" {
			encodeError(shared.ErrValidation("missing 'state' cookie-value"), logger, w)
			return
		}
		url, err := svc.GetExtOIDCRedirect(state)
		if err != nil {
			encodeError(err, logger, w)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func handleLoginOIDC(svc oidc.Service, jwtCookieName string, jwtCookieExpire int, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// state in the saved cookie
		state := c.Get(StateCookie, r)
		if state == "" {
			encodeError(shared.ErrValidation("missing 'state' cookie-value"), logger, w)
			return
		}
		// fetch the parameters provided by the OIDC handshake
		oidcState := queryParam(r, oidcStateParam)
		oidcCode := queryParam(r, oidcCodeParam)
		logger.Debug("got OIDC params", logging.LogV("state", oidcState), logging.LogV("code", oidcCode))

		token, url, err := svc.LoginOIDC(state, oidcState, oidcCode)
		if err != nil {
			encodeError(err, logger, w)
			return
		}

		// clean-up the OIDC state held in the cookie
		c.Del(StateCookie, w)

		// creat a new cookie instance, without a prefix
		jwtCookie := cookies.NewAppCookie(cookies.Settings{
			Path:   c.Settings.Path,
			Domain: c.Settings.Domain,
			Secure: c.Settings.Secure,
			Prefix: "",
		})
		jwtCookie.Set(jwtCookieName, token, jwtCookieExpire, w)
		logger.Debug("set jwt cookie", logging.LogV("cookie_name", jwtCookieName), logging.LogV("cookie_expiry", fmt.Sprintf("%d", jwtCookieExpire)))

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
