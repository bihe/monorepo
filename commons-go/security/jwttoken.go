package security // import "golang.binggl.net/commons/security"

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// --------------------------------------------------------------------------
// define custom claims
// --------------------------------------------------------------------------

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

type jwtClaims struct {
	jwt.StandardClaims
	Claims
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
