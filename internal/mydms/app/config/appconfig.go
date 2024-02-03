package config

import "golang.binggl.net/monorepo/pkg/config"

// AppConfig holds the application configuration
type AppConfig struct {
	config.BaseConfig
	Database  Database
	Upload    UploadConfig
	Filestore FileStore
}

// Database defines the connection string
type Database struct {
	ConnectionString string
}

// UploadConfig defines relevant values for the upload logic
type UploadConfig struct {
	// EndpointURL defines the upload endpoint api URL
	EndpointURL string
}

// FileStore holds configuration settings for the backend file store
type FileStore struct {
	Region string
	Bucket string
	Key    string
	Secret string
}
