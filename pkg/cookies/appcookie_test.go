package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

func Test_Cookie_Prefix(t *testing.T) {
	// arrange
	c := NewAppCookie(Settings{
		Path:   "/",
		Domain: "localhost",
		Secure: false,
		Prefix: "",
	})
	rec := httptest.NewRecorder()
	name := "test"
	value := "value"

	// act
	c.Set(name, value, 60, rec)

	// assert
	r := &http.Request{Header: http.Header{"Cookie": []string{rec.Header().Get("Set-Cookie")}}}
	cookie, err := r.Cookie(name)
	assert.NoError(t, err)
	assert.Equal(t, value, cookie.Value)
}

// TestDel tests the Del method of AppCookie
func TestDel(t *testing.T) {
	c := NewAppCookie(Settings{
		Path:   "/",
		Domain: "localhost",
		Secure: false,
	})
	name := "testcookie"
	recorder := httptest.NewRecorder()

	// Call the Del method
	c.Del(name, recorder)

	// Check if the cookie was deleted by inspecting Set-Cookie header
	setCookieHeader := recorder.Header().Get("Set-Cookie")
	assert.Contains(t, setCookieHeader, "testcookie=")
	assert.Contains(t, setCookieHeader, "Max-Age=0")
}

// TestGet tests the Get method of AppCookie
func TestGet(t *testing.T) {
	c := NewAppCookie(Settings{
		Path:   "/",
		Domain: "localhost",
		Secure: false,
	})
	name := "testcookie"
	value := "testvalue"
	recorder := httptest.NewRecorder()

	// Set a cookie using Set method
	c.Set(name, value, 60, recorder)

	// Create a request with the set cookie
	req := &http.Request{Header: http.Header{"Cookie": []string{recorder.Header().Get("Set-Cookie")}}}

	// Call the Get method and check if it retrieves the correct value
	retValue := c.Get(name, req)
	assert.Equal(t, value, retValue)

	// Check for a non-existent cookie
	retValue = c.Get("non_existent", req)
	assert.Empty(t, retValue)
}
