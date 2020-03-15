// Package config defines the customization/configuration of the application
package config

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// AppConfig holds the application configuration
type AppConfig struct {
	JWT            JwtSettings    `yaml:"jwt"`
	Log            LogConfig      `yaml:"logging"`
	Cookies        CookieSettings `yaml:"cookies"`
	Cors           CorsSettings   `yaml:"cors"`
	AssetDir       string         `yaml:"assetDir"`
	AssetPrefix    string         `yaml:"assetPrefix"`
	FrontendDir    string         `yaml:"frontendDir"`
	FrontendPrefix string         `yaml:"frontendPrefix"`
	ErrorPath      string         `yaml:"errorPath"`
	StartURL       string         `yaml:"startUrl"`
	Environment    string         `yaml:"environment"`
}

// JwtSettings settings for the application
type JwtSettings struct {
	JwtIssuer     string `yaml:"jwtIssuer"`
	JwtSecret     string `yaml:"jwtSecret"`
	CookieName    string `yaml:"cookieName"`
	LoginRedirect string `yaml:"loginRedirect"`
	Claim         Claim  `yaml:"claim"`
	CacheDuration string `yaml:"cacheDuration"`
}

// Claim defines the required claims
type Claim struct {
	Name  string   `yaml:"name"`
	URL   string   `yaml:"url"`
	Roles []string `yaml:"roles"`
}

// CookieSettings define how cookies should be written
type CookieSettings struct {
	Domain string `yaml:"domain"`
	Path   string `yaml:"path"`
	Secure bool   `yaml:"secure"`
}

// LogConfig is used to define settings for the logging process
type LogConfig struct {
	FilePath    string `yaml:"filePath"`
	RequestPath string `yaml:"requestPath"`
	LogLevel    string `yaml:"logLevel"`
}

// CorsSettings specifies the used settings
type CorsSettings struct {
	AllowedOrigins   []string `yaml:"origins"`
	AllowedMethods   []string `yaml:"methods"`
	AllowedHeaders   []string `yaml:"headers"`
	AllowCredentials bool     `yaml:"credentials"`
	MaxAge           int      `yaml:"maxAge"`
}

// GetSettings returns application configuration values
func GetSettings(r io.Reader) (*AppConfig, error) {
	var (
		c    AppConfig
		cont []byte
		err  error
	)
	if cont, err = ioutil.ReadAll(r); err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(cont, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
