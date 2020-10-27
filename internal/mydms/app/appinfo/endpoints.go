package appinfo

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"golang.binggl.net/monorepo/pkg/security"
)

// MakeGetAppInfoEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetAppInfoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetAppInfoRequest)
		appInfo, err := s.GetAppInfo(req.User)
		return GetAppInfoResponse{AppInfo: appInfo, Err: err}, nil
	}
}

// --------------------------------------------------------------------------
// Requests / Responses
// --------------------------------------------------------------------------

// GetAppInfoRequest combines the necessary parameters for a appinfo request
type GetAppInfoRequest struct {
	User *security.User
}

// GetAppInfoResponse is the appinfo response object
type GetAppInfoResponse struct {
	AppInfo AppInfo
	Err     error `json:"err,omitempty"`
}

// Failed implements endpoint.Failer.
func (r GetAppInfoResponse) Failed() error { return r.Err }
