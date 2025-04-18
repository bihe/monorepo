package text

import (
	"encoding/base64"
	"net/url"
)

// EncBase64SafePath encodes the input as base64.
// In addition the result is URL-encoded, because of the two extra params +,/ in the base64 "vocabulary".
// https://en.wikipedia.org/wiki/Base64
func EncBase64SafePath(input string) string {
	data := []byte(input)
	encodedString := base64.StdEncoding.EncodeToString(data)
	encodedString = url.QueryEscape(encodedString)
	return encodedString
}

// DecBase64SafePath decodes a base64 string, which was additionally URL-encoded to be used as web-URLs.
func DecBase64SafePath(input string) string {
	urldecoded, err := url.QueryUnescape(input)
	if err != nil {
		return ""
	}
	decode, err := base64.StdEncoding.DecodeString(urldecoded)
	if err != nil {
		return ""
	}
	return string(decode)
}

// EncBase64 turns the input into a base64 encoded string
func EncBase64(input string) string {
	data := []byte(input)
	encodedString := base64.StdEncoding.EncodeToString(data)
	return encodedString
}

// DecBase64 decodes the provided bas64 string
// if an error occurs an empty string is returned
func DecBase64(input string) string {
	decode, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return ""
	}
	return string(decode)
}
