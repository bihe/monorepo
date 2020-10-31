package crypter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"golang.binggl.net/monorepo/pkg/logging"
)

// --------------------------------------------------------------------------
// Endpoint definitions
// --------------------------------------------------------------------------

// Endpoints collects all of the endpoints that compose a service. It's meant to
// be used as a helper struct, to collect all of the endpoints into a single
// parameter.
// An endpoint is needed per service method.
type Endpoints struct {
	CrypterEndpoint endpoint.Endpoint
}

// NewEndpoints returns Endpoints which wrap the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func NewEndpoints(svc EncryptionService, logger logging.Logger) Endpoints {
	var cryptEndpoint endpoint.Endpoint
	{
		cryptEndpoint = MakeEncryptEndpoint(svc)
		cryptEndpoint = EndpointLoggingMiddleware(logger, "Encrypt")(cryptEndpoint)
	}
	return Endpoints{
		CrypterEndpoint: cryptEndpoint,
	}
}

// Encrypt implements the service interface, so Endpoints may be used as a service.
// This is primarily useful in the context of a client library.
func (e Endpoints) Encrypt(ctx context.Context, req Request) ([]byte, error) {
	resp, err := e.CrypterEndpoint(ctx, req)
	if err != nil {
		return nil, err
	}
	response := resp.(Response)
	return response.Payload, response.Err
}

// MakeEncryptEndpoint constructs an Encrypt endpoint wrapping the service.
func MakeEncryptEndpoint(s EncryptionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(Request)
		enc, err := s.Encrypt(ctx, req)
		return Response{Payload: enc, Err: err}, nil
	}
}

// --------------------------------------------------------------------------
// Logging middleware
// --------------------------------------------------------------------------

// EndpointLoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func EndpointLoggingMiddleware(logger logging.Logger, endpointName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			logger.Info(fmt.Sprintf("call endpoint: %s", endpointName))
			defer func(begin time.Time) {
				logger.Info(fmt.Sprintf("%s: endpoint stats", endpointName), logging.LogV("took", time.Since(begin).String()))
				if err != nil {
					logger.Error(fmt.Sprintf("%s: transport-error", endpointName), logging.ErrV(err))
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}
