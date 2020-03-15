package security // import "golang.binggl.net/commons/security"

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

// JwtTokenPayload is the parsed contents of the given token
type JwtTokenPayload struct {
	Type        string
	UserName    string
	Email       string
	Claims      []string
	UserID      string `json:"UserId"`
	DisplayName string
	Surname     string
	GivenName   string
	jwt.StandardClaims
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
				return JwtTokenPayload{}, fmt.Errorf("invalid issued created the token")
			}
			return *claims, nil
		}
	}
	return JwtTokenPayload{}, fmt.Errorf("could not parse JWT token")
}
