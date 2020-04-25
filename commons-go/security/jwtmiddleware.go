// Package security implements authentication / authorization by means of JWT tokens
package security // import "golang.binggl.net/commons/security"

import (
	"context"

	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	jwtAuth := NewJWTAuthorization(options, true)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err   error
			token string
			user  *User
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
					Err:     fmt.Errorf("security error because of missing JWT token"),
					Request: r,
					URL:     redirectURL,
				})
				return
			}

			token = cookie.Value
		}

		user, err = jwtAuth.EvaluateToken(token)
		if err != nil {
			commons.LogWithReq(r, logger, "security.handleJWT").Warnf("error during JWT token evaluation: %v", err)
			errRep.Negotiate(w, r, errors.SecurityError{
				Err:     fmt.Errorf("error during token evaluation: %v", err),
				Request: r,
			})
			return
		}
		ctx := NewContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
