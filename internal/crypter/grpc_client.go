package crypter

import (
	"context"
	"fmt"

	"golang.binggl.net/monorepo/proto"
	"google.golang.org/grpc"
)

// NewClient creates a new instance to interact with the EncryptionService
func NewClient(con *grpc.ClientConn) EncryptionService {
	return &GrpcClient{
		client: proto.NewCrypterClient(con),
	}
}

// GrpcClient implments the EncryptionService
type GrpcClient struct {
	client proto.CrypterClient
}

func (g *GrpcClient) Encrypt(ctx context.Context, req Request) ([]byte, error) {
	var pType proto.CrypterRequest_PayloadType
	switch req.Type {
	case PDF:
		pType = proto.CrypterRequest_PDF
	}

	res, err := g.client.Encrypt(ctx, &proto.CrypterRequest{
		AuthToken:    req.AuthToken,
		Type:         pType,
		NewPassword:  req.NewPass,
		InitPassword: req.InitPass,
		Payload:      req.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("the call to the 'Encrypt' method returned error: %v", err)
	}
	if res.Result == proto.CrypterResponse_ERROR {
		return nil, fmt.Errorf("the operation resulted in an error-outcome: '%s'", res.Message)
	}
	return res.Payload, nil
}
