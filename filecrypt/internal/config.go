package internal

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
}
