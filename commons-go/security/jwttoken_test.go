package security // import "golang.binggl.net/commons/security"

import (
	"testing"
)

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
