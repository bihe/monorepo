package mydms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	stdlog "log"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
	"golang.binggl.net/monorepo/internal/mydms/app/shared"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.binggl.net/monorepo/pkg/server"

	httptransport "github.com/go-kit/kit/transport/http"
	pkgerr "golang.binggl.net/monorepo/pkg/errors"
)

// HTTPHandlerOptions containts configuration settings relevant for the uses HTTP handler implementation
type HTTPHandlerOptions struct {
	BasePath     string
	ErrorPath    string
	JWTConfig    config.Security
	CookieConfig config.ApplicationCookies
	CorsConfig   config.CorsSettings
	AssetConfig  config.AssetSettings
}

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a profilesvc server.
func MakeHTTPHandler(e Endpoints, logger log.Logger, lLogger *logrus.Entry, opts HTTPHandlerOptions) http.Handler {
	r := server.SetupBasicRouter(opts.BasePath, opts.CookieConfig, opts.CorsConfig, opts.AssetConfig, lLogger)
	apiRouter := server.SetupSecureAPIRouter(opts.ErrorPath, opts.JWTConfig, opts.CookieConfig, lLogger)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	apiRouter.Mount("/appinfo", httptransport.NewServer(
		e.GetAppInfoEndpoint,
		decodeGetAppInfoRequest,
		encodeResponse,
		options...,
	))

	apiRouter.Mount("/documents/{id}", httptransport.NewServer(
		e.GetDocumentByIDEndpoint,
		decodeGetDocumentByIDRequest,
		encodeResponse,
		options...,
	))

	// we are mounting all APIs under /api/v1 path
	r.Mount("/api/v1", apiRouter)

	return r
}

// --------------------------------------------------------------------------
// encode / decode requests and responses
// --------------------------------------------------------------------------

func decodeGetAppInfoRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	user := ensureUser(r)
	return appinfo.GetAppInfoRequest{User: user}, nil
}

func decodeGetDocumentByIDRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return nil, fmt.Errorf("missing parameter")
	}
	return document.GetDocumentByIDRequest{
		ID: chi.URLParam(r, "id"),
	}, nil
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	return respondJSON(w, response)
}

// --------------------------------------------------------------------------------------------------------------------

const t = "about:blank"

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	var (
		pd          *pkgerr.ProblemDetail
		errNotFound *shared.NotFoundError
	)
	if errors.As(err, &errNotFound) {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "object cannot be found",
			Status: http.StatusNotFound,
			Detail: errNotFound.Error(),
		}
	} else {
		pd = &pkgerr.ProblemDetail{
			Type:   t,
			Title:  "cannot service the request",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
	}
	writeProblemJSON(w, pd)
}

func ensureUser(r *http.Request) *security.User {
	user, ok := security.UserFromContext(r.Context())
	if !ok || user == nil {
		panic("the sucurity context user is not available!")
	}
	return user
}

func respondJSON(w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func writeProblemJSON(w http.ResponseWriter, pd *pkgerr.ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(pd.Status)
	b, err := json.Marshal(pd)
	if err != nil {
		stdlog.Printf("could not marshal json %v\n", err)
	}
	_, err = w.Write(b)
	if err != nil {
		stdlog.Printf("could not write bytes using http.ResponseWriter: %v\n", err)
	}
}
