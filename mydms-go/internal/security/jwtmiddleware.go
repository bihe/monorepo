package security

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"

	"github.com/bihe/mydms/internal/errors"
	sec "golang.binggl.net/commons/security"
)

// JwtWithConfig returns the configured JWT Auth middleware
func JwtWithConfig(options JwtOptions) echo.MiddlewareFunc {
	cache := sec.NewMemCache(parseDuration(options.CacheDuration))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			var token string
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				token = strings.Replace(authHeader, "Bearer ", "", 1)
			}
			if token == "" {
				// fallback to get the token via the cookie
				var cookie *http.Cookie
				if cookie, err = c.Request().Cookie(options.CookieName); err != nil {
					// neither the header nor the cookie supplied a jwt token
					return errors.RedirectError{
						Status:  http.StatusUnauthorized,
						Err:     fmt.Errorf("invalid authentication, no JWT token present"),
						Request: c.Request(),
						URL:     options.RedirectURL,
					}
				}
				token = cookie.Value
			}

			// to speed up processing use the cache for token lookups
			var user sec.User
			u := cache.Get(token)
			if u != nil {
				// cache hit, put the user in the context
				sc := &ServerContext{Context: c, Identity: *u}
				return next(sc)
			}

			var payload sec.JwtTokenPayload
			if payload, err = sec.ParseJwtToken(token, options.JwtSecret, options.JwtIssuer); err != nil {
				log.Warnf("Could not decode the JWT token payload: %s", err)
				return errors.RedirectError{
					Status:  http.StatusUnauthorized,
					Err:     fmt.Errorf("invalid authentication, could not parse the JWT token: %v", err),
					Request: c.Request(),
					URL:     options.RedirectURL,
				}
			}
			var roles []string
			claim := options.RequiredClaim
			if roles, err = sec.Authorize(sec.Claim{Name: claim.Name, URL: claim.URL, Roles: claim.Roles}, payload.Claims); err != nil {
				log.Warnf("Insufficient permissions to access the resource: %s", err)
				return errors.RedirectError{
					Status:  http.StatusForbidden,
					Err:     fmt.Errorf("Invalid authorization: %v", err),
					Request: c.Request(),
					URL:     options.RedirectURL,
				}
			}

			user = sec.User{
				DisplayName:   payload.DisplayName,
				Email:         payload.Email,
				Roles:         roles,
				UserID:        payload.UserID,
				Username:      payload.UserName,
				Authenticated: true,
			}
			cache.Set(token, &user)
			sc := &ServerContext{Context: c, Identity: user}
			return next(sc)
		}
	}
}

func parseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(fmt.Sprintf("wrong value, cannot parse duration: %v", err))
	}
	return d
}
