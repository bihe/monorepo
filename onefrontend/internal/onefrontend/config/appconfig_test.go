package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const configYAML = `---
jwt:
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

cookies:
  domain: dot.com
  path: /path
  secure: true

logging:
  filePath: "/temp/file"
  requestPath: "/temp/request"
  logLevel: debug

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

assetDir: "./assets"
assetPrefix: "prefix"

frontendDir: "./ui"
frontendPrefix: "ui"

errorPath: error
startUrl: http://url
environment: Development
`

// TestConfigReader reads config settings from json
func TestConfigReader(t *testing.T) {
	reader := strings.NewReader(configYAML)
	config, err := GetSettings(reader)
	if err != nil {
		t.Error("Could not read.", err)
	}

	assert.Equal(t, "https://login.url.com", config.JWT.LoginRedirect)
	assert.Equal(t, "bookmarks", config.JWT.Claim.Name)
	assert.Equal(t, "secret", config.JWT.JwtSecret)
	assert.Equal(t, "10m", config.JWT.CacheDuration)

	assert.Equal(t, "/temp/file", config.Log.FilePath)
	assert.Equal(t, "/temp/request", config.Log.RequestPath)
	assert.Equal(t, "debug", config.Log.LogLevel)

	assert.Equal(t, "http://url", config.StartURL)
	assert.Equal(t, "Development", config.Environment)
	assert.Equal(t, "error", config.ErrorPath)

	assert.Equal(t, "./assets", config.AssetDir)
	assert.Equal(t, "prefix", config.AssetPrefix)

	assert.Equal(t, "./ui", config.FrontendDir)
	assert.Equal(t, "ui", config.FrontendPrefix)

	assert.Equal(t, "dot.com", config.Cookies.Domain)
	assert.Equal(t, "/path", config.Cookies.Path)
	assert.Equal(t, true, config.Cookies.Secure)

	assert.Equal(t, 500, config.Cors.MaxAge)
	assert.Equal(t, true, config.Cors.AllowCredentials)
	assert.Equal(t, []string{"Accept", "Authorization"}, config.Cors.AllowedHeaders)
	assert.Equal(t, []string{"GET", "POST"}, config.Cors.AllowedMethods)
	assert.Equal(t, []string{"*"}, config.Cors.AllowedOrigins)
}
