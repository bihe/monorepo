package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"golang.binggl.net/monorepo/pkg/logging"
)

// LoggerMiddleware is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return.
//
// this implementation was derived from the go-chi middleware https://github.com/go-chi/chi/blob/master/middleware/logger.go
type LoggerMiddleware struct {
	logger logging.Logger
}

// NewRequestLogger instantiates a new middleware to log requests
func NewRequestLogger(logger logging.Logger) *LoggerMiddleware {
	m := LoggerMiddleware{
		logger: logger,
	}
	return &m
}

// LoggerContext performs the middleware action
func (l *LoggerMiddleware) LoggerContext(next http.Handler) http.Handler {
	return handleLog(next, l.logger)
}

func handleLog(next http.Handler, logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			kvs    []logging.KeyValue
			scheme string
		)

		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
		kvs = append(kvs, logging.LogV("scheme", scheme),
			logging.LogV("method", r.Method),
			logging.LogV("host", r.Host),
			logging.LogV("requestURI", r.RequestURI),
			logging.LogV("proto", r.Proto),
			logging.LogV("remoteAddr", r.RemoteAddr),
		)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			kvs = append(kvs,
				logging.LogV("status", fmt.Sprintf("%d", ww.Status())),
				logging.LogV("bytes", fmt.Sprintf("%d", ww.BytesWritten())),
				logging.LogV("duration", fmt.Sprintf("%d", time.Since(t1))),
			)
			logger.InfoRequest("request", r, kvs...)
		}()

		next.ServeHTTP(ww, r)
	})
}
