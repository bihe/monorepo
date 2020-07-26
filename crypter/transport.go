package crypter

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"golang.binggl.net/monorepo/proto"
	"google.golang.org/grpc"
)

type grpcServer struct {
	encrypt grpctransport.Handler
}

// NewGRPCServer makes endpoints available as a gRPC CrypterServer.
func NewGRPCServer(endpoints Endpoints, logger log.Logger) proto.CrypterServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	return &grpcServer{
		encrypt: grpctransport.NewServer(
			endpoints.CrypterEndpoint,
			decodeGRPCEncryptRequest,
			encodeGRPCEncryptResponse,
			options...,
		),
	}
}

// NewGRPCClient returns a EncryptionService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) EncryptionService {
	// global client middlewares
	var options []grpctransport.ClientOption

	// Each individual endpoint is an grpc/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var crypterEndpoint endpoint.Endpoint
	{
		crypterEndpoint = grpctransport.NewClient(
			conn,
			"proto.Crypter",
			"Encrypt",
			encodeGRPCEncrytpRequest,
			decodeGRPCEncryptResponse,
			proto.CrypterResponse{},
			options...,
		).Endpoint()

	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return Endpoints{
		CrypterEndpoint: crypterEndpoint,
	}
}

// Encrypt implements the method of the proto definition
func (s *grpcServer) Encrypt(ctx context.Context, req *proto.CrypterRequest) (*proto.CrypterResponse, error) {
	_, rep, err := s.encrypt.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*proto.CrypterResponse), nil
}

// --------------------------------------------------------------------------
// translate between grpc objects and domain objects
// --------------------------------------------------------------------------

// decodeGRPCEncryptRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC encrypt request to a user-domain encrypt request. Primarily useful in a server.
func decodeGRPCEncryptRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*proto.CrypterRequest)
	return Request{
		AuthToken: req.AuthToken,
		Payload:   req.Payload,
		Type:      translateProtoEnum(req.Type),
		InitPass:  req.InitPassword,
		NewPass:   req.NewPassword,
	}, nil
}

// decodeGRPCEncryptResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC encrypt respone to a user-domain encrypt response. Primarily useful in a client.
func decodeGRPCEncryptResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	resp := grpcReply.(*proto.CrypterResponse)
	return Response{
		Payload: resp.Payload,
		Err:     responseToError(resp),
	}, nil
}

// encodeGRPCEncrytpRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain encrypt request to a gRPC encrypt request. Primarily useful in a client.
func encodeGRPCEncrytpRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(Request)
	return &proto.CrypterRequest{
		AuthToken:    req.AuthToken,
		Payload:      req.Payload,
		Type:         translateDomainEnum(req.Type),
		InitPassword: req.InitPass,
		NewPassword:  req.NewPass,
	}, nil
}

// encodeGRPCEncryptResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain encrypt response to a gRPC encrypt response. Primarily useful in a server.
func encodeGRPCEncryptResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(Response)

	var result proto.CrypterResponse_OperationResult
	result = proto.CrypterResponse_SUCCESS
	msg := ""
	if resp.Err != nil {
		result = proto.CrypterResponse_ERROR
		msg = resp.Err.Error()
	}

	return &proto.CrypterResponse{
		Payload: resp.Payload,
		Result:  result,
		Message: msg,
	}, nil
}

func translateProtoEnum(t proto.CrypterRequest_PayloadType) PayloadType {
	switch t {
	case proto.CrypterRequest_PDF:
		return PDF
	}
	// default
	return PDF
}

func translateDomainEnum(t PayloadType) proto.CrypterRequest_PayloadType {
	switch t {
	case PDF:
		return proto.CrypterRequest_PDF
	}
	// default
	return proto.CrypterRequest_PDF
}

func responseToError(resp *proto.CrypterResponse) error {
	switch resp.Result {
	case proto.CrypterResponse_ERROR:
		return fmt.Errorf(resp.Message)
	}
	return nil
}
