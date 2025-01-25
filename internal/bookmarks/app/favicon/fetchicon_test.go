package favicon

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed favicon.png
var pngFavicon []byte

//go:embed favicon.ico
var icoFavicon []byte

func TestFetchFavicon(t *testing.T) {

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
	mux.HandleFunc("/noFavicon", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := ` <html>
                <head>
                    <meta charset="utf-8">
                    <link rel="canonical" href="http://localhost/noFavicon">
                </head>
                <body>html</body>
            </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(icoFavicon)))
		if _, err := w.Write(icoFavicon); err != nil {
			t.Fatalf("%v", err)
		}

	})
	mux.HandleFunc("/wrong-mimetype", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(icoFavicon)))
		w.Header().Add("content-type", "application/octet-stream")
		if _, err := w.Write(icoFavicon); err != nil {
			t.Fatalf("%v", err)
		}

	})
	mux.HandleFunc("/singleFile/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(icoFavicon)))
		if _, err := w.Write(icoFavicon); err != nil {
			t.Fatalf("%v", err)
		}

	})
	mux.HandleFunc("/img/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(pngFavicon)))
		if _, err := w.Write(pngFavicon); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/pageRel/img/favicon32x32.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(pngFavicon)))
		if _, err := w.Write(pngFavicon); err != nil {
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
	mux.HandleFunc("/pathNoFile", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(pngFavicon)))
		if _, err := w.Write(pngFavicon); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/disposition", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-length", fmt.Sprintf("%d", len(pngFavicon)))
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("favicon-disposition.png"))
		if _, err := w.Write(pngFavicon); err != nil {
			t.Fatalf("%v", err)
		}
	})
	mux.HandleFunc("/missing_schema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		html := ` <html>
	        <head>
	            <meta charset="utf-8">
	            <link rel="icon" href=":///favicon.ico">
	        </head>
	        <body>html</body>
	    </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			t.Fatalf("%v", err)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// default, use basepath favicon.ico
	// ------------------------------------------------------------------
	content, err := GetFaviconFromURL(ts.URL)
	if err != nil {
		t.Errorf("could not get default favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(icoFavicon), len(content.Payload))

	// use html content1
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/pageAbs")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.png", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(pngFavicon), len(content.Payload))

	// use html content2
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/pageRel")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon32x32.png", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(pngFavicon), len(content.Payload))

	// use html content3
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/cdn")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.png", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(pngFavicon), len(content.Payload))

	// single file
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/singleFile/index.html")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(icoFavicon), len(content.Payload))

	// html parse error
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/parseErr")
	if err != nil {
		t.Errorf("could not get default favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(icoFavicon), len(content.Payload))

	// DefaultFaviconName because not filename in path
	// ------------------------------------------------------------------
	content, err = FetchURL(ts.URL+"/pathNoFile", FetchImage)
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(pngFavicon), len(content.Payload))

	// parse content-disposition
	// ------------------------------------------------------------------
	content, err = FetchURL(ts.URL+"/disposition", FetchImage)
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon-disposition.png", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(pngFavicon), len(content.Payload))

	// http error
	// ------------------------------------------------------------------
	_, err = GetFaviconFromURL(ts.URL + "/errorFavicon")
	if err == nil {
		t.Errorf("expected error")
	}

	// invalid url
	// ------------------------------------------------------------------
	_, err = GetFaviconFromURL("udp://this should be an invalid URL /")
	if err == nil {
		t.Errorf("expected error")
	}

	// wrong mime-type; expected image/*
	// ------------------------------------------------------------------
	_, err = FetchURL(ts.URL+"/wrong-mimetype", FetchImage)
	if err == nil {
		t.Error("error because of wrong image/* mime-type expected")
	}

	// valid HTML but no favicon
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/noFavicon")
	if err != nil {
		t.Errorf("valid HTML expected default favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(icoFavicon), len(content.Payload))

	// missing schema
	// ------------------------------------------------------------------
	content, err = GetFaviconFromURL(ts.URL + "/missing_schema")
	if err != nil {
		t.Errorf("could not get favicon: %v", err)
	}
	assert.Equal(t, "favicon.ico", content.FileName)
	assert.True(t, len(content.Payload) > 0)
	assert.Equal(t, len(icoFavicon), len(content.Payload))
}
