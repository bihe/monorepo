package agecrypt_test

import (
	"testing"

	"golang.binggl.net/monorepo/internal/core/app/agecrypt"
)

func Test_TextEncryptionDecryption(t *testing.T) {
	// happy path
	cryptText, err := agecrypt.EncryptStringPassphrase("input-string", "passphrase")
	if err != nil {
		t.Errorf("encrypt: did not expect an error: %v", err)
	}

	decrypted, err := agecrypt.DecryptStringPassphrase(cryptText, "passphrase")
	if err != nil {
		t.Errorf("decrypt: did not expect an error: %v", err)
	}

	if decrypted != "input-string" {
		t.Errorf("could not decrypt the text properly: wanted '%s', got '%s'", "input-string", decrypted)
	}

	// input validation encryption
	_, err = agecrypt.EncryptStringPassphrase("", "")
	if err == nil {
		t.Errorf("error expected")
	}
	_, err = agecrypt.EncryptStringPassphrase("input", "")
	if err == nil {
		t.Errorf("error expected")
	}

	// input validation decryption
	_, err = agecrypt.DecryptStringPassphrase("", "")
	if err == nil {
		t.Errorf("error expected")
	}
	_, err = agecrypt.DecryptStringPassphrase("input", "")
	if err == nil {
		t.Errorf("error expected")
	}

	// invalid decrypt input
	_, err = agecrypt.DecryptStringPassphrase("random/text", "passphrase")
	if err == nil {
		t.Errorf("error expected")
	}
}
