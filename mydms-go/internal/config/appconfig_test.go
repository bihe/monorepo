package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const configString = `{
    "security": {
        "jwtIssuer": "login.binggl.net",
        "jwtSecret": "secret",
	"cookieName": "login_token",
	"loginRedirect": "https://login.url.com",
        "claim": {
            "name": "bookmarks",
            "url": "http://localhost:3000",
            "roles": ["User", "Admin"]
	},
	"cacheDuration": "10m"
    },
    "database": {
	"connectionString": "./bookmarks.db"
    },
    "upload": {
        "allowedFileTypes": ["pdf","png"],
        "maxUploadSize": 1000,
        "UploadPath": "/PATH"
    },
    "logging": {
	"filePath": "/temp/file",
	"logLevel": "debug"
    },
    "filestore": {
        "region": "_REGION_",
        "bucket": "_BUCKET_NAME_",
        "key": "key",
        "secret": "secret"
    },
    "cors": {
	"origins": ["*"],
	"methods": ["GET", "POST"],
	"headers": ["Accept", "Authorization"],
	"credentials": true,
	"maxAge": 500
    }
}`

// TestConfigReader reads config settings from json
func TestConfigReader(t *testing.T) {
	reader := strings.NewReader(configString)
	config, err := GetSettings(reader)
	if err != nil {
		t.Error("Could not read.", err)
	}

	assert.Equal(t, "./bookmarks.db", config.DB.ConnStr)

	assert.Equal(t, "https://login.url.com", config.Sec.LoginRedirect)
	assert.Equal(t, "bookmarks", config.Sec.Claim.Name)
	assert.Equal(t, "secret", config.Sec.JwtSecret)
	assert.Equal(t, "10m", config.Sec.CacheDuration)

	assert.Equal(t, int64(1000), config.UP.MaxUploadSize)
	assert.Equal(t, "/PATH", config.UP.UploadPath)
	assert.Equal(t, 2, len(config.UP.AllowedFileTypes))

	assert.Equal(t, "/temp/file", config.Log.FilePath)
	assert.Equal(t, "debug", config.Log.LogLevel)

	assert.Equal(t, "_REGION_", config.Store.Region)
	assert.Equal(t, "_BUCKET_NAME_", config.Store.Bucket)
	assert.Equal(t, "_BUCKET_NAME_", config.Store.Bucket)
	assert.Equal(t, "key", config.Store.Key)
	assert.Equal(t, "secret", config.Store.Secret)

	assert.Equal(t, 500, config.Cors.MaxAge)
	assert.Equal(t, true, config.Cors.AllowCredentials)
	assert.Equal(t, []string{"Accept", "Authorization"}, config.Cors.AllowedHeaders)
	assert.Equal(t, []string{"GET", "POST"}, config.Cors.AllowedMethods)
	assert.Equal(t, []string{"*"}, config.Cors.AllowedOrigins)

}
