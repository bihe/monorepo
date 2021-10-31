package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/core/app/sites"
	"golang.binggl.net/monorepo/internal/core/app/store"
)

const adminRole = "A"
const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjcyMDE0NDUsImlhdCI6MTYyNjU5NjY0NSwiaXNzIjoiaXNzdWVyIiwianRpIjoiZmI1ZThjNDMtZDEwMi00MGU3LTljM2EtNTkwYzA3ODAzNzUwIiwic3ViIjoiaGVucmlrLmJpbmdnbEBnbWFpbC5jb20iLCJDbGFpbXMiOlsiQXxodHRwOi8vQXxBIl0sIkRpc3BsYXlOYW1lIjoiVXNlciBBIiwiRW1haWwiOiJ1c2VyQGEuY29tIiwiR2l2ZW5OYW1lIjoidXNlciIsIlN1cm5hbWUiOiJBIiwiVHlwZSI6ImxvZ2luLlVzZXIiLCJVc2VySWQiOiIxMiIsIlVzZXJOYW1lIjoidXNlckBhLmNvbSJ9.JcZ9-ImQieOWW1KaLJGR_Pqol2MQviFDdjqbIfhAlek"

var siteStruct = map[string][]store.UserSiteEntity{
	"user@a.com": {
		{
			Name:     "A",
			User:     "user@a.com",
			URL:      "http://A",
			PermList: "A",
		},
	},
	"user@b.com": {
		{
			Name:     "A",
			User:     "user@b.com",
			URL:      "http://A",
			PermList: "A;B",
		},
	},
}

func sitesHandler() http.Handler {
	return handlerWith(&handlerOps{
		siteSvc: sites.New(adminRole, store.NewMock(siteStruct)),
	})
}

func Test_GetSitesForUser_NoAuth(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sites", nil)

	// act
	sitesHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func Test_GetSitesForUser(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sites", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	sitesHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, 200, rec.Code)
	var s sites.UserSites
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Equal(t, "user@a.com", s.User)
	assert.Len(t, s.Sites, 1)
	assert.Equal(t, "A", s.Sites[0].Name)
}

func Test_GetUsersForSites(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sites/users/A", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	sitesHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, 200, rec.Code)
	var users []string
	if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
		t.Errorf("could not unmarshal: %v", err)
	}
	assert.Len(t, users, 2)
}

func Test_SaveSitesForUser(t *testing.T) {
	// arrange
	rec := httptest.NewRecorder()
	u := sites.UserSites{
		User: "user@c.com",
		Sites: []sites.SiteInfo{
			{
				Name: "C",
				URL:  "http://C",
				Perm: []string{"C"},
			},
		},
	}
	payload, _ := json.Marshal(u)
	req, _ := http.NewRequest("POST", "/api/v1/sites", bytes.NewBuffer(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	sitesHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, 201, rec.Code)

	// ----- invalid payload ----
	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/sites", bytes.NewBuffer(make([]byte, 0)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))

	// act
	sitesHandler().ServeHTTP(rec, req)

	// assert
	assert.Equal(t, 400, rec.Code)
}
