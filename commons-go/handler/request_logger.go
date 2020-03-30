package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"golang.binggl.net/commons"
)

// logCtxtKey is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type logCtxtKey int

// logKey is used to access the log-entry in the context
var logKey logCtxtKey

// LoggerMiddleware is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return.
//
// this implementation was derived from the go-chi middleware https://github.com/go-chi/chi/blob/master/middleware/logger.go
type LoggerMiddleware struct {
	log *log.Entry
}

// NewLoggerMiddleware creates a new instance using the provided options
func NewLoggerMiddleware(logger *log.Entry) *LoggerMiddleware {
	m := LoggerMiddleware{
		log: logger,
	}
	return &m
}

// LoggerContext performs the middleware action
func (j *LoggerMiddleware) LoggerContext(next http.Handler) http.Handler {
	return handleLog(next, j.log)
}

// WithLogEntry sets the in-context LogEntry for a request.
func WithLogEntry(r *http.Request, entry *log.Entry) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), logKey, entry))
	return r
}

func handleLog(next http.Handler, logger *log.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := newLogEntry(r, logger)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			write(entry, ww.Status(), ww.BytesWritten(), time.Since(t1))
		}()

		next.ServeHTTP(ww, WithLogEntry(r, entry))
	})
}

func newLogEntry(r *http.Request, logger *log.Entry) *log.Entry {
	entry := commons.LogWithReq(r, logger, "handler.LoggerContext")
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return entry.WithFields(log.Fields{
		"method":     r.Method,
		"scheme":     scheme,
		"host":       r.Host,
		"requestURI": r.RequestURI,
		"proto":      r.Proto,
		"remoteAddr": r.RemoteAddr,
	})
}

func write(entry *log.Entry, status, bytes int, elapsed time.Duration) {
	entry.WithFields(log.Fields{
		"status":   status,
		"bytes":    bytes,
		"duration": elapsed,
	}).Info("request")
}
