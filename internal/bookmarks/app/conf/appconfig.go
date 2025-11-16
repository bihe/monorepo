// Package config defines the customization/configuration of the application
package conf

import "golang.binggl.net/monorepo/pkg/config"

// AppConfig holds the application configuration
type AppConfig struct {
	config.BaseConfig
	Database          Database
	FaviconUploadPath string
	DefaultFavicon    string
	Upload            UploadSettings
}

// Database defines the connection string
type Database struct {
	ConnectionString string
	Dialect          string
}

// UploadSettings defines relevant values for the upload logic
type UploadSettings struct {
	// AllowedFileTypes is a list of mime-types allowed to be uploaded
	AllowedFileTypes []string
	// MaxUploadSize defines the maximum permissible file-size
	MaxUploadSize int64
	// UploadPath defines a directory where uploaded files are stored
	UploadPath string
}
