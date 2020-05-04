package server

import (
	"context"
	"fmt"

	"github.com/bihe/monorepo/filecrypt/internal"
	"github.com/bihe/monorepo/filecrypt/proto"
	"golang.binggl.net/commons"

	log "github.com/sirupsen/logrus"
)

// FileCrypter implements the RPC service: Crypter
type FileCrypter struct {
	Logger        *log.Entry
	BasePath      string
	TokenSettings internal.TokenSecurity
}

// Encrypt takes a CrypterRequest and encrypts the payload according to the rules of the supplied data
// the result of the operation is a CrypterResponse
func (f *FileCrypter) Encrypt(ctxt context.Context, req *proto.CrypterRequest) (*proto.CrypterResponse, error) {
	if req.AuthToken == "" {
		commons.LogWith(f.Logger, "server.Encrypt").Infof("missing authToken!")
		return nil, fmt.Errorf("no authentication token supplied")
	}

	commons.LogWith(f.Logger, "server.Encrypt").Infof("processing Encrypt request")
	return &proto.CrypterResponse{
		Payload: req.Payload,
		Result:  proto.CrypterResponse_SUCCESS,
		Message: "result",
	}, nil
}
