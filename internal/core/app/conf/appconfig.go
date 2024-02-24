package conf

import "golang.binggl.net/monorepo/pkg/config"

// AppConfig holds the application configuration
type AppConfig struct {
	config.BaseConfig
	Database Database
	Security Security
	OIDC     OAuthConfig
}

// Database defines the connection string
type Database struct {
	ConnectionString string
}

// Security settings for the application
type Security struct {
	JwtIssuer     string
	JwtSecret     string
	CookieName    string
	Expiry        int
	Claim         config.Claim
	CacheDuration string
	LoginRedirect string
}

// OAuthConfig is used to configure OAuth OpenID Connect
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Provider     string
	EndPointURL  string
}
