package security

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

// User is the authenticated principal extracted from the JWT token
type User struct {
	Username      string
	Roles         []string
	Email         string
	UserID        string
	DisplayName   string
	Authenticated bool
	Token         string
}

// String represents the User struct as a string
func (u User) String() string {
	return fmt.Sprintf("%s (%s)", u.DisplayName, u.Username)
}

// JwtTokenPayload is the parsed contents of the given token
type JwtTokenPayload struct {
	Type        string
	UserName    string
	Email       string
	Claims      []string
	UserID      string `json:"UserId"`
	DisplayName string
	Surname     string
	GivenName   string
	jwt.StandardClaims
}

// Claims defines custom JWT claims for the token
type Claims struct {
	Type        string   `json:"Type"`
	DisplayName string   `json:"DisplayName"`
	Email       string   `json:"Email"`
	UserID      string   `json:"UserId"`
	UserName    string   `json:"UserName"`
	GivenName   string   `json:"GivenName"`
	Surname     string   `json:"Surname"`
	Claims      []string `json:"Claims"`
}

// JwtOptions defines presets for the Authentication handler
// by the default the JWT token is fetched from the Authentication header
// as a fallback it is possible to fetch the token from a specific cookie
type JwtOptions struct {
	// JwtSecret is the jwt signing key
	JwtSecret string
	// JwtIssuer specifies identifies the principal that issued the token
	JwtIssuer string
	// CookieName specifies the HTTP cookie holding the token
	CookieName string
	// RequiredClaim to access the application
	RequiredClaim Claim
	// RedirectURL forwards the request to an external authentication service
	RedirectURL string
	// CacheDuration defines the duration to cache the JWT token result
	CacheDuration string
	// ErrorPath is used if html errors are returned to the client
	ErrorPath string
}

// Claim defines the authorization requirements
type Claim struct {
	// Name of the application
	Name string
	// URL of the application
	URL string
	// Roles possible roles
	Roles []string
}
