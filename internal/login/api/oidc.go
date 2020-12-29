package api

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.binggl.net/monorepo/internal/login/config"
	"golang.binggl.net/monorepo/internal/login/persistence"
	"golang.binggl.net/monorepo/pkg/cookies"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/security"
	"golang.org/x/oauth2"

	"github.com/coreos/go-oidc/v3/oidc"

	per "golang.binggl.net/monorepo/pkg/persistence"
)

// --------------------------------------------------------------------------
// OIDC wrapping
// --------------------------------------------------------------------------

// OIDCToken wraps the underlying implementation of oidc-token
type OIDCToken interface {
	GetClaims(v interface{}) error
}

type oidcToken struct {
	*oidc.IDToken
}

// GetClaims returns the token claims
func (t *oidcToken) GetClaims(v interface{}) error {
	return t.Claims(v)
}

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

// NewOIDC creates a config and a verifier for the OIDC interaction
func NewOIDC(c config.OAuthConfig) (oauthConfig *oauth2.Config, oauthVerifier OIDCVerifier) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, c.Provider)
	if err != nil {
		panic(fmt.Sprintf("could not create a new OIDC provider: %v", err))
	}
	oidcConfig := &oidc.Config{
		ClientID: c.ClientID,
	}
	ver := provider.Verifier(oidcConfig)
	oauthVerifier = &oidcVerifier{ver}
	oauthConfig = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  c.RedirectURL,
		Scopes:       openIDScope,
	}
	return
}

// --------------------------------------------------------------------------
// constants and defintions
// --------------------------------------------------------------------------

const stateParam = "state"
const codeParam = "code"
const idTokenParam = "id_token"
const siteParam = "~site"
const redirectParam = "~url"
const authFlowCookie = "auth_flow"
const authFlowSep = "|"

var openIDScope = []string{oidc.ScopeOpenID, "profile", "email"}

const templateDir = "templates"
const errorTemplate = "error.tmpl"

const cookieExpiry = 60

// oidcInitiateURL is used as a local hop to ensure that cookies are written to the local domain
const oidcInitiateURL = "/redirect-oidc"

const errorPath = "/error"

// HandleError returns a HTML template showing errors
func (a *loginAPI) HandleError(w http.ResponseWriter, r *http.Request) error {
	cookie := cookies.NewAppCookie(a.cookieSettings)
	var (
		msg       string
		err       string
		isError   bool
		isMessage bool
		init      sync.Once
		tmpl      *template.Template
		e         error
	)

	init.Do(func() {
		path := filepath.Join(a.basePath, templateDir, errorTemplate)
		tmpl, e = template.ParseFiles(path)
	})
	if e != nil {
		return e
	}

	// read (flash)
	err = cookie.Get(errors.FlashKeyError, r)
	if err != "" {
		isError = true
	}
	msg = cookie.Get(errors.FlashKeyInfo, r)
	if msg != "" {
		isMessage = true
	}

	// clear (flash)
	cookie.Del(errors.FlashKeyError, w)
	cookie.Del(errors.FlashKeyInfo, w)

	data := map[string]interface{}{
		"year":      time.Now().Year(),
		"appname":   "login.binggl.net",
		"version":   fmt.Sprintf("%s-%s", a.Version, a.Build),
		"isError":   isError,
		"error":     err,
		"isMessage": isMessage,
		"msg":       msg,
	}

	return tmpl.Execute(w, data)
}

// GetOIDCRedirectURL returns the URL used for additional round-trip to ensure that cookies are written
func (a *loginAPI) GetOIDCRedirectURL() string {
	return oidcInitiateURL
}

// HandleOIDCRedirect initiates the OAUTH flow by redirecting the authentication system
func (a *loginAPI) HandleOIDCRedirect(w http.ResponseWriter, r *http.Request) error {
	state := randToken()
	a.appCookie.Set(stateParam, state, cookieExpiry, w)
	a.logger.InfoRequest(fmt.Sprintf("HandleOIDCRedirect: initiate using state '%s'", state), r)
	http.Redirect(w, r, a.GetOIDCRedirectURL(), http.StatusTemporaryRedirect)
	return nil
}

// HandleAuthFlow initiates the authentication and redirects to a specific URL
func (a *loginAPI) HandleAuthFlow(w http.ResponseWriter, r *http.Request) error {
	state := randToken()
	a.appCookie.Set(stateParam, state, cookieExpiry, w)
	a.logger.InfoRequest(fmt.Sprintf("HandleAuthFlow: initiate using state '%s'", state), r)

	site, redirect := a.query(r, siteParam), a.query(r, redirectParam)
	if site == "" || redirect == "" {
		return errors.BadRequestError{Err: fmt.Errorf("missing or invalid parameters supplied"), Request: r}
	}
	a.appCookie.Set(authFlowCookie, fmt.Sprintf("%s%s%s", site, authFlowSep, redirect), cookieExpiry, w)
	http.Redirect(w, r, a.GetOIDCRedirectURL(), http.StatusTemporaryRedirect)
	return nil
}

// HandleOIDCRedirectFinal is responsible to set the state cookie for the OIDC interaction
func (a *loginAPI) HandleOIDCRedirectFinal(w http.ResponseWriter, r *http.Request) error {
	state := a.appCookie.Get(stateParam, r)
	if state == "" {
		a.logger.ErrorRequest(fmt.Sprintf("HandleOIDCRedirectFinal: empty state from cookie, referrer: '%s'", r.Referer()), r)
		return errors.BadRequestError{Err: fmt.Errorf("missing state, cannot initiate OIDC"), Request: r}
	}
	a.logger.InfoRequest(fmt.Sprintf("HandleOIDCRedirectFinal: initiate OIDC redirect using state: '%s'", state), r)
	http.Redirect(w, r, a.oauthConfig.AuthCodeURL(state), http.StatusFound)
	return nil
}

// HandleOIDCLogin performs the login/authentication of the OIDC context
// the user of the external authentication provider is checked against
// the database. If a match is found a token with the valid claims is created
// and a redirect is made to the defined URL
func (a *loginAPI) HandleOIDCLogin(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()

	// read the stateParam again
	state := a.appCookie.Get(stateParam, r)
	a.logger.InfoRequest(fmt.Sprintf("HandleOIDCLogin: got state param: %s", state), r)

	if a.query(r, stateParam) != state {
		return errors.BadRequestError{Err: fmt.Errorf("state did not match"), Request: r}
	}
	a.appCookie.Del(stateParam, w)

	// is this an auth/flow request
	var (
		authFlow       bool
		site, redirect string
	)
	authFlowParams := a.appCookie.Get(authFlowCookie, r)
	if authFlowParams != "" {
		a.logger.InfoRequest(fmt.Sprintf("HandleOIDCLogin: auth/flow login-mode"), r)
		parts := strings.Split(authFlowParams, "|")
		site = parts[0]
		redirect = parts[1]
		authFlow = true
	}
	a.appCookie.Del(authFlowCookie, w)

	oauth2Token, err := a.oauthConfig.Exchange(ctx, a.query(r, codeParam))
	if err != nil {
		return errors.ServerError{Err: fmt.Errorf("failed to exchange token: %v", err), Request: r}
	}
	rawIDToken, ok := oauth2Token.Extra(idTokenParam).(string)
	if !ok {
		return errors.ServerError{Err: fmt.Errorf("no id_token field in oauth2 token"), Request: r}
	}
	idToken, err := a.oauthVerifier.VerifyToken(ctx, rawIDToken)
	if err != nil {
		return errors.ServerError{Err: fmt.Errorf("failed to verify ID Token: %v", err), Request: r}
	}

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

	if err := idToken.GetClaims(&oidcClaims); err != nil {
		return errors.ServerError{Err: fmt.Errorf("claims error: %v", err), Request: r}
	}

	// the user was authenticated successfully, check if sites are available for the given user!
	success := true
	sites, err := a.repo.GetSitesByUser(oidcClaims.Email)
	if err != nil {
		a.logger.ErrorRequest(fmt.Sprintf("HandleOIDCLogin: successfull login by '%s' but error fetching sites! %v", oidcClaims.Email, err), r)
		success = false
	}

	if len(sites) == 0 {
		a.logger.ErrorRequest(fmt.Sprintf("HandleOIDCLogin: successfull login by '%s' but no sites availabel!", oidcClaims.Email), r)
		success = false
	}

	if authFlow {
		a.logger.InfoRequest(fmt.Sprintf("HandleOIDCLogin: auth/flow - check for specific site '%s'", site), r)
		success = false
		// check specific site
		for _, e := range sites {
			if e.Name == site {
				success = true
				break
			}
		}
	}

	if !success {
		a.appCookie.Set(errors.FlashKeyError, fmt.Sprintf("User '%s' is not allowed to login!", oidcClaims.Email), cookieExpiry, w)
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return nil
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
	token, err := security.CreateToken(a.jwt.JwtIssuer, []byte(a.jwt.JwtSecret), a.jwt.Expiry, claims)
	if err != nil {
		a.logger.ErrorRequest(fmt.Sprintf("HandleOIDCLogin: could not create a JWT token: %v", err), r)
		return errors.ServerError{Err: fmt.Errorf("error creating JWT: %v", err), Request: r}
	}

	login := persistence.Login{
		User:    oidcClaims.Email,
		Created: time.Now().UTC(),
		Type:    persistence.DIRECT,
	}

	if authFlow {
		login.Type = persistence.FLOW
	}

	err = a.repo.StoreLogin(login, per.Atomic{})
	if err != nil {
		a.logger.ErrorRequest(fmt.Sprintf("HandleOIDCLogin: the login could not be saved: %v", err), r)
		return errors.ServerError{Err: fmt.Errorf("error storing the login: %v", err), Request: r}
	}

	// set the cookie
	exp := a.jwt.Expiry * 24 * 3600
	a.setJWTCookie(a.jwt.CookieName, token, exp, w)

	redirectURL := a.jwt.LoginRedirect
	if authFlow {
		a.logger.InfoRequest(fmt.Sprintf("HandleOIDCLogin: auth/flow - redirect to specific URL: '%s'", redirect), r)
		redirectURL = redirect
	}

	// redirect to provided URL
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	return nil
}

// HandleLogout invalidates the authenticated user
func (a *loginAPI) HandleLogout(user security.User, w http.ResponseWriter, r *http.Request) error {
	a.logger.InfoRequest(fmt.Sprintf("HandleLogout: for user '%s'", user.Username), r)
	// remove the cookie by expiring it
	a.setJWTCookie(a.jwt.CookieName, "", -1, w)
	a.appCookie.Set(errors.FlashKeyInfo, fmt.Sprintf("User '%s' was logged-off!", user.Email), cookieExpiry, w)
	http.Redirect(w, r, a.jwt.LoginRedirect, http.StatusTemporaryRedirect)
	return nil
}

func (a *loginAPI) setJWTCookie(name, value string, exp int, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   exp,  /* exp in seconds */
		HttpOnly: true, // only let the api access those cookies
		Domain:   a.jwt.CookieDomain,
		Path:     a.jwt.CookiePath,
		Secure:   a.jwt.CookieSecure,
	}
	http.SetCookie(w, &cookie)
}

func randToken() string {
	u := uuid.New()
	return u.String()
}
