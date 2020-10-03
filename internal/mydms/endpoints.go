package mydms

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/internal/mydms/app/document"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	GetAppInfoEndpoint      endpoint.Endpoint
	GetDocumentByIDEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeServerEndpoints(ai appinfo.Service, doc document.Service, logger log.Logger) Endpoints {
	var appInfoEndopint endpoint.Endpoint
	{
		appInfoEndopint = appinfo.MakeGetAppInfoEndpoint(ai)
		appInfoEndopint = EndpointLoggingMiddleware(log.With(logger, "method", "GetAppInfo"))(appInfoEndopint)
	}

	var documentByIDEndpoint endpoint.Endpoint
	{
		documentByIDEndpoint = document.MakeGetDocumentByIDEndpoint(doc)
		documentByIDEndpoint = EndpointLoggingMiddleware(log.With(logger, "method", "GetDocumentByID"))(documentByIDEndpoint)
	}

	return Endpoints{
		GetAppInfoEndpoint:      appInfoEndopint,
		GetDocumentByIDEndpoint: documentByIDEndpoint,
	}
}

// --------------------------------------------------------------------------
// Logging middleware
// --------------------------------------------------------------------------

// EndpointLoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("transport_error", err, "took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)
		}
	}
}
