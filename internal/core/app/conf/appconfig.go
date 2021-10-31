package conf

import "golang.binggl.net/monorepo/pkg/config"

// AppConfig holds the application configuration
type AppConfig struct {
	config.BaseConfig
	Database Database
	Security Security
	OIDC     OAuthConfig
	Upload   UploadSettings
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

// UploadSettings defines relevant values for the upload logic
type UploadSettings struct {
	// AllowedFileTypes is a list of mime-types allowed to be uploaded
	AllowedFileTypes []string
	// MaxUploadSize defines the maximum permissible fiile-size
	MaxUploadSize int64
	// UploadPath defines a directory where uploaded files are stored
	UploadPath string
	// EncGrpcConn defines a connection to a encryption GRPC service
	EncGrpcConn string
}
