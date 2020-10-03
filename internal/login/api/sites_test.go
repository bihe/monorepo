package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/pkg/errors"
	"golang.binggl.net/monorepo/pkg/security"
)

func TestGetSites(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{}
	api.editRole = "role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/sites", api.Secure(api.HandleGetSites))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sites", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var u UserSites
	err := json.Unmarshal(rec.Body.Bytes(), &u)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}

	assert.Equal(t, "username", u.User)
	assert.Equal(t, true, u.Editable)
	assert.Equal(t, "site", u.Sites[0].Name)
}

func TestFailSites(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{fail: true}
	api.editRole = "role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/sites", api.Secure(api.HandleGetSites))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sites", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}
}

func TestSaveSites(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{fail: false}
	api.editRole = "role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/sites", api.Secure(api.HandleSaveSites))

	sites := []SiteInfo{{
		Name: "site",
		URL:  "http://example",
		Perm: []string{"role1", "role2"},
	}}
	payload, _ := json.Marshal(sites)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sites", bytes.NewBuffer(payload))

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSaveSitesFail(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{fail: true}
	api.editRole = "role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/sites", api.Secure(api.HandleSaveSites))

	sites := []SiteInfo{{
		Name: "site",
		URL:  "http://example",
		Perm: []string{"role1", "role2"},
	}}
	payload, _ := json.Marshal(sites)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sites", bytes.NewBuffer(payload))

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}
}

func TestSaveSitesNoPayload(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{fail: true}
	api.editRole = "role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/sites", api.Secure(api.HandleSaveSites))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sites", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}
}

func TestSaveSitesNotAllowed(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{fail: true}
	api.editRole = "missing-role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/sites", api.Secure(api.HandleSaveSites))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sites", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var p errors.ProblemDetail
	err := json.Unmarshal(rec.Body.Bytes(), &p)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}
}

func TestGetUsersForSite(t *testing.T) {
	r, api := newAPIRouter()

	api.repo = &mockRepository{}
	api.editRole = "role"

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := security.NewContext(r.Context(), &security.User{
				Username:    "username",
				Email:       "a.b@c.de",
				DisplayName: "displayname",
				Roles:       []string{"role"},
				UserID:      "12345",
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/sites/users/site", api.Secure(api.HandleGetUsersForSite))

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sites/users/site", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var u UserList
	err := json.Unmarshal(rec.Body.Bytes(), &u)
	if err != nil {
		t.Errorf("could not get valid json: %v", err)
	}

	assert.Equal(t, 2, u.Count)
	assert.Equal(t, "user1", u.Users[0])

	// fail
	api.repo = &mockRepository{fail: true}
	r.Get("/sites/users/site", api.Secure(api.HandleGetUsersForSite))

	rec = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/sites/users/site", nil)

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
