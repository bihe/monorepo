package agecrypt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// EncryptStringPassphrase uses the given passphrase to encrypt the input string with age
func EncryptStringPassphrase(input, passphrase string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("no input supplied")
	}
	if passphrase == "" {
		return "", fmt.Errorf("no passphrase supplied")
	}

	in := strings.NewReader(input)
	out := &bytes.Buffer{}

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return "", err
	}
	err = encrypt([]age.Recipient{recipient}, in, out)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

// this function was "literally" copied from https://github.com/FiloSottile/age/blob/main/cmd/age/age.go#L395
// because I hated it to convert io.Readers, io.Writers, io.ReadWriteClosers
func encrypt(recipients []age.Recipient, in io.Reader, out io.Writer) (err error) {
	var (
		w io.WriteCloser
	)

	a := armor.NewWriter(out)
	defer func() {
		if errDefer := a.Close(); errDefer != nil {
			err = errDefer
		}
	}()
	out = a

	w, err = age.Encrypt(out, recipients...)
	if err != nil {
		return fmt.Errorf("could not create encryption writer with passphrase; %v", err)
	}
	if _, err = io.Copy(w, in); err != nil {
		return fmt.Errorf("could not encrypt input; %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("could not close writer input; %v", err)
	}
	return nil
}

// DecryptStringPassphrase decrypts the given cryptText with the provided passphrase
func DecryptStringPassphrase(cryptText, passphrase string) (string, error) {
	if cryptText == "" {
		return "", fmt.Errorf("no cryptText supplied")
	}
	if passphrase == "" {
		return "", fmt.Errorf("no passphrase supplied")
	}

	in := strings.NewReader(cryptText)
	out := &bytes.Buffer{}

	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return "", err
	}
	err = decrypt([]age.Identity{identity}, in, out)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

// this function also was copied from https://github.com/FiloSottile/age/blob/main/cmd/age/age.go#L467
// to prevent the handling of io.Readers, io.Writers, io.ReadWriteClosers
func decrypt(identities []age.Identity, in io.Reader, out io.Writer) error {
	rr := bufio.NewReader(in)
	if start, _ := rr.Peek(len(armor.Header)); string(start) == armor.Header {
		in = armor.NewReader(rr)
	} else {
		in = rr
	}

	r, err := age.Decrypt(in, identities...)
	if err != nil {
		return fmt.Errorf("could not create decryption writer with passphrase; %v", err)
	}
	out.Write(nil) // trigger the lazyOpener even if r is empty
	if _, err := io.Copy(out, r); err != nil {
		return fmt.Errorf("could not decrypt input; %v", err)
	}
	return nil
}
