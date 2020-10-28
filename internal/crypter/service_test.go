package crypter_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"golang.binggl.net/monorepo/internal/crypter"
)

var tokenSecurity = crypter.TokenSecurity{
	JwtIssuer: "issuer",
	JwtSecret: "secret",
	Claim: crypter.Claim{
		Name:  "claim",
		URL:   "http://localhost:3000",
		Roles: []string{"role"},
	},
	CacheDuration: "10m",
}

var logger = log.NewLogfmtLogger(os.Stderr)

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
const testToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4NzAzOTI2NDcsImp0aSI6IjZmYWQ1YzAwLWZlZTItNDU5Yy1hYmFkLTIwNDU3Y2ZmM2Q4YSIsImlhdCI6MTU1OTc4Nzg0NywiaXNzIjoiaXNzdWVyIiwic3ViIjoidXNlciIsIlR5cGUiOiJsb2dpbi5Vc2VyIiwiRGlzcGxheU5hbWUiOiJEaXNwbGF5IE5hbWUiLCJFbWFpbCI6ImEuYkBjLmRlIiwiVXNlcklkIjoiMTIzNDUiLCJVc2VyTmFtZSI6ImEuYkBjLmRlIiwiR2l2ZW5OYW1lIjoiRGlzcGxheSIsIlN1cm5hbWUiOiJOYW1lIiwiQ2xhaW1zIjpbImNsYWltfGh0dHA6Ly9sb2NhbGhvc3Q6MzAwMHxyb2xlIl19.qUwvHXBmV_FuwLtykOnzu3AMbxSqrg82bQlAi3Nabyo"

const encryptedPDF = "../../testdata/encrypted.pdf"
const unencryptedPDF = "../../testdata/unencrypted.pdf"

func TestEncryption_Validate_Input(t *testing.T) {
	svc := crypter.NewService(logger, tokenSecurity)

	var requestTests = []struct {
		name string
		req  crypter.Request
	}{
		{
			name: "expect empty token",
			req:  crypter.Request{},
		},
		{
			name: "expect missing newpass",
			req: crypter.Request{
				AuthToken: "token",
			},
		},
		{
			name: "expect missing payload",
			req: crypter.Request{
				AuthToken: "token",
				NewPass:   "12345",
			},
		},
	}

	for _, tt := range requestTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Encrypt(context.TODO(), tt.req)
			if err == nil {
				t.Errorf("error expected")
			}
		})
	}

}

func TestFileCrypter_Expect_Invalid_TokenError(t *testing.T) {
	svc := crypter.NewService(logger, tokenSecurity)
	payload, err := ioutil.ReadFile(encryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}

	_, err = svc.Encrypt(context.TODO(), crypter.Request{
		AuthToken: "token",
		Payload:   payload,
		Type:      crypter.PDF,
		InitPass:  "12345",
		NewPass:   "54321",
	})
	if err == nil {
		t.Errorf("invalid token error expected")
	}
}

func TestFileCrypter_Encrypt_File(t *testing.T) {
	payload, err := ioutil.ReadFile(encryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	svc := crypter.NewService(logger, tokenSecurity)
	crypto_payload, err := svc.Encrypt(context.TODO(), crypter.Request{
		AuthToken: testToken,
		Payload:   payload,
		Type:      crypter.PDF,
		InitPass:  "12345",
		NewPass:   "54321",
	})
	if err != nil {
		t.Errorf("error returned: %v", err)
	}
	if len(crypto_payload) == 0 {
		t.Errorf("no success encrypting file")
	}
}

func TestFileCrypter_Encrypt_UnencryptedFile(t *testing.T) {
	payload, err := ioutil.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	svc := crypter.NewService(logger, tokenSecurity)
	crypto_payload, err := svc.Encrypt(context.TODO(), crypter.Request{
		AuthToken: testToken,
		Payload:   payload,
		Type:      crypter.PDF,
		InitPass:  "",
		NewPass:   "54321",
	})
	if err != nil {
		t.Errorf("error returned: %v", err)
	}
	if len(crypto_payload) == 0 {
		t.Errorf("no success encrypting file")
	}
}
