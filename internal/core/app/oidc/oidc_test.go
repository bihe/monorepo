package oidc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/oidc"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.org/x/oauth2"
)

var oauthConfig = conf.OAuthConfig{
	ClientID:     "client",
	ClientSecret: "secret",
	RedirectURL:  "/redirect",
	Provider:     "https://accounts.google.com",
}

var jwtConfig = conf.Security{
	JwtIssuer:  "issuer",
	JwtSecret:  "secret",
	CookieName: "cookie",
	Expiry:     500,
	Claim: config.Claim{
		Name:  "A",
		URL:   "http://urlA",
		Roles: []string{"a", "b"},
	},
	LoginRedirect: "http://redirect",
}

const adminRole = "ADMIN"

var Err = fmt.Errorf("error")
var userEmail = "a.b@c.de"
var oicdTestSites []store.UserSiteEntity = []store.UserSiteEntity{
	{
		Name:      "A",
		User:      userEmail,
		URL:       "http://urlA",
		PermList:  adminRole + ",a,b,c",
		CreatedAt: time.Now(),
	},
}

func withMockRepo(sites map[string][]store.UserSiteEntity) store.Repository {
	return store.NewMock(sites)
}

func newMockRepo() store.Repository {
	svc := make(map[string][]store.UserSiteEntity)
	svc[userEmail] = oicdTestSites
	return store.NewMock(svc)
}

func newOIDCConfig() (oidc.OIDCConfig, oidc.OIDCVerifier) {
	return oidc.NewConfigAndVerifier(oauthConfig)
}

func withOIDCConfig(oauthConfig conf.OAuthConfig) (oidc.OIDCConfig, oidc.OIDCVerifier) {
	return oidc.NewConfigAndVerifier(oauthConfig)
}

func Test_Panic_InvalidConfig(t *testing.T) {
	assert.Panics(t, func() {
		c, v := withOIDCConfig(conf.OAuthConfig{
			ClientID:     "client",
			ClientSecret: "secret",
			RedirectURL:  "/redirect",
			Provider:     "-1",
		})
		svc := oidc.New(c, v, jwtConfig, newMockRepo())
		_, _ = svc.PrepIntOIDCRedirect()
	})
}

func Test_PrepIntOIDCRedirect(t *testing.T) {
	c, v := newOIDCConfig()
	svc := oidc.New(c, v, jwtConfig, newMockRepo())

	url, state := svc.PrepIntOIDCRedirect()
	assert.Equal(t, oidc.OIDCInitiateURL, url)
	assert.True(t, state != "")
}

func Test_GetExtOIDCRedirect(t *testing.T) {
	c, v := newOIDCConfig()
	svc := oidc.New(c, v, jwtConfig, newMockRepo())
	state := "123456"

	url, err := svc.GetExtOIDCRedirect(state)
	assert.NoError(t, err)
	assert.Contains(t, url, "state="+state)

	// empty state produces an error
	_, err = svc.GetExtOIDCRedirect("")
	assert.Error(t, err)
}

// --------------------------------------------------------------------------
// mocking of OAUTH logic
// --------------------------------------------------------------------------

// the test-server is used so that the oidc implementation has an endpoint to interact with
// based on the paths the desired data is returned, which is later on used
func setupMockOAuthServer() (*httptest.Server, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		// Should return authorization code back to the user
		fmt.Println("mockOAuth --> /auth")
	})

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	id_token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE4MTYyMzkwMjIsImlzcyI6Imh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbSIsImF1ZCI6IkNMSUVOVF9JRCJ9.liGIFlDU_usFWxdgqpHuyr_DXSyWNSwiP1iwVd4QkJY"
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// Should return acccess token back to the user
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte(fmt.Sprintf("access_token=%s&scope=user&token_type=bearer&id_token=%s", token, id_token)))
		fmt.Println("mockOAuth --> /token")
	})

	server := httptest.NewServer(mux)

	return server, func() {
		server.Close()
	}
}

// mock the configuration and verifier for the test-cases
func newMockOIDCConfigAndVerifier(c conf.OAuthConfig, url string) (oidc.OIDCConfig, oidc.OIDCVerifier) {
	oidcVer := &mockVerifier{}
	oidcCfg := &mockConfig{
		config: &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			RedirectURL:  c.RedirectURL,
			Scopes:       []string{"email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  url + "/auth",
				TokenURL: url + "/token",
			},
		},
	}
	return oidcCfg, oidcVer
}

// --------------------------------------------------------------------------

type mockConfig struct {
	config *oauth2.Config
	fail   bool
}

func (o *mockConfig) AuthCodeURL(state string) string {
	return o.config.AuthCodeURL(state)
}

func (o *mockConfig) GetIDToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (string, error) {
	if o.fail {
		return "", Err
	}
	return "rawIdToken", nil
}

// --------------------------------------------------------------------------

type mockVerifier struct {
	fail      bool
	failToken bool
}

func (v *mockVerifier) VerifyToken(ctx context.Context, rawToken string) (oidc.OIDCToken, error) {
	if v.fail {
		return nil, Err
	}
	return &mockToken{v.failToken}, nil
}

// --------------------------------------------------------------------------

type mockToken struct {
	fail bool
}

func (t *mockToken) GetClaims(v interface{}) error {
	if t.fail {
		return Err
	}

	claims := `{
  "email": "%s",
  "email_verified": true,
  "name": "NAME",
  "picture": "PICTURE",
  "given_name": "GIVEN_NAME",
  "family_name": "FAMILY_NAME",
  "locale": "en",
  "sub": "1"
}`
	return json.Unmarshal([]byte(fmt.Sprintf(claims, userEmail)), v)
}

// --------------------------------------------------------------------------

func Test_OIDCLogin_Mock_Full(t *testing.T) {
	// arrange
	testSrv, closeSrv := setupMockOAuthServer()
	defer func() {
		closeSrv()
	}()
	c, v := newMockOIDCConfigAndVerifier(conf.OAuthConfig{
		ClientID:     "CLIENT_ID",
		ClientSecret: "CLIENT_SECRET",
		RedirectURL:  "REDIRECT_URL",
	}, testSrv.URL)

	repo := withMockRepo(map[string][]store.UserSiteEntity{
		userEmail: oicdTestSites,
	})

	// act
	svc := oidc.New(c, v, jwtConfig, repo)
	token, redirect, err := svc.LoginOIDC("123456", "123456", "code")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, jwtConfig.LoginRedirect, redirect)
	assert.True(t, len(token) > 1)

	// check validation errors
	_, _, err = svc.LoginOIDC("", "123456", "code")
	assert.Error(t, err)

	_, _, err = svc.LoginOIDC("123456", "", "code")
	assert.Error(t, err)

	_, _, err = svc.LoginOIDC("123456", "123456", "")
	assert.Error(t, err)

	_, _, err = svc.LoginOIDC("1", "2", "3")
	assert.Error(t, err)

	// no user
	svc = oidc.New(c, v, jwtConfig, withMockRepo(map[string][]store.UserSiteEntity{
		"A": {
			{
				Name:      "A",
				User:      userEmail,
				URL:       "http://urlA",
				PermList:  adminRole + ",a,b,c",
				CreatedAt: time.Now(),
			},
		},
	}))
	_, _, err = svc.LoginOIDC("123456", "123456", "code")
	assert.Error(t, err)

	// OIDC process errors - GetIDToken
	mockC, _ := c.(*mockConfig)
	mockC.fail = true
	svc = oidc.New(mockC, v, jwtConfig, newMockRepo())
	_, _, err = svc.LoginOIDC("123456", "123456", "code")
	assert.Error(t, err)

	// OIDC process errors - VerifyToken
	mockC.fail = false
	mockV, _ := v.(*mockVerifier)
	mockV.fail = true
	svc = oidc.New(c, mockV, jwtConfig, newMockRepo())
	_, _, err = svc.LoginOIDC("123456", "123456", "code")
	assert.Error(t, err)

	// // OIDC process errors - GetClaims
	mockC.fail = false
	mockV.fail = false
	mockV.failToken = true
	svc = oidc.New(c, mockV, jwtConfig, newMockRepo())
	_, _, err = svc.LoginOIDC("123456", "123456", "code")
	assert.Error(t, err)
}

func Test_OIDCLogin_Mock(t *testing.T) {
	// arrange
	testSrv, closeSrv := setupMockOAuthServer()
	defer func() {
		closeSrv()
	}()
	c, _ := oidc.NewConfigAndVerifier(conf.OAuthConfig{
		ClientID:     "CLIENT_ID",
		ClientSecret: "CLIENT_SECRET",
		RedirectURL:  "REDIRECT_URL",
		EndPointURL:  testSrv.URL,
		Provider:     "https://accounts.google.com",
	})
	v := &mockVerifier{}

	repo := withMockRepo(map[string][]store.UserSiteEntity{
		userEmail: oicdTestSites,
	})

	// act
	svc := oidc.New(c, v, jwtConfig, repo)
	token, redirect, err := svc.LoginOIDC("123456", "123456", "code")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, jwtConfig.LoginRedirect, redirect)
	assert.True(t, len(token) > 1)
}

func Test_OIDCSiteLogin_Mock(t *testing.T) {
	// arrange
	testSrv, closeSrv := setupMockOAuthServer()
	defer func() {
		closeSrv()
	}()
	c, v := newMockOIDCConfigAndVerifier(conf.OAuthConfig{
		ClientID:     "CLIENT_ID",
		ClientSecret: "CLIENT_SECRET",
		RedirectURL:  "REDIRECT_URL",
	}, testSrv.URL)

	repo := withMockRepo(map[string][]store.UserSiteEntity{
		userEmail: oicdTestSites,
	})

	// act
	svc := oidc.New(c, v, jwtConfig, repo)
	token, redirect, err := svc.LoginSiteOIDC("123456", "123456", "code", "A", "http://urlA/redirect")

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "http://urlA/redirect", redirect)
	assert.True(t, len(token) > 1)

	// wrong site
	_, _, err = svc.LoginSiteOIDC("123456", "123456", "code", "B", "https://redirectB")
	assert.Error(t, err)

	// wrong redirect
	_, _, err = svc.LoginSiteOIDC("123456", "123456", "code", "A", "http://urlB/redirect")
	assert.Error(t, err)

	// check validation errors
	_, _, err = svc.LoginSiteOIDC("", "123456", "code", "", "")
	assert.Error(t, err)

	_, _, err = svc.LoginSiteOIDC("123456", "123456", "code", "", "")
	assert.Error(t, err)

	_, _, err = svc.LoginSiteOIDC("123456", "123456", "code", "A", "")
	assert.Error(t, err)
}
