package security

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// RoleDelimiter specifies the element used to separate a list of roles
const RoleDelimiter = ";"

// keys used for the custom jwt-claims
const jwtType = "Type"
const userName = "UserName"
const displayName = "DisplayName"
const email = "Email"
const userID = "UserId"
const surname = "Surname"
const givenName = "GivenName"
const profileURL = "ProfileURL"
const claims = "Claims"

const jwtKeyAlgorithm = "HS256"

// JWTAuthorization handles authorization of supplied JWT tokens
type JWTAuthorization struct {
	Options JwtOptions
	Cache   *MemoryCache
}

// NewJWTAuthorization creates a new JWTAuthorization instance
// to speed up processing a cache in the form of a MemoryCache can be initialized
func NewJWTAuthorization(options JwtOptions, useCache bool) *JWTAuthorization {
	jwtAuth := &JWTAuthorization{}
	jwtAuth.Options = options
	if useCache {
		var duration = options.CacheDuration
		if duration == "" {
			duration = "10m"
		}
		jwtAuth.Cache = NewMemCache(parseDuration(duration))
	}
	return jwtAuth
}

// EvaluateToken parses the supplied JWT token and extracts a User object
func (j *JWTAuthorization) EvaluateToken(token string) (*User, error) {
	var (
		err       error
		user      *User
		roles     []string
		allClaims []Claim
		claim     Claim
		payload   JwtTokenPayload
	)

	if j.Cache != nil {
		user = j.Cache.Get(token)
	}
	if user != nil {
		return user, nil
	}

	if payload, err = ParseJwtToken(token, j.Options.JwtSecret, j.Options.JwtIssuer); err != nil {
		return nil, fmt.Errorf("could not parse the JWT token: %v", err)
	}
	claim = j.Options.RequiredClaim
	if roles, allClaims, err = Authorize(Claim{Name: claim.Name, URL: claim.URL, Roles: claim.Roles}, payload.Claims); err != nil {
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
		ProfileURL:    payload.ProfileURL,
		Claims:        allClaims, // add all existing claims to the user, the roles only specify the required/requested roles
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
func Authorize(required Claim, claims []string) (roles []string, allClaims []Claim, err error) {
	for _, claim := range claims {
		c := split(claim)
		allClaims = append(allClaims, c)
		ok, _ := compareURL(required.URL, c.URL)
		if required.Name == c.Name && matchRole(c.Roles, required.Roles) && ok {
			roles = append(roles, c.Roles...)
		}
	}
	if len(roles) == 0 {
		err = fmt.Errorf("supplied claims are not sufficient")
	}
	return
}

// ParseJwtToken parses, validates and extracts data from a jwt token
func ParseJwtToken(token, tokenSecret, issuer string) (JwtTokenPayload, error) {
	algo, err := jwa.KeyAlgorithmFrom(jwtKeyAlgorithm)
	if err != nil {
		return JwtTokenPayload{}, err
	}

	t, err := jwt.Parse([]byte(token), jwt.WithVerify(true), jwt.WithValidate(true), jwt.WithKey(algo, []byte(tokenSecret)))
	if err != nil {
		return JwtTokenPayload{}, err
	}
	if err = jwt.Validate(t, jwt.WithIssuer(issuer)); err != nil {
		return JwtTokenPayload{}, err
	}

	jwtId, _ := t.JwtID()
	jwtSubject, _ := t.Subject()
	jwtIssuer, _ := t.Issuer()
	jwtExpiry, _ := t.Expiration()
	jwtIssuedAt, _ := t.IssuedAt()

	return JwtTokenPayload{
		Type:        getTokenValueString(t, jwtType),
		UserName:    getTokenValueString(t, userName),
		Email:       getTokenValueString(t, email),
		UserID:      getTokenValueString(t, userID),
		DisplayName: getTokenValueString(t, displayName),
		Surname:     getTokenValueString(t, surname),
		GivenName:   getTokenValueString(t, givenName),
		ProfileURL:  getTokenValueString(t, profileURL),
		Claims:      getTokenValueSlice(t, claims),
		StandardClaims: StandardClaims{
			ID:        jwtId,
			Subject:   jwtSubject,
			Issuer:    jwtIssuer,
			ExpiresAt: jwtExpiry.Unix(),
			IssuedAt:  jwtIssuedAt.Unix(),
		},
	}, nil
}

// CreateToken uses the configuration and supplied parameter to create a new token
func CreateToken(issuer string, key []byte, expiry int, c Claims) (string, error) {
	algo, err := jwa.KeyAlgorithmFrom(jwtKeyAlgorithm)
	if err != nil {
		return "", err
	}

	defaultExp := 7
	if expiry == 0 {
		expiry = defaultExp
	}
	now := time.Now().UTC()
	exp := now.Add(time.Duration(expiry*24) * time.Hour)
	id := uuid.New()

	// use jwx to create a token
	t := jwt.New()
	// std. claims
	t.Set(jwt.JwtIDKey, id.String())
	t.Set(jwt.SubjectKey, c.Email)
	t.Set(jwt.IssuerKey, issuer)
	t.Set(jwt.ExpirationKey, exp.Unix())
	t.Set(jwt.IssuedAtKey, now.Unix())
	// custom claims
	t.Set(jwtType, c.Type)
	t.Set(displayName, c.DisplayName)
	t.Set(email, c.Email)
	t.Set(userID, c.UserID)
	t.Set(userName, c.UserName)
	t.Set(givenName, c.GivenName)
	t.Set(surname, c.Surname)
	t.Set(profileURL, c.ProfileURL)
	t.Set(claims, c.Claims)

	payload, err := jwt.Sign(t, jwt.WithKey(algo, key))
	if err != nil {
		return "", fmt.Errorf("could not create jwt token: %w", err)
	}
	return string(payload), nil
}

func getTokenValueString(t jwt.Token, key string) string {
	var value string
	err := t.Get(key, &value)
	if err == nil {
		return value
	}
	return ""
}

func getTokenValueSlice(t jwt.Token, key string) []string {
	var result = make([]string, 0)
	var s = make([]any, 0)
	err := t.Get(key, &s)
	if err == nil {
		for _, item := range s {
			result = append(result, item.(string))
		}

		return result
	}
	return result
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

func split(claim string) Claim {
	parts := strings.Split(claim, "|")
	if len(parts) == 3 {
		r := strings.Split(parts[2], ";")
		return Claim{Name: parts[0], URL: parts[1], Roles: r}
	}
	return Claim{}
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
		return false, fmt.Errorf("the urls do not match: '%s vs. %s'", urlA, urlB)
	}

	if normalizePath(urlA.Path) != normalizePath(urlB.Path) {
		return false, fmt.Errorf("the path of the urls does not match: '%s vs. %s'", urlA.Path, urlB.Path)
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
