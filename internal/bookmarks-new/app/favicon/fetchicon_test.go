package favicon

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchFavicon(t *testing.T) {
	favicon, _ := ioutil.ReadFile("../assets/favicon.ico")

	// setup a test-server
	// ------------------------------------------------------------------
	mux := http.NewServeMux()
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/errorFavicon", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := ` <html>
                <head>
                    <meta charset="utf-8">
                    <link rel="shortcut icon" href="/error">
                </head>
                <body>html</body>
            </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-lenght", fmt.Sprintf("%d", len(favicon)))
		if _, err := w.Write(favicon); err != nil {
			t.Fatalf("%v", err)
		}

	})
	mux.HandleFunc("/singleFile/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-lenght", fmt.Sprintf("%d", len(favicon)))
		if _, err := w.Write(favicon); err != nil {
			t.Fatalf("%v", err)
		}

	})
	mux.HandleFunc("/img/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-lenght", fmt.Sprintf("%d", len(favicon)))
		if _, err := w.Write(favicon); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/pageRel/img/favicon32x32.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-lenght", fmt.Sprintf("%d", len(favicon)))
		if _, err := w.Write(favicon); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/pageAbs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := ` <html>
                <head>
                    <meta charset="utf-8">
                    <link rel="shortcut icon" href="/img/favicon.png">
                </head>
                <body>html</body>
            </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/pageRel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := ` <html>
                <head>
                    <meta charset="utf-8">
                    <link rel="icon" href="./img/favicon32x32.png">
                </head>
                <body>html</body>
            </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/singleFile/index.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := ` <html>
	        <head>
	            <meta charset="utf-8">
	            <link rel="icon" href="favicon.ico">
	        </head>
	        <body>html</body>
	    </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/cdn", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := fmt.Sprintf(`<html>
                <head>
                    <meta charset="utf-8">
                    <link rel="icon" href="//%s/img/favicon.png">
                </head>
                <body>html</body>
            </html>`, r.Host)
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/parseErr", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := `\\\\\\\
                <head>
                    <meta charset="utf-8">
                    <link rel="icon ./img/favicon32x32.png>
                </html
                <bodyhtml</body>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// default, use basepath favicon.ico
	// ------------------------------------------------------------------
	fileName, payload, err := GetFaviconFromURL(ts.URL)
	if err != nil {
		t.Errorf("could not get default favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", fileName)
	assert.Equal(t, len(favicon), len(payload))

	// use html content1
	// ------------------------------------------------------------------
	fileName, payload, err = GetFaviconFromURL(ts.URL + "/pageAbs")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.png", fileName)
	assert.Equal(t, len(favicon), len(payload))

	// use html content2
	// ------------------------------------------------------------------
	fileName, payload, err = GetFaviconFromURL(ts.URL + "/pageRel")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon32x32.png", fileName)
	assert.Equal(t, len(favicon), len(payload))

	// use html content3
	// ------------------------------------------------------------------
	fileName, payload, err = GetFaviconFromURL(ts.URL + "/cdn")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.png", fileName)
	assert.Equal(t, len(favicon), len(payload))

	// single file
	// ------------------------------------------------------------------
	fileName, payload, err = GetFaviconFromURL(ts.URL + "/singleFile/index.html")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", fileName)
	assert.Equal(t, len(favicon), len(payload))

	// html parse error
	// ------------------------------------------------------------------
	fileName, payload, err = GetFaviconFromURL(ts.URL + "/parseErr")
	if err != nil {
		t.Errorf("could not get default favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", fileName)
	assert.Equal(t, len(favicon), len(payload))

	// http error
	// ------------------------------------------------------------------
	_, _, err = GetFaviconFromURL(ts.URL + "/errorFavicon")
	if err == nil {
		t.Errorf("expected error")
	}

	// invalid url
	// ------------------------------------------------------------------
	_, _, err = GetFaviconFromURL("udp://this should be an invalid URL /")
	if err == nil {
		t.Errorf("expected error")
	}

}
