package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const configYAML = `---
security:
  jwtIssuer: login.binggl.net
  jwtSecret: secret
  cookieName: login_token
  loginRedirect: https://login.url.com
  claim:
    name: bookmarks
    url: http://localhost:3000
    roles:
    - User
    - Admin
  cacheDuration: 10m

database:
  connectionString: "./bookmarks.db"
  dialect: mysql

logging:
  filePath: "/temp/file"
  requestPath: "/temp/request"
  logLevel: debug

cookies:
  domain: example.com
  path: "/"
  secure: true
  prefix: prefix

cors:
  origins:
  - "*"
  methods:
  - "GET"
  - "POST"
  headers:
  - "Accept"
  - "Authorization"
  credentials: true
  maxAge: 500

errorPath: error
environment: Development
faviconUploadPath: "./faviconpath"
defaultFavicon: "./favicon.ico"
`

// TestConfigReader reads config settings from json
func TestConfigReader(t *testing.T) {
	reader := strings.NewReader(configYAML)
	config, err := GetSettings(reader)
	if err != nil {
		t.Error("Could not read.", err)
	}
	assert.Equal(t, "./bookmarks.db", config.DB.ConnStr)
	assert.Equal(t, "mysql", config.DB.Dialect)

	assert.Equal(t, "https://login.url.com", config.Sec.LoginRedirect)
	assert.Equal(t, "bookmarks", config.Sec.Claim.Name)
	assert.Equal(t, "secret", config.Sec.JwtSecret)
	assert.Equal(t, "10m", config.Sec.CacheDuration)

	assert.Equal(t, "/temp/file", config.Log.FilePath)
	assert.Equal(t, "/temp/request", config.Log.RequestPath)
	assert.Equal(t, "debug", config.Log.LogLevel)

	assert.Equal(t, "example.com", config.Cookies.Domain)
	assert.Equal(t, "/", config.Cookies.Path)
	assert.Equal(t, "prefix", config.Cookies.Prefix)
	assert.Equal(t, true, config.Cookies.Secure)

	assert.Equal(t, "Development", config.Environment)
	assert.Equal(t, "error", config.ErrorPath)
	assert.Equal(t, "./faviconpath", config.FaviconPath)
	assert.Equal(t, "./favicon.ico", config.DefaultFavicon)

	assert.Equal(t, 500, config.Cors.MaxAge)
	assert.Equal(t, true, config.Cors.AllowCredentials)
	assert.Equal(t, []string{"Accept", "Authorization"}, config.Cors.AllowedHeaders)
	assert.Equal(t, []string{"GET", "POST"}, config.Cors.AllowedMethods)
	assert.Equal(t, []string{"*"}, config.Cors.AllowedOrigins)
}
