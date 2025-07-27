package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

const staticDirPath = "./testdata/static"
const publicPath = "/static/"

func TestServeStaticDir(t *testing.T) {
	r := chi.NewRouter()
	staticDir := http.Dir(staticDirPath)

	// Create test data directory and files
	err := os.MkdirAll(staticDirPath, 0755)
	assert.NoError(t, err)

	testFiles := []struct {
		name    string
		content string
		path    string
	}{
		{"index.html", "<html><body>Index</body></html>", "/"},
		{"test.txt", "Hello World", "/"},
	}

	for _, tf := range testFiles {
		f, err := os.Create(filepath.Join(staticDirPath, tf.path[1:], tf.name))
		assert.NoError(t, err)
		_, err = f.Write([]byte(tf.content))
		assert.NoError(t, err)
		f.Close()
	}

	ServeStaticDir(r, publicPath, staticDir)

	t.Run("Serving existing file", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/static/test.txt", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello World")
	})

	t.Run("Serving non-existing file should serve index.html", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/static/nonexistent", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "<html><body>Index</body></html>")
	})

	t.Run("Serving file with query parameters", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/static/test.txt?query=param", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Hello World")
	})

	t.Run("Serving non-existent directory should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("ServeStaticDir did not panic for non-existent static directory")
			}
		}()
		nonExistentDir := http.Dir("/nonexistent/path")
		ServeStaticDir(r, publicPath, nonExistentDir)
	})
}

func TestServeStaticFile(t *testing.T) {
	r := chi.NewRouter()
	publicPath := "/favicon.ico"

	// Create test data directory and files
	err := os.MkdirAll(staticDirPath, 0755)
	assert.NoError(t, err)

	faviconFilePath := filepath.Join(staticDirPath, "favicon.ico")
	testFiles := []struct {
		name    string
		content string
		path    string
	}{
		{"favicon.ico", "<icon data>", "/"},
	}

	for _, tf := range testFiles {
		f, err := os.Create(filepath.Join(staticDirPath, tf.path[1:], tf.name))
		assert.NoError(t, err)
		_, err = f.Write([]byte(tf.content))
		assert.NoError(t, err)
		f.Close()
	}

	ServeStaticFile(r, publicPath, faviconFilePath)

	t.Run("Serving existing static file", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "<icon data>")
	})

	t.Run("Serving non-existent static file should return 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/nonexistent.ico", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
