package security // import "golang.binggl.net/commons/security"

// User is the authenticated principal extracted from the JWT token
type User struct {
	Username      string
	Roles         []string
	Email         string
	UserID        string
	DisplayName   string
	Authenticated bool
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
