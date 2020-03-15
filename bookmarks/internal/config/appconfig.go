// Package config defines the customization/configuration of the application
package config

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Sec            Security           `yaml:"security"`
	DB             Database           `yaml:"database"`
	Log            LogConfig          `yaml:"logging"`
	Cookies        ApplicationCookies `yaml:"cookies"`
	Cors           CorsSettings       `yaml:"cors"`
	ErrorPath      string             `yaml:"errorPath"`
	Environment    string             `yaml:"environment"`
	FaviconPath    string             `yaml:"faviconUploadPath"`
	DefaultFavicon string             `yaml:"defaultFavicon"`
}

// Security settings for the application
type Security struct {
	JwtIssuer     string `yaml:"jwtIssuer"`
	JwtSecret     string `yaml:"jwtSecret"`
	CookieName    string `yaml:"cookieName"`
	LoginRedirect string `yaml:"loginRedirect"`
	Claim         Claim  `yaml:"claim"`
	CacheDuration string `yaml:"cacheDuration"`
}

// Database defines the connection string
type Database struct {
	ConnStr string `yaml:"connectionString"`
	Dialect string `yaml:"dialect"`
}

// Claim defines the required claims
type Claim struct {
	Name  string   `yaml:"name"`
	URL   string   `yaml:"url"`
	Roles []string `yaml:"roles"`
}

// LogConfig is used to define settings for the logging process
type LogConfig struct {
	FilePath    string `yaml:"filePath"`
	RequestPath string `yaml:"requestPath"`
	LogLevel    string `yaml:"logLevel"`
}

// ApplicationCookies defines values for cookies
type ApplicationCookies struct {
	Domain string `yaml:"domain"`
	Path   string `yaml:"path"`
	Secure bool   `yaml:"secure"`
	Prefix string `yaml:"prefix"`
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
