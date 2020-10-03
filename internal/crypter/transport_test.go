package crypter

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/proto"
)

var payload = make([]byte, 1)
var grpcRequest = proto.CrypterRequest{
	AuthToken:    "TOKEN",
	InitPassword: "INIT",
	NewPassword:  "NEW",
	Payload:      payload,
	Type:         proto.CrypterRequest_PDF,
}

type HandlerMock struct {
	fail bool
}

func (h *HandlerMock) ServeGRPC(ctx context.Context, request interface{}) (context.Context, interface{}, error) {
	if h.fail {
		return context.TODO(), nil, fmt.Errorf("error")
	}
	return context.TODO(), &proto.CrypterResponse{
		Message: "",
		Result:  proto.CrypterResponse_SUCCESS,
		Payload: payload,
	}, nil
}

func Test_Transport_Encrypt(t *testing.T) {
	srv := &grpcServer{
		encrypt: &HandlerMock{},
	}

	grpcResponse, err := srv.Encrypt(context.TODO(), &grpcRequest)
	if err != nil {
		t.Fatalf("could not encrypt: %v", err)
	}
	assert.Equal(t, payload, grpcResponse.Payload)
	assert.Equal(t, proto.CrypterResponse_SUCCESS, grpcResponse.Result)
}

func Test_Transport_Encrypt_Fail(t *testing.T) {
	srv := &grpcServer{
		encrypt: &HandlerMock{fail: true},
	}

	_, err := srv.Encrypt(context.TODO(), &grpcRequest)
	if err == nil {
		t.Fatalf("error expected")
	}
}

func Test_Transport_Decode_FromGrpcRequest(t *testing.T) {
	dec, err := decodeGRPCEncryptRequest(context.TODO(), &grpcRequest)
	if err != nil {
		t.Fatalf("error decoding request: %v", err)
	}
	decoded := dec.(Request)
	assert.Equal(t, "TOKEN", decoded.AuthToken)
	assert.Equal(t, "INIT", decoded.InitPass)
	assert.Equal(t, "NEW", decoded.NewPass)
	assert.Equal(t, payload, decoded.Payload)
	assert.Equal(t, PDF, decoded.Type)
}

func Test_Transport_Decode_FromGrpcResponse(t *testing.T) {
	grpcResponse := proto.CrypterResponse{
		Message: "MESSAGE",
		Payload: payload,
		Result:  proto.CrypterResponse_SUCCESS,
	}

	dec, err := decodeGRPCEncryptResponse(context.TODO(), &grpcResponse)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	decoded := dec.(Response)
	assert.Nil(t, decoded.Err)
	assert.Equal(t, payload, decoded.Payload)
}

func Test_Transport_Encode_ToGrpcRequest(t *testing.T) {
	request := Request{
		AuthToken: "TOKEN",
		InitPass:  "INIT",
		NewPass:   "NEW",
		Payload:   payload,
		Type:      PDF,
	}

	enc, err := encodeGRPCEncrytpRequest(context.TODO(), request)
	if err != nil {
		t.Fatalf("error encoding request: %v", err)
	}
	encoded := enc.(*proto.CrypterRequest)
	assert.Equal(t, "TOKEN", encoded.AuthToken)
	assert.Equal(t, "INIT", encoded.InitPassword)
	assert.Equal(t, "NEW", encoded.NewPassword)
	assert.Equal(t, payload, encoded.Payload)
	assert.Equal(t, proto.CrypterRequest_PDF, encoded.Type)
}

func Test_Transport_Encode_ToGrpcResponse(t *testing.T) {
	response := Response{
		Payload: payload,
		Err:     nil,
	}

	enc, err := encodeGRPCEncryptResponse(context.TODO(), response)
	if err != nil {
		t.Fatalf("error encoding response: %v", err)
	}
	encoded := enc.(*proto.CrypterResponse)
	assert.Equal(t, payload, encoded.Payload)
	assert.Equal(t, proto.CrypterResponse_SUCCESS, encoded.Result)
}
