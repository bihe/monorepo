package api

import (
	"fmt"
	"net/http"
	"strings"

	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/shared"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/logging"
)

const oidcStateParam = "state"
const oidcCodeParam = "code"
const stateCookieName = "state"
const siteParam = "~site"
const redirectParam = "~url"
const authFlowCookie = "auth_flow"
const authFlowSep = "|"
const cookieExpiry = 60

type OidcHandler struct {
	OidcSvc       oidc.Service
	Logger        logging.Logger
	Cookies       *cookies.AppCookie
	JwtCookieName string
	JwtExpiryDays int
}

// handlePrepIntOIDCRedirect creates a **state** value which is stored as a cookie
// the **state** value is used a a correlation token during the OIDC call
// to ensure that the cookie is written this intermediate hop is used (same host/damain/...) until the real OIDC redirect starts
// -- handlePrepIntOIDCRedirect -- (redirect) -->  handleGetExtOIDCRedirect
func (o OidcHandler) HandlePrepIntOIDCRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, state := o.OidcSvc.PrepIntOIDCRedirect()
		o.Cookies.Set(stateCookieName, state, 60, w)
		o.Logger.Debug("begin with state", logging.LogV("state_value", state), logging.LogV("cookie_name", stateCookieName))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// handleGetExtOIDCRedirect just reads the **state** cookie value (and returns an error if no **state** cookie is found)
// if the **state** is available the handler starts the OIDC process by redirecting to the first URL of the OIDC implementation
// -- handleGetExtOIDCRedirect -- (redirect) --> OIDC process flow
func (o OidcHandler) HandleGetExtOIDCRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := o.Cookies.Get(stateCookieName, r)
		o.Logger.Debug("retrieve state cookie", logging.LogV("cookie_name", stateCookieName), logging.LogV("cookie_value", state))
		if state == "" {
			encodeError(shared.ErrValidation("missing 'state' cookie-value"), o.Logger, w)
			return
		}
		url, err := o.OidcSvc.GetExtOIDCRedirect(state)
		if err != nil {
			encodeError(err, o.Logger, w)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// handleLoginOIDC is called via the OIDC process (the route is configured with the OIDC implementation as the return URL)
// the state/code parameters are retrieved state is compared to the **state** cookie value and code is used to fetch the token
// once successfull a redirect is performed to the configured URL or the site-URL
// -- handleLoginOIDC -- (redirect) --> configured application URL
func (o OidcHandler) HandleLoginOIDC(jwtCookieName string, jwtCookieExpire int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			token  string
			url    string
			err    error
			expSec int
		)

		// need the cookieExpiry in seconds
		expSec = jwtCookieExpire * 24 * 60 * 60

		// state in the saved cookie
		state := o.Cookies.Get(stateCookieName, r)
		if state == "" {
			encodeError(shared.ErrValidation("missing 'state' cookie-value"), o.Logger, w)
			return
		}
		// fetch the parameters provided by the OIDC handshake
		oidcState := queryParam(r, oidcStateParam)
		oidcCode := queryParam(r, oidcCodeParam)
		o.Logger.Debug("got OIDC params", logging.LogV("state", oidcState), logging.LogV("code", oidcCode))

		// check for site-login; this is indicated by the auth_flow cookie
		// if such a cookie is available perform a sit-login, otherwise the "std" login is performed
		siteCookie := o.Cookies.Get(authFlowCookie, r)
		if siteCookie != "" {
			// the auth-flow cookie is present, digest it
			parts := strings.Split(siteCookie, authFlowSep)
			if len(parts) != 2 {
				encodeError(shared.ErrValidation("invalid parameters to perform an auth-flow"), o.Logger, w)
				return
			}
			// there need to be two parts to make any sense
			site := parts[0]
			redirectURL := parts[1]
			o.Logger.Info("perform site login", logging.LogV("site", site), logging.LogV("url", redirectURL))

			token, url, err = o.OidcSvc.LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL)
			o.Cookies.Del(authFlowCookie, w)
		} else {
			// this is the "normal/std" behavior - evalulate the data proviced by the OIDC process
			// create a token and redirect to the configured URL
			token, url, err = o.OidcSvc.LoginOIDC(state, oidcState, oidcCode)
		}

		if err != nil {
			encodeError(err, o.Logger, w)
			return
		}

		if token == "" {
			o.Logger.Warn("no error, but empty token")
			encodeError(fmt.Errorf("a token could not be created"), o.Logger, w)
			return
		}
		o.Cookies.Del(stateCookieName, w)

		jwtCookie := cookies.NewAppCookie(cookies.Settings{
			Path:   o.Cookies.Settings.Path,
			Domain: o.Cookies.Settings.Domain,
			Secure: o.Cookies.Settings.Secure,
			Prefix: "",
		})
		jwtCookie.Set(jwtCookieName, token, expSec, w)
		o.Logger.Debug("set jwt cookie", logging.LogV("cookie_name", jwtCookieName), logging.LogV("cookie_expiry_sec", fmt.Sprintf("%d", expSec)))

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// handleAuthFlow is used to perform an OIDC login (all the same process as above) but in addition a site is provided which is used to check specific permisions
// if the check is successfull (and authorization is performend via OIDC) a redirect is performed to the specific application site
// the first step is to save the site/URL and then kick-off the OIDC process
// -- handleAuthFlow -- (redirect) --> handleGetExtOIDCRedirect
func (o OidcHandler) HandleAuthFlow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		site := queryParam(r, siteParam)
		redirect := queryParam(r, redirectParam)
		if site == "" || redirect == "" {
			o.Logger.Warn(fmt.Sprintf("either '%s' or '%s' param are missing!", siteParam, redirectParam))
			encodeError(shared.ErrValidation("missing parameters for auth-flow/site-login"), o.Logger, w)
			return
		}
		o.Cookies.Set(authFlowCookie, fmt.Sprintf("%s%s%s", site, authFlowSep, redirect), cookieExpiry, w)
		o.Logger.Debug("auth-flow data saved in cookie", logging.LogV("site", site), logging.LogV("url", redirect))

		url, state := o.OidcSvc.PrepIntOIDCRedirect()
		o.Cookies.Set(stateCookieName, state, cookieExpiry, w)
		o.Logger.Debug("begin with state", logging.LogV("state_value", state), logging.LogV("cookie_name", stateCookieName))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
