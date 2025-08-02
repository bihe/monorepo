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
	req := httptest.NewRequest("GET", "/crypter", nil)
	addJwtAuth(req)

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "Crypter")
}

func Test_Encrypt(t *testing.T) {
	th := templateHandler(&mockSiteService{})

	// arrange
	form := url.Values{}
	form.Add("crypter_passphrase", "test")
	form.Add("crypter_input", "hello, world - from unit-test")

	req := httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	// this is the indicator, that the provided HTML output has age-encrypted content
	assert.Contains(t, string(body), "-----END ENCRYPTED CONTENT-----")
}

const encryptedText = `-----BEGIN ENCRYPTED CONTENT-----
1Vfcq4zdPGAHL0OQss57YU23bBbATf9y
nP6V-oTdFCZTo-8LmXm_RFlX9PTcEFD7
3wtqGP6aqGY_8NF7M3n4hYd_TXdTSdTo
FNjJKHgS0fOsHiZb661kUv8AX3ctcR2O
bW_-XwEvN_yC5Bo9DDYUBACCaVD8dnT-
vvgUMFroHKY=
-----END ENCRYPTED CONTENT-----`

func Test_Decrypt(t *testing.T) {
	th := templateHandler(&mockSiteService{})
	// arrange
	form := url.Values{}
	form.Add("crypter_passphrase", "test")
	form.Add("crypter_output", encryptedText)
	req := httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
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
	form.Add("crypter_passphrase", "wrong-passphrase")
	form.Add("crypter_output", encryptedText)
	req := httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "no hmac found in decrypted result - operation invalid")
}

func Test_InputValidation(t *testing.T) {
	th := templateHandler(&mockSiteService{})

	// missing passphrase

	// arrange
	form := url.Values{}
	form.Add("crypter_passphrase", "")
	form.Add("crypter_input", "hello world")

	req := httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec := httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "a passphrase is needed")

	// max length of passphrase

	// arrange
	form = url.Values{}
	form.Add("crypter_passphrase", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaA")

	req = httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec = httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ = io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "the maximum length of the passphrase is 32 chars")

	// both text-input fields are filled

	// arrange
	form = url.Values{}
	form.Add("crypter_passphrase", "test")
	form.Add("crypter_input", "hello world")
	form.Add("crypter_output", "encrypted")

	req = httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	addJwtAuth(req)
	rec = httptest.NewRecorder()

	// act
	th.ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ = io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "both fields are filled / cannot decide what to do")

	// both text-input fields are empty

	// arrange
	form = url.Values{}
	form.Add("crypter_passphrase", "test")
	form.Add("crypter_input", "")
	form.Add("crypter_output", "")

	req = httptest.NewRequest("POST", "/crypter", strings.NewReader(form.Encode()))
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
