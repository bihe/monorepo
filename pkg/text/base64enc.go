package text

import (
	"encoding/base64"
	"strings"
)

// get rid of "special characters" of base64 +,/
var replacement = map[string]string{
	"+": "_",
	"/": "-",
}

// SafePathEscapeBase64 takes a base64 encoded string and "cleans" it for URL-path use
func SafePathEscapeBase64(input string) string {
	enc := input
	for k := range replacement {
		if strings.Contains(input, k) {
			enc = strings.ReplaceAll(enc, k, replacement[k])
		}
	}
	return enc
}

// EncBase64SafePath encodes the input as base64.
// In addition the result is corrected to not collide in a URL-path scenario because of the two extra params +,/ in the base64 "vocabulary".
// https://en.wikipedia.org/wiki/Base64
func EncBase64SafePath(input string) string {
	data := []byte(input)
	encodedString := base64.StdEncoding.EncodeToString(data)
	for k, v := range replacement {
		encodedString = strings.ReplaceAll(encodedString, k, v)
	}

	return encodedString
}

// DecBase64SafePath decodes a base64 string, which was additionally cleaned to be used as web-URLs.
func DecBase64SafePath(input string) string {
	cleaned := input
	for k := range replacement {
		cleaned = strings.ReplaceAll(cleaned, replacement[k], k)
	}
	decode, err := base64.StdEncoding.DecodeString(cleaned)
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
