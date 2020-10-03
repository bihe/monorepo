// Package crypter implements an encryption services.
// the main purpose is to encrypt files / binary payloads.
// initial implementation focuses on PDF files
package crypter

import "github.com/go-kit/kit/endpoint"

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

// --------------------------------------------------------------------------
// Configuration settings
// --------------------------------------------------------------------------

// Environment specifies operation modes
type Environment string

const (
	// Development is used in development
	Development Environment = "Development"
	// Production is used for deployments
	Production Environment = "Production"
)

// AppConfig holds the application configuration
type AppConfig struct {
	TokenSecurity TokenSecurity
	Logging       LogConfig
	Environment   Environment
	ServiceName   string
	HostID        string
}

// TokenSecurity settings for the application
type TokenSecurity struct {
	JwtIssuer     string
	JwtSecret     string
	Claim         Claim
	CacheDuration string
}

// Claim defines the required claims
type Claim struct {
	Name  string
	URL   string
	Roles []string
}

// LogConfig is used to define settings for the logging process
type LogConfig struct {
	FilePath string
	LogLevel string
	// GrayLogServer defines the address of a log-aggregator using Graylog
	GrayLogServer string
}
