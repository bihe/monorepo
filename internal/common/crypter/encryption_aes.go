package crypter

// @see: https://golang.org/src/crypto/cipher/example_test.go

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// encryptAES creates the encryption with a HMAC to ensure proper decryption
func encryptAES(text, passphrase string) ([]byte, error) {
	key := padKey([]byte(passphrase))
	hmac, err := getMAC([]byte(text), key)
	if err != nil {
		return nil, err
	}
	hmacString := base64.URLEncoding.EncodeToString(hmac)
	textEnc := base64.URLEncoding.EncodeToString([]byte(text))
	text = textEnc + "." + hmacString
	return encrypt(key, text)
}

// decryptAES anticipates a HMAC guard, otherwise the decryption will fail
func decryptAES(cipherText, passphrase string) ([]byte, error) {
	key := padKey([]byte(passphrase))
	decryptedPayload, err := decrypt(key, cipherText)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt secret: %s", err)
	}
	decrypted := string(decryptedPayload)
	index := strings.LastIndex(decrypted, ".")
	if index == -1 {
		return nil, fmt.Errorf("could not decrypt, invalid input")
	}

	textEnc := decrypted[:index]
	text, err := base64.URLEncoding.DecodeString(textEnc)
	if err != nil {
		return nil, fmt.Errorf("could not base64decode decrypted string: %s", err)
	}
	hmacString := decrypted[index+1:]
	hmac, err := base64.URLEncoding.DecodeString(hmacString)
	if err != nil {
		return nil, fmt.Errorf("could not base64decode hmacString: %s", err)
	}

	match, err := checkMAC([]byte(text), hmac, key)
	if err != nil {
		return nil, fmt.Errorf("could not check hmac: %s", err)
	}

	if !match {
		return nil, fmt.Errorf("the hmac of the decrypted text does not match")
	}

	return []byte(text), nil
}

func encrypt(key []byte, text string) ([]byte, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// https://pkg.go.dev/crypto/cipher#NewCBCEncrypter
	// CBC mode works on blocks so plain-texts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2.
	if len(plaintext)%aes.BlockSize != 0 {
		// add the remainder as long as it takes to have a full multiple of the block-size
		textLen := len(plaintext)
		remainder := textLen % aes.BlockSize
		for remainder != 0 {
			textLen += remainder
			remainder = textLen % aes.BlockSize
		}
		// initialize a "bigger" slices which is filled with 0-bytes
		// the 0-bytes act as teh padding
		padPlainText := make([]byte, textLen)
		copy(padPlainText, plaintext)
		plaintext = padPlainText
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return []byte(base64.URLEncoding.EncodeToString(ciphertext)), nil
}

func decrypt(key []byte, cryptoText string) ([]byte, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not base64decode ciphertext: %s", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// during encryption the plaintext was potentially using a padding
	// this padding needs to be removed to fully recover the plaintext
	// 0-bytes act as padding, so break if the first is found
	paddingStart := 0
	for i, b := range ciphertext {
		if b == 0 {
			paddingStart = i
			break
		}
	}
	ciphertext = ciphertext[0:paddingStart]
	return ciphertext, nil
}

func padKey(key []byte) []byte {
	paddedKey := make([]byte, 32)
	if len(key) <= 32 {
		copy(paddedKey[0:32], key[:])
	} else if len(key) > 32 {
		copy(paddedKey[0:32], key[:32])
	}
	return paddedKey
}

func checkMAC(message, messageMAC, key []byte) (bool, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(message)
	if err != nil {
		return false, err
	}
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC), nil
}

func getMAC(message, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(message)
	if err != nil {
		return nil, err
	}
	expectedMAC := mac.Sum(nil)
	return expectedMAC, nil
}
