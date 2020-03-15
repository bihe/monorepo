package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const configString = `{
	"security": {
		"jwtIssuer": "issuer",
		"jwtSecret": "secret",
		"expiry": 7,
		"cookieName": "cookie",
		"cookieDomain": "example.com",
		"cookiePath": "/",
		"cookieSecure": true,
		"claim": {
            "name": "login",
            "url": "http://localhost:3000",
            "roles": ["User", "Admin"]
		},
		"cacheDuration": "10m",
		"loginRedirect": "http://localhost:3000"
	},
	"database": {
		"connectionString": "./bookmarks.db"
    	},
    	"logging": {
		"filePath": "/temp/file",
		"requestPath": "/temp/request",
		"logLevel": "debug"
	},
	"oidc": {
		"clientID": "clientID",
		"clientSecret": "clientSecret",
		"redirectURL": "redirectURL"
	},
	"appCookies": {
		"domain": "example.com",
		"path": "/",
		"secure": true,
		"prefix": "prefix"
	},
	"cors": {
		"origins": ["*"],
		"methods": ["GET", "POST"],
		"headers": ["Accept", "Authorization"],
		"credentials": true,
		"maxAge": 500
	},
	"environment": "Development"
}`

// TestConfigReader reads config settings from json
func TestConfigReader(t *testing.T) {
	reader := strings.NewReader(configString)
	config, err := Read(reader)
	if err != nil {
		t.Error("Could not read.", err)
	}

	assert.Equal(t, "issuer", config.Sec.JwtIssuer)
	assert.Equal(t, 7, config.Sec.Expiry)
	assert.Equal(t, "secret", config.Sec.JwtSecret)
	assert.Equal(t, "cookie", config.Sec.CookieName)
	assert.Equal(t, "example.com", config.Sec.CookieDomain)
	assert.Equal(t, "/", config.Sec.CookiePath)
	assert.Equal(t, true, config.Sec.CookieSecure)
	assert.Equal(t, "10m", config.Sec.CacheDuration)
	assert.Equal(t, "http://localhost:3000", config.Sec.LoginRedirect)
	assert.Equal(t, "login", config.Sec.Claim.Name)
	assert.Equal(t, "http://localhost:3000", config.Sec.Claim.URL)
	assert.Equal(t, []string{"User", "Admin"}, config.Sec.Claim.Roles)

	assert.Equal(t, "./bookmarks.db", config.DB.ConnStr)

	assert.Equal(t, "/temp/file", config.Log.FilePath)
	assert.Equal(t, "/temp/request", config.Log.RequestPath)
	assert.Equal(t, "debug", config.Log.LogLevel)

	assert.Equal(t, "clientID", config.OIDC.ClientID)
	assert.Equal(t, "clientSecret", config.OIDC.ClientSecret)
	assert.Equal(t, "redirectURL", config.OIDC.RedirectURL)

	assert.Equal(t, "example.com", config.AppCookies.Domain)
	assert.Equal(t, "/", config.AppCookies.Path)
	assert.Equal(t, "prefix", config.AppCookies.Prefix)
	assert.Equal(t, true, config.AppCookies.Secure)

	assert.Equal(t, "Development", config.Environment)

	assert.Equal(t, 500, config.Cors.MaxAge)
	assert.Equal(t, true, config.Cors.AllowCredentials)
	assert.Equal(t, []string{"Accept", "Authorization"}, config.Cors.AllowedHeaders)
	assert.Equal(t, []string{"GET", "POST"}, config.Cors.AllowedMethods)
	assert.Equal(t, []string{"*"}, config.Cors.AllowedOrigins)
}
