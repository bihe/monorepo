package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.binggl.net/monorepo/pkg/logging"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// mock the logging.Logger

/*
type Logger interface {
	// implement the io.Closer interface
	io.Closer
	// Info logs a message and values with logging-level INFO
	Info(msg string, keyvals ...KeyValue)
	// Info logs a message and values with logging-level DEBUG
	Debug(msg string, keyvals ...KeyValue)
	// Warn logs a message and values with logging-level WARNING
	Warn(msg string, keyvals ...KeyValue)
	// Error logs a message and values with logging-level ERROR
	Error(msg string, keyvals ...KeyValue)
	// LogRequest is used to log a *http.Request (with a Trace-ID, if available)
	LogRequest(req *http.Request, keyvals ...KeyValue)
}
*/

type mockLogger struct {
	keyvals []logging.KeyValue
}

func (m *mockLogger) Info(msg string, keyvals ...logging.KeyValue) {
	m.keyvals = append(m.keyvals, keyvals...)
}

func (m *mockLogger) Debug(msg string, keyvals ...logging.KeyValue) {
	m.keyvals = append(m.keyvals, keyvals...)
}

func (m *mockLogger) Warn(msg string, keyvals ...logging.KeyValue) {
	m.keyvals = append(m.keyvals, keyvals...)
}

func (m *mockLogger) Error(msg string, keyvals ...logging.KeyValue) {
	m.keyvals = append(m.keyvals, keyvals...)
}

func (m *mockLogger) InfoRequest(msg string, req *http.Request, keyvals ...logging.KeyValue) {
	m.keyvals = append(m.keyvals, keyvals...)
}

func (m *mockLogger) ErrorRequest(msg string, req *http.Request, keyvals ...logging.KeyValue) {
	m.keyvals = append(m.keyvals, keyvals...)
}

func (m *mockLogger) Close() error {
	return nil
}

var _ logging.Logger = &mockLogger{}

func TestRequestLogger(t *testing.T) {
	logger := &mockLogger{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r := chi.NewRouter()
	mw := NewRequestLogger(logger)

	r.Use(mw.LoggerContext)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	})

	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	// the request-logger addes 9 entries
	assert.Equal(t, 9, len(logger.keyvals))

	var keys []string
	for _, kv := range logger.keyvals {
		v := kv.Read()
		keys = append(keys, v[0])
	}

	for _, key := range keys {
		if key != "scheme" && key != "method" && key != "host" && key != "requestURI" && key != "proto" && key != "remoteAddr" && key != "status" && key != "bytes" && key != "duration" {
			t.Errorf("request-logging did not return the correct keys %v", keys)
		}
	}

}
