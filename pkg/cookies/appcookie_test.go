package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookies(t *testing.T) {
	c := NewAppCookie(Settings{
		Path:   "/",
		Domain: "localhost",
		Secure: false,
		Prefix: "app_",
	})
	cookieName := "testcookie"
	expCookie := "test"
	rec := httptest.NewRecorder()

	c.Set(cookieName, expCookie, 60, rec)
	req := &http.Request{Header: http.Header{"Cookie": []string{rec.Header().Get("Set-Cookie")}}}
	v := c.Get(cookieName, req)
	if v != expCookie {
		t.Errorf("could not read cookie: %s, expected: %s, got %s", cookieName, expCookie, v)
	}

	v = c.Get("randomName", req)
	if v == expCookie {
		t.Errorf("the cookie should be empty for randomName: expected: '', got %s", v)
	}

	rec = httptest.NewRecorder()
	c.Del(cookieName, rec)
	req = &http.Request{Header: http.Header{"Cookie": []string{rec.Header().Get("Set-Cookie")}}}
	v = c.Get(cookieName, req)
	if v != "" {
		t.Errorf("could not clear the cookie, got %s", v)
	}

}
