package oidc

import (
	"context"
	"fmt"
	"time"

	"golang.binggl.net/monorepo/internal/gway/app/conf"
	"golang.binggl.net/monorepo/internal/gway/app/store"
	"golang.binggl.net/monorepo/pkg/security"

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
const OIDCInitiateURL = "/redirect-oidc"

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
// interface implementation
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
		return "", fmt.Errorf("invalid/empty state parameter supplied")
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
		err = fmt.Errorf("empty 'site' parameter supplied")
		return
	}
	if redirectURL == "" {
		err = fmt.Errorf("empty 'redirectURL' parameter supplied")
		return
	}
	return o.performOIDCLogin(state, oidcState, oidcCode, site, redirectURL)
}

func validate(state, oidcState, oidcCode string) (err error) {
	if state == "" {
		err = fmt.Errorf("invalid/empty 'state' parameter supplied")
		return
	}
	if oidcState == "" {
		err = fmt.Errorf("invalid/empty 'oidcState' parameter supplied")
		return
	}
	if oidcCode == "" {
		err = fmt.Errorf("invalid/empty 'oidcCode' parameter supplied")
		return
	}
	if state != oidcState {
		err = fmt.Errorf("the provided oidcState '%s' does not match the initial state '%s'", oidcState, state)
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
		err = fmt.Errorf("could not retrieve site-information for the given user '%s', %v", oidcClaims.Email, err)
		return
	}
	if len(sites) == 0 {
		// the user was loged-in successfully by the external auth provider. but by using the given Email
		// as the ID, we did not find ANY site which is associated to the ID, therefor the ID is not valid!
		err = fmt.Errorf("the given user '%s' is not allowed to access the system", oidcClaims.Email)
		return
	}

	// check if this is a site-login
	if siteLogin {
		var siteFound = false
		for _, e := range sites {
			if e.Name == site {
				siteFound = true
				break
			}
		}

		if !siteFound {
			err = fmt.Errorf("user '%s' is not allowed to login", oidcClaims.Email)
			return
		}
	}

	// create the token using the claims of the database
	var siteClaims []string
	for _, s := range sites {
		siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", s.Name, s.URL, s.PermList))
	}
	claims := security.Claims{
		Type:        "login.User",
		DisplayName: oidcClaims.DisplayName,
		Email:       oidcClaims.Email,
		UserID:      oidcClaims.UserID,
		UserName:    oidcClaims.Email,
		GivenName:   oidcClaims.GivenName,
		Surname:     oidcClaims.FamilyName,
		Claims:      siteClaims,
	}

	token, err = security.CreateToken(o.jwtConfig.JwtIssuer, []byte(o.jwtConfig.JwtSecret), o.jwtConfig.Expiry, claims)
	if err != nil {
		err = fmt.Errorf("could not create a JWT: %v", err)
		return
	}

	err = o.repo.InUnitOfWork(func(repo store.Repository) error {
		var t store.Logintype
		t = store.DIRECT
		if siteLogin {
			t = store.FLOW
		}
		return repo.StoreLogin(store.LoginsEntity{
			User:      oidcClaims.Email,
			CreatedAt: time.Now().UTC(),
			Type:      t,
		})
	})
	if err != nil {
		err = fmt.Errorf("could not store the the login: %v", err)
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
