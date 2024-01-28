package config

// Environment specifies operation modes
type Environment string

const (
	// Development is used in development
	Development Environment = "Development"
	// Production is used for deployments
	Production Environment = "Production"
	// Production is a specific environment used for testing
	Integration Environment = "Integration"
)

// BaseConfig holds the application configuration
type BaseConfig struct {
	Security    Security
	Logging     LogConfig
	Cors        CorsSettings
	Environment Environment
	Cookies     ApplicationCookies
	Assets      AssetSettings
	AppName     string
	HostID      string
	ErrorPath   string
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

// CorsSettings specifies the used settings
type CorsSettings struct {
	Origins     []string
	Methods     []string
	Headers     []string
	Credentials bool
	MaxAge      int
}

// ApplicationCookies defines values for cookies
type ApplicationCookies struct {
	Domain string
	Path   string
	Secure bool
	Prefix string
}

// AssetSettings define the prefix-name and path to application assets
type AssetSettings struct {
	AssetDir    string
	AssetPrefix string
}
