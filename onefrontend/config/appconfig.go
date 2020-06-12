// Package config defines the customization/configuration of the application
package config

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
	JWT            JwtSettings
	Logging        LogConfig
	Cookies        CookieSettings
	Cors           CorsSettings
	AssetDir       string
	AssetPrefix    string
	FrontendDir    string
	FrontendPrefix string
	ErrorPath      string
	StartURL       string
	Environment    Environment
	AppName        string
	HostID         string
}

// JwtSettings settings for the application
type JwtSettings struct {
	JwtIssuer     string
	JwtSecret     string
	CookieName    string
	LoginRedirect string
	Claim         Claim
	CacheDuration string
}

// Claim defines the required claims
type Claim struct {
	Name  string
	URL   string
	Roles []string
}

// CookieSettings define how cookies should be written
type CookieSettings struct {
	Domain string
	Path   string
	Secure bool
}

// LogConfig is used to define settings for the logging process
type LogConfig struct {
	FilePath    string
	RequestPath string
	LogLevel    string
}

// CorsSettings specifies the used settings
type CorsSettings struct {
	Origins     []string
	Methods     []string
	Headers     []string
	Credentials bool
	MaxAge      int
}
