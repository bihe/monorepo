package crypter

import (
	"context"
	"fmt"

	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/proto"
)

// GrpcServer is the implementation of proto.CrypterServer
type GrpcServer struct {
	proto.UnimplementedCrypterServer
	logger  logging.Logger
	service EncryptionService
}

// NewServer returns a new instance of the crypter.GrpcServer
func NewServer(svc EncryptionService, logger logging.Logger) *GrpcServer {
	return &GrpcServer{
		logger:  logger,
		service: svc,
	}
}

// Encrypt is the grpc method to encrypt a payload
func (g *GrpcServer) Encrypt(ctx context.Context, req *proto.CrypterRequest) (*proto.CrypterResponse, error) {
	encReq := Request{
		AuthToken: req.AuthToken,
		Payload:   req.Payload,
		Type:      translateType(req.Type),
		InitPass:  req.InitPassword,
		NewPass:   req.NewPassword,
	}

	enc, err := g.service.Encrypt(ctx, encReq)
	if err != nil {
		g.logger.Error("call to 'Encrypt' returned error", logging.ErrV(err))
		return nil, fmt.Errorf("error calling method 'Encrypt'; %v", err)
	}
	return &proto.CrypterResponse{
		Payload: enc,
		Result:  proto.CrypterResponse_SUCCESS,
		Message: "",
	}, nil

}

func translateType(t proto.CrypterRequest_PayloadType) PayloadType {
	var pt PayloadType
	switch t {
	case proto.CrypterRequest_PDF:
		pt = PDF
	}
	return pt
}
