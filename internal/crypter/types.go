// Package crypter implements an encryption services.
// the main purpose is to encrypt files / binary payloads.
// initial implementation focuses on PDF files
package crypter

import (
	"github.com/go-kit/kit/endpoint"
	"golang.binggl.net/monorepo/pkg/config"
)

// PayloadType enumerates the available payload types
type PayloadType string

const (
	// PDF is a payload of application/pdf
	PDF PayloadType = "PDF"
)

// --------------------------------------------------------------------------
// Define request/response types for Endpoints
// --------------------------------------------------------------------------

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = Response{}
)

// Request combines the parameters passed on to the encryption service
type Request struct {
	AuthToken string
	Payload   []byte
	Type      PayloadType
	InitPass  string
	NewPass   string
}

// Response collects the response values for the Encrypt method.
type Response struct {
	Payload []byte `json:"payload"`
	Err     error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

// Failed implements endpoint.Failer.
func (r Response) Failed() error { return r.Err }

// AppConfig defines configuration settings of the application
type AppConfig struct {
	config.BaseConfig
}
