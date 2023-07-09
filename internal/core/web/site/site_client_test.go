package site_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	services "golang.binggl.net/monorepo/internal/core/web/site"
)

const token = "token"
const siteDataJson = `{
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

func Test_GetSites(t *testing.T) {

	// setup a test-server
	// ------------------------------------------------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sites", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(siteDataJson)); err != nil {
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

func Test_SaveSites(t *testing.T) {

	// setup a test-server
	// ------------------------------------------------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sites", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/api/v1/sites/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Fatalf("%v", err)
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := services.GetSiteClient(ts.URL+"/api/v1/sites", token)
	var data services.UserSite
	if err := json.Unmarshal([]byte(siteDataJson), &data); err != nil {
		t.Errorf("could not unmarshal json: %v", err)
	}
	err := client.SaveSites(data)
	if err != nil {
		t.Errorf("could not get sites; %v", err)
	}

	// wrong endpoint
	// ------------------------------------------------------------------
	client = services.GetSiteClient("http://localhost", token)
	err = client.SaveSites(data)
	if err == nil {
		t.Errorf("error expected")
	}
	// wrong status
	// ------------------------------------------------------------------
	client = services.GetSiteClient(ts.URL+"/api/v1/sites/status", token)
	err = client.SaveSites(data)
	if err == nil {
		t.Errorf("error expected")
	}
}
