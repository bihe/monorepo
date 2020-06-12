package logging

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

const (
	// LogFieldNameFunction identifies the function in structured logging
	LogFieldNameFunction = "function"

	// RequestIDKey identifies a reqest in structured logging
	RequestIDKey = "requestID"

	// ApplicationNameKey identifies the application in structured logging
	ApplicationNameKey = "appName"

	// HostIDKey identifies the host in structured logging
	HostIDKey = "hostID"
)

// NewLog create a *log.Entry with fields appName and hostID
func NewLog(logger *log.Logger, appName, hostID string) *log.Entry {
	return logger.WithFields(log.Fields{
		ApplicationNameKey: appName,
		HostIDKey:          hostID,
	})
}

// Log create a *log.Entry
func Log(name string) *log.Entry {
	return log.WithField(LogFieldNameFunction, name)
}

// LogWith uses the supplied *log.Entry to create a new decorated entry
func LogWith(entry *log.Entry, name string) *log.Entry {
	return entry.WithField(LogFieldNameFunction, name)
}

// LogWithReq uses the supplied http.Request and *log.Entry to create a new decorated entry
func LogWithReq(req *http.Request, entry *log.Entry, name string) *log.Entry {
	reqID := middleware.GetReqID(req.Context())
	if reqID == "" {
		return entry.WithFields(log.Fields{
			LogFieldNameFunction: name,
		})
	}
	return entry.WithFields(log.Fields{
		LogFieldNameFunction: name,
		RequestIDKey:         reqID,
	})
}
