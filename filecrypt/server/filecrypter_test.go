package server

import (
	"context"
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
	"golang.binggl.net/monorepo/filecrypt"
	"golang.binggl.net/monorepo/proto"
)

var tokenSecurity = filecrypt.TokenSecurity{
	JwtIssuer: "issuer",
	JwtSecret: "secret",
	Claim: filecrypt.Claim{
		Name:  "claim",
		URL:   "http://localhost:3000",
		Roles: []string{"role"},
	},
	CacheDuration: "10m",
}

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

func TestFileCrypter_Validate_Input(t *testing.T) {
	crypter := NewFileCrypter("./", log.New().WithField("mode", "test"), filecrypt.TokenSecurity{})
	var requestTests = []struct {
		name string
		req  *proto.CrypterRequest
	}{
		{
			name: "expect empty token",
			req:  &proto.CrypterRequest{},
		},
		{
			name: "expect missing newpass",
			req: &proto.CrypterRequest{
				AuthToken: "token",
			},
		},
		{
			name: "expect missing payload",
			req: &proto.CrypterRequest{
				AuthToken:   "token",
				NewPassword: "12345",
			},
		},
	}

	for _, tt := range requestTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := crypter.Encrypt(context.TODO(), tt.req)
			if err == nil {
				t.Errorf("error expected")
			}
		})
	}
}

func TestFileCrypter_Expect_Invalid_TokenError(t *testing.T) {
	crypter := NewFileCrypter("./", log.New().WithField("mode", "test"), filecrypt.TokenSecurity{})
	payload, err := ioutil.ReadFile("./encrypted.pdf")
	if err != nil {
		t.Errorf("could not read testfile: %v", err)
	}

	_, err = crypter.Encrypt(context.TODO(), &proto.CrypterRequest{
		AuthToken:   "token",
		NewPassword: "12345",
		Payload:     payload,
	})
	if err == nil {
		t.Errorf("invalid token error expected")
	}
}

func TestFileCrypter_Encrypt_File(t *testing.T) {
	payload, err := ioutil.ReadFile("./encrypted.pdf")
	if err != nil {
		t.Errorf("could not read testfile: %v", err)
	}
	crypter := NewFileCrypter("./", log.New().WithField("mode", "test"), tokenSecurity)
	resp, err := crypter.Encrypt(context.TODO(), &proto.CrypterRequest{
		AuthToken:    testToken,
		InitPassword: "12345",
		NewPassword:  "54321",
		Payload:      payload,
		Type:         proto.CrypterRequest_PDF,
	})
	if err != nil {
		t.Errorf("error returned: %v", err)
	}
	if resp.Result != proto.CrypterResponse_SUCCESS {
		t.Errorf("no success encrypting file")
	}
}

func TestFileCrypter_Encrypt_UnencryptedFile(t *testing.T) {
	payload, err := ioutil.ReadFile("./unencrypted.pdf")
	if err != nil {
		t.Errorf("could not read testfile: %v", err)
	}
	crypter := NewFileCrypter("./", log.New().WithField("mode", "test"), tokenSecurity)
	resp, err := crypter.Encrypt(context.TODO(), &proto.CrypterRequest{
		AuthToken:    testToken,
		InitPassword: "",
		NewPassword:  "54321",
		Payload:      payload,
		Type:         proto.CrypterRequest_PDF,
	})
	if err != nil {
		t.Errorf("error returned: %v", err)
	}
	if resp.Result != proto.CrypterResponse_SUCCESS {
		t.Errorf("no success encrypting file")
	}
}
