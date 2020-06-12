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
	Security          Security
	Database          Database
	Logging           LogConfig
	Cookies           ApplicationCookies
	Cors              CorsSettings
	ErrorPath         string
	Environment       Environment
	FaviconUploadPath string
	DefaultFavicon    string
	AppName           string
	HostID            string
}

// Security settings for the application
type Security struct {
	JwtIssuer     string
	JwtSecret     string
	CookieName    string
	LoginRedirect string
	Claim         Claim
	CacheDuration string
}

// Database defines the connection string
type Database struct {
	ConnectionString string
	Dialect          string
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
}

// ApplicationCookies defines values for cookies
type ApplicationCookies struct {
	Domain string
	Path   string
	Secure bool
	Prefix string
}

// CorsSettings specifies the used settings
type CorsSettings struct {
	Origins     []string
	Methods     []string
	Headers     []string
	Credentials bool
	MaxAge      int
}
