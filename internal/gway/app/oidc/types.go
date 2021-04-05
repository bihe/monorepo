package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.binggl.net/monorepo/internal/gway"
	"golang.org/x/oauth2"
)

const idTokenParam = "id_token"

// --------------------------------------------------------------------------
// OIDC wrapping
// --------------------------------------------------------------------------

// NewConfigAndVerifier creates the necessary configuration for OIDC
func NewConfigAndVerifier(c gway.OAuthConfig) (OIDCConfig, OIDCVerifier) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, c.Provider)
	if err != nil {
		panic(fmt.Sprintf("could not create a new OIDC provider: %v", err))
	}
	ver := provider.Verifier(&oidc.Config{
		ClientID: c.ClientID,
	})
	oidcVer := &oidcVerifier{ver}
	oidcCfg := &oidcConfig{
		config: &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  c.RedirectURL,
			Scopes:       openIDScope,
		},
	}
	return oidcCfg, oidcVer
}

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
