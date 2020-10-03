package mydms

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/mydms/app/appinfo"
	"golang.binggl.net/monorepo/pkg/security"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	GetAppInfoEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeServerEndpoints(s appinfo.Service, logger log.Logger) Endpoints {
	var appInfoEndopint endpoint.Endpoint
	{
		appInfoEndopint = MakeGetAppInfoEndpoint(s)
		appInfoEndopint = EndpointLoggingMiddleware(log.With(logger, "method", "GetAppInfo"))(appInfoEndopint)
	}

	return Endpoints{
		GetAppInfoEndpoint: appInfoEndopint,
	}
}

// MakeGetAppInfoEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetAppInfoEndpoint(s appinfo.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getAppInfoRequest)
		appInfo, err := s.GetAppInfo(req.User)
		return getAppInfoResponse{AppInfo: appInfo, Err: err}, nil
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

// --------------------------------------------------------------------------
// Requests and Responses passed to endpoints
// --------------------------------------------------------------------------

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error.
type errorer interface {
	error() error
}

type getAppInfoRequest struct {
	User *security.User
}

type getAppInfoResponse struct {
	AppInfo appinfo.AppInfo
	Err     error `json:"err,omitempty"`
}

func (r getAppInfoResponse) error() error { return r.Err }
