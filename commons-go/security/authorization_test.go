package security // import "golang.binggl.net/commons/security"

import (
	"testing"
)

const failed = "Authorization failed."
const shouldFail = "Authorization should fail."

func TestAuthorization(t *testing.T) {
	reqClaim := Claim{Name: "test", URL: "http://a.b.c.de/f", Roles: []string{"user", "admin"}}
	var claims []string

	claims = append(claims, "a|http://1.com|nothing")
	claims = append(claims, "test|http://a.b.c.de/f|user")

	if _, err := Authorize(reqClaim, claims); err != nil {
		t.Error(failed, err)
	}

	claims = claims[0:1]
	if _, err := Authorize(reqClaim, claims); err == nil {
		t.Error(shouldFail, err)
	}

	claims = nil
	claims = append(claims, "a|http://1.com|nothing")
	claims = append(claims, "test|http://a.b.c.de/f|admin")

	if _, err := Authorize(reqClaim, claims); err != nil {
		t.Error(failed, err)
	}

	claims = []string{"test|http://a.b.c.de/nomatch|role"}
	if _, err := Authorize(reqClaim, claims); err == nil {
		t.Error(shouldFail, err)
	}

	claims = []string{"test|NOURLä|role"}
	if _, err := Authorize(reqClaim, claims); err == nil {
		t.Error(shouldFail, err)
	}

	claims = []string{"test|http://a.b.c.de|user"}
	reqClaim = Claim{Name: "test", URL: "http://a.b.c.de/", Roles: []string{"user"}}
	if _, err := Authorize(reqClaim, claims); err != nil {
		t.Error(failed, err)
	}
}
