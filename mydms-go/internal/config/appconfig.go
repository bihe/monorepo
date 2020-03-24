package config

// AppConfig holds the application configuration
type AppConfig struct {
	Security  Security
	Database  Database
	Logging   LogConfig
	Upload    UploadConfig
	Filestore FileStore
	Cors      CorsSettings
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

// UploadConfig defines relevant values for the upload logic
type UploadConfig struct {
	// AllowedFileTypes is a list of mime-types allowed to be uploaded
	AllowedFileTypes []string
	// MaxUploadSize defines the maximum permissible fiile-size
	MaxUploadSize int64
	// UploadPath defines a directory where uploaded files are stored
	UploadPath string
}

// FileStore holds configuration settings for the backend file store
type FileStore struct {
	Region string
	Bucket string
	Key    string
	Secret string
}

// CorsSettings specifies the used settings
type CorsSettings struct {
	Origins     []string
	Methods     []string
	Headers     []string
	Credentials bool
	MaxAge      int
}
