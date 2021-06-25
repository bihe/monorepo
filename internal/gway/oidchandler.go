package gway

import (
	"net/http"

	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/logging"
)

/*
type Service interface {
	// LoginOIDC evaluates the returend data from the OIDC interaction and creates a token and redirect url
	LoginOIDC(state, oidcState, oidcCode string) (token, url string, err error)
	// LoginSiteOIDC performs a login via OIDC but checks if the user has permissions for the given site and redirects to the defined url
	LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error)
}
*/

func handlePrepIntOIDCRedirect(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, state := svc.PrepIntOIDCRedirect()
		c.Set(StateCookie, state, 60, w)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func handleGetExtOIDCRedirect(svc oidc.Service, logger logging.Logger, c *cookies.AppCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := c.Get(StateCookie, r)
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
