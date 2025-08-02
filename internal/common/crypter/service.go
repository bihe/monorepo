// Package crypter implements an encryption services.
// the main purpose is to encrypt files and strings
package crypter

import (
	"context"
	"fmt"

	"golang.binggl.net/monorepo/pkg/logging"
)

// --------------------------------------------------------------------------
// Types definition
// --------------------------------------------------------------------------

// PayloadType enumerates the available payload types
type PayloadType string

const (
	// PDF is a payload of application/pdf
	PDF PayloadType = "PDF"
	// String payload to be used for encryption/decryption
	String PayloadType = "String"
)

// Request combines the parameters passed on to the encryption service
type Request struct {
	// Payload which is used for encryption or decryption
	Payload []byte
	// Type of the payload which influences the encryption used
	Type PayloadType
	// Password to use for the encryption/decryption operation
	Password string
	// InitPass if the payload is encrypted and the password should be changed
	InitPass string
}

// --------------------------------------------------------------------------
// EncryptionService definition
// --------------------------------------------------------------------------

// EncryptionService describes a service that encrypts payloads
type EncryptionService interface {
	// Encrypt performs the encryption of the provided payload
	Encrypt(ctx context.Context, req Request) ([]byte, error)
	// Decrypt uses the provided data to perform a decryption of the payload
	Decrypt(ctx context.Context, req Request) ([]byte, error)
}

// NewService returns a Service with all of the expected middlewares wired in.
func NewService(logger logging.Logger) EncryptionService {
	svc := &encryptionSvc{
		logger: logger,
	}
	return svc

}

// --------------------------------------------------------------------------
// Service implementation
// --------------------------------------------------------------------------

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ EncryptionService = &encryptionSvc{}
)

// implement the Service
type encryptionSvc struct {
	logger logging.Logger
}

// Encrypt performs the encryption operation for different file-types.
func (e *encryptionSvc) Encrypt(ctx context.Context, req Request) ([]byte, error) {

	if req.Password == "" {
		return nil, fmt.Errorf("cannot encrypt with empty password")
	}
	if len(req.Payload) == 0 {
		return nil, fmt.Errorf("no payload supplied")
	}

	// encrypt the payload per respective type
	switch req.Type {
	case PDF:
		return encryptPdfPayload(req.Payload, req.InitPass, req.Password)
	case String:
		return encryptAES(string(req.Payload), req.Password)
	}
	return nil, fmt.Errorf("the provided PayloadType '%s' is not supported for the operation 'Encrypt'", req.Type)
}

// Decrypt peforms decryption on the provided request
func (e *encryptionSvc) Decrypt(ctx context.Context, req Request) ([]byte, error) {
	if req.Password == "" {
		return nil, fmt.Errorf("cannot encrypt with empty password")
	}
	if len(req.Payload) == 0 {
		return nil, fmt.Errorf("no payload supplied")
	}

	// encrypt the payload per respective type
	switch req.Type {
	case String:
		return decryptAES(string(req.Payload), req.Password)
	}
	return nil, fmt.Errorf("the provided PayloadType '%s' is not supported for the operation 'Encrypt'", req.Type)
}
