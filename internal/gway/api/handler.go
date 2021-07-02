package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.binggl.net/monorepo/pkg/cookies"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"

	"golang.binggl.net/monorepo/internal/gway/app/conf"
	"golang.binggl.net/monorepo/internal/gway/app/oidc"
	"golang.binggl.net/monorepo/internal/gway/app/shared"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/server"
)

// HTTPHandlerOptions are used to configure the http handler setup
type HTTPHandlerOptions struct {
	BasePath  string
	ErrorPath string
	Config    conf.AppConfig
}

// MakeHTTPHandler creates a new handler implementation which is used together with the HTTP server
func MakeHTTPHandler(oidcSvc oidc.Service, logger logging.Logger, opts HTTPHandlerOptions) http.Handler {
	cookie := cookies.NewAppCookie(cookies.Settings{
		Path:   opts.Config.Cookies.Path,
		Domain: opts.Config.Cookies.Domain,
		Secure: opts.Config.Cookies.Secure,
		Prefix: opts.Config.Cookies.Prefix,
	})

	r := server.SetupBasicRouter(opts.BasePath, opts.Config.Cookies, opts.Config.Cors, opts.Config.Assets, logger)
	//apiRouter := server.SetupSecureAPIRouter(opts.ErrorPath, opts.JWTConfig, opts.CookieConfig, logger)

	// ---- OIDC routes ----
	r.Get("/start-oidc", handlePrepIntOIDCRedirect(oidcSvc, logger, cookie))
	r.Get(oidc.OIDCInitiateURL, handleGetExtOIDCRedirect(oidcSvc, logger, cookie))
	r.Get("/signin-oidc", handleLoginOIDC(oidcSvc, opts.Config.Security.CookieName, opts.Config.Security.Expiry, logger, cookie))
	r.Get("/auth/flow", handleAuthFlow(oidcSvc, logger, cookie))

	return r
}

// --------------------------------------------------------------------------
//   Internal helper functions
// --------------------------------------------------------------------------

func encodeError(err error, logger logging.Logger, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}

	t := "about:blank"
	var (
		pd            *pkgerr.ProblemDetail
		errNotFound   *shared.NotFoundError
		errValidation *shared.ValidationError
		errSecurity   *shared.SecurityError
	)
	if errors.As(err, &errNotFound) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "object cannot be found",
			Status: http.StatusNotFound,
			Detail: errNotFound.Error(),
		}
	} else if errors.As(err, &errValidation) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "error in parameter-validaton",
			Status: http.StatusBadRequest,
			Detail: errValidation.Error(),
		}
	} else if errors.As(err, &errSecurity) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "security error",
			Status: http.StatusForbidden,
			Detail: errSecurity.Error(),
		}
	} else {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "cannot service the request",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
	}
	writeProblemJSON(logger, w, pd)
}

func respondJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func writeProblemJSON(logger logging.Logger, w http.ResponseWriter, pd *pkgerr.ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(pd.Status)
	b, err := json.Marshal(pd)
	if err != nil {
		logger.Error(fmt.Sprintf("could not marshal json %v", err))
	}
	_, err = w.Write(b)
	if err != nil {
		logger.Error(fmt.Sprintf("could not write bytes using http.ResponseWriter: %v", err))
	}
}

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}
