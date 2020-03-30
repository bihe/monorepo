// Package security implements authentication / authorization by means of JWT tokens
package security // import "golang.binggl.net/commons/security"

import (
	"context"

	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.binggl.net/commons"
	"golang.binggl.net/commons/cookies"
	"golang.binggl.net/commons/errors"
)

// ctxtKey is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxtKey int

// userKey is used to access the authenticated user in the context
var userKey ctxtKey

// JwtMiddleware is used to authenticate a user based on a token
// the token is either retrieved by the well known Authorization header
// or fetched from a cookie
type JwtMiddleware struct {
	jwt    JwtOptions
	errRep *errors.ErrorReporter
	log    *log.Entry
}

// NewJwtMiddleware creates a new instance using the provided options
func NewJwtMiddleware(options JwtOptions, settings cookies.Settings, logger *log.Entry) *JwtMiddleware {
	m := JwtMiddleware{
		jwt: options,
		errRep: &errors.ErrorReporter{
			CookieSettings: settings,
		},
		log: logger,
	}
	return &m
}

// JwtContext performs the middleware action
func (j *JwtMiddleware) JwtContext(next http.Handler) http.Handler {
	return handleJWT(next, j.jwt, j.errRep, j.log)
}

// UserFromContext returns the User value stored in ctx, if any.
func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}

// NewContext returns a new Context that carries value u.
func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func handleJWT(next http.Handler, options JwtOptions, errRep *errors.ErrorReporter, logger *log.Entry) http.Handler {
	cache := NewMemCache(parseDuration(options.CacheDuration))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err   error
			token string
		)

		authHeader := r.Header.Get("Authorization")
		reqURL := r.URL.String()
		if reqURL != "" {
			reqURL = url.QueryEscape(reqURL)
		}
		// add request-Url and take care of url-encoding
		redirectURL := options.RedirectURL + "%26ref=" + reqURL

		if authHeader != "" {
			token = strings.Replace(authHeader, "Bearer ", "", 1)
		}
		if token == "" {
			// fallback to get the token via the cookie
			var cookie *http.Cookie
			if cookie, err = r.Cookie(options.CookieName); err != nil {
				commons.LogWithReq(r, logger, "security.handleJWT").Warnf("could not get token from header nor cookie: %v", err)
				// neither the header nor the cookie supplied a jwt token
				errRep.Negotiate(w, r, errors.RedirectError{
					Status:  http.StatusUnauthorized,
					Err:     fmt.Errorf("invalid authentication, no JWT token present"),
					Request: r,
					URL:     redirectURL,
				})
				return
			}

			token = cookie.Value
		}

		// to speed up processing use the cache for token lookups
		var user User
		u := cache.Get(token)
		if u != nil {
			// cache hit, put the user in the context
			ctx := context.WithValue(r.Context(), userKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		var payload JwtTokenPayload
		if payload, err = ParseJwtToken(token, options.JwtSecret, options.JwtIssuer); err != nil {
			commons.LogWithReq(r, logger, "security.handleJWT").Warnf("Could not decode the JWT token payload: %s", err)
			errRep.Negotiate(w, r, errors.RedirectError{
				Status:  http.StatusUnauthorized,
				Err:     fmt.Errorf("invalid authentication, could not parse the JWT token: %v", err),
				Request: r,
				URL:     redirectURL,
			})
			return
		}
		var roles []string
		claim := options.RequiredClaim
		if roles, err = Authorize(Claim{Name: claim.Name, URL: claim.URL, Roles: claim.Roles}, payload.Claims); err != nil {
			commons.LogWithReq(r, logger, "security.handleJWT").Warnf("Insufficient permissions to access the resource: %s", err)
			errRep.Negotiate(w, r, errors.RedirectError{
				Status:  http.StatusForbidden,
				Err:     fmt.Errorf("Invalid authorization: %v", err),
				Request: r,
				URL:     redirectURL,
			})
			return
		}

		user = User{
			DisplayName:   payload.DisplayName,
			Email:         payload.Email,
			Roles:         roles,
			UserID:        payload.UserID,
			Username:      payload.UserName,
			Authenticated: true,
		}
		cache.Set(token, &user)

		ctx := context.WithValue(r.Context(), userKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(fmt.Sprintf("wrong value, cannot parse duration: %v", err))
	}
	return d
}
