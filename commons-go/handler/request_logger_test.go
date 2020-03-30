package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func getLoggerMiddleware(w io.Writer) *LoggerMiddleware {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(w)
	entry := logger.WithField("mode", "test")
	return NewLoggerMiddleware(entry)
}

type logEntry struct {
	Appname    string
	Bytes      int
	Duration   int
	Function   string
	Host       string
	Hostid     string
	Level      string
	Method     string
	Msg        string
	Proto      string
	RemoteAddr string
	RequestURI string
	RequestID  string
	Scheme     string
	Status     int
	Time       string
}

func TestLoggerMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r := chi.NewRouter()
	var b strings.Builder
	logger := getLoggerMiddleware(&b)

	r.Use(logger.LoggerContext)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	logMsg := b.String()
	assert.NotEmpty(t, logMsg)
	var payload logEntry
	err := json.Unmarshal([]byte(logMsg), &payload)
	if err != nil {
		t.Errorf("could not unmarshall: %v", err)
	}

	assert.Equal(t, "example.com", payload.Host)
	assert.Equal(t, "handler.LoggerContext", payload.Function)
	assert.Equal(t, "/", payload.RequestURI)

}
