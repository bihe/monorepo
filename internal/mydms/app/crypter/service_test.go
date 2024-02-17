package crypter_test

import (
	"context"
	"os"
	"testing"

	"golang.binggl.net/monorepo/internal/mydms/app/crypter"
	"golang.binggl.net/monorepo/pkg/logging"
)

var logger = logging.NewNop()

const encryptedPDF = "../../../../testdata/encrypted.pdf"
const unencryptedPDF = "../../../../testdata/unencrypted.pdf"

func TestEncryption_Validate_Input(t *testing.T) {
	svc := crypter.NewService(logger)

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
			req:  crypter.Request{},
		},
		{
			name: "expect missing payload",
			req: crypter.Request{
				NewPass: "12345",
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

func TestFileCrypter_Encrypt_File(t *testing.T) {
	payload, err := os.ReadFile(encryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	svc := crypter.NewService(logger)
	crypto_payload, err := svc.Encrypt(context.TODO(), crypter.Request{
		Payload:  payload,
		Type:     crypter.PDF,
		InitPass: "12345",
		NewPass:  "54321",
	})
	if err != nil {
		t.Errorf("error returned: %v", err)
	}
	if len(crypto_payload) == 0 {
		t.Errorf("no success encrypting file")
	}
}

func TestFileCrypter_Encrypt_UnencryptedFile(t *testing.T) {
	payload, err := os.ReadFile(unencryptedPDF)
	if err != nil {
		t.Fatalf("could not read testfile: %v", err)
	}
	svc := crypter.NewService(logger)
	crypto_payload, err := svc.Encrypt(context.TODO(), crypter.Request{
		Payload:  payload,
		Type:     crypter.PDF,
		InitPass: "",
		NewPass:  "54321",
	})
	if err != nil {
		t.Errorf("error returned: %v", err)
	}
	if len(crypto_payload) == 0 {
		t.Errorf("no success encrypting file")
	}
}
