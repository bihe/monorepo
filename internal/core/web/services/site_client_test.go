package services_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.binggl.net/monorepo/internal/core/web/services"
)

const token = "token"
const siteResponse = `{
	"editable":true,
	"user":"a.b@cd.de",
	"userSites":[
		{
			"name":"bookmarks",
			"url":"https://bookmarks.net",
			"permissions":[
				"Admin",
				"User"
			]
		},
		{
			"name":"crypter",
			"url":"https://crypter.net",
			"permissions":[
				"User"
			]
		},
		{
			"name":"login",
			"url":"https://login.net",
			"permissions":[
				"Admin",
				"User"
			]
		}
	]
}`

func Test_SiteClient(t *testing.T) {

	// setup a test-server
	// ------------------------------------------------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sites", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(siteResponse)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/api/v1/sites/invalid", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("invalid")); err != nil {
			t.Fatalf("%v", err)
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := services.GetSiteClient(ts.URL+"/api/v1/sites", token)
	sites, err := client.GetSites()
	if err != nil {
		t.Errorf("could not get sites; %v", err)
	}
	if len(sites.UserSites) != 3 {
		t.Errorf("empty sites returned")
	}

	// unmarshal
	// ------------------------------------------------------------------
	client = services.GetSiteClient(ts.URL+"/api/v1/sites/invalid", token)
	_, err = client.GetSites()
	if err == nil {
		t.Errorf("error expected")
	}

	// wrong endpoint
	// ------------------------------------------------------------------
	client = services.GetSiteClient("http://localhost", token)
	_, err = client.GetSites()
	if err == nil {
		t.Errorf("error expected")
	}
}
