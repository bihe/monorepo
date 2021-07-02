package api

import (
	"fmt"
	"net/http"
	"strings"

	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/internal/gway/app/shared"
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

// handlePrepIntOIDCRedirect creates a **state** value which is stored as a cookie
// the **state** value is used a a correlation token during the OIDC call
// to ensure that the cookie is written this intermediate hop is used (same host/damain/...) until the real OIDC redirect starts
// -- handlePrepIntOIDCRedirect -- (redirect) -->  handleGetExtOIDCRedirect
func handlePrepIntOIDCRedirect(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, state := svc.PrepIntOIDCRedirect()
		c.Set(stateCookieName, state, 60, w)
		logger.Debug("begin with state", logging.LogV("state_value", state), logging.LogV("cookie_name", stateCookieName))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// handleGetExtOIDCRedirect just reads the **state** cookie value (and returns an error if no **state** cookie is found)
// if the **state** is available the handler starts the OIDC process by redirecting to the first URL of the OIDC implementation
// -- handleGetExtOIDCRedirect -- (redirect) --> OIDC process flow
func handleGetExtOIDCRedirect(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := c.Get(stateCookieName, r)
		logger.Debug("retrieve state cookie", logging.LogV("cookie_name", stateCookieName), logging.LogV("cookie_value", state))
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

// handleLoginOIDC is called via the OIDC process (the route is configured with the OIDC implementation as the return URL)
// the state/code parameters are retrieved state is compared to the **state** cookie value and code is used to fetch the token
// once successfull a redirect is performed to the configured URL or the site-URL
// -- handleLoginOIDC -- (redirect) --> configured application URL
func handleLoginOIDC(svc oidc.Service, jwtCookieName string, jwtCookieExpire int, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			token string
			url   string
			err   error
		)

		// state in the saved cookie
		state := c.Get(stateCookieName, r)
		if state == "" {
			encodeError(shared.ErrValidation("missing 'state' cookie-value"), logger, w)
			return
		}
		// fetch the parameters provided by the OIDC handshake
		oidcState := queryParam(r, oidcStateParam)
		oidcCode := queryParam(r, oidcCodeParam)
		logger.Debug("got OIDC params", logging.LogV("state", oidcState), logging.LogV("code", oidcCode))

		// check for site-login; this is indicated by the auth_flow cookie
		// if such a cookie is available perform a sit-login, otherwise the "std" login is performed
		siteCookie := c.Get(authFlowCookie, r)
		if siteCookie != "" {
			// the auth-flow cookie is present, digest it
			parts := strings.Split(siteCookie, authFlowSep)
			if len(parts) != 2 {
				encodeError(shared.ErrValidation("invalid parameters to perform an auth-flow"), logger, w)
				return
			}
			// there need to be two parts to make any sense
			site := parts[0]
			redirectURL := parts[1]
			logger.Info("perform site login", logging.LogV("site", site), logging.LogV("url", redirectURL))

			token, url, err = svc.LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL)
			c.Del(authFlowCookie, w)
		} else {
			// this is the "normal/std" behavior - evalulate the data proviced by the OIDC process
			// create a token and redirect to the configured URL
			token, url, err = svc.LoginOIDC(state, oidcState, oidcCode)
		}

		if err != nil {
			encodeError(err, logger, w)
			return
		}

		if token == "" {
			logger.Warn("no error, but empty token")
			encodeError(fmt.Errorf("a token could not be created"), logger, w)
			return
		}

		// clean-up the OIDC state held in the cookie
		c.Del(stateCookieName, w)

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

// handleAuthFlow is used to perform an OIDC login (all the same process as above) but in addition a site is provided which is used to check specific permisions
// if the check is successfull (and authorization is performend via OIDC) a redirect is performed to the specific application site
// the first step is to save the site/URL and then kick-off the OIDC process
// -- handleAuthFlow -- (redirect) --> handleGetExtOIDCRedirect
func handleAuthFlow(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		site := queryParam(r, siteParam)
		redirect := queryParam(r, redirectParam)
		if site == "" || redirect == "" {
			logger.Warn(fmt.Sprintf("either '%s' or '%s' param are missing!", siteParam, redirectParam))
			encodeError(shared.ErrValidation("missing parameters for auth-flow/site-login"), logger, w)
			return
		}
		c.Set(authFlowCookie, fmt.Sprintf("%s%s%s", site, authFlowSep, redirect), cookieExpiry, w)
		logger.Debug("auth-flow data saved in cookie", logging.LogV("site", site), logging.LogV("url", redirect))

		url, state := svc.PrepIntOIDCRedirect()
		c.Set(stateCookieName, state, cookieExpiry, w)
		logger.Debug("begin with state", logging.LogV("state_value", state), logging.LogV("cookie_name", stateCookieName))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
