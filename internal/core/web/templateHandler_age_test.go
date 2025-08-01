package web_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/pkg/security"
)

const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

type mockSiteService struct{}

func (*mockSiteService) GetSitesForUser(user security.User) (sites.UserSites, error) {
	return sites.UserSites{}, nil
}

func (*mockSiteService) GetUsersForSite(site string, user security.User) ([]string, error) {
	return make([]string, 0), nil
}

func (*mockSiteService) SaveSitesForUser(sites sites.UserSites, user security.User) error {
	return nil
}

// compile guard
var _ sites.Service = &mockSiteService{}

func Test_ShowStartPage(t *testing.T) {
	th := templateHandler(&mockSiteService{})

	// arrange
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/age", nil)
	addJwtAuth(req)

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "Age")
}

func Test_Encrypt(t *testing.T) {
	th := templateHandler(&mockSiteService{})

	// arrange
	form := url.Values{}
	form.Add("age_passphrase", "test")
	form.Add("age_input", "hello, world - from unit-tes")

	req := httptest.NewRequest("POST", "/age", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	// this is the indicator, that the provided HTML output has age-encrypted content
	assert.Contains(t, string(body), "-----END AGE ENCRYPTED FILE-----")
}

const encryptedText = `-----BEGIN AGE ENCRYPTED FILE-----
YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IHNjcnlwdCBrUm5OL3YvQVNiZXpMYjlq
aVY2anl3IDE4ClIrcHNtT1BvbHhjOExkKzl4bjh6dElFNWZncU14U3hIWmppU2VK
V29uNlEKLS0tIEZwSHlWa2Qzakxzajhrei9idy9vYkIxTWs3MzkxWFBQREtRdjFo
cmYvVlkKjk0tE6iPhl6qBaMDmKFsBwrm/l1i7uYjhjDbkKumxjWKoIgt6BTWz9tb
bgAT6zPh8j7cjIH1LBJQvFU0vA==
-----END AGE ENCRYPTED FILE-----`

func Test_Decrypt(t *testing.T) {
	th := templateHandler(&mockSiteService{})
	// arrange
	form := url.Values{}
	form.Add("age_passphrase", "test")
	form.Add("age_output", encryptedText)
	req := httptest.NewRequest("POST", "/age", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "hello, world - from unit-test")
}

func Test_DecryptWrongPassphrase(t *testing.T) {
	th := templateHandler(&mockSiteService{})
	// arrange
	form := url.Values{}
	form.Add("age_passphrase", "wrong-passphrase")
	form.Add("age_output", encryptedText)
	req := httptest.NewRequest("POST", "/age", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "could not create decryption writer with passphrase; no identity matched any of the recipients")
}

func Test_InputValidation(t *testing.T) {
	th := templateHandler(&mockSiteService{})

	// missing passphrase

	// arrange
	form := url.Values{}
	form.Add("age_passphrase", "")
	form.Add("age_input", "hello world")

	req := httptest.NewRequest("POST", "/age", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "a passphrase is needed")

	// both text-input fields are filled

	// arrange
	form = url.Values{}
	form.Add("age_passphrase", "test")
	form.Add("age_input", "hello world")
	form.Add("age_output", "encrypted")

	req = httptest.NewRequest("POST", "/age", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec = httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ = io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "both fields are filled / cannot decide what to do")

	// both text-input fields are emptz

	// arrange
	form = url.Values{}
	form.Add("age_passphrase", "test")
	form.Add("age_input", "")
	form.Add("age_output", "")

	req = httptest.NewRequest("POST", "/age", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec = httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ = io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "both fields are empty")
}

func addJwtAuth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
}
