package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Sec         Security           `json:"security"`
	DB          Database           `json:"database"`
	Log         LogConfig          `json:"logging"`
	OIDC        OAuthConfig        `json:"oidc"`
	AppCookies  ApplicationCookies `json:"appCookies"`
	Environment string             `json:"environment"`
	Cors        CorsSettings       `json:"cors"`
}

// Security settings for the application
type Security struct {
	JwtIssuer     string `json:"jwtIssuer"`
	JwtSecret     string `json:"jwtSecret"`
	Expiry        int    `json:"expiry"`
	CookieName    string `json:"cookieName"`
	CookieDomain  string `json:"cookieDomain"`
	CookiePath    string `json:"cookiePath"`
	CookieSecure  bool   `json:"cookieSecure"`
	Claim         Claim  `json:"claim"`
	CacheDuration string `json:"cacheDuration"`
	LoginRedirect string `json:"loginRedirect"`
}

// ApplicationCookies defines values for cookies
type ApplicationCookies struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
	Secure bool   `json:"secure"`
	Prefix string `json:"prefix"`
}

// Claim defines the required claims
type Claim struct {
	Name  string   `json:"name"`
	URL   string   `json:"url"`
	Roles []string `json:"roles"`
}

// Database defines the connection string
type Database struct {
	ConnStr string `json:"connectionString"`
}

// LogConfig is used to define settings for the logging process
type LogConfig struct {
	FilePath    string `json:"filePath"`
	RequestPath string `json:"requestPath"`
	LogLevel    string `json:"logLevel"`
}

// CorsSettings specifies the used settings
type CorsSettings struct {
	AllowedOrigins   []string `json:"origins"`
	AllowedMethods   []string `json:"methods"`
	AllowedHeaders   []string `json:"headers"`
	AllowCredentials bool     `json:"credentials"`
	MaxAge           int      `json:"maxAge"`
}

// OAuthConfig is used to configure OAuth OpenID Connect
type OAuthConfig struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"redirectURL"`
	Provider     string `json:"provider"`
}

// Read returns application configuration values
func Read(r io.Reader) (*AppConfig, error) {
	var (
		c    AppConfig
		cont []byte
		err  error
	)
	if cont, err = ioutil.ReadAll(r); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cont, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
