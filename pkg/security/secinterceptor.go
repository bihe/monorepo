package security

import (
	"fmt"
	"net/http"

	pkgerr "golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/logging"
)

// secInterceptor is used the redirect to an error page if it can be assumed
// that a human requested a URL and not an api-client.
// A middleware to "catch" security errors and present a human-readable form
// if the client requests "application/json" just use the the problem-json format
type SecInterceptor struct {
	Log           logging.Logger
	Auth          *JWTAuthorization
	Options       JwtOptions
	ErrorRedirect string
}

const httpHeaderAccept = "Accept"
const apiContentType = "application/json"

// HandleJWT calls the JWT validation and either presents an error or redirects to the ErrorRedirect
func (u SecInterceptor) HandleJWT(next http.Handler) http.Handler {
	jwtAuth := u.Auth
	options := u.Options
	logger := u.Log
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := Validate(r, jwtAuth, options, logger)
		if err != nil {
			u.Log.Debug("Custom JWT validation in SecInterceptor!")
			acc := r.Header.Get(httpHeaderAccept)
			if acc == apiContentType {
				u.Log.Debug("the client requests JSON, provide json error-message")
				pkgerr.WriteError(w, r, pkgerr.SecurityError{
					Err:     fmt.Errorf("security error because of missing JWT token"),
					Request: r,
					Status:  http.StatusUnauthorized,
				})
				return
			}
			// when the Accept header is not pure 'application/json' provide a human-readable response
			http.Redirect(w, r, u.ErrorRedirect, http.StatusFound)
			return

		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
