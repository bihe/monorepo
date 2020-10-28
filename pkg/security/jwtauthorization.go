package security

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// JWTAuthorization handles authorizaton of supplied JWT tokens
type JWTAuthorization struct {
	Options JwtOptions
	Cache   *MemoryCache
}

// NewJWTAuthorization creates a new JWTAuthorization instance
// to speed up processing a cache in the form of a MemoryCache can be initilized
func NewJWTAuthorization(options JwtOptions, useCache bool) *JWTAuthorization {
	jwtAuth := &JWTAuthorization{}
	jwtAuth.Options = options
	if useCache {
		jwtAuth.Cache = NewMemCache(parseDuration(options.CacheDuration))
	}
	return jwtAuth
}

// EvaluateToken parses the supplied JWT token and extracts a User object
func (j *JWTAuthorization) EvaluateToken(token string) (*User, error) {
	var (
		err  error
		user *User
	)

	if j.Cache != nil {
		user = j.Cache.Get(token)
	}
	if user != nil {
		return user, nil
	}

	var payload JwtTokenPayload
	if payload, err = ParseJwtToken(token, j.Options.JwtSecret, j.Options.JwtIssuer); err != nil {
		return nil, fmt.Errorf("could not parse the JWT token: %v", err)
	}
	var roles []string
	claim := j.Options.RequiredClaim
	if roles, err = Authorize(Claim{Name: claim.Name, URL: claim.URL, Roles: claim.Roles}, payload.Claims); err != nil {
		return nil, fmt.Errorf("insufficient permissions to access the resource: %v", err)
	}

	user = &User{
		DisplayName:   payload.DisplayName,
		Email:         payload.Email,
		Roles:         roles,
		UserID:        payload.UserID,
		Username:      payload.UserName,
		Authenticated: true,
		Token:         token, // add the token to call other services which need auth!
	}
	if j.Cache != nil {
		j.Cache.Set(token, user)
	}
	return user, nil
}

// --------------------------------------------------------------------------
// Exported functions
// --------------------------------------------------------------------------

// Authorize validates the given claims and verifies if
// they match the required claim
// a claim entry is in the form "name|url|role"
func Authorize(required Claim, claims []string) (roles []string, err error) {
	for _, claim := range claims {
		c := split(claim)
		ok, _ := compareURL(required.URL, c.URL)
		if required.Name == c.Name && matchRole(c.Roles, required.Roles) && ok {
			return c.Roles, nil
		}
	}
	return roles, fmt.Errorf("supplied claims are not sufficient")
}

// ParseJwtToken parses, validates and extracts data from a jwt token
func ParseJwtToken(token, tokenSecret, issuer string) (JwtTokenPayload, error) {
	parsed, err := jwt.ParseWithClaims(token, &JwtTokenPayload{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return JwtTokenPayload{}, err
	}
	if parsed.Valid {
		claims, ok := parsed.Claims.(*JwtTokenPayload)
		if ok {
			if validIssuer := claims.StandardClaims.VerifyIssuer(issuer, true); !validIssuer {
				return JwtTokenPayload{}, fmt.Errorf("invalid issuer created the token")
			}
			return *claims, nil
		}
	}
	return JwtTokenPayload{}, fmt.Errorf("could not parse JWT token")
}

// CreateToken uses the configuration and supplied parameter to create a new token
func CreateToken(issuer string, key []byte, expiry int, c Claims) (string, error) {
	defaultExp := 7
	if expiry == 0 {
		expiry = defaultExp
	}
	now := time.Now().UTC()
	exp := now.Add(time.Duration(expiry*24) * time.Hour)
	id := uuid.New()

	type jwtClaims struct {
		jwt.StandardClaims
		Claims
	}

	claims := jwtClaims{
		jwt.StandardClaims{
			Id:        id.String(),
			IssuedAt:  now.Unix(),
			ExpiresAt: exp.Unix(),
			Issuer:    issuer,
			Subject:   c.Email,
		},
		Claims{
			Type:        c.Type,
			DisplayName: c.DisplayName,
			Email:       c.Email,
			UserID:      c.UserID,
			UserName:    c.UserName,
			GivenName:   c.GivenName,
			Surname:     c.Surname,
			Claims:      c.Claims,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func parseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(fmt.Sprintf("wrong value, cannot parse duration: %v", err))
	}
	return d
}

func matchRole(a []string, b []string) bool {
	for _, r := range a {
		for _, s := range b {
			if s == r {
				return true
			}
		}
	}
	return false
}

func split(claim string) *Claim {
	parts := strings.Split(claim, "|")
	if len(parts) == 3 {
		r := strings.Split(parts[2], ";")
		return &Claim{Name: parts[0], URL: parts[1], Roles: r}
	}
	return &Claim{}
}

func compareURL(a, b string) (bool, error) {
	var (
		urlA *url.URL
		urlB *url.URL
		err  error
	)
	if urlA, err = url.Parse(a); err != nil {
		return false, err
	}
	if urlB, err = url.Parse(b); err != nil {
		return false, err
	}
	if urlA.Scheme != urlB.Scheme || urlA.Port() != urlB.Port() || urlA.Host != urlB.Host {
		return false, fmt.Errorf("The urls do not match: '%s vs. %s'", urlA, urlB)
	}

	if normalizePath(urlA.Path) != normalizePath(urlB.Path) {
		return false, fmt.Errorf("The path of the urls does not match: '%s vs. %s'", urlA.Path, urlB.Path)
	}
	return true, nil
}

func normalizePath(path string) string {
	if path != "" {
		end := path[len(path)-1:]
		if end == "/" {
			return path[:len(path)-1]
		}
	}
	return path
}
