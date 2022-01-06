package oidc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.binggl.net/monorepo/internal/core/app/conf"
	"golang.binggl.net/monorepo/internal/core/app/shared"
	"golang.binggl.net/monorepo/internal/core/app/store"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.org/x/oauth2"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
)

// --------------------------------------------------------------------------
// definitions and consts
// --------------------------------------------------------------------------

var openIDScope = []string{oidc.ScopeOpenID, "profile", "email"}

var oidcClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	DisplayName   string `json:"name"`
	PicURL        string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
	UserID        string `json:"sub"`
}

// OIDCInitiateURL is used as a local hop to ensure that cookies are written to the local domain
const OIDCInitiateURL = "/oidc/redirect"

// OIDCInitiateRoutingPath is the routing path used for the logic
const OIDCInitiateRoutingPath = "/redirect"

// --------------------------------------------------------------------------
//   Service defintion
// --------------------------------------------------------------------------

// Service defines the logic of the OIDC interaction
type Service interface {
	// PrepIntOIDCRedirect returns the internal OIDC hop and a random state param
	PrepIntOIDCRedirect() (url string, state string)
	// GetExtOIDCRedirect returns the external auth provider url which is used to start the OIDC interaction
	GetExtOIDCRedirect(state string) (url string, err error)
	// LoginOIDC evaluates the returend data from the OIDC interaction and creates a token and redirect url
	LoginOIDC(state, oidcState, oidcCode string) (token, url string, err error)
	// LoginSiteOIDC performs a login via OIDC but checks if the user has permissions for the given site and redirects to the defined url
	LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error)
}

// --------------------------------------------------------------------------
//   OIDC wrapping
// --------------------------------------------------------------------------

const idTokenParam = "id_token"

// NewConfigAndVerifier creates the necessary configuration for OIDC
func NewConfigAndVerifier(c conf.OAuthConfig) (OIDCConfig, OIDCVerifier) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, c.Provider)
	if err != nil {
		panic(fmt.Sprintf("could not create a new OIDC provider: %v", err))
	}
	ver := provider.Verifier(&oidc.Config{
		ClientID:             c.ClientID,
		SupportedSigningAlgs: []string{"RS256", "HS256"},
	})
	oidcVer := &oidcVerifier{ver}

	var endPoint oauth2.Endpoint
	// either the endpoint is "constructed" by the supplied endpoint-URL
	// or the specific provider-endpoint is used
	if c.EndPointURL != "" {
		endPoint = oauth2.Endpoint{
			AuthURL:  c.EndPointURL + "/auth",
			TokenURL: c.EndPointURL + "/token",
		}
	} else {
		endPoint = provider.Endpoint()
	}
	oidcCfg := &oidcConfig{
		config: &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint:     endPoint,
			RedirectURL:  c.RedirectURL,
			Scopes:       openIDScope,
		},
	}
	return oidcCfg, oidcVer
}

// --------------------------------------------------------------------------

// OIDCConfig holds the underlying oauth config which is necessary for the OIDC process
type OIDCConfig interface {
	// AuthCodeURL returns a URL to OAuth 2.0 provider's consent page
	AuthCodeURL(state string) string
	// GetIDToken converts an authorization code into a token.
	GetIDToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (string, error)
}

var _ OIDCConfig = (*oidcConfig)(nil)

type oidcConfig struct {
	config *oauth2.Config
}

func (o *oidcConfig) AuthCodeURL(state string) string {
	return o.config.AuthCodeURL(state)
}

func (o *oidcConfig) GetIDToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (string, error) {
	oauth2Token, err := o.config.Exchange(ctx, code, opts...)
	if err != nil {
		return "", fmt.Errorf("could not get the token, error in OIDC interaction: %v", err)
	}
	rawIDToken, ok := oauth2Token.Extra(idTokenParam).(string)
	if !ok {
		return "", fmt.Errorf("OIDC processing error, no 'id_token' field in exchanged oauth2 token")
	}
	return rawIDToken, nil
}

// --------------------------------------------------------------------------

// OIDCVerifier wraps the underlying implementation of oidc-verify
type OIDCVerifier interface {
	VerifyToken(ctx context.Context, rawToken string) (OIDCToken, error)
}

var _ OIDCVerifier = (*oidcVerifier)(nil)

type oidcVerifier struct {
	*oidc.IDTokenVerifier
}

func (v *oidcVerifier) VerifyToken(ctx context.Context, rawToken string) (OIDCToken, error) {
	t, err := v.Verify(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	return &oidcToken{t}, nil
}

// --------------------------------------------------------------------------

// OIDCToken wraps the underlying implementation of oidc-token
type OIDCToken interface {
	GetClaims(v interface{}) error
}

var _ OIDCToken = (*oidcToken)(nil)

type oidcToken struct {
	*oidc.IDToken
}

// GetClaims returns the token claims
func (t *oidcToken) GetClaims(v interface{}) error {
	return t.Claims(v)
}

// --------------------------------------------------------------------------
//   Service implementation
// --------------------------------------------------------------------------

type oidcService struct {
	oauthConfig   OIDCConfig
	oauthVerifier OIDCVerifier
	repo          store.Repository
	jwtConfig     conf.Security
}

// compile guard
var _ Service = &oidcService{}

// New create a new instance of the Service
func New(oidcConfig OIDCConfig, oidcVerifier OIDCVerifier, jwtConfig conf.Security, repo store.Repository) Service {
	return &oidcService{
		oauthConfig:   oidcConfig,
		oauthVerifier: oidcVerifier,
		repo:          repo,
		jwtConfig:     jwtConfig,
	}
}

func (o *oidcService) PrepIntOIDCRedirect() (url string, state string) {
	url = OIDCInitiateURL
	state = randToken()
	return
}

func (o *oidcService) GetExtOIDCRedirect(state string) (url string, err error) {
	if state == "" {
		return "", shared.ErrValidation("invalid/empty state parameter supplied")
	}
	return o.oauthConfig.AuthCodeURL(state), nil
}

func (o *oidcService) LoginOIDC(state, oidcState, oidcCode string) (token, url string, err error) {
	err = validate(state, oidcState, oidcCode)
	if err != nil {
		return
	}
	return o.performOIDCLogin(state, oidcState, oidcCode, "", "")
}

func (o *oidcService) LoginSiteOIDC(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error) {
	err = validate(state, oidcState, oidcCode)
	if err != nil {
		return
	}
	if site == "" {
		err = shared.ErrValidation("empty 'site' parameter supplied")
		return
	}
	if redirectURL == "" {
		err = shared.ErrValidation("empty 'redirectURL' parameter supplied")
		return
	}
	return o.performOIDCLogin(state, oidcState, oidcCode, site, redirectURL)
}

func validate(state, oidcState, oidcCode string) (err error) {
	if state == "" {
		err = shared.ErrValidation("invalid/empty 'state' parameter supplied")
		return
	}
	if oidcState == "" {
		err = shared.ErrValidation("invalid/empty 'oidcState' parameter supplied")
		return
	}
	if oidcCode == "" {
		err = shared.ErrValidation("invalid/empty 'oidcCode' parameter supplied")
		return
	}
	if state != oidcState {
		err = shared.ErrValidation(fmt.Sprintf("the provided oidcState '%s' does not match the initial state '%s'", oidcState, state))
		return
	}
	return
}

func (o *oidcService) performOIDCLogin(state, oidcState, oidcCode, site, redirectURL string) (token, url string, err error) {

	var siteLogin bool
	if site != "" {
		siteLogin = true
	}

	// retrieve the ext auth provider token via the supplied code of the OIDC redirect from the auth provider
	// parse the token and finally extract claims from the token
	ctx := context.Background()
	rawIDToken, err := o.oauthConfig.GetIDToken(ctx, oidcCode)
	if err != nil {
		err = fmt.Errorf("could not get the token, error in OIDC interaction: %v", err)
		return
	}
	idToken, err := o.oauthVerifier.VerifyToken(ctx, rawIDToken)
	if err != nil {
		err = fmt.Errorf("OIDC processing error, failed to verify ID Token: %v", err)
		return
	}
	if err = idToken.GetClaims(&oidcClaims); err != nil {
		err = fmt.Errorf("could not get claims from token: %v", err)
		return
	}

	// the authentication part is done at this point. we have an externally authenticated user with some valid
	// claims. now we use those claims to check if the user is allowed (authorized) to access our system
	// the user is known throughout the system by it's email address, so we use the retrieved email-address from
	// the claims as the ID
	var sites []store.UserSiteEntity
	sites, err = o.repo.GetSitesForUser(oidcClaims.Email)
	if err != nil {
		err = shared.ErrSecurity(fmt.Sprintf("could not retrieve site-information for the given user '%s', %v", oidcClaims.Email, err))
		return
	}
	if len(sites) == 0 {
		// the user was loged-in successfully by the external auth provider. but by using the given Email
		// as the ID, we did not find ANY site which is associated to the ID, therefor the ID is not valid!
		err = shared.ErrSecurity(fmt.Sprintf("the given user '%s' is not allowed to access the system", oidcClaims.Email))
		return
	}

	// check if this is a site-login
	if siteLogin {
		var siteFound = false
		var siteURL string
		for _, e := range sites {
			if e.Name == site {
				siteFound = true
				siteURL = e.URL
				break
			}
		}

		if !siteFound {
			err = shared.ErrSecurity(fmt.Sprintf("user '%s' is not allowed to login", oidcClaims.Email))
			return
		}

		// validate the redirectURL
		if !strings.HasPrefix(redirectURL, siteURL) {
			err = shared.ErrSecurity(fmt.Sprintf("the provided redirectURL '%s' is not allowed for the given site", redirectURL))
			return
		}
	}

	// create the token using the claims of the database
	var siteClaims []string
	for _, s := range sites {
		siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", s.Name, s.URL, s.PermList))
	}

	// the google OIDC implementation returns a profile picture URL
	// this url can be used to retrieve images with a given size. The size information might be changed in the frontend
	// therefor the "size-param" is removed
	picUrl := ""
	sep := strings.Index(oidcClaims.PicURL, "=")
	fmt.Printf("sep: %d\n", sep)
	if sep > 0 {
		picUrl = oidcClaims.PicURL[0:sep]
	}
	claims := security.Claims{
		Type:        "login.User",
		DisplayName: oidcClaims.DisplayName,
		Email:       oidcClaims.Email,
		UserID:      oidcClaims.UserID,
		UserName:    oidcClaims.Email,
		GivenName:   oidcClaims.GivenName,
		Surname:     oidcClaims.FamilyName,
		ProfileURL:  picUrl,
		Claims:      siteClaims,
	}

	token, err = security.CreateToken(o.jwtConfig.JwtIssuer, []byte(o.jwtConfig.JwtSecret), o.jwtConfig.Expiry, claims)
	if err != nil {
		err = fmt.Errorf("could not create a JWT: %v", err)
		return
	}

	var t store.Logintype
	t = store.DIRECT
	if siteLogin {
		t = store.FLOW
	}
	err = o.repo.StoreLogin(store.LoginsEntity{
		User:      oidcClaims.Email,
		CreatedAt: time.Now().UTC(),
		Type:      t,
	})
	if err != nil {
		err = fmt.Errorf("could not store the login: %v", err)
		return
	}
	url = o.jwtConfig.LoginRedirect
	if siteLogin {
		url = redirectURL
	}
	return
}

func randToken() string {
	u := uuid.New()
	return u.String()
}
