// Package config defines the customization/configuration of the application
package conf

import "golang.binggl.net/monorepo/pkg/config"

// AppConfig holds the application configuration
type AppConfig struct {
	config.BaseConfig
	Database          Database
	FaviconUploadPath string
	DefaultFavicon    string
}

// Database defines the connection string
type Database struct {
	ConnectionString string
	Dialect          string
}
