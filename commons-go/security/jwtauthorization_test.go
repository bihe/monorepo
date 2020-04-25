package security // import "golang.binggl.net/commons/security"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const failed = "Authorization failed."
const shouldFail = "Authorization should fail."
const testToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4NzAzOTI2NDcsImp0aSI6IjZmYWQ1YzAwLWZlZTItNDU5Yy1hYmFkLTIwNDU3Y2ZmM2Q4YSIsImlhdCI6MTU1OTc4Nzg0NywiaXNzIjoiaXNzdWVyIiwic3ViIjoidXNlciIsIlR5cGUiOiJsb2dpbi5Vc2VyIiwiRGlzcGxheU5hbWUiOiJEaXNwbGF5IE5hbWUiLCJFbWFpbCI6ImEuYkBjLmRlIiwiVXNlcklkIjoiMTIzNDUiLCJVc2VyTmFtZSI6ImEuYkBjLmRlIiwiR2l2ZW5OYW1lIjoiRGlzcGxheSIsIlN1cm5hbWUiOiJOYW1lIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3Q6MzAwMHxyb2xlIl19.qUwvHXBmV_FuwLtykOnzu3AMbxSqrg82bQlAi3Nabyo"

/*
{
  "exp": 1870392647,
  "jti": "6fad5c00-fee2-459c-abad-20457cff3d8a",
  "iat": 1559787847,
  "iss": "issuer",
  "sub": "user",
  "Type": "login.User",
  "DisplayName": "Display Name",
  "Email": "a.b@c.de",
  "UserId": "12345",
  "UserName": "a.b@c.de",
  "GivenName": "Display",
  "Surname": "Name",
  "Claims": [
    "claim|http://localhost:3000|role"
  ]
}
*/

func TestJWTAuthorization(t *testing.T) {
	var jwtOpts = JwtOptions{
		JwtSecret:  "secret",
		JwtIssuer:  "issuer",
		CookieName: cookie,
		RequiredClaim: Claim{
			Name:  "claim",
			URL:   "http://localhost:3000",
			Roles: []string{"role"},
		},
		RedirectURL:   "/redirect",
		CacheDuration: "10m",
	}
	var jwtAuth = NewJWTAuthorization(jwtOpts, false)

	user, err := jwtAuth.EvaluateToken(testToken)
	if err != nil {
		t.Errorf("could not parse JWT token: %v", err)
	}

	assert.Equal(t, "Display Name", user.DisplayName)
	assert.Equal(t, "a.b@c.de", user.Email)
	assert.Equal(t, true, user.Authenticated)
	assert.Equal(t, "12345", user.UserID)
	assert.Equal(t, "a.b@c.de", user.Username)
	assert.Equal(t, []string{"role"}, user.Roles)

	// wrong token
	_, err = jwtAuth.EvaluateToken("testTken")
	if err == nil {
		t.Errorf("the token should be invalid")
	}

	// wrong claims
	jwtOpts.RequiredClaim = Claim{
		Name:  "claim",
		URL:   "http://localhost:3000",
		Roles: []string{"role1"},
	}
	jwtAuth = NewJWTAuthorization(jwtOpts, false)
	_, err = jwtAuth.EvaluateToken(testToken)
	if err == nil {
		t.Errorf("should have authorization errror")
	}
}

func TestJWTAuthorizationCache(t *testing.T) {
	var jwtOpts = JwtOptions{
		JwtSecret:  "secret",
		JwtIssuer:  "issuer",
		CookieName: cookie,
		RequiredClaim: Claim{
			Name:  "claim",
			URL:   "http://localhost:3000",
			Roles: []string{"role"},
		},
		RedirectURL:   "/redirect",
		CacheDuration: "10m",
	}
	var jwtAuth = NewJWTAuthorization(jwtOpts, true)

	user, err := jwtAuth.EvaluateToken(testToken)
	if err != nil {
		t.Errorf("could not parse JWT token: %v", err)
	}

	assert.Equal(t, "Display Name", user.DisplayName)
	assert.Equal(t, "a.b@c.de", user.Email)
	assert.Equal(t, true, user.Authenticated)
	assert.Equal(t, "12345", user.UserID)
	assert.Equal(t, "a.b@c.de", user.Username)
	assert.Equal(t, []string{"role"}, user.Roles)

	// second call using the cache
	user, err = jwtAuth.EvaluateToken(testToken)
	if err != nil {
		t.Errorf("could not parse JWT token: %v", err)
	}

	assert.Equal(t, "Display Name", user.DisplayName)
	assert.Equal(t, "a.b@c.de", user.Email)
	assert.Equal(t, true, user.Authenticated)
	assert.Equal(t, "12345", user.UserID)
	assert.Equal(t, "a.b@c.de", user.Username)
	assert.Equal(t, []string{"role"}, user.Roles)
}

func TestAuthorization(t *testing.T) {
	reqClaim := Claim{Name: "test", URL: "http://a.b.c.de/f", Roles: []string{"user", "admin"}}
	var claims []string

	claims = append(claims, "a|http://1.com|nothing")
	claims = append(claims, "test|http://a.b.c.de/f|user")

	if _, err := Authorize(reqClaim, claims); err != nil {
		t.Error(failed, err)
	}

	claims = claims[0:1]
	if _, err := Authorize(reqClaim, claims); err == nil {
		t.Error(shouldFail, err)
	}

	claims = nil
	claims = append(claims, "a|http://1.com|nothing")
	claims = append(claims, "test|http://a.b.c.de/f|admin")

	if _, err := Authorize(reqClaim, claims); err != nil {
		t.Error(failed, err)
	}

	claims = []string{"test|http://a.b.c.de/nomatch|role"}
	if _, err := Authorize(reqClaim, claims); err == nil {
		t.Error(shouldFail, err)
	}

	claims = []string{"test|NOURLä|role"}
	if _, err := Authorize(reqClaim, claims); err == nil {
		t.Error(shouldFail, err)
	}

	claims = []string{"test|http://a.b.c.de|user"}
	reqClaim = Claim{Name: "test", URL: "http://a.b.c.de/", Roles: []string{"user"}}
	if _, err := Authorize(reqClaim, claims); err != nil {
		t.Error(failed, err)
	}
}

func TestJwtTokenParsing(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI3NzAyNmYxMWM2ODQ0YzVlOWEyYmM1YWNkNGQxYjljNyIsImlhdCI6MTU0MDcyODAxMiwiaXNzIjoibG9naW4uYS5iLmMuZGUiLCJleHAiOjE3NDEzMzI4MTIsInN1YiI6ImhlbnJpa0BhLmIuYy5kZSIsIlR5cGUiOiJsb2dpbi5Vc2VyIiwiVXNlck5hbWUiOiJ1c2VyLm5hbWUiLCJFbWFpbCI6ImhlbnJpa0BhLmIuYy5kZSIsIkNsYWltcyI6WyJib29rbWFya3N8aHR0cDovL2xvY2FsaG9zdDozMDAwfEFkbWluIl0sIlVzZXJJZCI6IjExMSIsIkRpc3BsYXlOYW1lIjoiVXNlcjEgTmFtZSIsIlN1cm5hbWUiOiJOYW1lIiwiR2l2ZW5OYW1lIjoiVXNlcjEifQ.GzR2RAEYyTazO6zaDcCWdSlvaso3TUW-Tem3-l0qnnw"
	secret := "secret"

	// valid token
	_, err := ParseJwtToken(token, secret, "login.a.b.c.de")
	if err != nil {
		t.Errorf("Could not parse jwt token %s", err)
	}

	// issue
	_, err = ParseJwtToken(token, secret, "wrong issued")
	if err == nil {
		t.Error("The issuer check should work!", err)
	}

	// expired token
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiI3NzAyNmYxMWM2ODQ0YzVlOWEyYmM1YWNkNGQxYjljNyIsImlhdCI6MTU0MDcyODAxMiwiaXNzIjoibG9naW4uYS5iLmMuZGUiLCJleHAiOjEzNDEzMzI4MTIsInN1YiI6ImhlbnJpa0BhLmIuYy5kZSIsIlR5cGUiOiJsb2dpbi5Vc2VyIiwiVXNlck5hbWUiOiJ1c2VyLm5hbWUiLCJFbWFpbCI6ImhlbnJpa0BhLmIuYy5kZSIsIkNsYWltcyI6WyJib29rbWFya3N8aHR0cDovL2xvY2FsaG9zdDozMDAwfEFkbWluIl0sIlVzZXJJZCI6IjExMSIsIkRpc3BsYXlOYW1lIjoiVXNlcjEgTmFtZSIsIlN1cm5hbWUiOiJOYW1lIiwiR2l2ZW5OYW1lIjoiVXNlcjEifQ.Tet0Umd7D6pB3URb5G-rGuZBTKpR3kQqD49tDQBzw3Q"
	_, err = ParseJwtToken(token, secret, "login.a.b.c.de")
	if err == nil {
		t.Error("The token was expired!", err)
	}
}

func TestJWTCreation(t *testing.T) {
	c := Claims{
		Type:        "type",
		DisplayName: "displayName",
		Email:       "a.b@c.de",
		UserID:      "1",
		UserName:    "a",
		GivenName:   "g",
		Surname:     "s",
		Claims: []string{
			"a|http://example.com|User;Admin",
			"b|http://example.net|User",
		},
	}

	token, err := CreateToken("issuer", []byte("secretsecretsecretsecret"), 1, c)
	if err != nil {
		t.Errorf("could not create token: %v", err)
	}
	if token == "" {
		t.Errorf("cannot get a token!")
	}
}
