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
	Security    Security
	Database    Database
	Logging     LogConfig
	OIDC        OAuthConfig
	AppCookies  ApplicationCookies
	Environment Environment
	Cors        CorsSettings
	AppName     string
	HostID      string
}

// Security settings for the application
type Security struct {
	JwtIssuer     string
	JwtSecret     string
	Expiry        int
	CookieName    string
	CookieDomain  string
	CookiePath    string
	CookieSecure  bool
	Claim         Claim
	CacheDuration string
	LoginRedirect string
}

// ApplicationCookies defines values for cookies
type ApplicationCookies struct {
	Domain string
	Path   string
	Secure bool
	Prefix string
}

// Claim defines the required claims
type Claim struct {
	Name  string
	URL   string
	Roles []string
}

// Database defines the connection string
type Database struct {
	ConnectionString string
}

// LogConfig is used to define settings for the logging process
type LogConfig struct {
	FilePath string
	LogLevel string
}

// CorsSettings specifies the used settings
type CorsSettings struct {
	Origins     []string
	Methods     []string
	Headers     []string
	Credentials bool
	MaxAge      int
}

// OAuthConfig is used to configure OAuth OpenID Connect
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Provider     string
}
